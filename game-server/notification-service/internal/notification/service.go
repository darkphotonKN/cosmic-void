package notification

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	commonutils "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, createNotification *CreateNotification) (*Notification, error)
	CreateTx(ctx context.Context, tx *sqlx.Tx, createNotification *CreateNotification) (*Notification, error)
	GetByUserID(ctx context.Context, request *QueryNotifications) ([]Notification, error)
	Update(ctx context.Context, request *UpdateNotification) error
	MarkAllAsReadByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

type service struct {
	db        *sqlx.DB
	repo      Repository
	inboxRepo InboxRepository
}

func NewService(db *sqlx.DB, repo Repository, inboxRepo InboxRepository) *service {
	return &service{
		db:        db,
		repo:      repo,
		inboxRepo: inboxRepo,
	}
}

func (s *service) ProcessMemberSignedUp(ctx context.Context, payload *commonconstants.MemberSignedUpEventPayload) error {
	eventID, err := uuid.Parse(payload.EventID)
	if err != nil {
		slog.Warn("invalid event UUID format",
			"event_id", payload.EventID,
			"err", err,
		)
		return err
	}

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
		EventType:        commonconstants.MemberSignedUpEvent,
		Title:            title,
		Message:          message,
		Data:             templateData,
	}

	return s.runInTx(ctx, eventID, commonconstants.MemberSignedUpEvent, func(tx *sqlx.Tx) error {
		_, err := s.repo.CreateTx(ctx, tx, createNotification)
		if err != nil {
			slog.Warn("Error occurred when creating new notification",
				"user_id", payload.UserID,
				"err", err,
			)
		}
		return err
	})
}

// runInTx opens a tx, records the event in the inbox, runs fn, and commits.
// If the event was already processed, it returns ErrAlreadyProcessed without running fn.
func (s *service) runInTx(ctx context.Context, eventID uuid.UUID, eventType string, fn func(tx *sqlx.Tx) error) (err error) {
	return commonutils.ExecTx(ctx, s.db, func(tx *sqlx.Tx) error {
		inserted, err := s.inboxRepo.MarkEventProcessed(ctx, tx, eventID, eventType)
		if err != nil {
			return fmt.Errorf("mark event processed: %w", err)
		}
		if !inserted {
			slog.Info("event already processed, skipping",
				"event_id", eventID,
				"event_type", eventType,
			)
			return commonconstants.ErrAlreadyProcessed
		}
		return fn(tx)
	})
}

func (s *service) ProcessItemCreated(ctx context.Context, payload *pb.ItemCreatedEvent) error {
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

func (s *service) ProcessGameEnded(ctx context.Context, payload *pb.MatchEndedEvent) error {
	// Notify all players about the game result
	for _, player := range payload.Players {
		playerID, err := uuid.Parse(player.MemberId)
		if err != nil {
			slog.Warn("invalid player UUID format",
				"member_id", player.MemberId,
				"err", err,
			)
			continue // Skip this player but continue with others
		}

		// Prepare template data
		templateData := map[string]any{
			"SessionId":     payload.SessionId,
			"Username":      player.Username,
			"Win":           player.Win,
			"FinalPosition": player.FinalPosition,
			"Kills":         player.Kills,
			"Deaths":        player.Deaths,
		}

		// Build notification message
		var title, message string
		if player.Escape {
			title = "Escaped!"
			message = fmt.Sprintf("Congratulations %s! You escaped successfully! Position: #%d, Kills: %d, Deaths: %d",
				player.Username, player.FinalPosition, player.Kills, player.Deaths)
		} else if player.Win {
			title = "Victory!"
			message = fmt.Sprintf("Congratulations %s! Last one standing! Position: #%d, Kills: %d, Deaths: %d",
				player.Username, player.FinalPosition, player.Kills, player.Deaths)
		} else {
			title = "Game Over"
			message = fmt.Sprintf("%s, match ended. Position: #%d, Kills: %d, Deaths: %d",
				player.Username, player.FinalPosition, player.Kills, player.Deaths)
		}

		createNotification := &CreateNotification{
			UserID:           playerID,
			NotificationType: "in_app",
			EventType:        "game.ended",
			Title:            title,
			Message:          message,
			Data:             templateData,
		}

		_, err = s.repo.Create(ctx, createNotification)
		if err != nil {
			slog.Warn("Error occurred when creating game ended notification",
				"user_id", player.MemberId,
				"session_id", payload.SessionId,
				"err", err,
			)
			// Continue with other players even if one fails
		}
	}

	slog.Info("Game ended notifications sent",
		"session_id", payload.SessionId,
		"players_count", len(payload.Players),
	)
	return nil
}

func (s *service) GetNotification(ctx context.Context, payload QueryNotifications) (*NotificationListResponse, error) {
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

// sigle read
func (s *service) MarkNotificationAsRead(ctx context.Context, payload *UpdateNotification) error {
	err := s.repo.Update(ctx, payload)
	if err != nil {
		slog.Error("Failed to mark notification as read",
			"notification_id", payload.ID,
			"user_id", payload.UserID,
			"error", err)
		return err
	}

	slog.Info("Notification marked as read successfully",
		"notification_id", payload.ID,
		"user_id", payload.UserID)
	return nil
}

func (s *service) MarkAllNotificationsAsRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	updatedCount, err := s.repo.MarkAllAsReadByUserID(ctx, userID)
	if err != nil {
		slog.Error("Failed to mark all notifications as read",
			"user_id", userID,
			"error", err)
		return 0, err
	}

	slog.Info("All notifications marked as read",
		"user_id", userID,
		"updated_count", updatedCount)

	return updatedCount, nil
}
