package member

import (
	"context"
	"errors"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

type Consumer struct {
	service Service
	channel *amqp.Channel
}

func NewConsumer(service Service, channel *amqp.Channel) *Consumer {
	return &Consumer{
		service: service,
		channel: channel,
	}
}

func (c *Consumer) SetupConsumer() error {
	// Declare topic exchange
	err := c.channel.ExchangeDeclare(
		commonconstants.AuthEventsExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Declare queue
	_, err = c.channel.QueueDeclare(
		commonconstants.AuthSignupQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		slog.Error("Failed to declare auth RPC queue", "error", err)
		return err
	}

	// Bind routing key to queue
	err = c.channel.QueueBind(
		commonconstants.AuthSignupQueue,
		commonconstants.AuthMemberCreate,
		commonconstants.AuthEventsExchange,
		false,
		nil,
	)
	if err != nil {
		slog.Error("Failed to bind routing key", "key", commonconstants.AuthMemberCreate, "error", err)
		return err
	}

	slog.Info("Auth RPC infrastructure setup complete",
		"exchange", commonconstants.AuthEventsExchange,
		"queue", commonconstants.AuthSignupQueue,
	)
	return nil
}

// Listen starts consuming RPC requests in a goroutine.
func (c *Consumer) Listen() {
	go c.consumeRequests()
	slog.Info("Auth RPC consumer listening for requests...")
}

func (c *Consumer) consumeRequests() {
	msgs, err := c.channel.Consume(
		commonconstants.AuthSignupQueue,
		"",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		slog.Error("Failed to register auth RPC consumer", "error", err)
		return
	}

	for msg := range msgs {
		c.handleRequest(msg)
	}
}

func (c *Consumer) handleRequest(msg amqp.Delivery) {
	ctx := context.Background()

	slog.Info("Received auth AMQP request",
		"routing_key", msg.RoutingKey,
		"correlation_id", msg.CorrelationId,
	)

	switch msg.RoutingKey {
	case commonconstants.AuthMemberCreate:
		_, err := c.handleCreateMember(ctx, msg.Body)
		if err != nil {
			slog.Error("Failed to create member", "error", err)
		}

	default:
		slog.Warn("Unknown routing key", "routing_key", msg.RoutingKey)
	}

	msg.Ack(false)
	slog.Info("Auth AMQP request processed",
		"routing_key", msg.RoutingKey,
		"correlation_id", msg.CorrelationId,
	)
}

func (c *Consumer) handleCreateMember(ctx context.Context, body []byte) ([]byte, error) {
	var req pb.CreateMemberRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		return nil, errors.New("failed to parse CreateMemberRequest")
	}

	member, err := c.service.CreateMember(ctx, &req)
	if err != nil {
		return nil, err
	}

	data, err := proto.Marshal(member)
	if err != nil {
		return nil, errors.New("failed to marshal CreateMember response")
	}
	return data, nil
}

