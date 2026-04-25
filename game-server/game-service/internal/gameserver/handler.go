package gameserver

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	authPb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel"
)

/**
* Handles all the management and maintenance of connections with client
**/
var tracer = otel.Tracer("game-service")

func (s *Server) HandleWebSocketConnection(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "service.HandleWebSocketConnection")
	defer span.End()

	userIdStr, ok := c.Get("userIdStr")
	slog.Debug("User ID from token and passed down with gin context", "userIdStr", userIdStr)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"statusCode": http.StatusUnauthorized, "message": "Unauthorized"})
		return
	}

	// verify member
	grpcPayload := &authPb.GetMemberRequest{
		Id: userIdStr.(string),
	}
	authClient := s.GetAuthClient()

	data, err := authClient.GetMember(c.Request.Context(), grpcPayload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user does not exist"})
		return
	}

	memberId, err := uuid.Parse(data.Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "memberId is not uuid"})
		return
	}

	// check if player is reconnecting by checking if they already exist in the players map
	s.mu.RLock()
	reconnectingPlayer, playerIsReconnecting := s.players[memberId]
	s.mu.RUnlock()

	var player *types.Player
	isReconnection := false

	if playerIsReconnecting {
		slog.Debug("Player is reconnecting",
			"username", reconnectingPlayer.Username,
			"game_session_id", reconnectingPlayer.CurrentGameSessionId,
			"connect_state", reconnectingPlayer.ConnectState)

		player = reconnectingPlayer
		isReconnection = true

		// mark as connected
		connected := constants.Connected
		player.ConnectState = &connected

		// update username (may have changed name)
		player.Username = data.Name
	} else {
		// new player connecting for the first time
		slog.Debug("Player connecting fresh.",
			"username", data.Name,
		)

		connected := constants.Connected
		player = &types.Player{
			ID:                   memberId,
			Username:             data.Name,
			CurrentGameSessionId: uuid.Nil,
			ConnectState:         &connected,
		}

		// add to players map
		s.mu.Lock()
		s.players[memberId] = player
		s.mu.Unlock()
	}

	// establish websocket connection
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("Error establishing websocket connection.",
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade connection"})
		return
	}

	// update conn to player mapping
	s.MapConnToPlayer(conn, *player)

	// create msgChan
	s.setupClientWriter(conn)

	if isReconnection && player.CurrentGameSessionId != uuid.Nil {
		slog.Debug("Sending reconnection success message.",
			"player_username", player.Username,
			"player_current_game_session_id", player.CurrentGameSessionId,
		)

		s.mu.RLock()
		msgChan, exists := s.msgChan[conn]
		s.mu.RUnlock()

		if exists {
			// send reconnected message
			msgChan <- types.Message{
				Action: "reconnected",
				Payload: map[string]interface{}{
					"message":    "Successfully reconnected",
					"session_id": player.CurrentGameSessionId.String(),
					"username":   player.Username,
				},
			}

			// important! send game_found message to let frontend enter game
			slog.Info("Sending game_found message after reconnection")
			msgChan <- types.Message{
				Action: "game_found",
				Payload: map[string]interface{}{
					"session_id": player.CurrentGameSessionId.String(),
				},
			}
		}
	}

	// handle connection messages
	go s.ServeConnectedPlayer(conn)
}

/**
* Serves each individually connected player.
**/
func (s *Server) ServeConnectedPlayer(conn *websocket.Conn) {
	for {
		slog.Debug("Listening for player messages")
		_, message, err := conn.ReadMessage()

		// --- Handle WebSocket Errors ---
		if err != nil {
			player, exists := s.GetPlayerFromConn(conn)
			playerInfo := "Unknown"
			playerID := uuid.Nil
			if exists {
				playerInfo = fmt.Sprintf("%s (ID: %s)", player.Username, player.ID)
				playerID = player.ID
			}

			slog.Error("\nWebSocket Connection Error\n",
				"player_info", playerInfo,
				"error", err,
			)

			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				slog.Error("Network disconnection for player. Keeping state for reconnection...",
					"player_info", playerInfo)

				if exists {
					// mark as reconnecting state
					s.markPlayerAsReconnecting(player)

					// only clean connection, keep player data, session, queue
					s.cleanUpConnectionOnly(conn)

					// start 30 second timer
					go s.handleReconnectionTimeout(playerID, 30*time.Second)
				} else {
					// no player info, directly close connection
					conn.Close()
				}

				return // exit this goroutine
			}

			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				slog.Info("Player closed connection normally (intentional leave)", "player_info", playerInfo)
				s.cleanUpClient(conn)
				return // exit this goroutine
			}

			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				slog.Info("Player navigated away, keeping state for 10s", "player_info", playerInfo)

				if exists {
					s.markPlayerAsReconnecting(player)
					s.cleanUpConnectionOnly(conn)
					go s.handleReconnectionTimeout(playerID, 10*time.Second)
				} else {
					conn.Close()
				}

				return
			}

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Warn("Unexpected disconnection, keeping state", "player_info", playerInfo)

				if exists {
					s.markPlayerAsReconnecting(player)
					s.cleanUpConnectionOnly(conn)
					go s.handleReconnectionTimeout(playerID, 30*time.Second)
				} else {
					conn.Close()
				}

				return
			}

			slog.Error("Unknown websocket error, cleaning up completely", "player_info", playerInfo, "error", err)
			s.cleanUpClient(conn)
			return
		}

		// --- Normal Message Processing ---
		slog.Debug("Message received from connected user", "message", string(message))

		slog.Debug("before decoding received message")

		// --- Client Connection Handling ---
		// Decodes Incoming client message and serves their unique connection its own goroutine

		// decode message to pre-defined json structure "GameMessage"
		var decodedMsg types.Message

		err = json.Unmarshal(message, &decodedMsg)

		if err != nil {
			slog.Error("Error when decoding payload", "error", err)

			conn.WriteJSON(types.Message{Action: "Error", Payload: map[string]interface{}{"error": "Your message to server was the incorrect format and could not be decoded as JSON."}})
			continue
		}

		// handle concurrent writes back to clients
		s.setupClientWriter(conn)

		clientPackage := types.ClientPackage{Message: decodedMsg, Conn: conn}

		slog.Debug("Sending clientPackage to message hub")

		// send message to MessageHub via an *unbuffered channel* for handling based on the type field.
		s.serverChan <- clientPackage
	}
}

/**
* Handles adding clients and creating gameMsgChans for handling connection writes
* back to the connected client.
*
* NOTE: Gorilla Websocket package only allows ONE CONCURRENT WRITER
* at a time, meaning its best to utilize unbuffered channels to prevent
* a single client from locking the entire server, and prevent race conditions
* where multiple writes to the same connection.
**/
func (s *Server) setupClientWriter(conn *websocket.Conn) {
	isNew := s.createMsgChan(conn)
	if !isNew {
		return
	}

	// get the message channel for this connection
	msgChan, err := s.getGameMsgChan(conn)

	if err != nil {
		slog.Error("Error getting message channel",
			"error", err,
		)
		return
	}

	// concurrently listen to all incoming messages over the channel to write game actions
	// back to the client
	go func() {
		defer func() {
			// ensure we recover from any panics in the writer goroutine
			if r := recover(); r != nil {
				slog.Info("clientWriter panic recovered",
					"r", r,
				)
			}
		}()

		for msg := range msgChan {
			err := conn.WriteJSON(msg)

			if err != nil {
				slog.Error("Error writing to client, connection likely closed",
					"error", err,
				)
				// channel will be closed by cleanUpClient, which will exit this goroutine
				return
			}
		}
	}()
}

/**
* Creates the unique game message channel for a specific connection for writing back
* from server to client. Only creates if it doesn't already exist.
**/
func (s *Server) createMsgChan(conn *websocket.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// if already exists, don't duplicate create
	if _, exists := s.msgChan[conn]; exists {
		return false
	}

	s.msgChan[conn] = make(chan interface{}, 10) // add buffer to avoid blocking
	return true
}

/**
* Gets the unique game message channel for a specific connection for writing back
* from server to client, validating that it exists.
**/
func (s *Server) getGameMsgChan(conn *websocket.Conn) (chan interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	channel, exists := s.msgChan[conn]

	if !exists {
		return nil, fmt.Errorf("Game message channel for this connection does not exist.")
	}

	return channel, nil
}

func (s *Server) markPlayerAsReconnecting(player *types.Player) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reconnecting := constants.Reconnecting
	player.ConnectState = &reconnecting

	// 同時更新 server 的 players map
	if existingPlayer, exists := s.players[player.ID]; exists {
		existingPlayer.ConnectState = &reconnecting
	}
}

func (s *Server) cleanUpConnectionOnly(conn *websocket.Conn) {
	s.mu.Lock()

	player, exists := s.connToPlayer[conn]
	if !exists {
		s.mu.Unlock()
		slog.Error("cleanUpConnectionOnly: connection not found",
			"conn", conn,
		)
		conn.Close()
		return
	}

	slog.Info("Cleaning up connection only for player - keeping game state",
		"player_username", player.Username,
	)

	if ch, exists := s.msgChan[conn]; exists {
		close(ch)
		delete(s.msgChan, conn)
	}

	// delete conn to player mapping (but don't delete player itself)
	delete(s.connToPlayer, conn)

	// Note: don't delete s.players[player.ID], keep player data
	// Note: don't remove from queue, keep queue state
	// Note: don't remove from session, keep game state

	s.mu.Unlock()

	conn.Close()
}

func (s *Server) handleReconnectionTimeout(playerID uuid.UUID, timeout time.Duration) {
	slog.Debug("Reconnection timer started",
		"player_id", playerID,
		"timeout", timeout,
	)

	time.Sleep(timeout)

	s.mu.RLock()
	player, exists := s.players[playerID]
	s.mu.RUnlock()

	if !exists {
		slog.Debug("Player already cleaned up", "player_id", playerID)
		return
	}

	// check player state
	if player.ConnectState != nil && *player.ConnectState == constants.Reconnecting {
		slog.Debug("Player failed to reconnect within timeout limit.",
			"player_username", player.Username,
			"timeout", timeout,
		)
		s.cleanUpPlayerCompletely(playerID)
	} else {
		slog.Debug("Player already reconnected",
			"player_username", player.Username)
	}
}

func (s *Server) cleanUpPlayerCompletely(playerID uuid.UUID) {
	s.mu.RLock()
	player, exists := s.players[playerID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	slog.Info("Completely cleaning up player", "username", player.Username)

	// 從 queue 移除
	s.queue.PlayerRemoveQueue(player)

	// 從 session 移除
	s.cleanUpPlayerFromSession(player)

	// 從 players map 移除
	s.mu.Lock()
	delete(s.players, playerID)
	s.mu.Unlock()
}

/**
* Cleans up all resources associated with a client connection.
* Called when connection is closed or errors out.
**/
func (s *Server) cleanUpClient(conn *websocket.Conn) {
	s.mu.Lock()

	// 獲取玩家資訊
	player, exists := s.connToPlayer[conn]

	if exists {
		slog.Info("Cleaning up client", "username", player.Username)
		// 從 queue 中移除玩家
		s.queue.PlayerRemoveQueue(player)
		return
	}

	slog.Info("Cleaning up client", "username", player.Username, "player_id", player.ID)

	// 從 queue 中移除玩家
	s.queue.PlayerRemoveQueue(player)

	// 關閉並刪除 msgChan
	if ch, exists := s.msgChan[conn]; exists {
		close(ch)
		delete(s.msgChan, conn)
		slog.Info("Closed message channel for player", "username", player.Username)
	}

	// delete conn to player mapping
	delete(s.connToPlayer, conn)

	// delete player mapping
	delete(s.players, player.ID)

	s.mu.Unlock()

	// remove player from game session (call after releasing lock to avoid deadlock)
	s.cleanUpPlayerFromSession(player)

	// 關閉 WebSocket 連線
	conn.Close()
}

/**
* Removes player from their game session and shuts down empty sessions.
* This is separated from cleanUpClient to avoid holding the server mutex for too long.
**/
func (s *Server) cleanUpPlayerFromSession(player *types.Player) {
	if player == nil {
		slog.Warn("cleanUpPlayerFromSession: player is nil, skipping")
		return
	}

	if player.CurrentGameSessionId == uuid.Nil {
		slog.Debug("Player is not in any session, skipping session cleanup", "username", player.Username)
		return
	}

	s.mu.RLock()
	playerSession, exists := s.sessions[player.CurrentGameSessionId]
	s.mu.RUnlock()

	if !exists {
		slog.Warn("Session not found for player", "session_id", player.CurrentGameSessionId, "username", player.Username)
		return
	}

	slog.Info("Removing player from session", "username", player.Username, "session_id", playerSession.ID)

	// remove player from session
	playerSession.RemovePlayer(player.ID.String())

	// check if session still has players
	remainingPlayers := playerSession.GetPlayerIDs()
	if len(remainingPlayers) == 0 {
		slog.Info("Session has no remaining players, shutting down", "session_id", playerSession.ID)

		// shutdown session
		playerSession.Shutdown()

		// remove from server's sessions map
		s.mu.Lock()
		delete(s.sessions, playerSession.ID)
		s.mu.Unlock()

		slog.Info("Session shut down and removed", "session_id", playerSession.ID)
	} else {
		slog.Info("Session still has remaining players", "session_id", playerSession.ID, "remaining", len(remainingPlayers))
	}
}
