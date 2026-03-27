package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type WallComponent struct {
	WallID        uuid.UUID
	Width, Height float64
}

func (d *WallComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeWall
}

func NewWallComponent(wallID uuid.UUID, width, height float64) *WallComponent {
	return &WallComponent{WallID: wallID, Width: width, Height: height}
}
