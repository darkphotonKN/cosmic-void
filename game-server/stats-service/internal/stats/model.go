package stats

import (
	"time"

	"github.com/google/uuid"
)

type PlayerStats struct {
	ID               uuid.UUID `db:"id" json:"id"`
	PlayerID         uuid.UUID `db:"player_id" json:"player_id"`
	Level            int32     `db:"level" json:"level"`
	XP               int32     `db:"xp" json:"xp"`
	TotalMatches     int32     `db:"total_matches" json:"total_matches"`
	Wins             int32     `db:"wins" json:"wins"`
	Losses           int32     `db:"losses" json:"losses"`
	Kills            int32     `db:"kills" json:"kills"`
	Deaths           int32     `db:"deaths" json:"deaths"`
	Assists          int32     `db:"assists" json:"assists"`
	KDRatio          float32   `db:"kd_ratio" json:"kd_ratio"`
	WinRate          float32   `db:"win_rate" json:"win_rate"`
	ItemsCollected   int32     `db:"items_collected" json:"items_collected"`
	DamageDealt      float32   `db:"damage_dealt" json:"damage_dealt"`
	DamageTaken      float32   `db:"damage_taken" json:"damage_taken"`
	PlayTimeSeconds  int32     `db:"play_time_seconds" json:"play_time_seconds"`
	CurrentStreak    int32     `db:"current_streak" json:"current_streak"`
	BestStreak       int32     `db:"best_streak" json:"best_streak"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

type MatchHistory struct {
	ID              uuid.UUID `db:"id" json:"id"`
	MatchID         uuid.UUID `db:"match_id" json:"match_id"`
	PlayerID        uuid.UUID `db:"player_id" json:"player_id"`
	Won             bool      `db:"won" json:"won"`
	Kills           int32     `db:"kills" json:"kills"`
	Deaths          int32     `db:"deaths" json:"deaths"`
	Assists         int32     `db:"assists" json:"assists"`
	DamageDealt     float32   `db:"damage_dealt" json:"damage_dealt"`
	DamageTaken     float32   `db:"damage_taken" json:"damage_taken"`
	ItemsCollected  int32     `db:"items_collected" json:"items_collected"`
	DurationSeconds int32     `db:"duration_seconds" json:"duration_seconds"`
	XPGained        int32     `db:"xp_gained" json:"xp_gained"`
	PlayedAt        time.Time `db:"played_at" json:"played_at"`
}

func (ps *PlayerStats) CalculateLevel() {
	levels := []int32{0, 100, 250, 450, 700, 1000, 1400, 1900, 2500, 3200}
	for i, threshold := range levels {
		if ps.XP < threshold {
			ps.Level = int32(i)
			return
		}
	}
	ps.Level = int32(len(levels))
}

func (ps *PlayerStats) CalculateRatios() {
	if ps.Deaths > 0 {
		ps.KDRatio = float32(ps.Kills) / float32(ps.Deaths)
	} else if ps.Kills > 0 {
		ps.KDRatio = float32(ps.Kills)
	}

	if ps.TotalMatches > 0 {
		ps.WinRate = float32(ps.Wins) / float32(ps.TotalMatches) * 100
	}
}