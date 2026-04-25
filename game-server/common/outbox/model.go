package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OutboxParams struct {
	RoutingKey string `db:"routing_key"`
	Exchange   string `db:"exchange"`
	Payload    []byte `db:"payload"`
}

type OutboxEvent struct {
	ID         uuid.UUID `db:"id"`
	RoutingKey string    `db:"routing_key"`
	Exchange   string    `db:"exchange"`
	Payload    []byte    `db:"payload"`
	CreatedAt  time.Time `db:"created_at"`
	// nil = pending, not nil = processed
	PublishedAt *time.Time `db:"published_at"`
}

// isp interface for all services
type OutboxPublisher interface {
	CreateOutbox(ctx context.Context, params OutboxParams) error
	CreateOutboxTx(ctx context.Context, tx *sqlx.Tx, params OutboxParams) error
}
