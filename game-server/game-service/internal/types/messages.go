package types

import (
	"fmt"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/gorilla/websocket"
)

/**
* Manages all message types for websocket connections.
**/

type Message struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"package"`
}

func (m *Message) ParsePayload() (interface{}, error) {

	payload, ok := m.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("payload base type was incorrect.")
	}

	switch constants.Action(m.Action) {
	case constants.ActionMove:
		// assert for the correct payload type when action == move
		parsedPayload := PlayerSessionMovePayload{
			PlayerSessionPayload: PlayerSessionPayload{
				SessionID: payload["sessionId"].(string),
				PlayerID:  payload["playerId"].(string),
			},
			Vx: payload["vx"].(float64),
			Vy: payload["vy"].(float64),
		}

		fmt.Printf("\n\npayload of action move was: %+v\n", parsedPayload)

		return parsedPayload, nil

	default:
		return nil, fmt.Errorf("No matching actions.")
	}

}

/**
* Payloads for players in ongoing games
**/
type PlayerSessionPayload struct {
	SessionID string `json:"sessionId"`
	PlayerID  string `json:"PlayerId"`
}

type PlayerSessionMovePayload struct {
	PlayerSessionPayload
	Vx float64 `json:"vx"`
	Vy float64 `json:"vy"`
}

/**
* Provides the abstraction for clients to interface with the websocket connections.
**/

type ClientPackage struct {
	Message Message
	Conn    *websocket.Conn
}
