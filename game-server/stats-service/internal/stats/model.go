package stats

import (
	"time"

	"github.com/google/uuid"
)

// PlayerMatchStats represents aggregated match statistics for a player
type PlayerMatchStats struct {
	ID                  uuid.UUID `db:"id" json:"id"`
	MemberID            uuid.UUID `db:"member_id" json:"member_id"`
	GamesPlayed         int32     `db:"games_played" json:"games_played"`
	Wins                int32     `db:"wins" json:"wins"`
	Losses              int32     `db:"losses" json:"losses"`
	Kills               int32     `db:"kills" json:"kills"`
	Deaths              int32     `db:"deaths" json:"deaths"`
	TimesPlacedTopThree int32     `db:"times_placed_top_three" json:"times_placed_top_three"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}

// PlayerRankingStats represents ranking and leaderboard data
type PlayerRankingStats struct {
	ID               uuid.UUID `db:"id" json:"id"`
	MemberID         uuid.UUID `db:"member_id" json:"member_id"`
	Username         string    `db:"username" json:"username"`
	Wins             int32     `db:"wins" json:"wins"`
	TopThrees        int32     `db:"top_threes" json:"top_threes"`
	Rating           int32     `db:"rating" json:"rating"`
	RankPosition     *int32    `db:"rank_position" json:"rank_position"`
	LastCalculatedAt time.Time `db:"last_calculated_at" json:"last_calculated_at"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// MatchHistory represents a single match record
type MatchHistory struct {
	ID             uuid.UUID `db:"id" json:"id"`
	SessionID      uuid.UUID `db:"session_id" json:"session_id"`
	MemberID       uuid.UUID `db:"member_id" json:"member_id"`
	Win            bool      `db:"win" json:"win"`
	FinalPosition  int32     `db:"final_position" json:"final_position"`
	Kills          int32     `db:"kills" json:"kills"`
	Deaths         int32     `db:"deaths" json:"deaths"`
	RatingBefore   *int32    `db:"rating_before" json:"rating_before"`
	RatingAfter    *int32    `db:"rating_after" json:"rating_after"`
	RatingChange   *int32    `db:"rating_change" json:"rating_change"`
	MatchStartedAt time.Time `db:"match_started_at" json:"match_started_at"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// updating player match stats
type UpdateStatsParams struct {
	MemberID            uuid.UUID `db:"member_id"`
	GamesPlayed         int       `db:"games_played"`
	Wins                int       `db:"wins"`
	Losses              int       `db:"losses"`
	Kills               int       `db:"kills"`
	Deaths              int       `db:"deaths"`
	TimesPlacedTopThree int       `db:"times_placed_top_three"`
}
