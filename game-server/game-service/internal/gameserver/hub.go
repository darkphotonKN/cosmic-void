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
	GetQueueChan() chan []*types.Player
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

			if gameActions[messageAction] {

				// get correct game session from payload
				testPlayerGameSession := uuid.MustParse("10000000-0000-0000-0000-000000000000")

				session, exists := h.sessionManager.GetGameSession(testPlayerGameSession)

				if !exists {
					// TODO: return to client game doesn't exist
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

				// NOTE: starts a new game
				// once enough players have joined.

				// TODO: NICK
				// Matchmaking system
				// for example player queue system ["nick", "kiki", "trump"]
				player, exists := h.sessionManager.GetPlayerFromConn(clientPackage.Conn)

				if !exists {
					fmt.Println("player is exist")
					continue
				}

				go h.sessionManager.AddPlayerToQueue(player)
				// goroutine to check ^ for 5 players

				// for {
				// 	select {
				// 	case queueResult := <-h.sessionManager.GetQueueChan():
				// 		for _, player := range queueResult {
				// 			fmt.Println("queue player", player.Username)
				// 		}
				// 		if len(queueResult) == 2 {
				// 			go h.sessionManager.CreateGameSession(queueResult)
				// 		}
				// 	}
				// }
				// NICK log "game started" if game found
				// TODO: add real conditional to start game session (Kranti)
				// start := true
				// select 等goroutin人數滿或時間到開始遊戲
				// if ok {
				// test players
				// testId := uuid.MustParse("00000000-0000-0000-0000-000000000001")
				// playerOne := types.Player{
				// 	ID:       testId,
				// 	Username: "testPlayerOne",
				// }

				// testIdTwo := uuid.MustParse("00000000-0000-0000-0000-000000000002")
				// playerTwo := types.Player{
				// 	ID:       testIdTwo,
				// 	Username: "testPlayerTwo",
				// }
				// testPlayerSlice := []*types.Player{&playerOne, &playerTwo}

				// 	go h.sessionManager.CreateGameSession(fullQueue)
				// }

				// give client message "game found!"
				// loop through found game player's id's, send them "game found"

			case constants.ActionLeaveQueue:
				// TODO: add client leaving queue
				fmt.Println("Leave game...")

			}
		case queueResult := <-h.sessionManager.GetQueueChan():
			for _, player := range queueResult {
				fmt.Println("queue player", player.Username)
			}
			// if len(queueResult) == 2 {
			fmt.Println("create game")
			go h.sessionManager.CreateGameSession(queueResult)
			// }
		}
	}
}
