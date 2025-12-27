package gameserver

import (
	"fmt"
	"sync"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/messaging"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/systems"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/**
* Core concurrent message orchestrator.
**/

type messageHub struct {
	sessionManager SessionManager
	gameSessionCh  chan types.Message
	sessions       map[string]*game.Session
	mu             sync.RWMutex
	sender         *messaging.MessageSender
}

type SessionManager interface {
	CreateGameSession(players []*types.Player) *game.Session
	GetGameSession(id uuid.UUID) (*game.Session, bool)
	GetServerChan() chan types.ClientPackage
	AddPlayerToQueue(*types.Player)
	GetPlayerFromConn(conn *websocket.Conn) (*types.Player, bool)
	GetMatchedChan() chan []*types.Player
	GetQueueStatusChan() chan systems.QueueStatus
}

func NewMessageHub(sessionManager SessionManager, sender *messaging.MessageSender) *messageHub {
	return &messageHub{
		sessionManager: sessionManager,
		sessions:       make(map[string]*game.Session),
		sender:         sender,
	}
}

/**
* Core goroutine hub to handle all incoming messages and orchestrate them
* to other parts of game.
**/
func (h *messageHub) Run() {
	fmt.Printf("\nInitializing message hub...\n\n")

	for {
		select {
		case clientPackage := <-h.sessionManager.GetServerChan():
			fmt.Printf("\nincoming message: %+v\n\n", clientPackage.Message)

			response := types.NewResponseBuilder()

			// handle message based on action
			var gameActions map[constants.Action]bool = map[constants.Action]bool{
				constants.ActionMove:   true,
				constants.ActionAttack: true,
			}

			messageAction := constants.Action(clientPackage.Message.Action)

			// --- GAME RELATED ACTIONS ---
			// any message sent from the client after a game session is initialized
			// will be propogated from the messsage hub to corresponding server.

			if gameActions[messageAction] {
				sessionID, err := clientPackage.Message.GetSessionID()

				if err != nil {
					// 傳入 conn 作為參數
					response.Error(
						clientPackage.Conn,
						clientPackage.Message.Action,
						constants.ErrorInvalidSessionID,
						"Invalid or missing session ID in payload",
					)
					continue
				}

				session, exists := h.sessionManager.GetGameSession(sessionID)

				if !exists {
					// 傳入 conn 作為參數
					response.Error(
						clientPackage.Conn,
						clientPackage.Message.Action,
						constants.ErrorSessionNotFound,
						fmt.Sprintf("Game session not found for session ID: %s", sessionID),
					)
					fmt.Printf("\ngame doesn't exist for this player, message: %+v\n\n", clientPackage.Message)
					continue
				}

				// propogate message to corresponding game
				session.MessageCh <- clientPackage
				continue
			}

			// --- MENU RELATED ACTIONS ---
			// These actions will be actions for before game initialization happens.
			switch messageAction {

			// NOTE: queues a player for a game
			case constants.ActionFindGame:
				fmt.Println("ActionFindGame")
				player, exists := h.sessionManager.GetPlayerFromConn(clientPackage.Conn)

				if !exists {
					// 傳入 conn 作為參數
					h.sender.SendToPlayer(player.ID, types.Message{
						Action: string(constants.ActionFindGame),
						Payload: map[string]interface{}{
							"message":   "Successfully joined matchmaking queue",
							"player_id": player.ID.String(),
							"username":  player.Username,
						},
					})
					fmt.Println("Player not found for connection")
					continue
				}

				// 將 player 加入 queue，QueueSystem 會透過 channel 處理
				// 配對成功後會自動呼叫 Server.onMatchFound callback
				h.sessionManager.AddPlayerToQueue(player)
				fmt.Printf("Player %s added to matchmaking queue\n", player.Username)

				// 傳入 conn 作為參數
				response.Success(clientPackage.Conn, clientPackage.Message.Action, map[string]interface{}{
					"message":   "Successfully joined matchmaking queue",
					"player_id": player.ID.String(),
					"username":  player.Username,
				})

			case constants.ActionLeaveQueue:
				player, exists := h.sessionManager.GetPlayerFromConn(clientPackage.Conn)
				if !exists {
					// 傳入 conn 作為參數
					response.Error(
						clientPackage.Conn,
						clientPackage.Message.Action,
						constants.ErrorPlayerNotFound,
						"Player not found for connection",
					)
					continue
				}

				// TODO: 實現離開隊列邏輯
				// h.sessionManager.RemovePlayerFromQueue(player)
				fmt.Println("Leave game...")

				// 傳入 conn 作為參數
				response.Success(clientPackage.Conn, clientPackage.Message.Action, map[string]interface{}{
					"message":   "Successfully left the queue",
					"player_id": player.ID.String(),
				})

			default:
				// 傳入 conn 作為參數
				response.Error(
					clientPackage.Conn,
					clientPackage.Message.Action,
					constants.ErrorInvalidPayload,
					fmt.Sprintf("Unknown action: %s", messageAction),
				)
			}

		// 監聯配對成功的 channel
		case matchedPlayers := <-h.sessionManager.GetMatchedChan():
			fmt.Printf("Received matched players, creating game session...\n")
			fmt.Println(matchedPlayers)
			session := h.sessionManager.CreateGameSession(matchedPlayers)
			h.sender.BroadcastToPlayerList(matchedPlayers,
				types.Message{
					Action: "game_found",
					Payload: map[string]any{
						"session_id": session.ID.String(),
					},
				})

		// 監聽排隊狀態更新
		case status := <-h.sessionManager.GetQueueStatusChan():
			fmt.Printf("Queue status update: %d/%d\n", status.Current, status.Total)
			h.sender.BroadcastToPlayerList(status.Players,
				types.Message{
					Action: "queue_status",
					Payload: map[string]any{
						"current": status.Current,
						"total":   status.Total,
					},
				})
		}
	}
}
