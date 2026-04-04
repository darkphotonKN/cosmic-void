package game

import (
	commonconstants "github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
)

type MatchConfig struct {
	players []*ecs.Entity
}

func CreateMatchProgressEntity(em *ecs.EntityManager) *ecs.Entity {
	entity := em.CreateEntity()
	entity.AddComponent(components.NewMatchProgressComponent(commonconstants.DefautMaxSessionPlayers))

	return entity
}

type PlayerConfig struct {
	MemberID      uuid.UUID
	Username      string
	X, Y          float64
	SkillName     string
	SkillLevel    int
	CurrentHealth int
	MaxHealth     int
	ItemName      string
	ItemQuantity  int
	Vx, Vy        float64
	ItemIDList    []uuid.UUID
	AttackActive  bool
	HasHit        bool
	Escape        bool
}

func CreatePlayerEntity(em *ecs.EntityManager, config PlayerConfig) *ecs.Entity {
	entity := em.CreateEntity()

	entity.AddComponent(components.NewPlayerComponent(config.MemberID, config.Username, config.HasHit, config.AttackActive, config.Escape))

	entity.AddComponent(components.NewTransformComponent(config.X, config.Y))

	entity.AddComponent(components.NewVelocityComponent(config.Vx, config.Vy, commonconstants.DefaultSpeed))

	entity.AddComponent(components.NewHealthComponent(config.CurrentHealth, config.MaxHealth))
	entity.AddComponent(components.NewSkillComponent(config.SkillName, config.SkillLevel))

	entity.AddComponent(components.NewItemIDListComponent(config.ItemIDList))

	entity.AddComponent(components.NewStatsComponent())

	return entity
}

type DoorConfig struct {
	X, Y, Width, Height float64
}

func CreateDoorEntity(em *ecs.EntityManager, config DoorConfig) *ecs.Entity {
	entity := em.CreateEntity()
	entity.AddComponent(components.NewDoorComponent(config.Width, config.Height))
	entity.AddComponent(components.NewTransformComponent(config.X, config.Y))
	entity.AddComponent(components.NewOpenableComponent(false)) // default closed

	return entity
}

type ContainerConfig struct {
	X, Y float64
}

func CreateContainerEntity(em *ecs.EntityManager, config ContainerConfig, itemIDList []uuid.UUID) *ecs.Entity {
	entity := em.CreateEntity()
	containerID := uuid.New()
	entity.AddComponent(components.NewContainerComponent(containerID))
	entity.AddComponent(components.NewTransformComponent(config.X, config.Y))
	entity.AddComponent(components.NewOpenableComponent(false)) // default false
	entity.AddComponent(components.NewItemIDListComponent(itemIDList))

	return entity
}

type WallConfig struct {
	X, Y, Width, Height float64
}

func CreateWallEntity(em *ecs.EntityManager, wallConfig WallConfig) *ecs.Entity {
	entity := em.CreateEntity()
	wallID := uuid.New()
	entity.AddComponent(components.NewWallComponent(wallID, wallConfig.Width, wallConfig.Height))
	entity.AddComponent(components.NewTransformComponent(wallConfig.X, wallConfig.Y))
	return entity
}

type ItemConfig struct {
	TemplateID      uuid.UUID
	ItemType        types.ItemType
	Name            string
	AttackPower     int
	CriticalRate    float64
	WeaponType      string
	DefenseRating   int
	MagicResistance int
	ArmorSlot       string
	HealingAmount   int
	ManaAmount      int
	BuffDuration    int
	BuyPrice        int
	SellPrice       int
	Description     string
}

func CreateItemEntity(em *ecs.EntityManager, itemconfig types.ItemConfig) *ecs.Entity {
	entity := em.CreateEntity()
	itemComp := components.NewItemComponent(itemconfig.TemplateID, itemconfig.ItemType, itemconfig.Name)

	itemComp.AttackPower = itemconfig.AttackPower
	itemComp.CriticalRate = itemconfig.CriticalRate
	itemComp.WeaponType = itemconfig.WeaponType
	itemComp.DefenseRating = itemconfig.DefenseRating
	itemComp.MagicResistance = itemconfig.MagicResistance
	itemComp.ArmorSlot = itemconfig.ArmorSlot
	itemComp.HealingAmount = itemconfig.HealingAmount
	itemComp.ManaAmount = itemconfig.ManaAmount
	itemComp.BuffDuration = itemconfig.BuffDuration
	itemComp.BuyPrice = itemconfig.BuyPrice
	itemComp.SellPrice = itemconfig.SellPrice
	itemComp.Description = itemconfig.Description

	entity.AddComponent(itemComp)

	return entity
}

type EscapeConfig struct {
	X, Y float64
}

func CreateEscapeDoorEntity(em *ecs.EntityManager, config EscapeConfig) *ecs.Entity {
	entity := em.CreateEntity()
	entity.AddComponent(components.NewEscapeDoorComponent())
	entity.AddComponent(components.NewLockableComponent(true))
	entity.AddComponent(components.NewTransformComponent(config.X, config.Y))
	entity.AddComponent(components.NewOpenableComponent(false))
	entity.AddComponent(components.NewInteractableComponent(commonconstants.DefaultInteractableRange))
	return entity
}

type SwitchConfig struct {
	X, Y     float64
	SwitchID int
}

func CreateSwitchEntity(em *ecs.EntityManager, config SwitchConfig) *ecs.Entity {
	entity := em.CreateEntity()
	entity.AddComponent(components.NewSwitchComponent(config.SwitchID))
	entity.AddComponent(components.NewTransformComponent(config.X, config.Y))
	entity.AddComponent(components.NewInteractableComponent(commonconstants.DefaultInteractableRange))
	return entity

}
