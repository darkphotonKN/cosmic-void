package game

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
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
	ItemName      string
	ItemQuantity  int
	Vx, Vy        float64
}

func CreatePlayerEntity(em *ecs.EntityManager, config PlayerConfig) *ecs.Entity {
	entity := em.CreateEntity()
	entity.AddComponent(components.NewPlayerComponent(config.UserID, config.Username))

	entity.AddComponent(components.NewTransformComponent(config.X, config.Y))

	entity.AddComponent(components.NewVelocityComponent(config.Vx, config.Vy, constants.DefaultSpeed))

	entity.AddComponent(components.NewHealthComponent(config.CurrentHealth, config.MaxHealth))
	entity.AddComponent(components.NewSkillComponent(config.SkillName, config.SkillLevel))

	entity.AddComponent(components.NewStatsComponent())

	return entity
}

type DoorConfig struct {
	X, Y float64
}

func CreateDoorEntity(em *ecs.EntityManager, config DoorConfig) *ecs.Entity {
	entity := em.CreateEntity()
	entity.AddComponent(components.NewDoorComponent())
	entity.AddComponent(components.NewTransformComponent(config.X, config.Y))
	entity.AddComponent(components.NewOpenableComponent(false)) // default false

	return entity
}
