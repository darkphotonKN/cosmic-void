package serializer

import (
	"context"
	"sync"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	pb "github.com/darkphotonKN/cosmic-void-server/game-service/proto/pb"
	"github.com/google/uuid"
)

/**
* The state serializer struct is in charge of serializing all complex game
* state in the form of entity and components into client consumable state.
**/
type StateSerializer struct {
	em             *ecs.EntityManager
	protoStatePool *sync.Pool
}

func NewStateSerializer(em *ecs.EntityManager) *StateSerializer {
	return &StateSerializer{
		em: em,
		protoStatePool: &sync.Pool{
			New: func() interface{} {
				return &pb.ClientGameState{
					OtherPlayers: make([]*pb.PlayerState, 0, 10),
					Items:        make([][]byte, 0, 20),
					Doors:        make([]*pb.DoorState, 0, 5),
					Containers:   make([]*pb.ContainerState, 0, 5),
				}
			},
		},
	}
}

// GetProtoState retrieves a pooled protobuf state object
func (s *StateSerializer) GetProtoState() *pb.ClientGameState {
	state := s.protoStatePool.Get().(*pb.ClientGameState)
	state.Reset() // Protobuf built-in method to clear all fields
	return state
}

// PutProtoState returns a protobuf state object to the pool
func (s *StateSerializer) PutProtoState(state *pb.ClientGameState) {
	s.protoStatePool.Put(state)
}

// SerializeToProto serializes game entities into protobuf format
// This is similar to Serialize() but outputs to protobuf instead of JSON types
func (s *StateSerializer) SerializeToProto(ctx context.Context, sessionID uuid.UUID, recipientPlayerID uuid.UUID, entities []*ecs.Entity, state *pb.ClientGameState) error {
	// Set session ID (UUID as bytes)
	state.SessionId = sessionID[:]

	for _, entity := range entities {
		// --- Player ---
		pc, isPlayer := entity.GetComponent(ecs.ComponentTypePlayer)

		if isPlayer {
			// Get all player components
			player := pc.(*components.PlayerComponent)
			tc, _ := entity.GetComponent(ecs.ComponentTypeTransform)
			transform := tc.(*components.TransformComponent)
			vc, _ := entity.GetComponent(ecs.ComponentTypeVelocity)
			velocity := vc.(*components.VelocityComponent)

			// Get player's inventory
			inventory := make([]*pb.ItemState, 0)
			itemIDListC, _ := entity.GetComponent(ecs.ComponentTypeItemIDList)
			itemIDList := itemIDListC.(*components.ItemIDListComponent)
			for _, itemID := range itemIDList.ItemIDs {
				itemEntity, exists := s.em.GetEntity(itemID)
				if exists {
					itemC, _ := itemEntity.GetComponent(ecs.ComponentTypeItem)
					item := itemC.(*components.ItemComponent)

					itemState := &pb.ItemState{
						ItemId:   itemID[:],
						EntityId: itemEntity.ID[:],
						Name:     item.ItemName,
						Quantity: 1,
					}

					populateProtoItemDetails(item, itemState)
					inventory = append(inventory, itemState)
				}
			}

			playerState := &pb.PlayerState{
				Id:       player.MemberID[:],
				EntityId: entity.ID[:],
				Username: player.Username,
				Position: &pb.Position{
					X: transform.X,
					Y: transform.Y,
				},
				Direction: &pb.PlayerDirection{
					Vx:    velocity.VX,
					Vy:    velocity.VY,
					Speed: velocity.Speed,
				},
				Inventory: inventory,
			}

			// Check if this is the recipient player
			if player.MemberID == recipientPlayerID {
				state.CurrentPlayer = playerState
			} else {
				state.OtherPlayers = append(state.OtherPlayers, playerState)
			}
		}

		// --- Containers ---
		containerComp, isContainer := entity.GetComponent(ecs.ComponentTypeContainer)
		if isContainer {
			container := containerComp.(*components.ContainerComponent)
			tc, _ := entity.GetComponent(ecs.ComponentTypeTransform)
			transform := tc.(*components.TransformComponent)

			isOpen := false
			openableC, hasOpenable := entity.GetComponent(ecs.ComponentTypeOpenable)
			if hasOpenable {
				openable := openableC.(*components.OpenableComponent)
				isOpen = openable.IsOpen
			}

			items := make([]*pb.ItemState, 0)
			itemIDListComp, hasItemIDList := entity.GetComponent(ecs.ComponentTypeItemIDList)
			if hasItemIDList {
				itemIDList := itemIDListComp.(*components.ItemIDListComponent)
				for _, itemID := range itemIDList.ItemIDs {
					itemEntity, exists := s.em.GetEntity(itemID)
					if exists {
						itemComp, hasItem := itemEntity.GetComponent(ecs.ComponentTypeItem)
						if hasItem {
							item := itemComp.(*components.ItemComponent)

							itemState := &pb.ItemState{
								ItemId:   itemID[:],
								EntityId: itemEntity.ID[:],
								Name:     item.ItemName,
								Quantity: 1,
							}

							populateProtoItemDetails(item, itemState)
							items = append(items, itemState)
						}
					}
				}
			}

			containerState := &pb.ContainerState{
				ContainerId: container.ContainerID[:],
				EntityId:    entity.ID[:],
				Position: &pb.Position{
					X: transform.X,
					Y: transform.Y,
				},
				IsOpen: isOpen,
				Items:  items,
			}
			state.Containers = append(state.Containers, containerState)
		}
	}

	return nil
}

// populateProtoItemDetails reads cached item details from ItemComponent
// No gRPC calls needed - all data is cached at item creation time
func populateProtoItemDetails(item *components.ItemComponent, itemState *pb.ItemState) {
	// Simply copy cached values from ItemComponent to protobuf ItemState
	itemState.Quantity = item.Quantity
	itemState.AttackPower = item.AttackPower
	itemState.Durability = item.Durability
	itemState.CriticalRate = item.CriticalRate
	itemState.WeaponType = item.WeaponType
	itemState.DefenseRating = item.DefenseRating
	itemState.ArmorSlot = item.ArmorSlot
	itemState.HealingAmount = item.HealingAmount
	itemState.ManaAmount = item.ManaAmount
	itemState.Description = item.Description
}
