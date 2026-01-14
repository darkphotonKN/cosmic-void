package stats

import (
	"context"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
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
	publishCh *amqp.Channel
}

func NewService(repo Repository, publishCh *amqp.Channel) *service {
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

		s.updatePlayerStats(ctx, player)
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

func (s *service) updatePlayerStats(ctx context.Context, player *pb.PlayerMatchResults) ([]*MatchHistory, error) {

	memberId, err := uuid.Parse(player.MemberId)
	if err != nil {
		slog.Info("Errored when attempting to get parse member id into UUID", "err", err)
		return nil, err
	}

	// TODO: recalculate averages, averages WIP
	matchHistoryData, err := s.getMatchHistory(ctx, memberId)

	if err != nil {
		slog.Info("Errored when attempting to get match history", "err", err)
		return nil, err
	}

	slog.Info("Match history data for player", "player", player, "data", matchHistoryData)

	// matchHistoryUpdated := append(matchHistoryData, &MatchHistory{})

	// newPlayerAverages := s.calculateMatchAverage(ctx, matchHistoryUpdated)

	// get current stats
	playerStats, err := s.repo.GetPlayerMatchStats(ctx, memberId)

	if err != nil {
		return nil, err
	}

	playerStats.GamesPlayed += 1
	playerStats.Kills += int(player.Kills)
	playerStats.Deaths += int(player.Deaths)
	// TODO: check if player won
	playerStats.Wins += 0
	playerStats.Losses += 1

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
		return nil, err
	}
	return nil, nil
}

func (s *service) getMatchHistory(ctx context.Context, memberID uuid.UUID) ([]*MatchHistory, error) {

	return nil, nil
}

func (s *service) calculateMatchAverage(ctx context.Context, matchHistory []*MatchHistory) (*PlayerMatchStats, error) {
	// TODO: calculate the new averages
	return nil, nil
}
