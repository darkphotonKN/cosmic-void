package stats

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	GetLeaderboard(ctx context.Context, req *pb.GetLeaderboardRequest) (*pb.GetLeaderboardResponse, error)
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

func (h *Handler) GetLeaderboard(ctx context.Context, req *pb.GetLeaderboardRequest) (*pb.GetLeaderboardResponse, error) {
	return h.service.GetLeaderboard(ctx, req)
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
