package game

import (
	"context"
	"log/slog"

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
	Name     string
	ItemTool grpcitems.ItemsClient
}

type PriceConfig struct {
	BaseBuyPrice  int32
	BaseSellPrice int32
}

func CreateItemEntity(em *ecs.EntityManager, itemconfig ItemConfig, priceconfig PriceConfig) *ecs.Entity {
	entity := em.CreateEntity()

	// Create item component with default values
	itemComponent := components.NewItemComponent(itemconfig.Name, itemconfig.ItemTool)

	// Fetch and cache item details once at creation time
	ctx := context.Background()
	if err := populateItemDetails(ctx, itemComponent); err != nil {
		slog.Error("Failed to populate item details at creation", "itemName", itemconfig.Name, "error", err)
	}

	entity.AddComponent(itemComponent)
	entity.AddComponent(components.NewPriceComponent(priceconfig.BaseBuyPrice, priceconfig.BaseSellPrice))

	return entity
}

// populateItemDetails fetches item details from item-service and caches them in ItemComponent
func populateItemDetails(ctx context.Context, item *components.ItemComponent) error {
	// Try to find as weapon
	weaponResponse, err := item.ItemTool.ListWeaponsWithTemplate(ctx)
	if err == nil {
		for _, weapon := range weaponResponse.Weapons {
			if weapon.ItemName == item.ItemName {
				item.AttackPower = weapon.AttackPower
				item.Durability = weapon.Durability
				item.CriticalRate = weapon.CriticalRate
				item.WeaponType = weapon.WeaponType
				item.Description = weapon.Description
				return nil
			}
		}
	}

	// Try to find as armor
	armorResponse, err := item.ItemTool.ListArmorsWithTemplate(ctx)
	if err == nil {
		for _, armor := range armorResponse.Armors {
			if armor.ItemName == item.ItemName {
				item.DefenseRating = armor.DefenseRating
				item.Durability = armor.Durability
				item.ArmorSlot = armor.ArmorSlot
				item.Description = armor.Description
				return nil
			}
		}
	}

	// Try to find as consumable
	consumableResponse, err := item.ItemTool.ListConsumablesWithTemplate(ctx)
	if err == nil {
		for _, consumable := range consumableResponse.Consumables {
			if consumable.ItemName == item.ItemName {
				item.HealingAmount = consumable.HealingAmount
				item.ManaAmount = consumable.ManaAmount
				item.Description = consumable.Description
				return nil
			}
		}
	}

	return nil
}
