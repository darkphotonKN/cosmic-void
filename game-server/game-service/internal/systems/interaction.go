package systems

import (
	"sync"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

// Pool for interactable entities slice
var interactableEntitiesPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate slice for typical number of interactable objects (doors, chests, etc.)
		slice := make([]*ecs.Entity, 0, 16)
		return &slice
	},
}

type InteractionSystem struct{}

func NewInterationSystem() *InteractionSystem {
	return &InteractionSystem{}
}

func (s *InteractionSystem) Update(entities []*ecs.Entity) {
	// Get interactable entities slice from pool
	interactablesPtr := interactableEntitiesPool.Get().(*[]*ecs.Entity)
	interactables := *interactablesPtr

	// Always return to pool when done
	defer func() {
		// Clear slice before returning to pool
		for i := range interactables {
			interactables[i] = nil
		}
		interactables = interactables[:0]
		*interactablesPtr = interactables
		interactableEntitiesPool.Put(interactablesPtr)
	}()

	// Step 1: Collect all interactable entities once (avoid nested loop)
	for _, entity := range entities {
		_, hasInteractable := entity.GetComponent(ecs.ComponentTypeInteractable)
		_, hasOpenable := entity.GetComponent(ecs.ComponentTypeOpenable)
		_, hasTransform := entity.GetComponent(ecs.ComponentTypeTransform)

		if hasInteractable && hasOpenable && hasTransform {
			interactables = append(interactables, entity)
		}
	}

	// Step 2: For each player, check against collected interactables
	for _, playerEntity := range entities {
		// validation for player
		_, hasPlayer := playerEntity.GetComponent(ecs.ComponentTypePlayer)
		transformComp, hasTransform := playerEntity.GetComponent(ecs.ComponentTypeTransform)

		if !hasTransform || !hasPlayer {
			continue
		}

		playerTransform := transformComp.(*components.TransformComponent)

		// Check each interactable object
		for _, interactableEntity := range interactables {
			interactableComp, _ := interactableEntity.GetComponent(ecs.ComponentTypeInteractable)
			openableComp, _ := interactableEntity.GetComponent(ecs.ComponentTypeOpenable)
			objectTransformComp, _ := interactableEntity.GetComponent(ecs.ComponentTypeTransform)

			interactable := interactableComp.(*components.InteractableComponent)
			openable := openableComp.(*components.OpenableComponent)
			objectTransform := objectTransformComp.(*components.TransformComponent)

			// Calculate distance using squared distance (avoid sqrt for performance)
			dx := playerTransform.X - objectTransform.X
			dy := playerTransform.Y - objectTransform.Y
			distanceSquared := dx*dx + dy*dy
			rangeSquared := interactable.Range * interactable.Range

			// Check if within interaction range
			if distanceSquared <= rangeSquared {
				// object close enough to be interacted
				// flip object's openable state to open or closed
				openable.IsOpen = !openable.IsOpen
			}
		}
	}
}
