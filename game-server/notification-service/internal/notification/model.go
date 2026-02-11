package notification

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID               uuid.UUID      `json:"id" db:"id"`
	UserID           uuid.UUID      `json:"user_id" db:"user_id"`
	NotificationType string         `json:"notification_type" db:"notification_type"` // email, push, in_app
	EventType        string         `json:"event_type" db:"event_type"`               // member.signedup, game.match.ended
	Title            string         `json:"title" db:"title"`
	Message          string         `json:"message" db:"message"`
	Data             map[string]any `json:"data" db:"data"` // Extra event data (JSONB)
	Read             bool           `json:"read" db:"read"`
	Sent             bool           `json:"sent" db:"sent"`
	SentAt           *time.Time     `json:"sent_at" db:"sent_at"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
}

// DB

type CreateNotification struct {
	UserID           uuid.UUID      `json:"user_id" db:"user_id"`
	NotificationType string         `json:"notification_type" db:"notification_type"`
	EventType        string         `json:"event_type" db:"event_type"`
	Title            string         `json:"title" db:"title"`
	Message          string         `json:"message" db:"message"`
	Data             map[string]any `json:"data" db:"data"`
}

type QueryNotifications struct {
	UserID uuid.UUID `json:"user_id" db:"user_id"`
	Limit  *int      `json:"limit" db:"limit"`
	Offset *int      `json:"offset" db:"offset"`
}

type UpdateNotification struct {
	ID     uuid.UUID `json:"id" db:"id"`
	UserID uuid.UUID `json:"user_id" db:"user_id"`
	Read   bool      `json:"read" db:"read"`
}

// service

type GetNotificationRequest struct {
	UserID string `json:"user_id" db:"user_id"`
	Limit  *int   `json:"limit" db:"limit"`
	Offset *int   `json:"offset" db:"offset"`
}

type NotificationList struct {
	UserID  uuid.UUID      `json:"user_id" db:"user_id"`
	Title   string         `json:"title" db:"title"`
	Message string         `json:"message" db:"message"`
	Data    map[string]any `json:"data" db:"data"`
}

type NotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
	Total         int            `json:"total"`
}

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
