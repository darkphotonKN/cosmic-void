package stats

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ServiceInterface defines what the handler needs from the service
type ServiceInterface interface {
	CreatePlayerMatchStats(ctx context.Context, req *pb.CreatePlayerMatchStatsRequest) (*pb.PlayerMatchStats, error)
	CreatePlayerRankingStats(ctx context.Context, req *pb.CreatePlayerRankingStatsRequest) (*pb.PlayerRankingStats, error)
	CreateMatchHistory(ctx context.Context, req *pb.CreateMatchHistoryRequest) (*pb.MatchHistory, error)
}

type Handler struct {
	pb.UnimplementedStatsServiceServer
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{
		service: service,
	}
}

// CreatePlayerMatchStats handles the gRPC request to create player match stats
func (h *Handler) CreatePlayerMatchStats(ctx context.Context, req *pb.CreatePlayerMatchStatsRequest) (*pb.PlayerMatchStats, error) {
	if req.MemberId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "member ID is required")
	}

	stats, err := h.service.CreatePlayerMatchStats(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create player match stats: %v", err)
	}

	return stats, nil
}

// CreatePlayerRankingStats handles the gRPC request to create player ranking stats
func (h *Handler) CreatePlayerRankingStats(ctx context.Context, req *pb.CreatePlayerRankingStatsRequest) (*pb.PlayerRankingStats, error) {
	if req.MemberId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "member ID is required")
	}

	if req.Username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username is required")
	}

	stats, err := h.service.CreatePlayerRankingStats(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create player ranking stats: %v", err)
	}

	return stats, nil
}

// CreateMatchHistory handles the gRPC request to create match history
func (h *Handler) CreateMatchHistory(ctx context.Context, req *pb.CreateMatchHistoryRequest) (*pb.MatchHistory, error) {
	if req.SessionId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "session ID is required")
	}

	if req.MemberId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "member ID is required")
	}

	if req.MatchStartedAt == nil {
		return nil, status.Errorf(codes.InvalidArgument, "match started at is required")
	}

	history, err := h.service.CreateMatchHistory(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create match history: %v", err)
	}

	return history, nil
}

