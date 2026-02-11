package stats

import (
	"context"
	"log/slog"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonutils "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository interface defines what the service needs from the repository
type Repository interface {
	UpsertPlayerMatchStats(ctx context.Context, params *UpdateStatsParams) (*PlayerMatchStats, error)
	UpsertPlayerRankingStats(ctx context.Context, params *UpdatePlayerRankingsParams) (*PlayerRankingStats, error)
	CreateMatchHistory(ctx context.Context, history *MatchHistory) error
	GetPlayerMatchStats(ctx context.Context, memberID uuid.UUID) (*PlayerMatchStats, error)
	GetPlayerRankingStats(ctx context.Context, memberID uuid.UUID) (*PlayerRankingStats, error)
}

type service struct {
	repo      Repository
	publishCh commonbroker.Publisher
	db        *sqlx.DB
}

func NewService(repo Repository, publishCh commonbroker.Publisher, db *sqlx.DB) *service {
	return &service{
		repo:      repo,
		publishCh: publishCh,
		db:        db,
	}
}

/**
* Runs all the relevant processes after a match is completed, updating the
* relavant tables.
**/
func (s *service) ProcessMatchCompleted(ctx context.Context, req *pb.MatchEndedEvent) (*ProcessMatchCompletedResponse, error) {
	slog.Info("Processing match completed",
		"session_id", req.SessionId,
		"match_started_at", req.MatchStartedAt.AsTime(),
		"match_ended_at", req.MatchEndedAt.AsTime(),
	)

	for _, playerResults := range req.Players {
		slog.Info("Player match outcome",
			"playerResults", playerResults,
		)

		// wrap transaction for business critical sync up between players match
		// history and their personal stats
		err := commonutils.ExecTx(ctx, s.db, func(tx *sqlx.Tx) error {
			// -- player stats --
			err := s.updatePlayerStats(ctx, playerResults)
			if err != nil {
				slog.Error("error when updating match stats", "memberID", playerResults.MemberId, "error", err)
				return err
			}

			// -- leaderboards --
			err = s.updateDenormalizedLeaderboard(ctx, playerResults)
			if err != nil {
				slog.Error("error when updating leaderboard stats", "memberID", playerResults.MemberId, "error", err)
				return err
			}

			// -- match history --
			sessionId, err := uuid.Parse(req.SessionId)
			if err != nil {
				slog.Error("error parsing session id", "sessionId", req.SessionId, "error", err)
				return err
			}

			memberId, err := uuid.Parse(playerResults.MemberId)
			if err != nil {
				slog.Error("error parsing member id", "memberId", playerResults.MemberId, "error", err)
				return err
			}

			matchHistory := &MatchHistory{
				SessionID:      sessionId,
				MemberID:       memberId,
				Win:            playerResults.Win,
				FinalPosition:  int(playerResults.FinalPosition),
				Kills:          int(playerResults.Kills),
				Deaths:         int(playerResults.Deaths),
				MatchStartedAt: req.MatchStartedAt.AsTime(),
			}

			err = s.repo.CreateMatchHistory(ctx, matchHistory)
			if err != nil {
				slog.Error("error creating match history", "memberID", playerResults.MemberId, "error", err)
				return err
			}

			return nil
		})

		// transaction errored, continue to next player
		// (roll back automagically handled in helper already)
		if err != nil {
			slog.Error("error occured during match processing transaction, rolledback.",
				"error", err)
			continue
		}
	}

	// TODO: call auth service for player information

	return &ProcessMatchCompletedResponse{
		Success: true,
		Message: "Match data processed successfully",
	}, nil
}

func (s *service) updatePlayerStats(ctx context.Context, player *pb.PlayerMatchResult) error {
	memberId, err := uuid.Parse(player.MemberId)
	if err != nil {
		slog.Info("Errored when attempting to get parse member id into UUID", "err", err)
		return err
	}

	// TODO: recalculate averages, averages migration and update WIP
	matchHistoryData, err := s.getMatchHistory(ctx, memberId)

	if err != nil {
		slog.Info("Errored when attempting to get match history", "err", err)
		return err
	}

	slog.Info("Match history data for player", "player", player, "data", matchHistoryData)

	playerStats, err := s.repo.GetPlayerMatchStats(ctx, memberId)

	if err != nil {
		return err
	}

	if playerStats == nil {
		// initialize struct, players first time
		playerStats = &PlayerMatchStats{
			MemberID:            memberId,
			GamesPlayed:         0,
			Wins:                0,
			Losses:              0,
			Kills:               0,
			Deaths:              0,
			TimesPlacedTopThree: 0,
		}
	}

	playerStats.GamesPlayed += 1
	playerStats.Kills += int(player.Kills)
	playerStats.Deaths += int(player.Deaths)

	// TODO: check if player won
	if player.Win {
		playerStats.Wins += 1
	} else if player.FinalPosition == 5 {
		playerStats.Losses += 1
	}

	// update aggregate stats
	_, err = s.repo.UpsertPlayerMatchStats(ctx, &UpdateStatsParams{
		MemberID:            playerStats.MemberID,
		GamesPlayed:         playerStats.GamesPlayed,
		Wins:                playerStats.Wins,
		Losses:              playerStats.Losses,
		Kills:               playerStats.Kills,
		Deaths:              playerStats.Deaths,
		TimesPlacedTopThree: playerStats.TimesPlacedTopThree,
	})

	if err != nil {
		return err
	}
	return nil
}

func (s *service) getMatchHistory(ctx context.Context, memberID uuid.UUID) ([]*MatchHistory, error) {

	return nil, nil
}

func (s *service) calculateMatchAverage(ctx context.Context, matchHistory []*MatchHistory) (*PlayerMatchStats, error) {
	// TODO: calculate the new averages
	return nil, nil
}

/**
* Handles setting up and updating the denormalized leaderboard stats.
**/
func (s *service) updateDenormalizedLeaderboard(ctx context.Context, results *pb.PlayerMatchResult) error {
	memberId, err := uuid.Parse(results.MemberId)
	if err != nil {
		slog.Info("Errored when attempting to get parse member id into UUID", "err", err)
		return err
	}

	stats, err := s.repo.GetPlayerMatchStats(ctx, memberId)

	if err != nil {
		return err
	}

	wins := stats.Wins
	topThree := stats.TimesPlacedTopThree
	if results.Win {
		wins += 1
	}
	if results.FinalPosition >= 3 {
		topThree += 1
	}

	// recalculate rank position
	rankingStats, err := s.repo.GetPlayerRankingStats(ctx, memberId)

	slog.Debug("getting ranking stats from GetPlayerRankingStats", "rankingStats", rankingStats)

	var ranking *int
	if rankingStats != nil {
		ranking = rankingStats.RankPosition
	}

	statsParam := &UpdatePlayerRankingsParams{
		MemberID:         memberId,
		Username:         results.Username,
		Wins:             wins,
		TopThrees:        topThree,
		Rating:           0,
		RankPosition:     ranking,
		LastCalculatedAt: time.Now(),
	}

	playerRankingStats, err := s.repo.UpsertPlayerRankingStats(ctx, statsParam)

	if err != nil {
		return err
	}

	slog.Debug("after upsert for player ranking stats", "playerRankingStats", playerRankingStats)

	return nil
}

/**
* Updates the player rankings leaderboard.
**/
func (s *service) UpdatePlayerRankings(ctx context.Context, updateData *pb.MemberProfileUpdatedEvent) error {
	memberIdUUID, err := uuid.Parse(updateData.MemberId)
	if err != nil {
		slog.Error("error when parsing updateData.MemberId as uuid", "memberID", updateData.MemberId, "error", err)
	}
	_, err = s.repo.UpsertPlayerRankingStats(ctx, &UpdatePlayerRankingsParams{
		MemberID:  memberIdUUID,
		AvatarUrl: updateData.AvatarUrl,
	})

	if err != nil {
		slog.Error("error when updating player rankings", "memberID", updateData.MemberId, "error", err)
		return err
	}

	return nil
}
