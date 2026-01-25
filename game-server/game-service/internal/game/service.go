package game

import (
	"context"
	"encoding/json"
	"log/slog"

	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	commontypes "github.com/darkphotonKN/cosmic-void-server/common/types"
	amqp "github.com/rabbitmq/amqp091-go"
)

type service struct {
	publishCh commonbroker.Publisher
}

func NewService(publishCh commonbroker.Publisher) *service {
	return &service{
		publishCh: publishCh,
	}
}

func (s *service) PublishMatchComplete(ctx context.Context, data *commontypes.MatchEndState) error {
	slog.Info("Publishing game match ended.")

	// TODO: update this to protobuf with contract
	dataJSON, err := json.Marshal(data)

	err = s.publishCh.PublishWithContext(
		ctx,
		commonconstants.GameMatchEndedEvent,
		"",
		commonbroker.Message{
			ContentType:  "application/json",
			Body:         dataJSON,
			DeliveryMode: amqp.Persistent,
		})

	if err != nil {
		slog.Error("Error publishing game match end event", "error", err)
		return err
	}

	return nil
}
