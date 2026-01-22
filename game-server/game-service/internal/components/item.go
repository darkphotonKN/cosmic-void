package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type ItemComponent struct {
	ItemName string
	Quantity int
	ItemTool interface{}
}

func (i *ItemComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeItem
}
func NewItemComponent(itemName string, quantity int, itemTool interface{}) *ItemComponent {
	return &ItemComponent{ItemName: itemName, Quantity: quantity, ItemTool: itemTool}
}
