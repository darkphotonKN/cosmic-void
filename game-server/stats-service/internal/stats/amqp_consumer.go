package stats

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConsumerService defines what the consumer needs from the service
type ConsumerService interface {
	CreateMatchHistory(ctx context.Context, req *pb.CreateMatchHistoryRequest) (*pb.MatchHistory, error)
	CreatePlayerMatchStats(ctx context.Context, req *pb.CreatePlayerMatchStatsRequest) (*pb.PlayerMatchStats, error)
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

// MatchCompletedEvent represents the event payload when a match is completed
type MatchCompletedEvent struct {
	SessionID      string    `json:"session_id"`
	MemberID       string    `json:"member_id"`
	Win            bool      `json:"win"`
	FinalPosition  int32     `json:"final_position"`
	Kills          int32     `json:"kills"`
	Deaths         int32     `json:"deaths"`
	RatingBefore   *int32    `json:"rating_before,omitempty"`
	RatingAfter    *int32    `json:"rating_after,omitempty"`
	RatingChange   *int32    `json:"rating_change,omitempty"`
	MatchStartedAt time.Time `json:"match_started_at"`
}

// Listen starts consuming messages from the configured queues
func (c *Consumer) Listen() {
	// Set up queue for match completed events
	_, err := c.channel.QueueDeclare(
		"stats.match.completed", // queue name
		true,                    // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		slog.Error("Failed to declare queue", "error", err)
		return
	}

	// Start consuming match completed events
	go c.consumeMatchCompleted()

	slog.Info("Stats consumer listening for events...")
}

// consumeMatchCompleted handles match completion events
func (c *Consumer) consumeMatchCompleted() {
	msgs, err := c.channel.Consume(
		"stats.match.completed", // queue name
		"",                      // consumer
		false,                   // auto-ack (set to false for manual ack)
		false,                   // exclusive
		false,                   // no-local
		false,                   // no-wait
		nil,                     // args
	)
	if err != nil {
		slog.Error("Failed to register consumer", "error", err)
		return
	}

	for msg := range msgs {
		// Parse the message
		var event MatchCompletedEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			slog.Error("Failed to parse match completed event", "error", err)
			msg.Nack(false, false) // Negative acknowledgment
			continue
		}

		// Create match history record
		ctx := context.Background()
		req := &pb.CreateMatchHistoryRequest{
			SessionId:      event.SessionID,
			MemberId:       event.MemberID,
			Win:            event.Win,
			FinalPosition:  event.FinalPosition,
			Kills:          event.Kills,
			Deaths:         event.Deaths,
			RatingBefore:   event.RatingBefore,
			RatingAfter:    event.RatingAfter,
			RatingChange:   event.RatingChange,
			MatchStartedAt: timestamppb.New(event.MatchStartedAt),
		}

		_, err = c.service.CreateMatchHistory(ctx, req)
		if err != nil {
			slog.Error("Failed to create match history",
				"error", err,
				"session_id", event.SessionID,
				"member_id", event.MemberID,
			)
			msg.Nack(false, true) // Negative acknowledgment with requeue
			continue
		}

		// Successfully processed
		msg.Ack(false)
		slog.Info("Match history created",
			"session_id", event.SessionID,
			"member_id", event.MemberID,
		)
	}
}

// Helper method to set up AMQP exchange and bindings
func SetupAMQPInfrastructure(channel *amqp.Channel) error {
	// Declare the exchange
	err := channel.ExchangeDeclare(
		"game.events", // exchange name
		"topic",       // exchange type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return err
	}

	// Declare the queue
	_, err = channel.QueueDeclare(
		"stats.match.completed", // queue name
		true,                    // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		return err
	}

	// Bind the queue to the exchange
	err = channel.QueueBind(
		"stats.match.completed", // queue name
		"match.completed",       // routing key
		"game.events",           // exchange
		false,                   // no-wait
		nil,                     // args
	)
	if err != nil {
		return err
	}

	return nil
}

// StatsUpdateEvent represents an event to update player stats
type StatsUpdateEvent struct {
	MemberID            uuid.UUID `json:"member_id"`
	GamesPlayed         int32     `json:"games_played"`
	Wins                int32     `json:"wins"`
	Losses              int32     `json:"losses"`
	Kills               int32     `json:"kills"`
	Deaths              int32     `json:"deaths"`
	TimesPlacedTopThree int32     `json:"times_placed_top_three"`
}