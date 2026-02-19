package stats

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	commonutils "github.com/darkphotonKN/cosmic-void-server/common/utils"
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

	if rows.Next() {
		err = rows.StructScan(&updated)

		if err != nil {
			return nil, commonutils.AnalyzeDBErr(err)
		}
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

func (r *repository) UpsertPlayerRankingStats(ctx context.Context, stats *UpdatePlayerRankingsParams) (*PlayerRankingStats, error) {

	query := `
		INSERT INTO player_ranking_stats (
			member_id, username, wins, top_threes, avatar_url,
			rating, rank_position
		) VALUES (
			:member_id, 
			:username,
			:wins,
			:top_threes,
			:avatar_url,
			:rating,
			:rank_position
		)
		ON CONFLICT (member_id)
		DO UPDATE SET
				username = EXCLUDED.username,
				wins = EXCLUDED.wins,
				top_threes = EXCLUDED.top_threes,
				avatar_url = EXCLUDED.avatar_url,
				rating = EXCLUDED.rating,
				rank_position = EXCLUDED.rank_position,
				last_calculated_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		RETURNING *;
	`

	rows, err := r.DB.NamedQueryContext(ctx, query, stats)
	if err != nil {
		slog.Error("Could not run upsert to PlayerRankings successfully.",
			"member_id", stats.MemberID,
			"error", err,
		)
		return nil, commonutils.AnalyzeDBErr(err)
	}

	defer rows.Close()

	var updated PlayerRankingStats

	if rows.Next() {
		err = rows.StructScan(&updated)

		if err != nil {
			slog.Error("Could not scan to row.",
				"member_id", stats.MemberID,
				"error", err,
			)
			return nil, commonutils.AnalyzeDBErr(err)
		}

	}

	return &updated, nil
}

func (r *repository) GetPlayerRankings(ctx context.Context, params *GetPlayerRankings) ([]*PlayerRankingStats, error) {
	query := `
		SELECT
			id,
			member_id,
			username,
			wins,
			top_threes,
			avatar_url,
			rating,
			rank_position,
			last_calculated_at,
			created_at,
			updated_at
		FROM player_ranking_stats
		ORDER BY rating DESC, wins DESC, top_threes DESC
		LIMIT $1 OFFSET $2
	`

	var playerRankings []*PlayerRankingStats

	err := r.DB.SelectContext(ctx, &playerRankings, query, params.limit, params.offset)
	if err != nil {
		slog.Error("failed to get player rankings", "limit", params.limit, "offset", params.offset, "error", err)
		return nil, commonutils.AnalyzeDBErr(err)
	}

	slog.Info("successfully retrieved player rankings", "count", len(playerRankings), "limit", params.limit, "offset", params.offset)
	return playerRankings, nil
}

func (r *repository) GetPlayerRankingStats(ctx context.Context, memberID uuid.UUID) (*PlayerRankingStats, error) {
	query := `
		SELECT
			id,
			member_id,
			username,
			wins,
			top_threes,
			rating,
			rank_position,
			last_calculated_at,
			created_at,
			updated_at
		FROM player_ranking_stats
		WHERE member_id = $1
	`

	var playerRankingStats PlayerRankingStats

	err := r.DB.GetContext(ctx, &playerRankingStats, query, memberID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		slog.Error("Error getting player ranking stats", "memberID", memberID, "error", err)
		return nil, err
	}

	return &playerRankingStats, nil
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

// Transaction-aware repository methods

// UpsertPlayerMatchStatsTx updates player match stats within a transaction
func (r *repository) UpsertPlayerMatchStatsTx(ctx context.Context, tx *sqlx.Tx, params *UpdateStatsParams) (*PlayerMatchStats, error) {
	query := `
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
			$1, $2, $3, $4, $5, $6, $7,
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
	`

	var updated PlayerMatchStats
	err := tx.QueryRowContext(ctx, query,
		params.MemberID,
		params.GamesPlayed,
		params.Wins,
		params.Losses,
		params.Kills,
		params.Deaths,
		params.TimesPlacedTopThree,
	).Scan(
		&updated.ID,
		&updated.MemberID,
		&updated.GamesPlayed,
		&updated.Wins,
		&updated.Losses,
		&updated.Kills,
		&updated.Deaths,
		&updated.TimesPlacedTopThree,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err != nil {
		slog.Info("Errored when attempting to update player stats", "err", err)
		return nil, err
	}

	return &updated, nil
}

// GetPlayerMatchStatsTx retrieves player match stats within a transaction
func (r *repository) GetPlayerMatchStatsTx(ctx context.Context, tx *sqlx.Tx, memberID uuid.UUID) (*PlayerMatchStats, error) {
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

	err := tx.GetContext(ctx, &playerMatchStats, query, memberID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		slog.Info("Errored when attempting to getting player stats", "memberID", memberID, "error", err)
		return nil, err
	}

	return &playerMatchStats, nil
}

// UpsertPlayerRankingStatsTx updates player ranking stats within a transaction
func (r *repository) UpsertPlayerRankingStatsTx(ctx context.Context, tx *sqlx.Tx, stats *UpdatePlayerRankingsParams) (*PlayerRankingStats, error) {
	query := `
		INSERT INTO player_ranking_stats (
			member_id, username, wins, top_threes, avatar_url,
			rating, rank_position
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		ON CONFLICT (member_id)
		DO UPDATE SET
				username = EXCLUDED.username,
				wins = EXCLUDED.wins,
				top_threes = EXCLUDED.top_threes,
				avatar_url = EXCLUDED.avatar_url,
				rating = EXCLUDED.rating,
				rank_position = EXCLUDED.rank_position,
				last_calculated_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		RETURNING *;
	`

	var updated PlayerRankingStats
	err := tx.QueryRowContext(ctx, query,
		stats.MemberID,
		stats.Username,
		stats.Wins,
		stats.TopThrees,
		stats.AvatarUrl,
		stats.Rating,
		stats.RankPosition,
	).Scan(
		&updated.ID,
		&updated.MemberID,
		&updated.Username,
		&updated.Wins,
		&updated.TopThrees,
		&updated.AvatarUrl,
		&updated.Rating,
		&updated.RankPosition,
		&updated.LastCalculatedAt,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err != nil {
		slog.Error("Could not run upsert to PlayerRankings successfully.",
			"member_id", stats.MemberID,
			"error", err,
		)
		return nil, commonutils.AnalyzeDBErr(err)
	}

	return &updated, nil
}

// GetPlayerRankingStatsTx retrieves player ranking stats within a transaction
func (r *repository) GetPlayerRankingStatsTx(ctx context.Context, tx *sqlx.Tx, memberID uuid.UUID) (*PlayerRankingStats, error) {
	query := `
		SELECT
			id,
			member_id,
			username,
			wins,
			top_threes,
			rating,
			rank_position,
			last_calculated_at,
			created_at,
			updated_at
		FROM player_ranking_stats
		WHERE member_id = $1
	`

	var playerRankingStats PlayerRankingStats

	err := tx.GetContext(ctx, &playerRankingStats, query, memberID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		slog.Error("Error getting player ranking stats", "memberID", memberID, "error", err)
		return nil, err
	}

	return &playerRankingStats, nil
}

// CreateMatchHistoryTx creates a new match history record within a transaction
func (r *repository) CreateMatchHistoryTx(ctx context.Context, tx *sqlx.Tx, history *MatchHistory) error {
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

	_, err := tx.NamedExecContext(ctx, query, history)
	if err != nil {
		return fmt.Errorf("failed to create match history: %w", err)
	}

	return nil
}
