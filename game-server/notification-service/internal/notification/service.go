package notification

import (
	"context"
	"log/slog"

	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Repository interface {
	Create(ctx context.Context, createNotification *CreateNotification) (*Notification, error)
	GetByUserID(ctx context.Context, request *QueryNotifications) ([]Notification, error)
	Update(ctx context.Context, request *UpdateNotification) error
}

type Service struct {
	repo      Repository
	publishCh commonbroker.Publisher
}

func NewService(repo Repository, publishCh commonbroker.Publisher) *Service {
	return &Service{
		repo:      repo,
		publishCh: publishCh,
	}
}

func (s *Service) Create(ctx context.Context, notification *UserCreateNotification) (*Notification, error) {
	if notification.Title == "" {
		slog.Warn("notification creation failed: missing title",
			"user_id", notification.UserID,
			"type", notification.EventType,
		)
		return nil, status.Errorf(codes.InvalidArgument, "Name field is required")
	}

	id, err := uuid.Parse(notification.UserID)
	if err != nil {
		slog.Warn("invalid member UUID format",
			"user_id", notification.UserID,
			"err", err,
		)
		return nil, err
	}

	createNotification := &CreateNotification{
		UserID:           id,
		NotificationType: notification.NotificationType,
		EventType:        notification.EventType,
		Title:            notification.Title,
		Message:          notification.Message,
		Data:             notification.Data,
	}

	newNotification, err := s.repo.Create(ctx, createNotification)
	if err != nil {
		slog.Warn("Error occurred when creating new notification",
			"user_id", notification.UserID,
			"err", err,
		)
		return nil, err
	}

	return newNotification, nil
}
