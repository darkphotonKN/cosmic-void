package broker

import "context"

type Queue struct {
	Name      string
	Messages  int
	Consumers int
}

/**
* Internal custom version of Message, the unit used to hold pblished messages through
* message brokers.
**/
type Message struct {
	ContentType   string
	Body          []byte
	DeliveryMode  uint8
	CorrelationId string
	Headers       map[string]interface{}
}

const (
	Persistent uint8 = 2
	Transient  uint8 = 1
)

/**
* Acknowledger provides an abstraction for acknowledging or rejecting delivered
* messages, decoupled from specific message broker implementations.
**/
type Acknowledger interface {
	Ack(multiple bool) error
	Nack(multiple, requeue bool) error
}

/**
* Delivery is the abstracted unit for consumed messages from a message broker,
* symmetric to Message which is used for publishing.
**/
type Delivery struct {
	Body          []byte
	RoutingKey    string
	CorrelationId string
	ReplyTo       string
	Headers       map[string]interface{}
	Acknowledger  Acknowledger
}

/**
* The abstracted interface to provide decoupling between services using the MB to
* publish messages in order to easily swap them out for testing.
**/
type Publisher interface {
	PublishWithContext(_ context.Context, exchange, key string, msg Message) error
	ExchangeDeclare(name, kind string) error
	QueueDeclare(name string) (Queue, error)
	QueueBind(queue, key, exchange string) error
	Consume(queue string, autoAck bool) (<-chan Delivery, error)
}
