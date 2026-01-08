package systems

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
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

		// clamp position to map boundaries
		if transform.X < constants.PlayerRadius {
			transform.X = constants.PlayerRadius
		}
		if transform.X > constants.MapWidth-constants.PlayerRadius {
			transform.X = constants.MapWidth - constants.PlayerRadius
		}
		if transform.Y < constants.PlayerRadius {
			transform.Y = constants.PlayerRadius
		}
		if transform.Y > constants.MapHeight-constants.PlayerRadius {
			transform.Y = constants.MapHeight - constants.PlayerRadius
		}
	}
}
