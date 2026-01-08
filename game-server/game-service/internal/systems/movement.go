package systems

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

// Map boundaries (should match frontend)
const (
	MapWidth  = 1200
	MapHeight = 800
	PlayerRadius = 20 // player sprite radius for boundary collision
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

		// clamp position to map boundaries
		if transform.X < PlayerRadius {
			transform.X = PlayerRadius
		}
		if transform.X > MapWidth-PlayerRadius {
			transform.X = MapWidth - PlayerRadius
		}
		if transform.Y < PlayerRadius {
			transform.Y = PlayerRadius
		}
		if transform.Y > MapHeight-PlayerRadius {
			transform.Y = MapHeight - PlayerRadius
		}
	}
}
