package systems

import (
	"math"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type MovementSystem struct{}

func NewMovementSystem() *MovementSystem {
	return &MovementSystem{}
}

// NOTE: this runs every game tick
func (s *MovementSystem) Update(deltaTime float64, entities []*ecs.Entity) {
	// O(n) spatial hashing for collision + entity lookup
	entitiesMap := make(map[int]*ecs.Entity, 0)
	entityByID := make(map[uuid.UUID]*ecs.Entity, len(entities))
	for _, entity := range entities {

		entityByID[entity.ID] = entity
		transformComp, hasTransform := entity.GetComponent(ecs.ComponentTypeTransform)
		_, hasVelocity := entity.GetComponent(ecs.ComponentTypeVelocity)

		if !hasTransform || !hasVelocity {
			continue
		}
		// type assertion
		transform := transformComp.(*components.TransformComponent)

		entityCellX := int(transform.X / (2 * constants.PlayerRadius))
		entityCellY := int(transform.Y / (2 * constants.PlayerRadius))

		key := entityCellX<<8 | entityCellY

		entitiesMap[key] = entity
	}

	for _, entity := range entitiesMap {
		targetEntity := entity
		transformComp, _ := targetEntity.GetComponent(ecs.ComponentTypeTransform)
		velocityComp, _ := targetEntity.GetComponent(ecs.ComponentTypeVelocity)
		// type assertion
		transform := transformComp.(*components.TransformComponent)
		velocity := velocityComp.(*components.VelocityComponent)
		// calculate if there are other nearby entities in the 9-grid cells around this targetentity
		cellX := int(transform.X / (2 * constants.PlayerRadius))
		cellY := int(transform.Y / (2 * constants.PlayerRadius))

		newX := transform.X + velocity.VX*velocity.Speed*deltaTime
		newY := transform.Y + velocity.VY*velocity.Speed*deltaTime

		// check collision in 9-grid and resolve position by hashmap
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				cellKey := (cellX+i)<<8 | (cellY + j)
				if other, ok := entitiesMap[cellKey]; ok {
					if other.ID == targetEntity.ID {
						continue
					}
					resolvedX, resolvedY, collided := resolveCollision(newX, newY, other)
					if collided {
						newX = resolvedX
						newY = resolvedY
					}
				}
			}
		}

		playerC, _ := targetEntity.GetComponent(ecs.ComponentTypePlayer)
		player := playerC.(*components.PlayerComponent)

		// tick down cooldown
		if player.AttackCooldown > 0 {
			player.AttackCooldown -= deltaTime
		}

		// attack
		if player.AttackActive && player.AttackCooldown <= 0 && player.AttackTargetEntityID != uuid.Nil {
			if enemyEntity, ok := entityByID[player.AttackTargetEntityID]; ok {
				enemyTransformC, hasTransform := enemyEntity.GetComponent(ecs.ComponentTypeTransform)
				if hasTransform {
					enemyTransform := enemyTransformC.(*components.TransformComponent)
					dx := newX - enemyTransform.X
					dy := newY - enemyTransform.Y
					distance := math.Hypot(dx, dy)

					attackRange := float64(60)
					if distance <= attackRange {
						enemyHealthC, hasHealth := enemyEntity.GetComponent(ecs.ComponentTypeHealth)
						if hasHealth {
							enemyHealth := enemyHealthC.(*components.HealthComponent)
							enemyHealth.CurrentHealth -= 10
						}
					}
				}
			}
			player.HasHit = false
			player.AttackActive = false
			player.AttackTargetEntityID = uuid.Nil
			player.AttackCooldown = 0.5
		}

		// clamp position to map boundaries
		if newX < constants.PlayerRadius {
			newX = constants.PlayerRadius
		}
		if newX > constants.MapWidth-constants.PlayerRadius {
			newX = constants.MapWidth - constants.PlayerRadius
		}
		if newY < constants.PlayerRadius {
			newY = constants.PlayerRadius
		}
		if newY > constants.MapHeight-constants.PlayerRadius {
			newY = constants.MapHeight - constants.PlayerRadius
		}
		// update position based on velocity
		transform.X = newX
		transform.Y = newY

	}

}

// resolveCollision checks collision and returns resolved position
// if there's a collision, pushes position to just not overlapping
func resolveCollision(newX, newY float64, other *ecs.Entity) (float64, float64, bool) {
	otherTransformComp, hasTransform := other.GetComponent(ecs.ComponentTypeTransform)
	_, hasVelocity := other.GetComponent(ecs.ComponentTypeVelocity)

	if !hasTransform || !hasVelocity {
		return newX, newY, false
	}

	otherTransform := otherTransformComp.(*components.TransformComponent)

	dx := newX - otherTransform.X
	dy := newY - otherTransform.Y
	distance := math.Hypot(dx, dy)

	minDist := 2 * constants.PlayerRadius

	if distance < minDist && distance > 0 {
		// collision detected, push to just not overlapping position
		ratio := minDist / distance
		resolvedX := otherTransform.X + dx*ratio
		resolvedY := otherTransform.Y + dy*ratio
		return resolvedX, resolvedY, true
	}

	return newX, newY, false
}
