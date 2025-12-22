package serializer

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
)

/**
* The state serializer struct is in charge of serializing all complex game
* state in the form of entity and components into client consumable state.
**/
type StateSerializer struct {
}

// represents the state that client receives
type ClientGameState struct {
	SessionID uuid.UUID `json:"session_id"`
	Players   []*types.PlayerState
	Items     []string // TODO: update with item entity converted into struct format
	Doors     []*types.DoorState
}

func (s *StateSerializer) Serialize(sessionID uuid.UUID, entities map[uuid.UUID]*ecs.Entity) (*ClientGameState, error) {
	state := &ClientGameState{
		SessionID: sessionID,
		Players:   make([]*types.PlayerState, 0),
		Items:     make([]string, 0),
		Doors:     make([]*types.DoorState, 0),
	}

	for entityID, entity := range entities {
		// --- Player ---
		pc, isPlayer := entity.GetComponent(ecs.ComponentTypePlayer)
		if isPlayer {
			// -- get all player components --
			playerComponent := pc.(*components.PlayerComponent)
			tc, _ := entity.GetComponent(ecs.ComponentTypeTransform)
			transformComponent := tc.(*components.TransformComponent)
			vc, _ := entity.GetComponent(ecs.ComponentTypeVelocity)
			velocityComponent := vc.(*components.VelocityComponent)

			state.Players = append(state.Players, &types.PlayerState{
				ID:       playerComponent.UserID,
				EntityID: entityID,
				Username: playerComponent.Username,
				Position: &types.Position{
					X: transformComponent.X,
					Y: transformComponent.Y,
				},
				Direction: &types.PlayerDirection{
					VX:    velocityComponent.VX,
					VY:    velocityComponent.VY,
					Speed: velocityComponent.Speed,
				},
			})

		}

		// --- Doors ---

		// --- Items ---
		// TODO: add this
	}

	return state, nil
}
