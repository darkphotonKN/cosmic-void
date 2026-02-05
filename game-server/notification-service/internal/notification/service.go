package notification

import (
	"context"
	"log/slog"

	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, createNotification *CreateNotification) (*Notification, error)
	GetByUserID(ctx context.Context, request *QueryNotifications) ([]Notification, error)
	Update(ctx context.Context, request *UpdateNotification) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) ProcessMemberSignedUp(ctx context.Context, payload *commonconstants.MemberSignedUpEventPayload) error {
	id, err := uuid.Parse(payload.UserID)
	if err != nil {
		slog.Warn("invalid member UUID format",
			"user_id", payload.UserID,
			"err", err,
		)
		return err
	}
	templateData := map[string]any{
		"Name":  payload.Name,
		"Email": payload.Email,
	}
	title, message, notificationType, err := RenderTemplate("member.signedup", templateData)

	if err != nil {
		slog.Error("Failed to render template", "error", err)
		return err
	}

	createNotification := &CreateNotification{
		UserID:           id,
		NotificationType: notificationType,
		EventType:        "member.signedup",
		Title:            title,
		Message:          message,
		Data:             templateData,
	}

	_, err = s.repo.Create(ctx, createNotification)
	if err != nil {
		slog.Warn("Error occurred when creating new notification",
			"user_id", payload.UserID,
			"err", err,
		)
	}
	return err
}
