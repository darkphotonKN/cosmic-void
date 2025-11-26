package game

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

func CreatePlayerEntity(em *ecs.EntityManager, userID, username string, x, y float64) *ecs.Entity {
	entity := em.CreateEntity()
	entity.AddComponent(components.NewTransformComponent(x, y))
	entity.AddComponent(components.NewPlayerComponent(userID, username))
	return entity
}
