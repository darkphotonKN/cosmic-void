package types

import (
	"github.com/gorilla/websocket"
)

/**
* Manages all message types for websocket connections.
**/

type Message struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"package"`
}

// TODO: NICK add payload parse logic here.
func (m *Message) verifyPayload(action string) {
	switch action {
	case "move":

	}

}

/**
* Provides the abstraction for clients to interface with the websocket connections.
**/

type ClientPackage struct {
	Message Message
	Conn    *websocket.Conn
}
