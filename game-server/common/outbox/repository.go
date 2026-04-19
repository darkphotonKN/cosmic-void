package outbox

import (
	"context"
	"log/slog"

	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
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

func (r *repo) CreateOutbox(ctx context.Context, params OutboxParams) error {
	query := `
		INSERT INTO outbox(routing_key, exchange, payload)
		VALUES(:routing_key, :exchange, :payload)
	`

	_, err := r.db.NamedExecContext(ctx, query, params)

	if err != nil {
		slog.Error("Error occured when attempting to create outbox",
			"err", err)
		return commonhelpers.AnalyzeDBErr(err)
	}

	return nil
}

func (r *repo) GetOldestUnpublishedOutboxItem(ctx context.Context) (*OutboxEvent, error) {
	var outboxItem OutboxEvent

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
	LIMIT 1
	`

	err := r.db.GetContext(ctx, &outboxItem, query)

	if err != nil {
		slog.Error("Error occured when attempting to retrive from outbox table",
			"err", err)
		return nil, commonhelpers.AnalyzeDBErr(err)
	}

	return &outboxItem, nil
}
