package systems

import (
	"math"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

type MovementSystem struct{}

func NewMovementSystem() *MovementSystem {
	return &MovementSystem{}
}
// resolveCollision 檢查碰撞並返回解析後的位置
// 如果有碰撞，會將位置推到剛好不重疊的地方
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
		// 碰撞了，推到剛好不重疊的位置
		ratio := minDist / distance
		resolvedX := otherTransform.X + dx*ratio
		resolvedY := otherTransform.Y + dy*ratio
		return resolvedX, resolvedY, true
	}

	return newX, newY, false
}

// NOTE: this runs every game tick
func (s *MovementSystem) Update(deltaTime float64, entities []*ecs.Entity) {

	// O(n)
	// player 碰撞
	entitiesMap := make(map[int]*ecs.Entity, 0)
	for _, entity := range entities {
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
		// 計算這個targetentity的周圍的cell 9宮格內是否有其他接近的entity
		cellX := int(transform.X / (2 * constants.PlayerRadius))
		cellY := int(transform.Y / (2 * constants.PlayerRadius))

		newX := transform.X + velocity.VX*velocity.Speed*deltaTime
		newY := transform.Y + velocity.VY*velocity.Speed*deltaTime

		// 檢查 9 宮格內的碰撞，並解析位置
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
