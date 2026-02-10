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
	type DbNotification struct {
		ID               uuid.UUID  `json:"id" db:"id"`
		UserID           uuid.UUID  `json:"user_id" db:"user_id"`
		NotificationType string     `json:"notification_type" db:"notification_type"` // email, push, in_app
		EventType        string     `json:"event_type" db:"event_type"`               // member.signedup, game.match.ended
		Title            string     `json:"title" db:"title"`
		Message          string     `json:"message" db:"message"`
		Data             []byte     `json:"data" db:"data"` // Extra event data (JSONB)
		Read             bool       `json:"read" db:"read"`
		Sent             bool       `json:"sent" db:"sent"`
		SentAt           *time.Time `json:"sent_at" db:"sent_at"`
		CreatedAt        time.Time  `json:"created_at" db:"created_at"`
		UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	}

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

	// Define a temporary struct for database scanning (Data as []byte)
	type DbNotification struct {
		ID               uuid.UUID  `db:"id"`
		UserID           uuid.UUID  `db:"user_id"`
		NotificationType string     `db:"notification_type"`
		EventType        string     `db:"event_type"`
		Title            string     `db:"title"`
		Message          string     `db:"message"`
		Data             []byte     `db:"data"` // JSONB from DB
		Read             bool       `db:"read"`
		Sent             bool       `db:"sent"`
		SentAt           *time.Time `db:"sent_at"`
		CreatedAt        time.Time  `db:"created_at"`
		UpdatedAt        time.Time  `db:"updated_at"`
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
