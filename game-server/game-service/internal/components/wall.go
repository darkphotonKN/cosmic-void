package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type WallComponent struct {
	HouseID       uuid.UUID
	WallID        uuid.UUID
	Width, Height float64
}

func (d *WallComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeWall
}

func NewWallComponent(houseID uuid.UUID, wallID uuid.UUID, width, height float64) *WallComponent {
	return &WallComponent{HouseID: houseID, WallID: wallID, Width: width, Height: height}
}
