package gameserver

/**
* Manages all message types for websocket connections.
**/

type Message struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"package"`
}
