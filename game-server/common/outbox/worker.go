package outbox

import (
	"context"
	"log/slog"
	"time"

	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
)

/**
* Houses the struct and logic of the workers that process the outbox
* event messages for publishing.
**/

type OutboxWorker struct {
	workCycle       time.Duration
	batchCount      int
	db              *sqlx.DB
	outboxRetriever OutboxRetriever
	publishCh       commonbroker.Publisher
}

type OutboxRetriever interface {
	GetPendingOutboxItems(ctx context.Context, tx *sqlx.Tx, limit int) ([]*OutboxEvent, error)
	UpdateOutboxToPublished(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) error
}

func NewOutboxWorker(workCycle time.Duration, batchCount int, db *sqlx.DB, outboxRetriever OutboxRetriever, publishCh commonbroker.Publisher) *OutboxWorker {
	return &OutboxWorker{
		workCycle:       workCycle,
		batchCount:      batchCount,
		db:              db,
		outboxRetriever: outboxRetriever,
		publishCh:       publishCh,
	}
}

/**
* Sets up the cancel and publish for select which initiates the workers
**/
func (w *OutboxWorker) Run(ctx context.Context) {
	timer := time.NewTicker(w.workCycle)

	slog.Info("Initiating outbox workers..")

	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			w.PublishOutboxEvents(ctx)

			// cancelled
		case <-ctx.Done():

			return
		}
	}
}

/**
* Publishes a single event from an outbox item pulled from the outbox table.
**/
func (w *OutboxWorker) PublishOutboxEvents(ctx context.Context) error {
	tx, err := w.db.BeginTxx(ctx, nil)
	if err != nil {
		slog.Error("Failed to begin outbox tx", "err", err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	outboxEvts, err := w.outboxRetriever.GetPendingOutboxItems(ctx, tx, w.batchCount)
	if err != nil {
		slog.Error("Error when attempting to retrieve latest outbox event.",
			"err", err)
		return err
	}

	if len(outboxEvts) == 0 {
		return nil
	}

	for _, evt := range outboxEvts {
		err = w.publishCh.PublishWithContext(
			ctx,
			evt.Exchange,
			evt.RoutingKey,
			commonbroker.Message{
				ContentType:  "application/protobuf",
				Body:         evt.Payload,
				DeliveryMode: amqp.Persistent,
			})

		if err != nil {
			slog.Error("Error occured when attempting to publish event from outbox item.",
				"outbox_id", evt.ID,
				"err", err)
			continue
		}

		if err := w.outboxRetriever.UpdateOutboxToPublished(ctx, tx, evt.ID); err != nil {
			slog.Error("Error when updating outbox to published.",
				"outbox_id", evt.ID,
				"err", err)
			continue
		}

		slog.Debug("successfully published outbox event",
			"event_id", evt.ID,
			"event", evt.RoutingKey,
			"event_exchange", evt.Exchange,
		)
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Failed to commit outbox tx", "err", err)
		return err
	}

	return nil
}
