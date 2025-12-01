package gameserver

import (
	"fmt"
	"sync"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/systems"
	"github.com/google/uuid"
)

/**
* Core concurrent message orchestrator.
**/

type messageHub struct {
	serverChan chan ClientPackage
	sessions   map[string]*game.Session
	mu         sync.RWMutex
}

func NewMessageHub(serverChan chan ClientPackage) *messageHub {
	return &messageHub{
		serverChan: serverChan,
		sessions:   make(map[string]*game.Session),
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
		case clientPackage := <-h.serverChan:
			// handle message based on action
			fmt.Printf("\nincoming message: %+v\n\n", clientPackage.Message)

			switch clientPackage.Message.Action {
			// action == start game
			case "kiki_join":
				go h.handlePlayerJoin(clientPackage)

			// NOTE: queues a player for a game
			case "queue":

				// NOTE: starts a new game
				// once enough players have joined.

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

/**
* Handles all workings inside a single game session.
* NOTE: this method runs in a goroutine.
**/
func (h *messageHub) startGameSession(players []*game.Player, message Message) {
	game := game.NewSession("123")

	for _, player := range players {
		game.AddPlayer(player.ID, player.Username)
	}

	// update game loop
	// TODO: add once per second ticket
	entities := game.EntityManager.GetAllEntities()
	movementSys := systems.MovementSystem{}
	movementSys.Update(float64(1), entities)
}

func (h *messageHub) handlePlayerJoin(clientPackage ClientPackage) {
	roomID := "room-1"
	testIdTwo := uuid.MustParse("0000-0000-0000-0002")
	h.mu.RLock()
	newGame, exists := h.sessions[roomID]
	h.mu.RUnlock()

	h.mu.Lock()
	if !exists {
		newGame = game.NewSession(roomID)
		h.sessions[roomID] = newGame
		fmt.Println("Created new game session: ", newGame)
	}
	h.mu.Unlock()

	newGame.AddPlayer(testIdTwo, "player1")
}
