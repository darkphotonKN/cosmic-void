package game

import (
	"context"

	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	amqp "github.com/rabbitmq/amqp091-go"
)

type service struct {
	publishCh *amqp.Channel
}

func NewService() *service {
	return &service{}
}
func (s *service) UpdateMatchComplete(ctx context.Context) error {

	// publish as game match ended event
	err := s.publishCh.PublishWithContext(
		ctx,
		commonconstants.GameMatchEndedEvent,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        nil,
			// persist message
			DeliveryMode: amqp.Persistent,
		})
	return nil
}
