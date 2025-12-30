package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

type ContainerComponent struct {
}

func (d *ContainerComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeContainer
}

func NewContainerComponent() *ContainerComponent {
	return &ContainerComponent{}
}
