package game

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type PlayerConfig struct {
	UserID        uuid.UUID
	Username      string
	X, Y          float64
	SkillName     string
	SkillLevel    int
	CurrentHealth int
	MaxHealth     int
	Strength      int
	ItemName      string
	ItemQuantity  int
	Vx, Vy        float64
	Speed         float64
}

func CreatePlayerEntity(em *ecs.EntityManager, config PlayerConfig) *ecs.Entity {
	entity := em.CreateEntity()
	entity.AddComponent(components.NewPlayerComponent(config.UserID, config.Username))

	entity.AddComponent(components.NewTransformComponent(config.X, config.Y))

	entity.AddComponent(components.NewVelocityComponent(config.Vx, config.Vy, config.Speed))

	entity.AddComponent(components.NewHealthComponent(config.CurrentHealth, config.MaxHealth))

	return entity
}
