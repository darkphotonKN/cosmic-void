package stats

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	commontypes "github.com/darkphotonKN/cosmic-void-server/common/types"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConsumerService defines what the consumer needs from the service
type ConsumerService interface {
	ProcessMatchCompleted(ctx context.Context, req *pb.ProcessMatchCompletedRequest) (*pb.ProcessMatchCompletedResponse, error)
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

const statsMatchEndedQueue = "stats.game.match.ended"

// Listen starts consuming messages from the configured queues
func (c *Consumer) Listen() {
	// Start consuming match completed events
	go c.consumeMatchCompleted()

	slog.Info("Stats consumer listening for events...")
}

// consumeMatchCompleted handles match completion events
func (c *Consumer) consumeMatchCompleted() {
	msgs, err := c.channel.Consume(
		statsMatchEndedQueue, // queue name
		"",                   // consumer
		false,                // auto-ack (set to false for manual ack)
		false,                // exclusive
		false,                // no-local
		false,                // no-wait
		nil,                  // args
	)

	if err != nil {
		slog.Error("Failed to register consumer", "error", err)
		return
	}

	for msg := range msgs {
		var event commontypes.MatchEndState
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			slog.Error("Failed to parse match completed event", "error", err)
			msg.Nack(false, false)
			continue
		}

		slog.Info("Unmarshalled MatchEndState extracted from event.", "event type", commonconstants.GameMatchEndedEvent, "event data", event)

		ctx := context.Background()
		req := &pb.ProcessMatchCompletedRequest{
			SessionId:      event.SessionID.String(),
			MatchStartedAt: timestamppb.New(event.MatchStartedAt),
			MatchEndedAt:   timestamppb.New(event.MatchEndedAt),
			Players:        make([]*pb.PlayerMatchResults, len(event.PlayerMatchResults)),
		}

		for i, player := range event.PlayerMatchResults {
			req.Players[i] = &pb.PlayerMatchResults{
				MemberId:      player.MemberID,
				Username:      player.Username,
				Win:           player.Win,
				FinalPosition: player.FinalPosition,
				Kills:         player.Kills,
				Deaths:        player.Deaths,
			}
		}

		response, err := c.service.ProcessMatchCompleted(ctx, req)

		if err != nil {
			slog.Error("Failed to process match completed",
				"error", err,
				"session_id", event.SessionID,
			)

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
			"session_id", event.SessionID,
			"players_processed", response.PlayersProcessed,
			"success", response.Success,
			"message", response.Message,
		)
	}
}

// Helper method to set up AMQP exchange and bindings
func SetupAMQPInfrastructure(channel *amqp.Channel) error {
	err := channel.ExchangeDeclare(
		commonconstants.GameMatchEndedEvent, // exchange name
		"fanout",                            // exchange type
		true,                                // durable
		false,                               // auto-deleted
		false,                               // internal
		false,                               // no-wait
		nil,                                 // arguments
	)
	if err != nil {
		return err
	}

	// Declare the queue
	_, err = channel.QueueDeclare(
		statsMatchEndedQueue, // queue name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		slog.Error("Failed to declare queue", "error", err)
		return err
	}

	// Bind the queue to the exchange
	err = channel.QueueBind(
		statsMatchEndedQueue,                // queue name
		"match.completed",                   // routing key
		commonconstants.GameMatchEndedEvent, // exchange
		false,                               // no-wait
		nil,                                 // args
	)
	if err != nil {
		return err
	}

	return nil
}
