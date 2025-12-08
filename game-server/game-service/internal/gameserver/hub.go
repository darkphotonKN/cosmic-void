package gameserver

import (
	"fmt"
	"sync"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
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
}

func NewMessageHub(sessionManager SessionManager) *messageHub {
	return &messageHub{
		sessionManager: sessionManager,
		sessions:       make(map[string]*game.Session),
	}
}

/**
* core goroutine hub to handle all incoming messages and orchestrate them
* to other parts of game.
**/
func (h *messageHub) Run() {
	fmt.Printf("\nInitializing message hub...\n\n")

	for {
		select {
		case clientPackage := <-h.sessionManager.GetServerChan():
			// handle message based on action
			fmt.Printf("\nincoming message: %+v\n\n", clientPackage.Message)

			var gameActions map[constants.Action]bool = map[constants.Action]bool{
				constants.ActionMove:   true,
				constants.ActionAttack: true,
			}

			messageAction := constants.Action(clientPackage.Message.Action)

			// --- GAME RELATED ACTIONS ---
			// any message sent from the client after a game session is initialized
			// will be propogated from the messsage hub to corresponding server.

			messageParsedPayload, err := clientPackage.Message.ParsePayload()
			if err != nil {
				fmt.Printf("client message could not be parsed, err: %v\n", err)
				// TODO: add error handling to client
			}

			if gameActions[messageAction] {
				// get correct game session from payload
				// TODO: mismatching types
				sessionIDStr := messageParsedPayload.(types.PlayerSessionPayload).SessionID
				sessionID := uuid.MustParse(sessionIDStr)

				fmt.Printf("\nSessionID parsed was: %s\n\n", sessionIDStr)
				session, exists := h.sessionManager.GetGameSession(sessionID)

				if !exists {
					// TODO: return to client game doesn't exist

					fmt.Printf("\ngame doesn't exist for this player, message: %+v\n\n", clientPackage.Message)
					continue
				}

				// propogate message to corresponding game
				session.MessageCh <- clientPackage.Message
			}

			// --- MENU RELATED ACTIONS ---
			// These actions will be actions for before game initialization happens.
			switch messageAction {

			// NOTE: queues a player for a game
			case constants.ActionFindGame:
				player, exists := h.sessionManager.GetPlayerFromConn(clientPackage.Conn)
				if !exists {
					fmt.Println("Player not found for connection")
					continue
				}

				// 將 player 加入 queue，QueueSystem 會透過 channel 處理
				// 配對成功後會自動呼叫 Server.onMatchFound callback
				h.sessionManager.AddPlayerToQueue(player)
				fmt.Printf("Player %s added to matchmaking queue\n", player.Username)

			case constants.ActionLeaveQueue:
				// TODO: add client leaving queue
				fmt.Println("Leave game...")

			}

		// 監聽配對成功的 channel
		case matchedPlayers := <-h.sessionManager.GetMatchedChan():
			fmt.Printf("Received matched players, creating game session...\n")
			go h.sessionManager.CreateGameSession(matchedPlayers)
		}
	}
}
