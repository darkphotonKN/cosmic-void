package notification

import (
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ConsumerService defines what the consumer needs from the service
// This interface specifies methods that will be called when processing events
type ConsumerService interface {
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
	// TODO: Add queue declarations for item-related events here
	// Example:
	// _, err := c.channel.QueueDeclare(
	// 	"items.item.acquired", // queue name
	// 	true,                  // durable
	// 	false,                 // delete when unused
	// 	false,                 // exclusive
	// 	false,                 // no-wait
	// 	nil,                   // arguments
	// )
	// if err != nil {
	// 	slog.Error("Failed to declare queue", "error", err)
	// 	return
	// }

	// TODO: Start consuming item-related events
	// go c.consumeItemAcquired()
	// go c.consumeItemUsed()

	slog.Info("Notification consumer listening for events...")
}

// TODO: Add consumer methods for item-related events
// Example:
// func (c *Consumer) consumeItemAcquired() {
// 	msgs, err := c.channel.Consume(
// 		"items.item.acquired", // queue name
// 		"",                    // consumer
// 		false,                 // auto-ack
// 		false,                 // exclusive
// 		false,                 // no-local
// 		false,                 // no-wait
// 		nil,                   // args
// 	)
//
// 	if err != nil {
// 		slog.Error("Failed to register consumer", "error", err)
// 		return
// 	}
//
// 	for msg := range msgs {
// 		// Parse and process the event
// 		// Call c.service.ProcessItemAcquired(...)
// 		msg.Ack(false)
// 	}
// }

// SetupAMQPInfrastructure sets up AMQP exchanges, queues, and bindings
func SetupAMQPInfrastructure(channel *amqp.Channel) error {
	// TODO: Declare exchanges for item events
	// Example:
	// err := channel.ExchangeDeclare(
	// 	"items.events", // exchange name
	// 	"topic",        // exchange type
	// 	true,           // durable
	// 	false,          // auto-deleted
	// 	false,          // internal
	// 	false,          // no-wait
	// 	nil,            // arguments
	// )
	// if err != nil {
	// 	return err
	// }

	// TODO: Declare queues and bindings
	// Example for item acquired events:
	// _, err = channel.QueueDeclare(
	// 	"items.item.acquired", // queue name
	// 	true,                  // durable
	// 	false,                 // delete when unused
	// 	false,                 // exclusive
	// 	false,                 // no-wait
	// 	nil,                   // arguments
	// )
	// if err != nil {
	// 	return err
	// }
	//
	// err = channel.QueueBind(
	// 	"items.item.acquired", // queue name
	// 	"item.acquired",       // routing key
	// 	"items.events",        // exchange
	// 	false,                 // no-wait
	// 	nil,                   // args
	// )

	slog.Info("Notification AMQP infrastructure setup completed")
	return nil
}
