package gameserver

import (
	"fmt"
	"sync"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
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

				sessionID, err := clientPackage.Message.GetSessionID()

				if err != nil {
					// TODO: send back to client the error
					continue
				}

				session, exists := h.sessionManager.GetGameSession(*sessionID)

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

				// NOTE: starts a new game
				// once enough players have joined.

				// TODO: NICK
				// Matchmaking system
				// for example player queue system ["nick", "kiki", "trump"]
				// goroutine to check ^ for 5 players

				// NICK log "game started" if game found
				// TODO: add real conditional to start game session (Kranti)
				start := true
				if start {
					// test players
					testId := uuid.MustParse("00000000-0000-0000-0000-000000000001")
					playerOne := types.Player{
						ID:       testId,
						Username: "testPlayerOne",
					}

					testIdTwo := uuid.MustParse("00000000-0000-0000-0000-000000000002")
					playerTwo := types.Player{
						ID:       testIdTwo,
						Username: "testPlayerTwo",
					}
					testPlayerSlice := []*types.Player{&playerOne, &playerTwo}

					go h.sessionManager.CreateGameSession(testPlayerSlice)
				}

				// give client message "game found!"
				// loop through found game player's id's, send them "game found"

			case constants.ActionLeaveQueue:
				// TODO: add client leaving queue
				fmt.Println("Leave game...")

			}
		}
	}
}
