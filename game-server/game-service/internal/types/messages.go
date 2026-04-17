package types

import (
	"fmt"
	"log/slog"

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
	Error   *string                `json:"error,omitempty"`
}

/**
* Provides the abstraction for clients to interface with the websocket connections.
**/

type ClientPackage struct {
	Message Message
	Conn    *websocket.Conn
}

// represents entire game state that client receives
type ClientGameState struct {
	SessionID     uuid.UUID          `json:"session_id"`
	CurrentPlayer *PlayerState       `json:"current_player"` // The recipient's player state
	OtherPlayers  []*PlayerState     `json:"other_players"`  // All other players
	Items         []uuid.UUID        `json:"items"`          // TODO: update with item entity converted into struct format
	Doors         []*DoorState       `json:"doors"`
	Walls         []*WallState       `json:"walls"`
	Containers    []*ContainerState  `json:"containers"`
	EscapeDoor    []*EscapeDoorState `json:"escape_doors"`
	Equipment     *EquipmentState    `json:"equipment"`
	Switch        []*SwitchState     `json:"switches"`
	EscapedCount  int                `json:"escaped_count"`
}

type BackendGameState struct {
	SessionID    uuid.UUID
	Players      map[uuid.UUID]*PlayerState
	Items        []uuid.UUID
	Doors        []*DoorState
	Walls        []*WallState
	Containers   []*ContainerState
	EscapeDoor   []*EscapeDoorState
	Equipment    *EquipmentState
	Switch       []*SwitchState
	EscapedCount int
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

		slog.Debug("payload of action move", "payload", parsedPayload)

		return parsedPayload, nil

	case constants.ActionInteract:
		parsedPayload := PlayerSessionInteractPayload{
			PlayerSessionPayload: PlayerSessionPayload{
				SessionID: m.Payload["session_id"].(string),
				PlayerID:  m.Payload["player_id"].(string),
			},
			EntityID: m.Payload["entity_id"].(string),
		}

		slog.Debug("payload of action interact", "payload", parsedPayload)

		return parsedPayload, nil

	case constants.ActionAttack:
		parsedPayload := PlayerSectionAttackPayload{
			PlayerSessionPayload: PlayerSessionPayload{
				SessionID: m.Payload["session_id"].(string),
				PlayerID:  m.Payload["player_id"].(string),
			},
			EnemyEntityID: m.Payload["enemy_entity_id"].(string),
		}

		slog.Debug("payload of action attack", "payload", parsedPayload)

		return parsedPayload, nil

	case constants.ActionEquip, constants.ActionUnequip:
		parsedPayload := PlayerEquipPayload{
			PlayerSessionPayload: PlayerSessionPayload{
				SessionID: m.Payload["session_id"].(string),
				PlayerID:  m.Payload["player_id"].(string),
			},
			ItemEntityID: m.Payload["item_entity_id"].(string),
		}

		slog.Debug("payload of action equip / unequip", "payload", parsedPayload)

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
		slog.Debug("SessionID does not exist in the payload")
		return uuid.Nil, fmt.Errorf("SessionID does not exist in the payload.")
	}

	sessionID, err := uuid.Parse(sessionIDStr)

	if err != nil {
		slog.Debug("SessionID in payload is not a UUID")
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

type PlayerSectionAttackPayload struct {
	PlayerSessionPayload
	EnemyEntityID string `json:"enemy_entity_id"`
}

type PlayerEquipPayload struct {
	PlayerSessionPayload
	ItemEntityID string `json:"item_entity_id"`
}
