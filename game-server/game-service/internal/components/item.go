package components

import (
	grpcitems "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/items"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

type ItemComponent struct {
	ItemName string
	ItemTool grpcitems.ItemsClient
}

func (i *ItemComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeItem
}
func NewItemComponent(itemName string, itemTool grpcitems.ItemsClient) *ItemComponent {
	return &ItemComponent{ItemName: itemName, ItemTool: itemTool}
}
