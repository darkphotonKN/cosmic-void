package gameserver

import (
	"fmt"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/systems"
	"github.com/google/uuid"
)

/**
* Core concurrent message orchestrator.
**/

type messageHub struct {
	sessionManager SessionManager
	gameSessionCh chan Message
}

type SessionManager interface {
	CreateGameSession(players []*game.Player) *game.Session
	GetGameSession(id uuid.UUID) (*game.Session, bool)
	GetServerChan() chan ClientPackage
}

func NewMessageHub(sessionManager SessionManager) *messageHub {
	return &messageHub{
		sessionManager: sessionManager,
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

			// TODO: Nick here
			// get player's game
			// game := getPlayerGame(clientPackage.payload.ID)

			switch clientPackage.Message.Action {

			// --- MENU RELATED ACTIONS ---

			// NOTE: queues a player for a game
			case "find_game":

				// NOTE: starts a new game
				// once enough players have joined.

			// TODO: NICK
			// Matchmaking system
			// for example player queue system ["nick", "kiki", "trump"]
			// goroutine to check ^ for 5 players

			// NICK log "game started" if game found
			// TODO: add start game session (Kranti)

			// give client message "game found!"
			// loop through found game player's id's, send them "game found"


			// --- GAME RELATED ACTIONS ---

			case "start_game":
				// test players
				testId := uuid.MustParse("0000-0000-0000-0001")
				playerOne := game.Player{
					ID:       testId,
					Username: "testPlayerOne",
				}

				testIdTwo := uuid.MustParse("0000-0000-0000-0002")
				playerTwo := game.Player{
					ID:       testIdTwo,
					Username: "testPlayerTwo",
				}

				testPlayerSlice := []*game.Player{&playerOne, &playerTwo}

				go h.startGameSession(testPlayerSlice, clientPackage.Message)
			}
		}
	}
}

const framerate = 1

/**
* Handles all workings inside a single game session.
* NOTE: this method runs in a goroutine.
**/
func (h *messageHub) startGameSession(players []*game.Player, message Message) {
	newGameSession := h.sessionManager.CreateGameSession(players)

	// --- client actions ---
	clientAction := <- 

	// update game loop
	ticker := time.NewTicker((1 * time.Second) / framerate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			// --- update game loop ---
			fmt.Println("update game loop")
			entities := newGameSession.EntityManager.GetAllEntities()
			movementSys := systems.MovementSystem{}
			movementSys.Update(float64(1), entities)
		}
	}
}
