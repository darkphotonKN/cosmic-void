package components

import (
	grpcitems "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/items"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

type ItemComponent struct {
	ItemName string
	ItemTool grpcitems.ItemsClient

	// Cached item details (populated once at creation)
	Quantity      int32
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

func (i *ItemComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeItem
}
func NewItemComponent(itemName string, itemTool grpcitems.ItemsClient) *ItemComponent {
	return &ItemComponent{
		ItemName: itemName,
		ItemTool: itemTool,
		Quantity: 1, // Default quantity
	}
}
