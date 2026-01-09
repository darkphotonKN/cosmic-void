package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type ContainerComponent struct {
	ID uuid.UUID
}

func (d *ContainerComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeContainer
}

func NewContainerComponent(id uuid.UUID) *ContainerComponent {
	return &ContainerComponent{ID: id}
}
