package broker

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

/*
The Connect function establishes a connection to your RabbitMQ server and sets up
the exchanges needed for your service communication.
*/

func Connect(user, pass, host, port string) (*amqp.Channel, func() error) {
	address := fmt.Sprintf("amqp://%s:%s@%s:%s", user, pass, host, port)

	conn, err := amqp.Dial(address)

	if err != nil {
		log.Fatal(err)
	}

	ch, err := conn.Channel()

	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	return ch, ch.Close
}

func DeclareExchange(ch *amqp.Channel, exchangeName, exchangeType string) error {
	err := ch.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		log.Fatal("Failed to declare exchange: %v", exchangeName)
		return err
	}

	log.Printf("Declared exchange: %v", exchangeName)
	return nil
}
