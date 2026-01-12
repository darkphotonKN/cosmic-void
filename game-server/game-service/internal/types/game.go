package types

import (
	"time"

	"github.com/google/uuid"
)

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

type MatchEndState struct {
	SessionID          uuid.UUID `json:"session_id"`
	MatchStartedAt     time.Time `json:"match_started_at"`
	MatchEndedAt       time.Time `json:"match_ended_at"`
	PlayerMatchResults []*PlayerMatchResults
}

type PlayerMatchResults struct {
	MemberID      string `json:"member_id"`
	Username      string `json:"username"`
	Win           bool   `json:"win"`
	FinalPosition int32  `json:"final_position"`
	Kills         int32  `json:"kills"`
	Deaths        int32  `json:"deaths"`
}
