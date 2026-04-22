package notification

import (
	"context"

	commonutils "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type InboxRepository interface {
	MarkEventProcessed(ctx context.Context, tx *sqlx.Tx, eventID uuid.UUID, eventType string) (bool, error)
}

type inboxRepository struct {
	db *sqlx.DB
}

func NewInboxRepository(db *sqlx.DB) *inboxRepository {
	return &inboxRepository{db: db}
}

// MarkEventProcessed inserts a row into processed_events.
// Returns true if the row was inserted (new event), false if it already existed (duplicate).
// Must run inside the same tx as the business side effect.
func (r *inboxRepository) MarkEventProcessed(ctx context.Context, tx *sqlx.Tx, eventID uuid.UUID, eventType string) (bool, error) {
	query := `
		INSERT INTO processed_events (event_id, event_type)
		VALUES ($1, $2)
		ON CONFLICT (event_id, event_type) DO NOTHING
	`
	result, err := tx.ExecContext(ctx, query, eventID, eventType)
	if err != nil {
		return false, commonutils.AnalyzeDBErr(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows == 1, nil
}
