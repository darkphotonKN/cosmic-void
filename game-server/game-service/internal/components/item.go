package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type ItemComponent struct {
	TemplateID      uuid.UUID
	ItemType        string  // "weapon", "armor", "consumable"
	Name            string

	// Weapon stats
	AttackPower     int
	CriticalRate    float64
	WeaponType      string

	// Armor stats
	DefenseRating   int
	MagicResistance int
	ArmorSlot       string

	// Consumable stats
	HealingAmount   int
	ManaAmount      int
	BuffDuration    int

	// Shared
	Durability      int
	BuyPrice        int
	SellPrice       int
	Description     string
}

func (i *ItemComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeItem
}

func NewItemComponent(templateID uuid.UUID, itemType string, name string) *ItemComponent {
	return &ItemComponent{
		TemplateID: templateID,
		ItemType:   itemType,
		Name:       name,
	}
}
