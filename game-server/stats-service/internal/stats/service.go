package stats

import (
	"context"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Repository interface defines what the service needs from the repository
type Repository interface {
	CreatePlayerMatchStats(ctx context.Context, stats *PlayerMatchStats) error
	CreatePlayerRankingStats(ctx context.Context, stats *PlayerRankingStats) error
	CreateMatchHistory(ctx context.Context, history *MatchHistory) error
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

// ProcessMatchCompleted processes entire match result data
func (s *service) ProcessMatchCompleted(ctx context.Context, req *pb.ProcessMatchCompletedRequest) (*pb.ProcessMatchCompletedResponse, error) {
	slog.Info("Processing match completed",
		"session_id", req.SessionId,
		"match_started_at", req.MatchStartedAt.AsTime(),
		"match_ended_at", req.MatchEndedAt.AsTime(),
	)

	for _, player := range req.Players {
		slog.Info("Player match outcome",
			"player", player,
		)
	}

	// TODO: implement seperate updates to relevant tables

	return &pb.ProcessMatchCompletedResponse{
		Success: true,
		Message: "Match data processed successfully",
	}, nil
}
