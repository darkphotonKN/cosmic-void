package stats

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type repository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{
		DB: db,
	}
}

/**
* update the baseline player stats but increasing or decreasing the aggregate values.
**/
func (r *repository) UpsertPlayerMatchStats(ctx context.Context, params *UpdateStatsParams) (*PlayerMatchStats, error) {
	rows, err := r.DB.NamedQueryContext(ctx, `
	INSERT INTO player_match_stats (
			member_id,
			games_played,
			wins,
			losses,
			kills,
			deaths,
			times_placed_top_three,
			created_at,
			updated_at
	)
	VALUES (
			:member_id,
			:games_played,
			:wins,
			:losses,
			:kills,
			:deaths,
			:times_placed_top_three,
			CURRENT_TIMESTAMP,
			CURRENT_TIMESTAMP
	)
	ON CONFLICT (member_id)
	DO UPDATE SET
			games_played = EXCLUDED.games_played,
			wins = EXCLUDED.wins,
			losses = EXCLUDED.losses,
			kills = EXCLUDED.kills,
			deaths = EXCLUDED.deaths,
			times_placed_top_three = EXCLUDED.times_placed_top_three,
			updated_at = CURRENT_TIMESTAMP
	RETURNING *;
`, params)

	if err != nil {
		slog.Info("Errored when attempting to update player stats", "err", err)
		return nil, err
	}

	defer rows.Close()

	var updated PlayerMatchStats

	if err := rows.StructScan(&updated); err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *repository) GetPlayerMatchStats(ctx context.Context, memberID uuid.UUID) (*PlayerMatchStats, error) {
	query := `
	SELECT 
			id,
			member_id,
			games_played,
			wins,
			losses,
			kills,
			deaths,
			times_placed_top_three,
			created_at,
			updated_at
	FROM player_match_stats
	WHERE member_id = $1
	`

	var playerMatchStats PlayerMatchStats

	err := r.DB.GetContext(ctx, &playerMatchStats, query, memberID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		slog.Info("Errored when attempting to getting player stats", "memberID", memberID, "error", err)
		return nil, err
	}

	return &playerMatchStats, nil
}

func (r *repository) CreatePlayerRankingStats(ctx context.Context, stats *PlayerRankingStats) error {
	stats.ID = uuid.New()

	query := `
		INSERT INTO player_ranking_stats (
			id, member_id, username, wins, top_threes,
			rating, rank_position
		) VALUES (
			:id, :member_id, :username, :wins, :top_threes,
			:rating, :rank_position
		)`

	_, err := r.DB.NamedExecContext(ctx, query, stats)
	if err != nil {
		return fmt.Errorf("failed to create player ranking stats: %w", err)
	}

	return nil
}

// CreateMatchHistory creates a new match history record
func (r *repository) CreateMatchHistory(ctx context.Context, history *MatchHistory) error {
	history.ID = uuid.New()

	query := `
		INSERT INTO match_history (
			id, session_id, member_id, win, final_position,
			kills, deaths, rating_before, rating_after,
			rating_change, match_started_at
		) VALUES (
			:id, :session_id, :member_id, :win, :final_position,
			:kills, :deaths, :rating_before, :rating_after,
			:rating_change, :match_started_at
		)`

	_, err := r.DB.NamedExecContext(ctx, query, history)
	if err != nil {
		return fmt.Errorf("failed to create match history: %w", err)
	}

	return nil
}
