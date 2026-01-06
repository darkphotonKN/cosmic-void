package gameserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	authPb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/**
* Handles all the management and maintenance of connections with client
**/

func (s *Server) HandleWebSocketConnection(c *gin.Context) {
	userIdStr, ok := c.Get("userIdStr")
	fmt.Printf("User ID: %s\n", userIdStr)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"statusCode": http.StatusUnauthorized, "message": "Unauthorized"})
		return
	}
	// verifay member
	grpcPayload := &authPb.GetMemberRequest{
		Id: userIdStr.(string),
	}
	// 取得 auth client
	authClient := s.GetAuthClient()

	// 調用 GetMember
	data, err := authClient.GetMember(c.Request.Context(), grpcPayload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user is not exist"})
	}

	memberId, err := uuid.Parse(data.Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "memberId is not uuid"})
	}
	player := &types.Player{
		ID:       memberId,
		Username: data.Name,
	}
	fmt.Println("player:", player)

	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		fmt.Println("Error establishing websocket connection.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade connection"})
		return
	}

	s.MapConnToPlayer(conn, *player)

	// 立即建立 msgChan，確保重連後能收到 server 訊息
	s.setupClientWriter(conn)

	// handle each connected client's messages concurrently
	go s.ServeConnectedPlayer(conn)
}

/**
* Serves each individual connected player.
**/
func (s *Server) ServeConnectedPlayer(conn *websocket.Conn) {
	// removes client and closes connection
	defer func() {
		fmt.Println("Connection closed, cleaning up client...")
		s.cleanUpClient(conn)
	}()

	for {
		fmt.Println("Listening for user messages...")
		_, message, err := conn.ReadMessage()

		fmt.Printf("\nMessage received from connected user: %s\n\n", string(message))

		// --- clean up connection ---
		if err != nil {
			// get player info for logging
			player, exists := s.GetPlayerFromConn(conn)
			playerInfo := "Unknown"
			if exists {
				playerInfo = fmt.Sprintf("%s (ID: %s)", player.Username, player.ID)
			}

			// Unexpected Error
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("Abnormal error occurred with player %s. Closing connection. Error: %s\n", playerInfo, err)
				break
			}

			// Close Error
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				fmt.Printf("Connection closed for player %s. Error: %s\n", playerInfo, err)
				break
			}

			// General Error
			fmt.Printf("General error occurred during connection with player %s: %s\n", playerInfo, err)
			break
		}

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
* at a time, meaning its best to utilize *unbuffered* channels to prevent
* a single client from locking the entire server, and prevent race conditions
* where multiple writes to the same connection.
**/
func (s *Server) setupClientWriter(conn *websocket.Conn) {
	// 只在 channel 不存在時才創建並啟動 goroutine
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
				fmt.Printf("clientWriter panic recovered: %v\n", r)
			}
		}()

		for msg := range msgChan {
			// TEST: remove after testing
			fmt.Printf("\nclientWriter writing back to client message:\n\n%+v\n\n", msg)

			err := conn.WriteJSON(msg)

			if err != nil {
				fmt.Printf("Error writing to client, connection likely closed: %s\n", err)
				// channel will be closed by cleanUpClient, which will exit this goroutine
				return
			}
		}
		fmt.Println("clientWriter goroutine exiting (channel closed)")
	}()

}

/**
* Creates the unique game message channel for a specific connection for writing back
* from server to client. Only creates if it doesn't already exist.
**/
func (s *Server) createMsgChan(conn *websocket.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果已經存在，不要重複創建
	if _, exists := s.msgChan[conn]; exists {
		return false
	}

	s.msgChan[conn] = make(chan interface{}, 10) // 加入緩衝避免阻塞
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

/**
* Cleans up all resources associated with a client connection.
* Called when connection is closed or errors out.
**/
func (s *Server) cleanUpClient(conn *websocket.Conn) {
	s.mu.Lock()

	// 獲取玩家資訊
	player, exists := s.connToPlayer[conn]
	if !exists {
		s.mu.Unlock()
		fmt.Println("cleanUpClient: connection not found in connToPlayer map")
		conn.Close()
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

	// 刪除 conn -> player 映射
	delete(s.connToPlayer, conn)

	// 刪除 player 映射
	delete(s.players, player.ID)

	s.mu.Unlock()

	// 從遊戲 session 中移除玩家 (在釋放鎖之後調用,避免死鎖)
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

	s.mu.RLock()
	var playerSession *game.Session
	for _, session := range s.sessions {
		for _, pid := range session.GetPlayerIDs() {
			if pid == player.ID {
				playerSession = session
				break
			}
		}
		if playerSession != nil {
			break
		}
	}
	s.mu.RUnlock()

	if playerSession == nil {
		fmt.Printf("Player %s is not in any active session\n", player.Username)
		return
	}

	fmt.Printf("Removing player %s from session %s\n", player.Username, playerSession.ID)

	// 從 session 移除玩家
	playerSession.RemovePlayer(player.ID.String())

	// 檢查 session 是否還有玩家
	remainingPlayers := playerSession.GetPlayerIDs()
	if len(remainingPlayers) == 0 {
		fmt.Printf("Session %s has no remaining players, shutting down...\n", playerSession.ID)

		// 關閉 session
		playerSession.Shutdown()

		// 從 server 的 sessions map 移除
		s.mu.Lock()
		delete(s.sessions, playerSession.ID)
		s.mu.Unlock()

		fmt.Printf("Session %s successfully shut down and removed\n", playerSession.ID)
	} else {
		fmt.Printf("Session %s still has %d player(s) remaining\n", playerSession.ID, len(remainingPlayers))
	}
}
