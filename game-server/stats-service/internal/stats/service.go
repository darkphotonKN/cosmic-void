package stats

import (
	"context"
	"fmt"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/types/known/timestamppb"
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

// CreatePlayerMatchStats creates a new player match stats record
func (s *service) CreatePlayerMatchStats(ctx context.Context, req *pb.CreatePlayerMatchStatsRequest) (*pb.PlayerMatchStats, error) {
	memberID, err := uuid.Parse(req.MemberId)
	if err != nil {
		return nil, fmt.Errorf("invalid member ID: %w", err)
	}

	stats := &PlayerMatchStats{
		MemberID:            memberID,
		GamesPlayed:         req.GamesPlayed,
		Wins:                req.Wins,
		Losses:              req.Losses,
		Kills:               req.Kills,
		Deaths:              req.Deaths,
		TimesPlacedTopThree: req.TimesPlacedTopThree,
	}

	if err := s.repo.CreatePlayerMatchStats(ctx, stats); err != nil {
		return nil, err
	}

	return &pb.PlayerMatchStats{
		Id:                  stats.ID.String(),
		MemberId:            stats.MemberID.String(),
		GamesPlayed:         stats.GamesPlayed,
		Wins:                stats.Wins,
		Losses:              stats.Losses,
		Kills:               stats.Kills,
		Deaths:              stats.Deaths,
		TimesPlacedTopThree: stats.TimesPlacedTopThree,
		CreatedAt:           timestamppb.New(stats.CreatedAt),
		UpdatedAt:           timestamppb.New(stats.UpdatedAt),
	}, nil
}

// CreatePlayerRankingStats creates a new player ranking stats record
func (s *service) CreatePlayerRankingStats(ctx context.Context, req *pb.CreatePlayerRankingStatsRequest) (*pb.PlayerRankingStats, error) {
	memberID, err := uuid.Parse(req.MemberId)
	if err != nil {
		return nil, fmt.Errorf("invalid member ID: %w", err)
	}

	stats := &PlayerRankingStats{
		MemberID:     memberID,
		Username:     req.Username,
		Wins:         req.Wins,
		TopThrees:    req.TopThrees,
		Rating:       req.Rating,
		RankPosition: req.RankPosition,
	}

	if err := s.repo.CreatePlayerRankingStats(ctx, stats); err != nil {
		return nil, err
	}

	response := &pb.PlayerRankingStats{
		Id:               stats.ID.String(),
		MemberId:         stats.MemberID.String(),
		Username:         stats.Username,
		Wins:             stats.Wins,
		TopThrees:        stats.TopThrees,
		Rating:           stats.Rating,
		LastCalculatedAt: timestamppb.New(stats.LastCalculatedAt),
		CreatedAt:        timestamppb.New(stats.CreatedAt),
		UpdatedAt:        timestamppb.New(stats.UpdatedAt),
	}

	if stats.RankPosition != nil {
		response.RankPosition = stats.RankPosition
	}

	return response, nil
}

// CreateMatchHistory creates a new match history record
func (s *service) CreateMatchHistory(ctx context.Context, req *pb.CreateMatchHistoryRequest) (*pb.MatchHistory, error) {
	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	memberID, err := uuid.Parse(req.MemberId)
	if err != nil {
		return nil, fmt.Errorf("invalid member ID: %w", err)
	}

	history := &MatchHistory{
		SessionID:      sessionID,
		MemberID:       memberID,
		Win:            req.Win,
		FinalPosition:  req.FinalPosition,
		Kills:          req.Kills,
		Deaths:         req.Deaths,
		RatingBefore:   req.RatingBefore,
		RatingAfter:    req.RatingAfter,
		RatingChange:   req.RatingChange,
		MatchStartedAt: req.MatchStartedAt.AsTime(),
	}

	if err := s.repo.CreateMatchHistory(ctx, history); err != nil {
		return nil, err
	}

	response := &pb.MatchHistory{
		Id:             history.ID.String(),
		SessionId:      history.SessionID.String(),
		MemberId:       history.MemberID.String(),
		Win:            history.Win,
		FinalPosition:  history.FinalPosition,
		Kills:          history.Kills,
		Deaths:         history.Deaths,
		MatchStartedAt: timestamppb.New(history.MatchStartedAt),
		CreatedAt:      timestamppb.New(history.CreatedAt),
	}

	if history.RatingBefore != nil {
		response.RatingBefore = history.RatingBefore
	}
	if history.RatingAfter != nil {
		response.RatingAfter = history.RatingAfter
	}
	if history.RatingChange != nil {
		response.RatingChange = history.RatingChange
	}

	return response, nil
}
