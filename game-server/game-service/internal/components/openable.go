package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

type OpenableComponent struct {
	isOpen bool
}

func (p *OpenableComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypePlayer
}

func NewOpenableComponent(isOpen bool) *OpenableComponent {
	return &OpenableComponent{isOpen: isOpen}
}
