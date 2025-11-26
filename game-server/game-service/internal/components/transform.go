package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type TransformComponent struct {
	X float64
	Y float64
}

func (t *TransformComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeTransform
}

func NewTransformComponent(x, y float64) *TransformComponent {
	return &TransformComponent{X: x, Y: y}
}
