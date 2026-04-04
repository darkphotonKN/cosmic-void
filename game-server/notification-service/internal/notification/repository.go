package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	commonutils "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, createNotification *CreateNotification) (*Notification, error) {
	now := time.Now()
	data, _ := json.Marshal(createNotification.Data)

	notificationModel := &DbNotification{
		ID:               uuid.New(),
		UserID:           createNotification.UserID,
		NotificationType: createNotification.NotificationType,
		EventType:        createNotification.EventType,
		Title:            createNotification.Title,
		Message:          createNotification.Message,
		Data:             data,
		Read:             false,
		Sent:             false,
		SentAt:           nil,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	query := `
	INSERT INTO notifications (id, user_id, notification_type, event_type, title, message, data, read, sent, sent_at, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	RETURNING id, user_id, notification_type, event_type, title, message, data, read, sent, sent_at, created_at, updated_at

	`

	err := r.db.QueryRowx(
		query,
		notificationModel.ID,
		notificationModel.UserID,
		notificationModel.NotificationType,
		notificationModel.EventType,
		notificationModel.Title,
		notificationModel.Message,
		notificationModel.Data,
		notificationModel.Read,
		notificationModel.Sent,
		notificationModel.SentAt,
		notificationModel.CreatedAt,
		notificationModel.UpdatedAt,
	).StructScan(notificationModel)

	if err != nil {
		return nil, commonutils.AnalyzeDBErr(err)
	}

	// Convert DbNotification (with []byte Data) back to Notification (with map[string]any Data)
	var dataMap map[string]any
	if len(notificationModel.Data) > 0 {
		if err := json.Unmarshal(notificationModel.Data, &dataMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal notification data: %w", err)
		}
	}

	notification := &Notification{
		ID:               notificationModel.ID,
		UserID:           notificationModel.UserID,
		NotificationType: notificationModel.NotificationType,
		EventType:        notificationModel.EventType,
		Title:            notificationModel.Title,
		Message:          notificationModel.Message,
		Data:             dataMap,
		Read:             notificationModel.Read,
		Sent:             notificationModel.Sent,
		SentAt:           notificationModel.SentAt,
		CreatedAt:        notificationModel.CreatedAt,
		UpdatedAt:        notificationModel.UpdatedAt,
	}

	return notification, nil

}

func (r *repository) GetByUserID(ctx context.Context, request *QueryNotifications) ([]Notification, error) {

	query := `
	SELECT id, user_id, notification_type, event_type, title, message, data, read, sent, sent_at, created_at, updated_at
	FROM notifications
	WHERE user_id = $1
	ORDER BY created_at DESC
	`
	paramCount := 1
	params := []any{request.UserID}

	if request.Limit != nil {
		paramCount++
		params = append(params, request.Limit)
		query += fmt.Sprintf("\nLIMIT $%d", paramCount)
	}

	if request.Offset != nil {
		paramCount++
		params = append(params, request.Offset)
		query += fmt.Sprintf("\nOFFSET $%d", paramCount)
	}

	var dbNotifications []DbNotification
	err := r.db.SelectContext(ctx, &dbNotifications, query, params...)

	if err != nil {
		return nil, commonutils.AnalyzeDBErr(err)
	}

	// Convert []DbNotification to []Notification (unmarshal []byte Data to map[string]any)
	notifications := make([]Notification, len(dbNotifications))
	for i, dbNotif := range dbNotifications {
		var dataMap map[string]any
		if len(dbNotif.Data) > 0 {
			if err := json.Unmarshal(dbNotif.Data, &dataMap); err != nil {
				return nil, fmt.Errorf("failed to unmarshal notification data for id %s: %w", dbNotif.ID, err)
			}
		}

		notifications[i] = Notification{
			ID:               dbNotif.ID,
			UserID:           dbNotif.UserID,
			NotificationType: dbNotif.NotificationType,
			EventType:        dbNotif.EventType,
			Title:            dbNotif.Title,
			Message:          dbNotif.Message,
			Data:             dataMap,
			Read:             dbNotif.Read,
			Sent:             dbNotif.Sent,
			SentAt:           dbNotif.SentAt,
			CreatedAt:        dbNotif.CreatedAt,
			UpdatedAt:        dbNotif.UpdatedAt,
		}
	}

	return notifications, nil

}

func (r *repository) Update(ctx context.Context, request *UpdateNotification) error {
	query := `
	UPDATE notifications
	SET read = :read, updated_at = NOW()
	WHERE id = :id
	AND user_id = :user_id
	`
	result, err := r.db.NamedExecContext(ctx, query, request)
	if err != nil {
		return commonutils.AnalyzeDBErr(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("notification %s does not belong to user %s", request.ID, request.UserID)
	}
	return nil
}

func (r *repository) MarkAllAsReadByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `
      UPDATE notifications
      SET read = true, updated_at = NOW()
      WHERE user_id = $1
      AND read = false
      `
	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return 0, commonutils.AnalyzeDBErr(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}
