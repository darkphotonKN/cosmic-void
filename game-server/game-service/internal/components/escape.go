package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

type EscapeDoorComponent struct {
}

func (e *EscapeDoorComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeEscapeDoor
}

func NewEscapeDoorComponent() *EscapeDoorComponent {
	return &EscapeDoorComponent{}
}
