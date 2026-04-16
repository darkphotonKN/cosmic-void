package items

import (
	"errors"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

type Consumer struct {
	service ConsumerService
	channel *amqp.Channel
}

type ConsumerService interface {
	ProcessItemsExtracted(req *pb.ItemsExtractedEvent) error
}

func (c *Consumer) Listen() {
	go c.consumeItemsExtracted()

	slog.Info("Items consumer listening for events...")
}

func NewConsumer(service ConsumerService, ch *amqp.Channel) *Consumer {
	return &Consumer{
		service: service,
		channel: ch,
	}
}

func (c *Consumer) consumeItemsExtracted() {
	msgs, err := c.channel.Consume(
		commonconstants.ItemsGameItemsExtractedQueue,
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
		var itemsExtracted pb.ItemsExtractedEvent
		slog.Debug("Raw message received",
			"body_length", len(msg.Body),
			"content_type", msg.ContentType,
			"body_preview", string(msg.Body[:min(100, len(msg.Body))]),
		)

		if err := proto.Unmarshal(msg.Body, &itemsExtracted); err != nil {
			slog.Error("Failed to parse items extracted event", "error", err)
			msg.Nack(false, false) // dlq
			continue
		}

		slog.Info("after itemsExtractedEvent was emitted, consumed and proto unmarshalled",
			"items_extracted", itemsExtracted)

		// redis SETNX check if eventID has been processed before

		err := c.service.ProcessItemsExtracted(&itemsExtracted)

		if err != nil {
			if errors.Is(err, commonconstants.ErrTransient) {
				slog.Error("Items service could not process items extracted due to transient error. Requeuing message",
					"err", err,
				)
				// retry
				msg.Nack(false, true)
				continue
			}

			slog.Error("Items service could not process items extracted.",
				"items_extracted", itemsExtracted,
				"err", err,
			)

			msg.Nack(false, false)
			continue
		}

		msg.Ack(false)
	}
}

// Helper method to set up AMQP exchange and bindings
func SetupAMQPInfrastructure(channel *amqp.Channel) error {

	// --- Items Extracted Event ---

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

	// declare the queue
	_, err = channel.QueueDeclare(
		commonconstants.ItemsGameItemsExtractedQueue, // queue name
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
		commonconstants.ItemsGameItemsExtractedQueue, // queue name
		commonconstants.ItemsExtracted,               // routing key
		commonconstants.GameEventsExchange,           // exchange
		false,                                        // no-wait
		nil,                                          // args
	)

	if err != nil {
		return err
	}

	slog.Info("Items AMQP infrastructure setup complete",
		"exchange", commonconstants.GameEventsExchange,
		"queue", commonconstants.ItemsGameItemsExtractedQueue,
	)

	return nil
}
