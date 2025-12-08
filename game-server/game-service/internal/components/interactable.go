package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type InteractableComponent struct {
	// determines how far away something is interactable
	Range float64
}

func (i *InteractableComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypePlayer
}

func NewInteractableComponent(interactableRange float64) *InteractableComponent {
	return &InteractableComponent{Range: interactableRange}
}
