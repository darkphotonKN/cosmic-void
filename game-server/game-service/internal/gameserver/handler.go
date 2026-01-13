package gameserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	authPb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
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

	// 驗證 member
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

	// ==========================================
	// 檢查是否為重連
	// ==========================================
	s.mu.RLock()
	existingPlayer, playerExists := s.players[memberId]
	s.mu.RUnlock()

	var player *types.Player
	isReconnection := false

	if playerExists {
		// ✅ 這是重連!重用現有的 player 資料
		fmt.Printf("🔄 Player %s is reconnecting! (Session: %s, State: %v)\n",
			existingPlayer.Username,
			existingPlayer.CurrentGameSessionId,
			existingPlayer.ConnectState)

		player = existingPlayer
		isReconnection = true

		// 標記為已連線
		connected := constants.Connected
		player.ConnectState = &connected

		// 更新 username (可能改名了)
		player.Username = data.Name

	} else {
		// 新玩家第一次連線
		fmt.Printf("✨ New player connecting: %s\n", data.Name)

		connected := constants.Connected
		player = &types.Player{
			ID:                   memberId,
			Username:             data.Name,
			CurrentGameSessionId: uuid.Nil,
			ConnectState:         &connected,
		}

		// 加入 players map
		s.mu.Lock()
		s.players[memberId] = player
		s.mu.Unlock()
	}

	// 建立 WebSocket 連線
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Error establishing websocket connection.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade connection"})
		return
	}

	// 更新 conn -> player 映射
	s.MapConnToPlayer(conn, *player)

	// 建立 msgChan
	s.setupClientWriter(conn)

	// ==========================================
	// 如果是重連,發送重連成功訊息和遊戲資訊
	// ==========================================
	if isReconnection && player.CurrentGameSessionId != uuid.Nil {
		fmt.Printf("📤 Sending reconnection success message to %s (Session: %s)\n",
			player.Username, player.CurrentGameSessionId)

		s.mu.RLock()
		msgChan, exists := s.msgChan[conn]
		s.mu.RUnlock()

		if exists {
			// 發送 reconnected 訊息
			msgChan <- types.Message{
				Action: "reconnected",
				Payload: map[string]interface{}{
					"message":    "Successfully reconnected",
					"session_id": player.CurrentGameSessionId.String(),
					"username":   player.Username,
				},
			}

			// 重要！發送 game_found 訊息讓前端進入遊戲
			fmt.Println("Sending game_found message after reconnection...")
			msgChan <- types.Message{
				Action: "game_found",
				Payload: map[string]interface{}{
					"session_id": player.CurrentGameSessionId.String(),
				},
			}
		}
	}

	// 處理連線訊息
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

			// ==========================================
			// 情況 1: 網路斷線 (CloseAbnormalClosure 1006)
			// ==========================================
			// 處理:保留狀態,等待 30 秒重連
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				fmt.Printf("Network disconnection for %s. Keeping state for reconnection...\n", playerInfo)

				if exists {
					// 標記為重連狀態
					s.markPlayerAsReconnecting(player)

					// 只清理連線,保留玩家資料、session、queue
					s.cleanUpConnectionOnly(conn)

					// 啟動 30 秒計時器
					go s.handleReconnectionTimeout(playerID, 30*time.Second)
				} else {
					// 沒有玩家資訊,直接關閉連線
					conn.Close()
				}

				return // 退出此 goroutine
			}

			// ==========================================
			// 情況 2: 正常關閉 (CloseNormalClosure 1000)
			// ==========================================
			// 處理:完全清理所有資源
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				fmt.Printf("Player %s closed connection normally (intentional leave)\n", playerInfo)
				s.cleanUpClient(conn)
				return // 退出此 goroutine
			}

			// ==========================================
			// 情況 3: 離開頁面 (CloseGoingAway 1001)
			// ==========================================
			// 處理:保留狀態,等待 10 秒重連 (可能是頁面重新整理)
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

			// ==========================================
			// 情況 4: 其他 Unexpected Errors
			// ==========================================
			// 處理:視為暫時斷線,等待重連
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

			// ==========================================
			// 情況 5: 其他未知錯誤
			// ==========================================
			// 處理:完全清理
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
* 標記玩家為重連狀態
**/
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

/**
* 只清理連線,保留玩家資料、session、queue (用於重連情況)
**/
func (s *Server) cleanUpConnectionOnly(conn *websocket.Conn) {
	s.mu.Lock()

	// 獲取玩家資訊
	player, exists := s.connToPlayer[conn]
	if !exists {
		s.mu.Unlock()
		fmt.Println("cleanUpConnectionOnly: connection not found")
		conn.Close()
		return
	}

	fmt.Printf("Cleaning up connection only for player: %s (keeping game state)\n", player.Username)

	// 關閉並刪除 msgChan
	if ch, exists := s.msgChan[conn]; exists {
		close(ch)
		delete(s.msgChan, conn)
	}

	// 刪除 conn -> player 映射 (但不刪除 player 本身)
	delete(s.connToPlayer, conn)

	// 注意:不刪除 s.players[player.ID]  ← 保留玩家資料
	// 注意:不從 queue 移除              ← 保留排隊狀態
	// 注意:不從 session 移除            ← 保留遊戲狀態

	s.mu.Unlock()

	// 關閉 WebSocket 連線
	conn.Close()
}

/**
* 處理重連超時 - 如果玩家在指定時間內沒有重連,完全清理資源
**/
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

	// 檢查玩家狀態
	if player.ConnectState != nil && *player.ConnectState == constants.Reconnecting {
		fmt.Printf("Player %s failed to reconnect within %v. Cleaning up...\n", player.Username, timeout)
		s.cleanUpPlayerCompletely(playerID)
	} else {
		fmt.Printf("Player %s already reconnected\n", player.Username)
	}
}

/**
* 完全清理玩家 (根據 player ID)
**/
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
