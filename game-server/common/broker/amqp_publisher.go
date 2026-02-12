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

func (p *AmqpPublisher) ExchangeDeclare(name, kind string) error {
	p.ch.ExchangeDeclare(name, kind, true, false, false, false, nil)
	return nil
}

func (p *AmqpPublisher) QueueDeclare(name string) (Queue, error) {
	queue, err := p.ch.QueueDeclare(name, true, false, false, false, nil)
	if err != nil {
		return Queue{}, err
	}
	custmQueue := Queue{
		Name:      queue.Name,
		Messages:  queue.Messages,
		Consumers: queue.Consumers,
	}
	return custmQueue, nil
}

func (p *AmqpPublisher) QueueBind(queue, key, exchange string) error {
	return p.ch.QueueBind(queue, key, exchange, false, nil)
}

// amqpAcknowledger adapts amqp.Delivery's Ack/Nack to the broker.Acknowledger interface.
type amqpAcknowledger struct {
	d amqp.Delivery
}

func (a *amqpAcknowledger) Ack(multiple bool) error {
	return a.d.Ack(multiple)
}

func (a *amqpAcknowledger) Nack(multiple, requeue bool) error {
	return a.d.Nack(multiple, requeue)
}

func (p *AmqpPublisher) Consume(queue string, autoAck bool) (<-chan Delivery, error) {
	amqpCh, err := p.ch.Consume(queue, "", autoAck, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	deliveries := make(chan Delivery)
	go func() {
		defer close(deliveries)
		for d := range amqpCh {
			deliveries <- Delivery{
				Body:          d.Body,
				RoutingKey:    d.RoutingKey,
				CorrelationId: d.CorrelationId,
				ReplyTo:       d.ReplyTo,
				Headers:       d.Headers,
				Acknowledger:  &amqpAcknowledger{d: d},
			}
		}
	}()

	return deliveries, nil
}
