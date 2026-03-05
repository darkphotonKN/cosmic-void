package game

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	grpcitems "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/items"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

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
}

func CreatePlayerEntity(em *ecs.EntityManager, config PlayerConfig) *ecs.Entity {
	entity := em.CreateEntity()
	entity.AddComponent(components.NewPlayerComponent(config.MemberID, config.Username))

	entity.AddComponent(components.NewTransformComponent(config.X, config.Y))

	entity.AddComponent(components.NewVelocityComponent(config.Vx, config.Vy, constants.DefaultSpeed))

	entity.AddComponent(components.NewHealthComponent(config.CurrentHealth, config.MaxHealth))
	entity.AddComponent(components.NewSkillComponent(config.SkillName, config.SkillLevel))

	entity.AddComponent(components.NewItemIDListComponent(config.ItemIDList))

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

type ContainerConfig struct {
	X, Y float64
}

func CreateContainerEntity(em *ecs.EntityManager, containerconfig ContainerConfig, itemIDList []uuid.UUID) *ecs.Entity {
	entity := em.CreateEntity()
	containerID := uuid.New()
	entity.AddComponent(components.NewContainerComponent(containerID))
	entity.AddComponent(components.NewTransformComponent(containerconfig.X, containerconfig.Y))
	entity.AddComponent(components.NewOpenableComponent(false)) // default false
	entity.AddComponent(components.NewItemIDListComponent(itemIDList))

	return entity
}

type ItemConfig struct {
	Name          string
	ItemTool      grpcitems.ItemsClient
	AttackPower   int32
	Durability    int32
	CriticalRate  float32
	WeaponType    string
	DefenseRating int32
	ArmorSlot     string
	HealingAmount int32
	ManaAmount    int32
	Description   string
}

type PriceConfig struct {
	BaseBuyPrice  int32
	BaseSellPrice int32
}

func CreateItemEntity(em *ecs.EntityManager, itemconfig ItemConfig, priceconfig PriceConfig) *ecs.Entity {
	entity := em.CreateEntity()
	itemComp := components.NewItemComponent(itemconfig.Name, itemconfig.ItemTool)
	itemComp.AttackPower = itemconfig.AttackPower
	itemComp.Durability = itemconfig.Durability
	itemComp.CriticalRate = itemconfig.CriticalRate
	itemComp.WeaponType = itemconfig.WeaponType
	itemComp.DefenseRating = itemconfig.DefenseRating
	itemComp.ArmorSlot = itemconfig.ArmorSlot
	itemComp.HealingAmount = itemconfig.HealingAmount
	itemComp.ManaAmount = itemconfig.ManaAmount
	itemComp.Description = itemconfig.Description
	entity.AddComponent(itemComp)
	entity.AddComponent(components.NewPriceComponent(priceconfig.BaseBuyPrice, priceconfig.BaseSellPrice))

	return entity
}
