package outbox

import (
	"context"
	"log/slog"
	"time"

	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

/**
* Houses the struct and logic of the workers that process the outbox
* event messages for publishing.
**/

type OutboxWorker struct {
	workCycle       time.Duration
	batchCount      *int
	outboxRetriever OutboxRetriever
	publishCh       commonbroker.Publisher
}

type OutboxRetriever interface {
	GetPendingOutboxItems(ctx context.Context, limit *int) ([]*OutboxEvent, error)
	UpdateOutboxToPublished(ctx context.Context, id uuid.UUID) error
}

func NewOutboxWorker(workCycle time.Duration, batchCount int, outboxRetriever OutboxRetriever, publishCh commonbroker.Publisher) *OutboxWorker {
	return &OutboxWorker{
		workCycle:       workCycle,
		batchCount:      &batchCount,
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
	outboxEvts, err := w.outboxRetriever.GetPendingOutboxItems(ctx, w.batchCount)

	if err != nil {
		slog.Error("Error when attempting to retrieve latest outbox event.",
			"err", err)
		return err
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

		// no error, rabbitmq acknowledged, recieved, update status.
		go func(evtId uuid.UUID) {
			var offsetTime = time.Second * 10
			var retries = 0

			for {
				err := w.outboxRetriever.UpdateOutboxToPublished(ctx, evtId)

				if err != nil {
					if commonhelpers.IsTransientError(err) && retries < 5 {
						slog.Error("Transient error when attempting to update outbox to published.",
							"err", err)

						time.Sleep(offsetTime)

						// exponential backoff
						offsetTime = offsetTime * 2
						retries += 1
						continue
					}

					slog.Error("Error when attempting to update outbox to published.",
						"err", err)

					return
				}

				return
			}
		}(evt.ID)

		// update worked, exit goroutine
		slog.Debug("successfully published outbox event",
			"event_id", evt.ID,
			"event", evt.RoutingKey,
			"event_exchange", evt.Exchange,
		)
	}

	return nil
}
