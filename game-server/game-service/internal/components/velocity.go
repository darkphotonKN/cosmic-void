package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type VelocityComponent struct {
	VX      float64
	VY      float64
	SpeedV0 float64
}

func (v *VelocityComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeVelocity
}

func NewVelocityComponent(VX, VY, SpeedV0 float64) *VelocityComponent {
	return &VelocityComponent{VX: VX, VY: VY, SpeedV0: SpeedV0}
}
