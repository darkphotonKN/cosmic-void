package gameserver

import "fmt"

/**
* Core concurrent message orchestrator.
**/

type messageHub struct {
	msgChan chan ClientPackage
}

func NewMessageHub(msgChan chan ClientPackage) *messageHub {
	return &messageHub{
		msgChan: msgChan,
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
		case clientPackage := <-h.msgChan:
			// handle message based on action

			fmt.Printf("\nincoming message: %+v\n\n", clientPackage.Message)
		}
	}
}
