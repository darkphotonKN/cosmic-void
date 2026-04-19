package outbox

import "context"

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
}

func (s *service) CreateOutbox(ctx context.Context, params OutboxParams) error {
	return s.repo.CreateOutbox(ctx, params)
}

func (s *service) GetOldestUnpublishedOutboxItem(ctx context.Context) (*OutboxEvent, error) {
	return s.repo.GetOldestUnpublishedOutboxItem(ctx)
}
