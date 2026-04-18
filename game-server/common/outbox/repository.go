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

func (r *repo) CreateOutbox(ctx context.Context, params OutboxParams) error {
	query := `
		INSERT INTO outbox(routing_key, exchange, event_type, payload)
		VALUES(:routing_key, :exchange, event_type, :payload)
	`

	_, err := r.db.NamedExecContext(ctx, query, params)

	if err != nil {
		slog.Error("Error occured when attempting to create outbox",
			"err", err)
		return commonhelpers.AnalyzeDBErr(err)
	}

	return nil
}
