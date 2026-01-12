package game

import (
	"context"
	"fmt"

	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	amqp "github.com/rabbitmq/amqp091-go"
)

type service struct {
	publishCh *amqp.Channel
}

func NewService() *service {
	return &service{}
}

func (s *service) PublishMatchComplete(ctx context.Context, data *types.MatchEndState) error {

	// publish as game match ended event
	err := s.publishCh.PublishWithContext(
		ctx,
		commonconstants.GameMatchEndedEvent,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         nil,
			DeliveryMode: amqp.Persistent,
		})

	if err != nil {
		fmt.Printf("\nError when publishing game match end event: %v\n\n", err)
		return err
	}

	return nil
}
