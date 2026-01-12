package stats

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service defines what the handler needs from the service
type Service interface {
	ProcessMatchCompleted(ctx context.Context, req *pb.ProcessMatchCompletedRequest) (*pb.ProcessMatchCompletedResponse, error)
}

type Handler struct {
	pb.UnimplementedStatsServiceServer
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ProcessMatchCompleted handles the gRPC request to process entire match results
func (h *Handler) ProcessMatchCompleted(ctx context.Context, req *pb.ProcessMatchCompletedRequest) (*pb.ProcessMatchCompletedResponse, error) {
	// Validate request
	if req.SessionId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "session ID is required")
	}

	if req.MatchStartedAt == nil {
		return nil, status.Errorf(codes.InvalidArgument, "match started at is required")
	}

	if req.MatchEndedAt == nil {
		return nil, status.Errorf(codes.InvalidArgument, "match ended at is required")
	}

	if len(req.Players) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "at least one player is required")
	}

	if req.TotalPlayers != int32(len(req.Players)) {
		return nil, status.Errorf(codes.InvalidArgument, "total players count does not match players array length")
	}

	// Validate each player
	for i, player := range req.Players {
		if player.MemberId == "" {
			return nil, status.Errorf(codes.InvalidArgument, "player %d member ID is required", i+1)
		}
		if player.Username == "" {
			return nil, status.Errorf(codes.InvalidArgument, "player %d username is required", i+1)
		}
		if player.FinalPosition < 1 {
			return nil, status.Errorf(codes.InvalidArgument, "player %d final position must be >= 1", i+1)
		}
		if player.Kills < 0 {
			return nil, status.Errorf(codes.InvalidArgument, "player %d kills cannot be negative", i+1)
		}
		if player.Deaths < 0 {
			return nil, status.Errorf(codes.InvalidArgument, "player %d deaths cannot be negative", i+1)
		}
	}

	// Process the match
	response, err := h.service.ProcessMatchCompleted(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to process match completed: %v", err)
	}

	return response, nil
}

// Placeholder implementations for read operations (not implemented yet)
func (h *Handler) GetPlayerMatchStats(ctx context.Context, req *pb.GetPlayerMatchStatsRequest) (*pb.PlayerMatchStats, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

func (h *Handler) GetPlayerRankingStats(ctx context.Context, req *pb.GetPlayerRankingStatsRequest) (*pb.PlayerRankingStats, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

func (h *Handler) GetMatchHistory(ctx context.Context, req *pb.GetMatchHistoryRequest) (*pb.GetMatchHistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

func (h *Handler) GetLeaderboard(ctx context.Context, req *pb.GetLeaderboardRequest) (*pb.GetLeaderboardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}
