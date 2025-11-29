package gameserver

import (
	"encoding/json"
	"fmt"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/systems"
	"github.com/google/uuid"
)

/**
* Core concurrent message orchestrator.
**/

type messageHub struct {
	serverChan chan ClientPackage
}

func NewMessageHub(serverChan chan ClientPackage) *messageHub {
	return &messageHub{
		serverChan: serverChan,
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

			// TODO: Nick here
			// get player's game
			// game := getPlayerGame(clientPackage.payload.ID)

			switch clientPackage.Message.Action {

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
