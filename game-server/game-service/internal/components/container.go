package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type ContainerComponent struct {
	ContainerID uuid.UUID
}

func (d *ContainerComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeContainer
}

func NewContainerComponent(containerID uuid.UUID) *ContainerComponent {
	return &ContainerComponent{ContainerID: containerID}
}
