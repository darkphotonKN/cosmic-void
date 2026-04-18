package outbox

import (
	"time"

	"github.com/google/uuid"
)

type OutboxParams struct {
	RoutingKey  string    `db:"routing_key"`
	Exchange    string    `db:"exchange"`
	EventType   string    `db:"event_type"`
	Payload     []byte    `db:"payload"`
	PublishedAt time.Time `db:"published_at"`
}

type OutboxEvent struct {
	ID         uuid.UUID
	RoutingKey string
	Exchange   string
	EventType  string
	Payload    []byte
	CreatedAt  time.Time
	// nil = pending, not nil = processed
	PublishedAt *time.Time
}
