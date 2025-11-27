package systems

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

type MovementSystem struct{}

func NewMovementSystem() *MovementSystem {
	return &MovementSystem{}
}

// NOTE: this runs every game tick
func (s *MovementSystem) Update(deltaTime float64, entities []*ecs.Entity) {
	for _, entity := range entities {

		transformComp, hasTransform := entity.GetComponent(ecs.ComponentTypeTransform)
		velocityComp, hasVelocity := entity.GetComponent(ecs.ComponentTypeVelocity)

		if !hasTransform || !hasVelocity {
			continue
		}

		// type assertion
		transform := transformComp.(*components.TransformComponent)
		velocity := velocityComp.(*components.VelocityComponent)

		// update position based on velocity
		transform.X += velocity.VX * velocity.Speed * deltaTime
		transform.Y += velocity.VY * velocity.Speed * deltaTime
	}
}
