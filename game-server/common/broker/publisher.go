package broker

import "context"

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
* The abstracted interface to provide decoupling between services using the MB to
* publish messages in order to easily swap them out for testing.
**/
type Publisher interface {
	PublishWithContext(_ context.Context, exchange, key string, msg Message) error
}
