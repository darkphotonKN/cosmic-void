package stats

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type service struct {
	repo      Repository
	publishCh *amqp.Channel
}

func NewService(repo Repository, publishCh *amqp.Channel) *service {
	return &service{
		repo:      repo,
		publishCh: publishCh,
	}
}

func (s *service) GetPlayerStats(ctx context.Context, playerID uuid.UUID) (*PlayerStats, error) {
	return nil, fmt.Errorf("not implemented")
}

