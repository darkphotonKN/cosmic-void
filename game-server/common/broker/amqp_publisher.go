package broker

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

/**
* Adapter for rabbitmq version of our publisher.
**/
type AmqpPublisher struct {
	ch *amqp.Channel
}

func NewAmqpPublisher(ch *amqp.Channel) *AmqpPublisher {
	return &AmqpPublisher{ch: ch}
}

func (p *AmqpPublisher) PublishWithContext(_ context.Context, exchange, key string, msg Message) error {

	amqpMessageParam := amqp.Publishing{
		ContentType:   msg.ContentType,
		Body:          msg.Body,
		DeliveryMode:  msg.DeliveryMode,
		CorrelationId: msg.CorrelationId,
		Headers:       msg.Headers,
	}

	return p.ch.Publish(exchange, key, false, false, amqpMessageParam)
}