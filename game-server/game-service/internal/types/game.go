package types

import "github.com/google/uuid"

type Player struct {
	ID       uuid.UUID
	Username string
}

type PlayerState struct {
	ID        uuid.UUID        `json:"id"`
	EntityID  uuid.UUID        `json:"entity_id"`
	Username  string           `json:"username"`
	Position  *Position        `json:"position"`
	Direction *PlayerDirection `json:"direction"`
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
