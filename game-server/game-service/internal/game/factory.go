package game

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

type PlayerConfig struct {
	UserID        string
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
	SpeedV0       float64
}

func CreatePlayerEntity(em *ecs.EntityManager, config PlayerConfig) *ecs.Entity {
	entity := em.CreateEntity()
	entity.AddComponent(components.NewPlayerComponent(config.UserID, config.Username))

	entity.AddComponent(components.NewItemComponent(config.ItemName, config.ItemQuantity))
	entity.AddComponent(components.NewTransformComponent(config.X, config.Y))
	entity.AddComponent(components.NewVelocityComponent(config.Vx, config.Vy, config.SpeedV0))

	entity.AddComponent(components.NewHealthComponent(config.CurrentHealth, config.MaxHealth))
	entity.AddComponent(components.NewAttackComponent(config.Strength))
	entity.AddComponent(components.NewSkillComponent(config.SkillName, config.SkillLevel))

	entity.AddComponent(components.NewStatsComponent())


	return entity
}
