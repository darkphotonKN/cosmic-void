package game

import (
	"github.com/google/uuid"
)

type CreateRoomParams struct {
	Name       string    `db:"name" json:"name"`
	CreatorID  uuid.UUID `db:"creator_id" json:"creator_id"`
	MaxPlayers int       `db:"max_players" json:"max_players"`
	GameMode   string    `db:"game_mode" json:"game_mode"`
}

type JoinRoomParams struct {
	RoomID uuid.UUID `db:"room_id" json:"room_id"`
	UserID uuid.UUID `db:"user_id" json:"user_id"`
	X      float64   `db:"x" json:"x"`
	Y      float64   `db:"y" json:"y"`
}

type UpdatePlayerPositionParams struct {
	PlayerID  uuid.UUID `db:"id" json:"id"`
	X         float64   `db:"x" json:"x"`
	Y         float64   `db:"y" json:"y"`
	VelocityX float64   `db:"velocity_x" json:"velocity_x"`
	VelocityY float64   `db:"velocity_y" json:"velocity_y"`
}

type UpdatePlayerHealthParams struct {
	PlayerID uuid.UUID `db:"id" json:"id"`
	Health   int       `db:"health" json:"health"`
}

type UpdatePlayerScoreParams struct {
	PlayerID uuid.UUID `db:"id" json:"id"`
	Score    int       `db:"score" json:"score"`
}