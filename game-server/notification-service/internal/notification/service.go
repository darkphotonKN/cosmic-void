package notification

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
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

func (s *Service) ProcessItemCreated(ctx context.Context, payload *pb.ItemCreatedEvent) error {
	templateData := map[string]any{
		"UserId":   payload.UserId,
		"ItemName": payload.Name,
		"ItemType": payload.ItemType,
	}

	// For item.created, we notify admins (not using template since it's admin-specific)
	title := "新物品模板已建立"
	message := fmt.Sprintf("管理員已建立新物品模板：%s (類型：%s)", payload.Name, payload.ItemType)

	createNotification := &CreateNotification{
		UserID:           uuid.MustParse(payload.UserId), // Special ID for admin broadcast
		NotificationType: "in_app",
		EventType:        "item.created",
		Title:            title,
		Message:          message,
		Data:             templateData,
	}

	_, err := s.repo.Create(ctx, createNotification)
	if err != nil {
		slog.Warn("Error occurred when creating admin notification",
			"event_type", "item.created",
			"err", err,
		)

	}
	return err
}

func (s *Service) GetNotification(ctx context.Context, payload QueryNotifications) (*NotificationListResponse, error) {
	notificationList, err := s.repo.GetByUserID(ctx, &payload)
	if err != nil {
		slog.Error("Get notification list failed", "user_id", payload.UserID)
		return nil, err
	}

	sendNotification := &NotificationListResponse{
		Notifications: notificationList,
		Total:         len(notificationList),
	}

	return sendNotification, nil
}
