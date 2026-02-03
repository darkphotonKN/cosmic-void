package notification

import (
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
)

type Service struct {
	repo      *Repository
	publishCh commonbroker.Publisher
}

func NewService(repo *Repository, publishCh commonbroker.Publisher) *Service {
	return &Service{
		repo:      repo,
		publishCh: publishCh,
	}
}
