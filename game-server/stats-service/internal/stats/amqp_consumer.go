package stats

import (
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
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

func (c *Consumer) Listen() {
	go func() {
		slog.Info("Stats consumer listening for events...")
	}()
}