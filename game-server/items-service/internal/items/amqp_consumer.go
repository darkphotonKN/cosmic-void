package items

import (
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
)

type Consumer struct {
	service ConsumerService
	channel *amqp.Channel
}

type ConsumerService interface {
	CreatePlayerLoadout() error
	CreateItemInstance() error
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
