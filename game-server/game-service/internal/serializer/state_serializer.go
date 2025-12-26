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

func (s *StateSerializer) Serialize(sessionID uuid.UUID, entities map[uuid.UUID]*ecs.Entity) (*types.ClientGameState, error) {
	state := &types.ClientGameState{
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
			player := pc.(*components.PlayerComponent)
			tc, _ := entity.GetComponent(ecs.ComponentTypeTransform)
			transform := tc.(*components.TransformComponent)
			vc, _ := entity.GetComponent(ecs.ComponentTypeVelocity)
			velocity := vc.(*components.VelocityComponent)

			state.Players = append(state.Players, &types.PlayerState{
				ID:       player.UserID,
				EntityID: entityID,
				Username: player.Username,
				Position: &types.Position{
					X: transform.X,
					Y: transform.Y,
				},
				Direction: &types.PlayerDirection{
					VX:    velocity.VX,
					VY:    velocity.VY,
					Speed: velocity.Speed,
				},
			})
		}

		// --- Doors ---

		// --- Items ---
		// TODO: add this
	}

	return state, nil
}
