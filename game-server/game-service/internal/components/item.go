package components

import (
	"github.com/darkphotonKN/cosmic-void-server/common/constants/types"
	grpcitems "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/items"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

type ItemComponent struct {
	ItemName      string
	AttackPower   int32
	Durability    int32
	CriticalRate  float32
	WeaponType    string
	DefenseRating int32
	ArmorSlot     string
	HealingAmount int32
	ManaAmount    int32
	Description   string
	// ItemTool is kept for potential future use but no longer used during serialization.
	ItemTool grpcitems.ItemsClient
}

func (i *ItemComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeItem
}

func NewItemComponent(itemConfig types.ItemConfig, itemTool grpcitems.ItemsClient) *ItemComponent {
	return &ItemComponent{
		ItemName:      itemConfig.Name,
		AttackPower:   itemConfig.AttackPower,
		Durability:    itemConfig.Durability,
		CriticalRate:  itemConfig.CriticalRate,
		WeaponType:    itemConfig.WeaponType,
		DefenseRating: itemConfig.DefenseRating,
		ArmorSlot:     itemConfig.ArmorSlot,
		HealingAmount: itemConfig.HealingAmount,
		ManaAmount:    itemConfig.ManaAmount,
		Description:   itemConfig.Description,
		ItemTool:      itemTool,
	}
}
