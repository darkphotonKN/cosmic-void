package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type ItemIDListComponent struct {
	ItemIDs []uuid.UUID
}

func (d *ItemIDListComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeItemIDList
}

func NewItemIDListComponent(itemIDs []uuid.UUID) *ItemIDListComponent {
	return &ItemIDListComponent{
		ItemIDs: itemIDs, // ✓ use parameter
	}
}
