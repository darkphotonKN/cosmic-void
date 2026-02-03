package notification

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID               uuid.UUID      `db:"id"`
	UserID           uuid.UUID      `db:"user_id"`
	NotificationType string         `db:"notification_type"` // email, push, in_app
	EventType        string         `db:"event_type"`        // member.signedup, game.match.ended
	Title            string         `db:"title"`
	Message          string         `db:"message"`
	Data             map[string]any `db:"data"` // Extra event data (JSONB)
	Read             bool           `db:"read"`
	Sent             bool           `db:"sent"`
	SentAt           *time.Time     `db:"sent_at"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
}

type CreateNotification struct {
	UserID           uuid.UUID      `db:"user_id"`
	NotificationType string         `db:"notification_type"`
	EventType        string         `db:"event_type"`
	Title            string         `db:"title"`
	Message          string         `db:"message"`
	Data             map[string]any `db:"data"`
}

type QueryNotifications struct {
	UserID uuid.UUID `json:"user_id" db:"user_id"`
	Limit  *int      `json:"limit" db:"limit"`
	Offset *int      `json:"offset" db:"offset"`
}
