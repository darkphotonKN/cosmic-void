package systems

import (
	"math"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

type InteractionSystem struct{}

func NewInterationSystem() *InteractionSystem {
	return &InteractionSystem{}
}

func (s *InteractionSystem) Update(entities []*ecs.Entity) {
	for _, entity := range entities {
		// validation for player
		_, hasPlayer := entity.GetComponent(ecs.ComponentTypePlayer)
		transformComp, hasTransform := entity.GetComponent(ecs.ComponentTypeTransform)

		if !hasTransform || !hasPlayer {
			continue
		}

		transform := transformComp.(*components.TransformComponent)

		// --- check door within range ---

		for _, entity := range entities {
			interactableComp, hasInteractable := entity.GetComponent(ecs.ComponentTypeInteractable)
			openableComp, hasOpenable := entity.GetComponent(ecs.ComponentTypeOpenable)
			doorTransformComp, hasTransform := entity.GetComponent(ecs.ComponentTypeTransform)

			if !hasInteractable || !hasOpenable || !hasTransform {
				continue
			}

			interactable := interactableComp.(*components.InteractableComponent)
			openable := openableComp.(*components.OpenableComponent)
			doorTransform := doorTransformComp.(*components.TransformComponent)

			// area in a circle around point from main entity (player)

			// calculate range via range provided by interactable
			xDiff := math.Pow(transform.X-doorTransform.X, 2)
			yDiff := math.Pow(transform.Y-doorTransform.Y, 2)
			distanceBetween := math.Sqrt(xDiff + yDiff)

			// too far
			if distanceBetween > interactable.Range {
				continue
			}

			// object close enough to be interacted
			// flip object's openable state to open or closed
			openable.IsOpen = !openable.IsOpen
		}
	}
}
