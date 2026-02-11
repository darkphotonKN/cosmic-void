package member

import (
	"context"
	"errors"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"google.golang.org/protobuf/proto"
)

type Consumer struct {
	service   Service
	publishCh commonbroker.Publisher
}

func NewConsumer(service Service, publishCh commonbroker.Publisher) *Consumer {
	return &Consumer{
		service:   service,
		publishCh: publishCh,
	}
}

func (c *Consumer) SetupConsumer() error {
	// Declare topic exchange
	err := c.publishCh.ExchangeDeclare(
		commonconstants.AuthEventsExchange,
		"topic",
	)
	if err != nil {
		return err
	}

	// Declare queue
	_, err = c.publishCh.QueueDeclare(
		commonconstants.StatsAuthQueue,
	)
	if err != nil {
		slog.Error("Failed to declare auth RPC queue", "error", err)
		return err
	}

	// Bind routing keys to queue
	routingKeys := []string{
		commonconstants.AuthMemberCreate,
		commonconstants.AuthMemberLogin,
	}

	for _, key := range routingKeys {
		err = c.publishCh.QueueBind(
			commonconstants.StatsAuthQueue,
			key,
			commonconstants.AuthEventsExchange,
		)
		if err != nil {
			slog.Error("Failed to bind routing key", "key", key, "error", err)
			return err
		}
	}

	slog.Info("Auth RPC infrastructure setup complete",
		"exchange", commonconstants.AuthEventsExchange,
		"queue", commonconstants.StatsAuthQueue,
	)
	return nil
}

// Listen starts consuming RPC requests in a goroutine.
func (c *Consumer) Listen() {
	go c.consumeRequests()
	slog.Info("Auth RPC consumer listening for requests...")
}

func (c *Consumer) consumeRequests() {
	msgs, err := c.publishCh.Consume(
		commonconstants.StatsAuthQueue,
		false, // manual ack
	)
	if err != nil {
		slog.Error("Failed to register auth RPC consumer", "error", err)
		return
	}

	for msg := range msgs {
		c.handleRequest(msg)
	}
}

func (c *Consumer) handleRequest(msg commonbroker.Delivery) {
	ctx := context.Background()

	slog.Info("Received auth RPC request",
		"routing_key", msg.RoutingKey,
		"correlation_id", msg.CorrelationId,
		"reply_to", msg.ReplyTo,
	)

	var responseBody []byte
	var rpcErr error

	switch msg.RoutingKey {
	case commonconstants.AuthMemberCreate:
		responseBody, rpcErr = c.handleCreateMember(ctx, msg.Body)

	case commonconstants.AuthMemberLogin:
		responseBody, rpcErr = c.handleLoginMember(ctx, msg.Body)

	default:
		slog.Warn("Unknown routing key for auth RPC", "routing_key", msg.RoutingKey)
		c.replyError(msg, "unknown routing key: "+msg.RoutingKey)
		msg.Acknowledger.Ack(false)
		return
	}

	if rpcErr != nil {
		slog.Error("Auth RPC handler error",
			"routing_key", msg.RoutingKey,
			"error", rpcErr,
		)
		c.replyError(msg, rpcErr.Error())
		msg.Acknowledger.Ack(false)
		return
	}

	// Publish success response to reply queue
	err := c.publishCh.PublishWithContext(ctx, "", msg.ReplyTo,
		commonbroker.Message{
			ContentType:   "application/protobuf",
			Body:          responseBody,
			CorrelationId: msg.CorrelationId,
		})
	if err != nil {
		slog.Error("Failed to publish RPC reply", "error", err)
	}

	msg.Acknowledger.Ack(false)
	slog.Info("Auth RPC request processed successfully",
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

func (c *Consumer) handleLoginMember(ctx context.Context, body []byte) ([]byte, error) {
	var req pb.LoginRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		return nil, errors.New("failed to parse LoginRequest")
	}

	response, err := c.service.LoginMember(ctx, &req)
	if err != nil {
		return nil, err
	}

	data, err := proto.Marshal(response)
	if err != nil {
		return nil, errors.New("failed to marshal LoginMember response")
	}
	return data, nil
}

// replyError sends an error response back to the reply queue.
func (c *Consumer) replyError(msg commonbroker.Delivery, errMsg string) {
	ctx := context.Background()
	err := c.publishCh.PublishWithContext(ctx, "", msg.ReplyTo,
		commonbroker.Message{
			ContentType:   "application/protobuf",
			CorrelationId: msg.CorrelationId,
			Headers: map[string]interface{}{
				"x-error-code":    int32(1),
				"x-error-message": errMsg,
			},
		})
	if err != nil {
		slog.Error("Failed to publish RPC error reply", "error", err)
	}
}
