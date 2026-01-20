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
	em *ecs.EntityManager
}

func NewStateSerializer(em *ecs.EntityManager) *StateSerializer {
	return &StateSerializer{em: em}
}

func (s *StateSerializer) Serialize(sessionID uuid.UUID, recipientPlayerID uuid.UUID, entities []*ecs.Entity) (*types.ClientGameState, error) {
	state := &types.ClientGameState{
		SessionID:     sessionID,
		CurrentPlayer: nil,
		OtherPlayers:  make([]*types.PlayerState, 0),
		Items:         make([]uuid.UUID, 0),
		Doors:         make([]*types.DoorState, 0),
		Containers:    make([]*types.ContainerState, 0),
	}

	for _, entity := range entities {
		// --- Player ---
		pc, isPlayer := entity.GetComponent(ecs.ComponentTypePlayer)

		if isPlayer {
			// -- get all player components --
			player := pc.(*components.PlayerComponent)
			tc, _ := entity.GetComponent(ecs.ComponentTypeTransform)
			transform := tc.(*components.TransformComponent)
			vc, _ := entity.GetComponent(ecs.ComponentTypeVelocity)
			velocity := vc.(*components.VelocityComponent)
			// get player's inventory
			inventory := []*types.ItemState{}
			itemIDListC, _ := entity.GetComponent(ecs.ComponentTypeItemIDList)
			itemIDList := itemIDListC.(*components.ItemIDListComponent)
			for _, itemID := range itemIDList.ItemIDs {
				itemEntity, exists := s.em.GetEntity(itemID)
				if exists {
					itemC, _ := itemEntity.GetComponent(ecs.ComponentTypeItem)
					item := itemC.(*components.ItemComponent)
					inventory = append(inventory, &types.ItemState{
						ItemID:   itemID,
						EntityID: itemEntity.ID,
						Name:     item.ItemName,
						Quantity: item.Quantity,
					})
				}
			}
			playerState := &types.PlayerState{
				ID:       player.UserID,
				EntityID: entity.ID,
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
				Inventory: inventory,
			}

			// Check if this is the recipient player
			if player.UserID == recipientPlayerID {
				state.CurrentPlayer = playerState
			} else {
				state.OtherPlayers = append(state.OtherPlayers, playerState)
			}
		}

		// --- Interactables ---

		// -- Doors --

		// -- Containers --
		_, isContainer := entity.GetComponent(ecs.ComponentTypeContainer)
		if isContainer {
			tc, _ := entity.GetComponent(ecs.ComponentTypeTransform)
			transform := tc.(*components.TransformComponent)

			isOpen := false
			openableC, hasOpenable := entity.GetComponent(ecs.ComponentTypeOpenable)
			if hasOpenable {
				openable := openableC.(*components.OpenableComponent)
				isOpen = openable.IsOpen
			}

			items := make([]*types.ItemState, 0)
			itemIDListComp, hasItemIDList := entity.GetComponent(ecs.ComponentTypeItemIDList)
			if hasItemIDList {
				itemIDList := itemIDListComp.(*components.ItemIDListComponent)
				for _, itemID := range itemIDList.ItemIDs {
					itemEntity, exists := s.em.GetEntity(itemID)
					if exists {
						itemComp, hasItem := itemEntity.GetComponent(ecs.ComponentTypeItem)
						if hasItem {
							item := itemComp.(*components.ItemComponent)
							itemState := &types.ItemState{
								ItemID:   itemID,
								EntityID: itemEntity.ID,
								Name:     item.ItemName,
								Quantity: item.Quantity,
							}
							items = append(items, itemState)
						}
					}
				}
			}

			containerState := &types.ContainerState{
				EntityID: entity.ID,
				Position: &types.Position{
					X: transform.X,
					Y: transform.Y,
				},
				IsOpen: isOpen,
				Items:  items,
			}
			state.Containers = append(state.Containers, containerState)
		}
		// --- Items ---
		// TODO: add this after item entity is added
	}

	return state, nil
}
