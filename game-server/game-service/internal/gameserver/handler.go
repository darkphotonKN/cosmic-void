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
		c.JSON(http.StatusBadRequest, gin.H{"error": "user is not exist"})
		return
	}

	memberId, err := uuid.Parse(data.Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "memberId is not uuid"})
		return
	}

	s.mu.RLock()
	existingPlayer, playerExists := s.players[memberId]
	s.mu.RUnlock()

	var player *types.Player
	isReconnection := false

	if playerExists {
		fmt.Printf("Player %s is reconnecting! (Session: %s, State: %v)\n",
			existingPlayer.Username,
			existingPlayer.CurrentGameSessionId,
			existingPlayer.ConnectState)

		player = existingPlayer
		isReconnection = true

		// mark as connected
		connected := constants.Connected
		player.ConnectState = &connected

		// update username (may have changed name)
		player.Username = data.Name

	} else {
		// new player connecting for the first time
		fmt.Printf("New player connecting: %s\n", data.Name)

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
		fmt.Println("Error establishing websocket connection.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade connection"})
		return
	}

	// update conn to player mapping
	s.MapConnToPlayer(conn, *player)

	// create msgChan
	s.setupClientWriter(conn)

	if isReconnection && player.CurrentGameSessionId != uuid.Nil {
		fmt.Printf("📤 Sending reconnection success message to %s (Session: %s)\n",
			player.Username, player.CurrentGameSessionId)

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
			fmt.Println("Sending game_found message after reconnection...")
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
* Serves each individual connected player.
**/
func (s *Server) ServeConnectedPlayer(conn *websocket.Conn) {
	for {
		fmt.Println("Listening for user messages...")
		_, message, err := conn.ReadMessage()

		// --- Handle WebSocket Errors ---
		if err != nil {
			// Get player info for logging
			player, exists := s.GetPlayerFromConn(conn)
			playerInfo := "Unknown"
			playerID := uuid.Nil
			if exists {
				playerInfo = fmt.Sprintf("%s (ID: %s)", player.Username, player.ID)
				playerID = player.ID
			}

			fmt.Printf("\n[WebSocket Error] Player: %s, Error: %v\n", playerInfo, err)

			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				fmt.Printf("Network disconnection for %s. Keeping state for reconnection...\n", playerInfo)

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
				fmt.Printf("Player %s closed connection normally (intentional leave)\n", playerInfo)
				s.cleanUpClient(conn)
				return // exit this goroutine
			}

			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				fmt.Printf("Player %s navigated away. Keeping state for 10s...\n", playerInfo)

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
				fmt.Printf("Unexpected disconnection for %s. Keeping state...\n", playerInfo)

				if exists {
					s.markPlayerAsReconnecting(player)
					s.cleanUpConnectionOnly(conn)
					go s.handleReconnectionTimeout(playerID, 30*time.Second)
				} else {
					conn.Close()
				}

				return
			}

			fmt.Printf("Unknown error for %s: %v. Cleaning up completely.\n", playerInfo, err)
			s.cleanUpClient(conn)
			return
		}

		// --- Normal Message Processing ---
		fmt.Printf("\nMessage received from connected user: %s\n\n", string(message))

		fmt.Println("before decoding received message")

		// --- Client Connection Handling ---
		// Decodes Incoming client message and serves their unique connection its own goroutine

		// decode message to pre-defined json structure "GameMessage"
		var decodedMsg types.Message

		err = json.Unmarshal(message, &decodedMsg)

		if err != nil {
			fmt.Println("Error when decoding payload.")

			conn.WriteJSON(types.Message{Action: "Error", Payload: map[string]interface{}{"error": "Your message to server was the incorrect format and could not be decoded as JSON."}})
			continue
		}

		// handle concurrent writes back to clients
		s.setupClientWriter(conn)

		clientPackage := types.ClientPackage{Message: decodedMsg, Conn: conn}

		fmt.Println("Sending clientPackage to message hub.")

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
		return // channel 已存在，不需要重複設置
	}

	// get the message channel for this connection
	msgChan, err := s.getGameMsgChan(conn)

	if err != nil {
		fmt.Printf("Error getting message channel: %s\n", err)
		// This shouldn't happen since we just created it, but log and return if it does
		return
	}

	// concurrently listen to all incoming messages over the channel to write game actions
	// back to the client
	go func() {
		defer func() {
			// ensure we recover from any panics in the writer goroutine
			if r := recover(); r != nil {
				// fmt.Printf("clientWriter panic recovered: %v\n", r)
			}
		}()

		for msg := range msgChan {
			// TEST: remove after testing
			// fmt.Printf("\nclientWriter writing back to client message:\n\n%+v\n\n", msg)

			err := conn.WriteJSON(msg)

			if err != nil {
				fmt.Printf("Error writing to client, connection likely closed: %s\n", err)
				// channel will be closed by cleanUpClient, which will exit this goroutine
				return
			}
		}
		// fmt.Println("clientWriter goroutine exiting (channel closed)")
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

	fmt.Printf("Player %s marked as reconnecting\n", player.Username)
}

func (s *Server) cleanUpConnectionOnly(conn *websocket.Conn) {
	s.mu.Lock()

	player, exists := s.connToPlayer[conn]
	if !exists {
		s.mu.Unlock()
		fmt.Println("cleanUpConnectionOnly: connection not found")
		conn.Close()
		return
	}

	fmt.Printf("Cleaning up connection only for player: %s (keeping game state)\n", player.Username)

	if ch, exists := s.msgChan[conn]; exists {
		close(ch)
		delete(s.msgChan, conn)
	}

	// delete conn to player mapping (but don't delete player itself)
	delete(s.connToPlayer, conn)

	// note: don't delete s.players[player.ID] <- keep player data
	// note: don't remove from queue <- keep queue state
	// note: don't remove from session <- keep game state

	s.mu.Unlock()

	conn.Close()
}

func (s *Server) handleReconnectionTimeout(playerID uuid.UUID, timeout time.Duration) {
	fmt.Printf("Reconnection timer started for player %s (timeout: %v)\n", playerID, timeout)

	time.Sleep(timeout)

	s.mu.RLock()
	player, exists := s.players[playerID]
	s.mu.RUnlock()

	if !exists {
		fmt.Printf("Player %s already cleaned up\n", playerID)
		return
	}

	// check player state
	if player.ConnectState != nil && *player.ConnectState == constants.Reconnecting {
		fmt.Printf("Player %s failed to reconnect within %v. Cleaning up...\n", player.Username, timeout)
		s.cleanUpPlayerCompletely(playerID)
	} else {
		fmt.Printf("Player %s already reconnected\n", player.Username)
	}
}

func (s *Server) cleanUpPlayerCompletely(playerID uuid.UUID) {
	s.mu.RLock()
	player, exists := s.players[playerID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	fmt.Printf("🗑️  Completely cleaning up player: %s\n", player.Username)

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
		fmt.Printf("Cleaning up client: %s\n", player.Username)
		// 從 queue 中移除玩家
		s.queue.PlayerRemoveQueue(player)
		return
	}

	fmt.Printf("Cleaning up client: %s (ID: %s)\n", player.Username, player.ID)

	// 從 queue 中移除玩家
	s.queue.PlayerRemoveQueue(player)

	// 關閉並刪除 msgChan
	if ch, exists := s.msgChan[conn]; exists {
		close(ch)
		delete(s.msgChan, conn)
		fmt.Printf("Closed message channel for player %s\n", player.Username)
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
		fmt.Println("cleanUpPlayerFromSession: player is nil, skipping")
		return
	}

	if player.CurrentGameSessionId == uuid.Nil {
		fmt.Printf("Player %s is not in any session, skipping session cleanup\n", player.Username)
		return
	}

	s.mu.RLock()
	playerSession, exists := s.sessions[player.CurrentGameSessionId]
	s.mu.RUnlock()

	if !exists {
		fmt.Printf("Session %s not found for player %s\n", player.CurrentGameSessionId, player.Username)
		return
	}

	fmt.Printf("Removing player %s from session %s\n", player.Username, playerSession.ID)

	// remove player from session
	playerSession.RemovePlayer(player.ID.String())

	// check if session still has players
	remainingPlayers := playerSession.GetPlayerIDs()
	if len(remainingPlayers) == 0 {
		fmt.Printf("Session %s has no remaining players, shutting down...\n", playerSession.ID)

		// shutdown session
		playerSession.Shutdown()

		// remove from server's sessions map
		s.mu.Lock()
		delete(s.sessions, playerSession.ID)
		s.mu.Unlock()

		fmt.Printf("Session %s successfully shut down and removed\n", playerSession.ID)
	} else {
		fmt.Printf("Session %s still has %d player(s) remaining\n", playerSession.ID, len(remainingPlayers))
	}
}
