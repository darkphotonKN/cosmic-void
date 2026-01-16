package game

import (
	"context"
	"fmt"
	"log/slog"

	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	commontypes "github.com/darkphotonKN/cosmic-void-server/common/types"
	amqp "github.com/rabbitmq/amqp091-go"
)

type service struct {
	publishCh *amqp.Channel
}

func NewService(publishCh *amqp.Channel) *service {
	return &service{
		publishCh: publishCh,
	}
}

func (s *service) PublishMatchComplete(ctx context.Context, data *commontypes.MatchEndState) error {
	slog.Info("Publishing game match ended.")

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
