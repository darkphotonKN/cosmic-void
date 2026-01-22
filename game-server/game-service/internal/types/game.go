package types

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/google/uuid"
)

type Player struct {
	ID                   uuid.UUID
	Username             string
	CurrentGameSessionId uuid.UUID
	ConnectState         *constants.ConnectState
}

type PlayerState struct {
	ID        uuid.UUID        `json:"id"`
	EntityID  uuid.UUID        `json:"entity_id"`
	Username  string           `json:"username"`
	Position  *Position        `json:"position"`
	Direction *PlayerDirection `json:"direction"`
	Inventory []*ItemState     `json:"inventory"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type PlayerDirection struct {
	VX    float64 `json:"vx"`
	VY    float64 `json:"vy"`
	Speed float64 `json:"speed"`
}

type DoorState struct {
	EntityID uuid.UUID `json:"entity_id"`
	Position Position
	IsOpen   bool `json:"is_open"`
}

type ItemState struct {
	ItemID   uuid.UUID `json:"item_id"`
	EntityID uuid.UUID `json:"entity_id"`
	Name     string    `json:"name"`
	Quantity int       `json:"quantity"`
}
type ContainerState struct {
	ContainerID uuid.UUID    `json:"container_id"`
	EntityID    uuid.UUID    `json:"entity_id"`
	Position    *Position    `json:"position"`
	IsOpen      bool         `json:"is_open"`
	Items       []*ItemState `json:"items"`
}
