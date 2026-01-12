package types

import (
	"time"

	"github.com/google/uuid"
)

// ============ Game Sessions ============
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

// ============ End Game Sessions ============
