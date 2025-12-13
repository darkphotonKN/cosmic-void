package gameserver

import (
	"fmt"
	"sync"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
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
}

type SessionManager interface {
	CreateGameSession(players []*types.Player) *game.Session
	GetGameSession(id uuid.UUID) (*game.Session, bool)
	GetServerChan() chan types.ClientPackage
	AddPlayerToQueue(*types.Player)
	GetPlayerFromConn(conn *websocket.Conn) (*types.Player, bool)
	GetMatchedChan() chan []*types.Player
	GetQueueStatusChan() chan systems.QueueStatus
	SendToPlayer(playerID uuid.UUID, msg types.Message)
}

func NewMessageHub(sessionManager SessionManager) *messageHub {
	return &messageHub{
		sessionManager: sessionManager,
		sessions:       make(map[string]*game.Session),
	}
}

/**
* Core goroutine hub to handle all incoming messages and orchestrate them
* to other parts of game.
**/
func (h *messageHub) Run() {
	fmt.Printf("\nInitializing message hub...\n\n")
	defer fmt.Printf("EXIST HUB")
	for {

		// time.Sleep(time.Second * 2)
		fmt.Println("\n123gg")
		select {
		case clientPackage := <-h.sessionManager.GetServerChan():
			// handle message based on action
			fmt.Printf("\nincoming message: %+v\n\n", clientPackage.Message)

			response := types.NewResponseBuilder(clientPackage.Conn)

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
					response.Error(
						clientPackage.Message.Action,
						constants.ErrorInvalidSessionID,
						"Invalid or missing session ID in payload",
					)
					continue
				}

				session, exists := h.sessionManager.GetGameSession(sessionID)

				if !exists {
					response.Error(
						clientPackage.Message.Action,
						constants.ErrorSessionNotFound,
						fmt.Sprintf("Game session not found for session ID: %s", sessionID),
					)
					fmt.Printf("\ngame doesn't exist for this player, message: %+v\n\n", clientPackage.Message)
					continue
				}

				// propogate message to corresponding game
				session.MessageCh <- clientPackage
			}

			// --- MENU RELATED ACTIONS ---
			// These actions will be actions for before game initialization happens.
			switch messageAction {

			// NOTE: queues a player for a game
			case constants.ActionFindGame:
				fmt.Println("ActionFindGame")
				player, exists := h.sessionManager.GetPlayerFromConn(clientPackage.Conn)

				if !exists {
					response.Error(
						clientPackage.Message.Action,
						constants.ErrorPlayerNotFound,
						"Player not found for connection",
					)
					fmt.Println("Player not found for connection")
					continue
				}

				// 將 player 加入 queue，QueueSystem 會透過 channel 處理
				// 配對成功後會自動呼叫 Server.onMatchFound callback
				h.sessionManager.AddPlayerToQueue(player)
				fmt.Printf("Player %s added to matchmaking queue\n", player.Username)

				response.Success(clientPackage.Message.Action, map[string]interface{}{
					"message":   "Successfully joined matchmaking queue",
					"player_id": player.ID.String(),
					"username":  player.Username,
				})

			case constants.ActionLeaveQueue:
				player, exists := h.sessionManager.GetPlayerFromConn(clientPackage.Conn)
				if !exists {
					response.Error(
						clientPackage.Message.Action,
						constants.ErrorPlayerNotFound,
						"Player not found for connection",
					)
					continue
				}

				// TODO: 實現離開隊列邏輯
				// h.sessionManager.RemovePlayerFromQueue(player)
				fmt.Println("Leave game...")

				response.Success(clientPackage.Message.Action, map[string]interface{}{
					"message":   "Successfully left the queue",
					"player_id": player.ID.String(),
				})

			default:
				response.Error(
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
			for _, player := range matchedPlayers {
				h.sessionManager.SendToPlayer(player.ID, types.Message{
					Action: "game_found",
					Payload: map[string]interface{}{
						"sessionID": session.ID.String(),
					},
				})
			}

		// 監聽排隊狀態更新
		case status := <-h.sessionManager.GetQueueStatusChan():
			fmt.Printf("Queue status update: %d/%d\n", status.Current, status.Total)
			for _, player := range status.Players {
				h.sessionManager.SendToPlayer(player.ID, types.Message{
					Action: "queue_status",
					Payload: map[string]interface{}{
						"current": status.Current,
						"total":   status.Total,
					},
				})
			}
		}

	}
}
