package gameserver

import (
	"fmt"
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

			// action == start game

		}
	}
}
