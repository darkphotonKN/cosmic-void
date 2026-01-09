package stats

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service interface {
	GetPlayerStats(ctx context.Context, playerID uuid.UUID) (*PlayerStats, error)
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

func (h *Handler) GetPlayerStats(ctx context.Context, req *pb.GetPlayerStatsRequest) (*pb.PlayerStats, error) {
	playerID, err := uuid.Parse(req.PlayerId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid player ID: %v", err)
	}

	stats, err := h.service.GetPlayerStats(ctx, playerID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get player stats: %v", err)
	}

	return &pb.PlayerStats{
		Id:              stats.ID.String(),
		PlayerId:        stats.PlayerID.String(),
		Level:           stats.Level,
		Xp:              stats.XP,
		TotalMatches:    stats.TotalMatches,
		Wins:            stats.Wins,
		Losses:          stats.Losses,
		Kills:           stats.Kills,
		Deaths:          stats.Deaths,
		Assists:         stats.Assists,
		KdRatio:         stats.KDRatio,
		WinRate:         stats.WinRate,
		ItemsCollected:  stats.ItemsCollected,
		DamageDealt:     stats.DamageDealt,
		DamageTaken:     stats.DamageTaken,
		PlayTimeSeconds: stats.PlayTimeSeconds,
		CurrentStreak:   stats.CurrentStreak,
		BestStreak:      stats.BestStreak,
		CreatedAt:       timestamppb.New(stats.CreatedAt),
		UpdatedAt:       timestamppb.New(stats.UpdatedAt),
	}, nil
}

func (h *Handler) UpdatePlayerStats(ctx context.Context, req *pb.UpdatePlayerStatsRequest) (*pb.PlayerStats, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (h *Handler) GetLeaderboard(ctx context.Context, req *pb.GetLeaderboardRequest) (*pb.LeaderboardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (h *Handler) GetMatchHistory(ctx context.Context, req *pb.GetMatchHistoryRequest) (*pb.MatchHistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (h *Handler) RecordMatch(ctx context.Context, req *pb.RecordMatchRequest) (*pb.MatchResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

