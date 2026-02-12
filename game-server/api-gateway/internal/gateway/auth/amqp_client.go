package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"
)

var amqpTracer = otel.Tracer("api-gateway-amqp")

type AmqpAuthHandler struct {
	ch       *amqp.Channel
	exchange string
}

func NewAmqpAuthHandler(ch *amqp.Channel) *AmqpAuthHandler {
	return &AmqpAuthHandler{
		ch:       ch,
		exchange: commonconstants.AuthEventsExchange,
	}
}

// RpcCall sends a protobuf-encoded request via RabbitMQ and waits for a reply.
func (h *AmqpAuthHandler) RpcCall(ctx context.Context, routingKey string, payload []byte) ([]byte, error) {
	// Create exclusive, auto-delete reply queue
	replyQueue, err := h.ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare reply queue: %w", err)
	}

	// Start consuming from reply queue
	msgs, err := h.ch.Consume(replyQueue.Name, "", true, true, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to consume reply queue: %w", err)
	}

	// Generate correlation ID to match request with reply
	correlationId := uuid.New().String()

	// Publish message with ReplyTo and CorrelationId
	err = h.ch.PublishWithContext(ctx, h.exchange, routingKey, false, false,
		amqp.Publishing{
			ContentType:   "application/protobuf",
			Body:          payload,
			ReplyTo:       replyQueue.Name,
			CorrelationId: correlationId,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to publish rpc message: %w", err)
	}

	// Wait for reply with timeout
	timeout := time.After(30 * time.Second)
	for {
		select {
		case msg := <-msgs:
			if msg.CorrelationId == correlationId {
				// Check for error in headers
				if msg.Headers != nil {
					if _, ok := msg.Headers["x-error-code"]; ok {
						errMsg := "unknown error"
						if m, ok := msg.Headers["x-error-message"]; ok {
							errMsg = m.(string)
						}
						return nil, fmt.Errorf("%s", errMsg)
					}
				}
				return msg.Body, nil
			}
		case <-timeout:
			return nil, fmt.Errorf("rpc call timed out after 30s")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// CreateMemberAmqpHandler handles member signup via RabbitMQ RPC
func (h *AmqpAuthHandler) CreateMemberAmqpHandler(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := amqpTracer.Start(ctx, "amqp.CreateMember")
	defer span.End()

	var req pb.CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"message":    "Error parsing payload as JSON",
		})
		return
	}

	// Marshal protobuf request
	payload, err := proto.Marshal(&req)
	if err != nil {
		slog.Error("Failed to marshal CreateMemberRequest", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"message":    "Internal server error",
		})
		return
	}

	// RPC call via RabbitMQ
	data, err := h.RpcCall(ctx, commonconstants.AuthMemberCreate, payload)
	if err != nil {
		slog.Error("RPC call failed for CreateMember", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"message":    err.Error(),
		})
		return
	}

	// Unmarshal protobuf response
	var member pb.Member
	if err := proto.Unmarshal(data, &member); err != nil {
		slog.Error("Failed to unmarshal CreateMember response", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"message":    "Failed to parse response from auth service",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"message":    "Successfully created user",
		"result":     &member,
	})
}

// LoginMemberAmqpHandler handles member login via RabbitMQ RPC
func (h *AmqpAuthHandler) LoginMemberAmqpHandler(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := amqpTracer.Start(ctx, "amqp.LoginMember")
	defer span.End()

	var req pb.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"message":    fmt.Sprintf("Error parsing payload as JSON: %s", err),
		})
		return
	}

	// Marshal protobuf request
	payload, err := proto.Marshal(&req)
	if err != nil {
		slog.Error("Failed to marshal LoginRequest", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"message":    "Internal server error",
		})
		return
	}

	// RPC call via RabbitMQ
	data, err := h.RpcCall(ctx, commonconstants.AuthMemberLogin, payload)
	if err != nil {
		slog.Error("RPC call failed for LoginMember", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"message":    err.Error(),
		})
		return
	}

	// Unmarshal protobuf response
	var response pb.LoginResponse
	if err := proto.Unmarshal(data, &response); err != nil {
		slog.Error("Failed to unmarshal LoginMember response", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"message":    "Failed to parse response from auth service",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"message":    "Successfully logged in",
		"result":     &response,
	})
}
