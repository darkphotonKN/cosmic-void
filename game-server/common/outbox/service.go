package outbox

import (
	"context"

	"github.com/google/uuid"
)

type service struct {
	repo Repository
}

func NewService(repo Repository) *service {
	return &service{
		repo: repo,
	}
}

type Repository interface {
	CreateOutbox(ctx context.Context, params OutboxParams) error
	GetOldestUnpublishedOutboxItem(ctx context.Context) (*OutboxEvent, error)
	UpdateOutboxToPublished(ctx context.Context, id uuid.UUID) error
}

func (s *service) CreateOutbox(ctx context.Context, params OutboxParams) error {
	return s.repo.CreateOutbox(ctx, params)
}

func (s *service) UpdateOutboxToPublished(ctx context.Context, id uuid.UUID) error {
	return s.repo.UpdateOutboxToPublished(ctx, id)
}

func (s *service) GetOldestUnpublishedOutboxItem(ctx context.Context) (*OutboxEvent, error) {
	return s.repo.GetOldestUnpublishedOutboxItem(ctx)
}
