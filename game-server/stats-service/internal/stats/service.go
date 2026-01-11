package stats

import (
	"context"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Repository interface defines what the service needs from the repository
type RepositoryInterface interface {
	CreatePlayerMatchStats(ctx context.Context, stats *PlayerMatchStats) error
	CreatePlayerRankingStats(ctx context.Context, stats *PlayerRankingStats) error
	CreateMatchHistory(ctx context.Context, history *MatchHistory) error
}

type service struct {
	repo      RepositoryInterface
	publishCh *amqp.Channel
}

func NewService(repo RepositoryInterface, publishCh *amqp.Channel) *service {
	return &service{
		repo:      repo,
		publishCh: publishCh,
	}
}

// ProcessMatchCompleted processes entire match result data
func (s *service) ProcessMatchCompleted(ctx context.Context, req *pb.ProcessMatchCompletedRequest) (*pb.ProcessMatchCompletedResponse, error) {
	slog.Info("Processing match completed",
		"session_id", req.SessionId,
		"total_players", req.TotalPlayers,
		"match_started_at", req.MatchStartedAt.AsTime(),
		"match_ended_at", req.MatchEndedAt.AsTime(),
	)

	// Log all player outcomes
	for i, player := range req.Players {
		slog.Info("Player match outcome",
			"player_index", i+1,
			"member_id", player.MemberId,
			"username", player.Username,
			"win", player.Win,
			"final_position", player.FinalPosition,
			"kills", player.Kills,
			"deaths", player.Deaths,
		)
	}

	// TODO: Implement the actual processing logic here
	// - Create/update player match stats (aggregated stats)
	// - Create/update player ranking stats (for leaderboard)
	// - Create individual match history records
	// - Calculate rating changes
	// - Update rankings

	return &pb.ProcessMatchCompletedResponse{
		Success:          true,
		Message:          "Match data processed successfully",
		PlayersProcessed: req.TotalPlayers,
	}, nil
}
