package stats

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

// CreatePlayerMatchStats creates a new player match stats record
func (r *Repository) CreatePlayerMatchStats(ctx context.Context, stats *PlayerMatchStats) error {
	stats.ID = uuid.New()

	query := `
		INSERT INTO player_match_stats (
			id, member_id, games_played, wins, losses,
			kills, deaths, times_placed_top_three
		) VALUES (
			:id, :member_id, :games_played, :wins, :losses,
			:kills, :deaths, :times_placed_top_three
		)`

	_, err := r.DB.NamedExecContext(ctx, query, stats)
	if err != nil {
		return fmt.Errorf("failed to create player match stats: %w", err)
	}

	return nil
}

// CreatePlayerRankingStats creates a new player ranking stats record
func (r *Repository) CreatePlayerRankingStats(ctx context.Context, stats *PlayerRankingStats) error {
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
func (r *Repository) CreateMatchHistory(ctx context.Context, history *MatchHistory) error {
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