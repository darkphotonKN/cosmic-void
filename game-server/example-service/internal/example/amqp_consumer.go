package example

import (
	"encoding/json"
	"fmt"
	"log"

	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	amqp "github.com/rabbitmq/amqp091-go"
)

type consumer struct {
	service   Service
	publishCh *amqp.Channel
}

func NewConsumer(service Service, ch *amqp.Channel) *consumer {
	return &consumer{service: service, publishCh: ch}
}

func (c *consumer) Listen() {
	go c.exampleCreatedEventListener()

	fmt.Println("Notification consumer started - listening for create example events.")
}

func (c *consumer) exampleCreatedEventListener() {
	queueName := fmt.Sprintf("example.%s", commonconstants.ExampleCreatedEvent)

	// declare our unique queue that listens and waits for ExampleCreatedEvent to be published from example service
	queue, err := c.publishCh.QueueDeclare(queueName, true, false, false, false, nil)

	if err != nil {
		log.Fatal(err)
	}

	// bind to the exchange that will publish ExampleCreateEvent events
	err = c.publishCh.QueueBind(
		queue.Name,
		"",
		commonconstants.ExampleCreatedEvent,
		false,
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	// consume messages, delivers messages from the queue
	msgs, err := c.publishCh.Consume(queue.Name, "", true, false, false, false, nil)

	// start a goroutine to listen for events
	go func() {
		for msg := range msgs {
			var createdExample *CreateExampleEvent

			err := json.Unmarshal(msg.Body, &createdExample)
			if err != nil {
				fmt.Printf("Error when unmarshalling exampl event created body: %s\n", err.Error())
			}

			fmt.Printf("\nsuccessfully received event message: %+v\n\n", createdExample)
		}
	}()
}
