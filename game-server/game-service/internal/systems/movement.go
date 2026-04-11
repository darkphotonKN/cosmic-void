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

	wallEntities := make(map[uuid.UUID]*ecs.Entity, 0)
	doorEntities := make(map[uuid.UUID]*ecs.Entity, 0)
	for _, entity := range entities {

		entityByID[entity.ID] = entity
		transformComp, hasTransform := entity.GetComponent(ecs.ComponentTypeTransform)
		_, hasVelocity := entity.GetComponent(ecs.ComponentTypeVelocity)

		_, hasWallComp := entity.GetComponent(ecs.ComponentTypeWall)
		if hasWallComp {
			wallEntities[entity.ID] = entity
		}

		_, hasDoorComp := entity.GetComponent(ecs.ComponentTypeDoor)
		if hasDoorComp {
			doorEntities[entity.ID] = entity
		}

		if hasTransform && hasVelocity {
			// type assertion
			transform := transformComp.(*components.TransformComponent)

			entityCellX := int(transform.X / (2 * constants.PlayerRadius))
			entityCellY := int(transform.Y / (2 * constants.PlayerRadius))

			key := entityCellX<<8 | entityCellY

			entitiesMap[key] = entity
		}

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
		dx := velocity.VX * velocity.Speed * deltaTime
		dy := velocity.VY * velocity.Speed * deltaTime

		// split-axis swept collision, X first
		minTx := 1.0

		// walls
		for _, wallEntity := range wallEntities {
			wallC, _ := wallEntity.GetComponent(ecs.ComponentTypeWall)
			wallTransformComp, _ := wallEntity.GetComponent(ecs.ComponentTypeTransform)
			wallTransform := wallTransformComp.(*components.TransformComponent)
			wall := wallC.(*components.WallComponent)

			tX := SweptX(transform.X, transform.Y, wallTransform.X, wallTransform.Y, wall.Width, wall.Height, dx)
			if tX < minTx {
				minTx = tX
			}
		}

		// doors (skip if open)
		for _, doorEntity := range doorEntities {
			openableC, hasOpenable := doorEntity.GetComponent(ecs.ComponentTypeOpenable)
			if hasOpenable {
				openable := openableC.(*components.OpenableComponent)
				if openable.IsOpen {
					continue
				}
			}

			doorC, _ := doorEntity.GetComponent(ecs.ComponentTypeDoor)
			doorTransformComp, _ := doorEntity.GetComponent(ecs.ComponentTypeTransform)
			doorTransform := doorTransformComp.(*components.TransformComponent)
			door := doorC.(*components.DoorComponent)

			tX := SweptX(transform.X, transform.Y, doorTransform.X, doorTransform.Y, door.Width, door.Height, dx)
			if tX < minTx {
				minTx = tX
			}
		}

		newX := transform.X + dx*minTx

		// split-axis swept collision, Y second
		minTy := 1.0

		// walls
		for _, wallEntity := range wallEntities {
			wallC, _ := wallEntity.GetComponent(ecs.ComponentTypeWall)
			wallTransformComp, _ := wallEntity.GetComponent(ecs.ComponentTypeTransform)
			wallTransform := wallTransformComp.(*components.TransformComponent)
			wall := wallC.(*components.WallComponent)

			tY := SweptY(newX, transform.Y, wallTransform.X, wallTransform.Y, wall.Width, wall.Height, dy)
			if tY < minTy {
				minTy = tY
			}
		}

		// doors, skip if open
		for _, doorEntity := range doorEntities {
			openableC, hasOpenable := doorEntity.GetComponent(ecs.ComponentTypeOpenable)
			if hasOpenable {
				openable := openableC.(*components.OpenableComponent)
				if openable.IsOpen {
					continue
				}
			}

			doorC, _ := doorEntity.GetComponent(ecs.ComponentTypeDoor)
			doorTransformComp, _ := doorEntity.GetComponent(ecs.ComponentTypeTransform)
			doorTransform := doorTransformComp.(*components.TransformComponent)
			door := doorC.(*components.DoorComponent)

			tY := SweptY(newX, transform.Y, doorTransform.X, doorTransform.Y, door.Width, door.Height, dy)
			if tY < minTy {
				minTy = tY
			}
		}

		newY := transform.Y + dy*minTy

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

		// depenetration: push player out of walls after entity collision resolve
		for _, wallEntity := range wallEntities {
			wallC, _ := wallEntity.GetComponent(ecs.ComponentTypeWall)
			wallTransformComp, _ := wallEntity.GetComponent(ecs.ComponentTypeTransform)
			wallTransform := wallTransformComp.(*components.TransformComponent)
			wall := wallC.(*components.WallComponent)

			newX, newY = depenetrate(newX, newY, wallTransform.X, wallTransform.Y, wall.Width, wall.Height)
		}
		for _, doorEntity := range doorEntities {
			openableC, hasOpenable := doorEntity.GetComponent(ecs.ComponentTypeOpenable)
			if hasOpenable {
				openable := openableC.(*components.OpenableComponent)
				if openable.IsOpen {
					continue
				}
			}
			doorC, _ := doorEntity.GetComponent(ecs.ComponentTypeDoor)
			doorTransformComp, _ := doorEntity.GetComponent(ecs.ComponentTypeTransform)
			doorTransform := doorTransformComp.(*components.TransformComponent)
			door := doorC.(*components.DoorComponent)

			newX, newY = depenetrate(newX, newY, doorTransform.X, doorTransform.Y, door.Width, door.Height)
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

func SweptX(playerX, playerY, wallX, wallY, wallW, wallH, dx float64) float64 {
	if dx == 0 {
		return 1
	}

	wallLeft := wallX
	wallRight := wallX + wallW
	wallTop := wallY
	wallBottom := wallY + wallH

	playerLeft := playerX - constants.PlayerRadius
	playerRight := playerX + constants.PlayerRadius
	playerTop := playerY - constants.PlayerRadius
	playerBottom := playerY + constants.PlayerRadius

	var tEnter float64

	if playerBottom <= wallTop {
		return 1
	}

	if playerTop >= wallBottom {
		return 1
	}

	// 往右走

	if dx > 0 {
		tEnter = (wallLeft - playerRight) / dx
	}
	// 往左走
	if dx < 0 {
		tEnter = (wallRight - playerLeft) / dx
	}

	if tEnter >= 1 || tEnter < 0 {
		return 1
	}
	return tEnter
}

func SweptY(playerX, playerY, wallX, wallY, wallW, wallH, dy float64) float64 {

	if dy == 0 {
		return 1
	}
	wallLeft := wallX
	wallRight := wallX + wallW
	wallTop := wallY
	wallBottom := wallY + wallH

	playerLeft := playerX - constants.PlayerRadius
	playerRight := playerX + constants.PlayerRadius
	playerTop := playerY - constants.PlayerRadius
	playerBottom := playerY + constants.PlayerRadius

	var tEnter float64
	if playerRight <= wallLeft {
		return 1
	}

	if playerLeft >= wallRight {
		return 1
	}

	// 往下走
	if dy > 0 {
		tEnter = (wallTop - playerBottom) / dy
	}
	// 往上走
	if dy < 0 {
		tEnter = (wallBottom - playerTop) / dy
	}

	if tEnter >= 1 || tEnter < 0 {
		return 1
	}
	return tEnter
}

func depenetrate(playerX, playerY, wallX, wallY, wallW, wallH float64) (float64, float64) {
	playerLeft := playerX - constants.PlayerRadius
	playerRight := playerX + constants.PlayerRadius
	playerTop := playerY - constants.PlayerRadius
	playerBottom := playerY + constants.PlayerRadius

	wallRight := wallX + wallW
	wallBottom := wallY + wallH

	// no overlap
	if playerRight <= wallX || playerLeft >= wallRight ||
		playerBottom <= wallY || playerTop >= wallBottom {
		return playerX, playerY
	}

	// overlap in each direction
	overlapLeft := playerRight - wallX
	overlapRight := wallRight - playerLeft
	overlapTop := playerBottom - wallY
	overlapBottom := wallBottom - playerTop

	// push out in the direction of smallest overlap
	minOverlap := overlapLeft
	pushX := -overlapLeft
	pushY := 0.0

	if overlapRight < minOverlap {
		minOverlap = overlapRight
		pushX = overlapRight
		pushY = 0
	}
	if overlapTop < minOverlap {
		minOverlap = overlapTop
		pushX = 0
		pushY = -overlapTop
	}
	if overlapBottom < minOverlap {
		pushX = 0
		pushY = overlapBottom
	}

	return playerX + pushX, playerY + pushY
}
