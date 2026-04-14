package stats

import (
	"context"
	"errors"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

// ConsumerService defines what the consumer needs from the service
type ConsumerService interface {
	ProcessMatchCompleted(ctx context.Context, req *pb.MatchEndedEvent) (*ProcessMatchCompletedResponse, error)
	UpdatePlayerRankings(ctx context.Context, updateData *pb.MemberProfileUpdatedEvent) error
}

type Consumer struct {
	service ConsumerService
	channel *amqp.Channel
}

func NewConsumer(service ConsumerService, channel *amqp.Channel) *Consumer {
	return &Consumer{
		service: service,
		channel: channel,
	}
}

// Listen starts consuming messages from the configured queues
func (c *Consumer) Listen() {
	// Start consuming match completed events
	go c.consumeMatchCompleted()
	go c.consumeProfileUpdated()

	slog.Info("Stats consumer listening for events...")
}

func (c *Consumer) consumeProfileUpdated() {
	msgs, err := c.channel.Consume(
		commonconstants.StatsAuthProfileUpdatedQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		slog.Error("Failed to register consumer", "error", err)
		return
	}

	for msg := range msgs {
		var memberUpdated pb.MemberProfileUpdatedEvent

		if err := proto.Unmarshal(msg.Body, &memberUpdated); err != nil {
			slog.Error("Failed to parse member profile updated event", "error", err)
			msg.Nack(false, false)
			continue
		}

		slog.Info("after member profile updated data unmarshal",
			"memberUpdated", memberUpdated)

		err := c.service.UpdatePlayerRankings(context.Background(), &memberUpdated)

		if err != nil {
			if errors.Is(err, commonconstants.ErrTransient) {
				slog.Warn("Transient error, retrying", "error", err, "member_id", memberUpdated.MemberId)
				msg.Nack(false, true)
				continue
			}

			slog.Error("Failed to update member profile data after consuming memberProfileUpdatedEvent event", "error", err)
			msg.Nack(false, false)
			continue
		}

		msg.Ack(false)

		slog.Info("PlayerRankings updated successfully following member profile update",
			"member_id", memberUpdated.MemberId,
			"success", true,
			"message", "PlayerRankings updated successfully following member profile update",
		)
	}
}

// consumeMatchCompleted handles match completion events
func (c *Consumer) consumeMatchCompleted() {
	msgs, err := c.channel.Consume(
		commonconstants.StatsGameMatchEndedQueue, // queue name
		"",                                       // consumer
		false,                                    // auto-ack
		false,                                    // exclusive
		false,                                    // no-local
		false,                                    // no-wait
		nil,                                      // args
	)

	if err != nil {
		slog.Error("Failed to register consumer", "error", err)
		return
	}

	for msg := range msgs {
		var event pb.MatchEndedEvent

		slog.Debug("Raw message received",
			"body_length", len(msg.Body),
			"content_type", msg.ContentType,
			"body_preview", string(msg.Body[:min(100, len(msg.Body))]),
		)

		if err := proto.Unmarshal(msg.Body, &event); err != nil {
			slog.Error("Failed to parse match completed event", "error", err)
			msg.Nack(false, false)
			continue
		}

		slog.Info("Unmarshalled MatchEndState extracted from event.", "event type", commonconstants.GameMatchEnded, "event data", event)

		ctx := context.Background()

		response, err := c.service.ProcessMatchCompleted(ctx, &event)

		if err != nil {
			slog.Error("Failed to process match completed",
				"error", err,
				"session_id", event.SessionId,
			)

			// TODO: check retry count in header

			// retry if error is transient
			if errors.Is(err, commonconstants.ErrTransient) {
				msg.Nack(false, true)
			} else {
				msg.Nack(false, false) // TODO: setup DLQ
			}

			continue
		}

		// successfully processed
		msg.Ack(false)
		slog.Info("Match completed processed successfully",
			"session_id", event.SessionId,
			"success", response.Success,
			"message", response.Message,
		)
	}
}

// Helper method to set up AMQP exchange and bindings
func SetupAMQPInfrastructure(channel *amqp.Channel) error {

	// --- Game Match Ended ---

	err := channel.ExchangeDeclare(
		commonconstants.GameEventsExchange, // exchange name
		"topic",                            // exchange type
		true,                               // durable
		false,                              // auto-deleted
		false,                              // internal
		false,                              // no-wait
		nil,                                // arguments
	)
	if err != nil {
		return err
	}

	// Declare the queue
	_, err = channel.QueueDeclare(
		commonconstants.StatsGameMatchEndedQueue, // queue name
		true,                                     // durable
		false,                                    // delete when unused
		false,                                    // exclusive
		false,                                    // no-wait
		nil,                                      // arguments
	)
	if err != nil {
		slog.Error("Failed to declare queue", "error", err)
		return err
	}

	// Bind the queue to the exchange
	err = channel.QueueBind(
		commonconstants.StatsGameMatchEndedQueue, // queue name
		commonconstants.GameMatchEnded,           // routing key
		commonconstants.GameEventsExchange,       // exchange
		false,                                    // no-wait
		nil,                                      // args
	)
	if err != nil {
		return err
	}

	// --- Member Profile Update ---

	err = channel.ExchangeDeclare(
		commonconstants.AuthEventsExchange, // exchange name
		"topic",                            // exchange type
		true,                               // durable
		false,                              // auto-deleted
		false,                              // internal
		false,                              // no-wait
		nil,                                // arguments
	)
	if err != nil {
		return err
	}

	// Declare the queue
	_, err = channel.QueueDeclare(
		commonconstants.StatsAuthProfileUpdatedQueue, // queue name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		slog.Error("Failed to declare queue", "error", err)
		return err
	}

	// Bind the queue to the exchange
	err = channel.QueueBind(
		commonconstants.StatsAuthProfileUpdatedQueue, // queue name
		commonconstants.MemberProfileUpdated,         // routing key
		commonconstants.AuthEventsExchange,           // exchange
		false,                                        // no-wait
		nil,                                          // args
	)
	if err != nil {
		return err
	}

	return nil
}
