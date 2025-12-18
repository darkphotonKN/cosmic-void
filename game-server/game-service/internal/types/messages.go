package types

import (
	"fmt"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/**
* Manages all message types for websocket connections.
**/

type Message struct {
	Action  string                 `json:"action"`
	Payload map[string]interface{} `json:"payload"`
}

/**
* Provides the abstraction for clients to interface with the websocket connections.
**/

type ClientPackage struct {
	Message Message
	Conn    *websocket.Conn
}

type ServerResponse struct {
	Action  string                 `json:"action"`
	Payload map[string]interface{} `json:"payload"`
	Success bool                   `json:"success,omitempty"`
	Error   *ErrorResponse         `json:"error,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (m *Message) ParsePayload() (interface{}, error) {

	switch constants.Action(m.Action) {
	case constants.ActionMove:
		parsedPayload := PlayerSessionMovePayload{
			PlayerSessionPayload: PlayerSessionPayload{
				SessionID: m.Payload["session_id"].(string),
				PlayerID:  m.Payload["player_id"].(string),
			},
			Vx: m.Payload["vx"].(float64),
			Vy: m.Payload["vy"].(float64),
		}

		fmt.Printf("\n\npayload of action move was: %+v\n", parsedPayload)

		return parsedPayload, nil

	case constants.ActionInteract:
		parsedPayload := PlayerSessionInteractPayload{
			PlayerSessionPayload: PlayerSessionPayload{
				SessionID: m.Payload["session_id"].(string),
				PlayerID:  m.Payload["player_id"].(string),
			},
			EntityID: m.Payload["entity_id"].(string),
		}

		fmt.Printf("\n\npayload of action interact was: %+v\n", parsedPayload)

		return parsedPayload, nil
	default:
		return nil, fmt.Errorf("No matching actions.")
	}

}

/**
* helper to extract sessionID.
**/
func (m *Message) GetSessionID() (uuid.UUID, error) {
	sessionIDStr, ok := m.Payload["session_id"].(string)

	if !ok {
		fmt.Printf("SessionID does not exist in the payload.")
		return uuid.Nil, fmt.Errorf("SessionID does not exist in the payload.")
	}

	sessionID, err := uuid.Parse(sessionIDStr)

	if err != nil {
		fmt.Printf("SessionID in payload is not a UUID.")
		return uuid.Nil, fmt.Errorf("SessionID in payload is not a UUID.")
	}

	return sessionID, nil
}

/**
* Payloads for players in ongoing games
**/
type PlayerSessionPayload struct {
	SessionID string `json:"session_id"`
	PlayerID  string `json:"player_id"`
}

type PlayerSessionMovePayload struct {
	PlayerSessionPayload
	Vx float64 `json:"vx"`
	Vy float64 `json:"vy"`
}

type PlayerSessionInteractPayload struct {
	PlayerSessionPayload
	EntityID string `json:"entity_id"`
}
