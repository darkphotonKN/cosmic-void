package notification

import (
	"context"
	"encoding/json"
	"log/slog"

	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ConsumerService defines what the consumer needs from the service
// This interface specifies methods that will be called when processing events
type ConsumerService interface {
	ProcessMemberSignedUp(ctx context.Context, payload *commonconstants.MemberSignedUpEventPayload) error
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
	go c.consumeMemberSignedUp()
	slog.Info("Notification consumer listening for events...")
}

func (c *Consumer) consumeMemberSignedUp() {
	msgs, err := c.channel.Consume(
		commonconstants.NotificationMemberSignedUpQueue, // queue name
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		slog.Error("Failed to register consumer", "error", err)
		return
	}

	slog.Info("Started consuming member.signedup events")

	for msg := range msgs {
		var payload commonconstants.MemberSignedUpEventPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			slog.Error("Failed to parse event", "error", err)
			msg.Nack(false, false) // RabbitMQ：處理失敗，不要重試
			continue
		}
		slog.Info("Received member signed up event",
			"user_id", payload.UserID,
			"name", payload.Name,
		)

		ctx := context.Background()
		if err := c.service.ProcessMemberSignedUp(ctx, &payload); err != nil {
			slog.Error("Failed to process event", "error", err)
			msg.Nack(false, true) // RabbitMQ：處理失敗，請重試（requeue）
			continue
		}

		// successfully processed (ACK)
		msg.Ack(false)
		slog.Info("Member signed up notification created successfully",
			"user_id", payload.UserID,
		)
	}

}

func SetupAMQPInfrastructure(channel *amqp.Channel) error {
	// TODO: Declare exchanges for auth events
	err := channel.ExchangeDeclare(
		"auth.events", // exchange name
		"topic",       // exchange type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		slog.Error("Failed to declare exchange", "error", err)
		return err
	}

	// TODO: Declare queues and bindings
	_, err = channel.QueueDeclare(
		commonconstants.NotificationMemberSignedUpQueue, // queue name
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
	// QueueBind
	err = channel.QueueBind(
		commonconstants.NotificationMemberSignedUpQueue, // queue name
		commonconstants.MemberSignedUpEvent,             // routing key
		"auth.events",                                   // exchange
		false,                                           // no-wait
		nil,                                             // args
	)
	if err != nil {
		slog.Error("Failed to bind queue", "error", err)
		return err
	}

	slog.Info("Notification AMQP infrastructure setup completed")
	return nil
}
