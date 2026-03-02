package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"
)

var amqpTracer = otel.Tracer("api-gateway-amqp")

type AmqpAuthClient struct {
	ch       *amqp.Channel
	exchange string
}

func NewAmqpAuthClient(ch *amqp.Channel) *AmqpAuthClient {
	return &AmqpAuthClient{
		ch:       ch,
		exchange: commonconstants.AuthEventsExchange,
	}
}

// RpcCall sends a protobuf-encoded request via RabbitMQ and waits for a reply.
func (h *AmqpAuthClient) RpcCallNoWaitResponse(ctx context.Context, routingKey string, payload []byte) error {

	// Generate correlation ID to match request with reply
	correlationId := uuid.New().String()

	// Publish message with ReplyTo and CorrelationId
	err := h.ch.PublishWithContext(ctx, h.exchange, routingKey, false, false,
		amqp.Publishing{
			ContentType:   "application/protobuf",
			Body:          payload,
			CorrelationId: correlationId,
		})
	if err != nil {
		return fmt.Errorf("failed to publish rpc message: %w", err)
	}

	return nil
}

// SignupHandler handles member signup via AMQP (fire-and-forget).
func (h *AmqpAuthClient) SignupHandler(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := amqpTracer.Start(ctx, "amqp.Signup")
	defer span.End()

	var req pb.CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"message":    "Error parsing payload as JSON",
		})
		return
	}

	body, err := proto.Marshal(&req)
	if err != nil {
		slog.Error("Failed to marshal CreateMemberRequest", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"message":    "Internal server error",
		})
		return
	}

	if err := h.RpcCallNoWaitResponse(ctx, commonconstants.AuthMemberCreate, body); err != nil {
		slog.Error("Failed to publish signup message", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"message":    "Failed to process signup request",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"statusCode": http.StatusAccepted,
		"message":    "Signup request accepted",
	})
}
