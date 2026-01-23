package game

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	commontypes "github.com/darkphotonKN/cosmic-void-server/common/types"
	amqp "github.com/rabbitmq/amqp091-go"
)

type service struct {
	publishCh *amqp.Channel
}

type MessagePublisher interface {
	Publish(exchange string, key string, mandatory bool, immediate bool, msg Publishing) error
	PublishWithContext(_ context.Context, exchange, key string, mandatory, immediate bool, msg Publishing) error
}

func NewService(publishCh *amqp.Channel) *service {
	return &service{
		publishCh: publishCh,
	}
}

func (s *service) PublishMatchComplete(ctx context.Context, data *commontypes.MatchEndState) error {
	slog.Info("Publishing game match ended.")

	dataJSON, err := json.Marshal(data)

	err = s.publishCh.PublishWithContext(
		ctx,
		commonconstants.GameMatchEndedEvent,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         dataJSON,
			DeliveryMode: amqp.Persistent,
		})

	if err != nil {
		fmt.Printf("\nError when publishing game match end event: %v\n\n", err)
		return err
	}

	return nil
}
