package stats

import (
	"context"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	"github.com/google/uuid"
)

// Repository interface defines what the service needs from the repository
type Repository interface {
	UpsertPlayerMatchStats(ctx context.Context, params *UpdateStatsParams) (*PlayerMatchStats, error)
	CreatePlayerRankingStats(ctx context.Context, stats *PlayerRankingStats) error
	CreateMatchHistory(ctx context.Context, history *MatchHistory) error
	GetPlayerMatchStats(ctx context.Context, memberID uuid.UUID) (*PlayerMatchStats, error)
}

type service struct {
	repo      Repository
	publishCh commonbroker.Publisher
}

func NewService(repo Repository, publishCh commonbroker.Publisher) *service {
	return &service{
		repo:      repo,
		publishCh: publishCh,
	}
}

/**
* Runs all the relevant processes after a match is completed, updating the
* relavant tables.
**/
func (s *service) ProcessMatchCompleted(ctx context.Context, req *pb.ProcessMatchCompletedRequest) (*pb.ProcessMatchCompletedResponse, error) {
	slog.Info("Processing match completed",
		"session_id", req.SessionId,
		"match_started_at", req.MatchStartedAt.AsTime(),
		"match_ended_at", req.MatchEndedAt.AsTime(),
	)

	// TODO: implement seperate updates to relevant tables

	for _, player := range req.Players {
		slog.Info("Player match outcome",
			"player", player,
		)

		err := s.updatePlayerStats(ctx, player)
		if err != nil {
			slog.Error("error when updating match stats", "memberID", player.MemberId, "error", err)
		}
	}

	// TODO: call auth service for player information

	// TODO: update denormazlied ranking table
	// - add update param struct for repo updates
	// s.repo.CreatePlayerRankingStats(ctx, )

	return &pb.ProcessMatchCompletedResponse{
		Success: true,
		Message: "Match data processed successfully",
	}, nil
}

func (s *service) updatePlayerStats(ctx context.Context, player *pb.PlayerMatchResults) error {
	memberId, err := uuid.Parse(player.MemberId)
	if err != nil {
		slog.Info("Errored when attempting to get parse member id into UUID", "err", err)
		return err
	}

	// TODO: recalculate averages, averages WIP
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
