package notification

import (
	"context"
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
	notificationModel := &Notification{
		ID:               uuid.New(),
		UserID:           createNotification.UserID,
		NotificationType: createNotification.NotificationType,
		EventType:        createNotification.EventType,
		Title:            createNotification.Title,
		Message:          createNotification.Message,
		Data:             createNotification.Data,
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

	return notificationModel, nil

}

func (r *repository) GetByUserID(ctx context.Context, request *QueryNotifications) ([]Notification, error) {

	query := `
	SELECT id, user_id, notification_type, event_type, title, message, data, read, sent, sent_at, created_at, updated_at
	FROM notifications
	WHERE user_id = $1
	ORDER BY created_at DESC
	`
	paramCount := 1
	params := []interface{}{request.UserID}

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

	var notifications []Notification
	err := r.db.SelectContext(ctx, &notifications, query, params...)

	if err != nil {
		return nil, commonutils.AnalyzeDBErr(err)
	}

	return notifications, nil

}

func (r *repository) Update(ctx context.Context, request *UpdateNotification) error {
	query := `
	UPDATE notifications
	SET read = :read
	WHERE id = :id
	AND user_id = :user_id
	`
	_, err := r.db.NamedExec(query, request)
	if err != nil {
		return commonutils.AnalyzeDBErr(err)
	}
	return nil
}
