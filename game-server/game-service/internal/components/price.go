package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type PriceComponent struct {
	BaseBuyPrice  int32
	BaseSellPrice int32
}

func (t *PriceComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypePrice
}

func NewPriceComponent(BaseBuyPrice, BaseSellPrice int32) *PriceComponent {
	return &PriceComponent{BaseBuyPrice: BaseBuyPrice, BaseSellPrice: BaseSellPrice}
}
