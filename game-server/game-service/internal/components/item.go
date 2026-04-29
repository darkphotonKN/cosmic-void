package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
)

type ItemComponent struct {
	TemplateID uuid.UUID
	ItemType   types.ItemType // "weapon", "armor", "consumable"
	Name       string

	// Weapon stats
	AttackPower  int
	CriticalRate float64
	WeaponType   string

	// Armor stats
	DefenseRating   int
	MagicResistance int
	ArmorSlot       types.ArmorSlot

	// Consumable stats
	HealingAmount int
	ManaAmount    int
	BuffDuration  int

	// Shared
	BuyPrice    int
	SellPrice   int
	Description string

	// InstanceID is the item_instances.id this component was hydrated from.
	// nil for world items (chests / in-match drops) that have no DB row yet.
	InstanceID *uuid.UUID
}

func (i *ItemComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeItem
}

func NewItemComponent(templateID uuid.UUID, itemType types.ItemType, name string) *ItemComponent {
	return &ItemComponent{
		TemplateID: templateID,
		ItemType:   itemType,
		Name:       name,
	}
}
