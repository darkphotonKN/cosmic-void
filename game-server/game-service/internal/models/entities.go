package models

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	CreatorID      uuid.UUID `json:"creator_id" db:"creator_id"`
	MaxPlayers     int       `json:"max_players" db:"max_players"`
	CurrentPlayers int       `json:"current_players" db:"current_players"`
	GameMode       string    `json:"game_mode" db:"game_mode"`
	Status         string    `json:"status" db:"status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type Player struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	RoomID    uuid.UUID `json:"room_id" db:"room_id"`
	X         float64   `json:"x" db:"x"`
	Y         float64   `json:"y" db:"y"`
	VelocityX float64   `json:"velocity_x" db:"velocity_x"`
	VelocityY float64   `json:"velocity_y" db:"velocity_y"`
	Health    int       `json:"health" db:"health"`
	Score     int       `json:"score" db:"score"`
	IsAlive   bool      `json:"is_alive" db:"is_alive"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
}