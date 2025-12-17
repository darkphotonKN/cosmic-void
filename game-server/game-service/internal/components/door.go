package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type DoorComponent struct {
}

func (d *DoorComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeDoor
}

func NewDoorComponent() *DoorComponent {
	return &DoorComponent{}
}
