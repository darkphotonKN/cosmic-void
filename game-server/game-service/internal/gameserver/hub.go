package gameserver

/**
* Core concurrent message orchestrator.
**/

type messageHub struct {
}

func NewMessageHub() *messageHub {
	return &messageHub{}
}

/**
* core goroutine hub to handle all incoming messages and orchestrate them
* to other parts of game.
**/
func (h *messageHub) RunHub() {
}
