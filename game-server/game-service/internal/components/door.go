package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type DoorComponent struct {
	Width  float64
	Height float64
}

func (d *DoorComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeDoor
}

func NewDoorComponent(width, height float64) *DoorComponent {
	return &DoorComponent{Width: width, Height: height}
}
