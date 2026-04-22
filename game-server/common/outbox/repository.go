package outbox

import (
	"context"
	"log/slog"

	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type repo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *repo {
	return &repo{
		db: db,
	}
}

const createOutboxQuery = `
	INSERT INTO outbox(routing_key, exchange, payload)
	VALUES(:routing_key, :exchange, :payload)
`

func (r *repo) CreateOutbox(ctx context.Context, params OutboxParams) error {
	_, err := r.db.NamedExecContext(ctx, createOutboxQuery, params)

	if err != nil {
		slog.Error("Error occured when attempting to create outbox",
			"err", err)
		return commonhelpers.AnalyzeDBErr(err)
	}

	return nil
}

func (r *repo) CreateOutboxTx(ctx context.Context, tx *sqlx.Tx, params OutboxParams) error {
	_, err := tx.NamedExecContext(ctx, createOutboxQuery, params)

	if err != nil {
		slog.Error("Error occured when attempting to create outbox in tx",
			"err", err)
		return commonhelpers.AnalyzeDBErr(err)
	}

	return nil
}

func (r *repo) UpdateOutboxToPublished(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE 
			outbox
		SET published_at = NOW()
		WHERE id = $1
	`
	results, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		slog.Error("Error occured when attempting to update published on outbox item",
			"err", err)
		return commonhelpers.AnalyzeDBErr(err)
	}

	rows, _ := results.RowsAffected()
	if rows == 0 {
		return commonconstants.ErrOutboxItemNotFound
	}

	return nil
}

func (r *repo) GetUnpublishedOutboxItems(ctx context.Context, limit *int) ([]*OutboxEvent, error) {

	defaultLimit := 20
	if limit == nil {
		limit = &defaultLimit
	}

	var outboxItem []*OutboxEvent

	query := `
	SELECT 
		id,
		routing_key,
		exchange,
		payload,
		created_at
	FROM outbox
	WHERE published_at IS NULL
	ORDER BY created_at ASC
	`

	err := r.db.SelectContext(ctx, &outboxItem, query)

	if err != nil {
		slog.Error("Error occured when attempting to retrive from outbox table",
			"err", err)
		return nil, commonhelpers.AnalyzeDBErr(err)
	}

	return outboxItem, nil
}
