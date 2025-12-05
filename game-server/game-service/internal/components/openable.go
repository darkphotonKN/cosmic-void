package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

type OpenableComponent struct {
	IsOpen bool
}

func (o *OpenableComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeOpenable
}

func NewOpenableComponent(isOpen bool) *OpenableComponent {
	return &OpenableComponent{IsOpen: isOpen}
}
