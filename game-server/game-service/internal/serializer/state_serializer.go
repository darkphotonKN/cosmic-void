package serializer

import (
	"context"
	"sync"

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
	em               *ecs.EntityManager
	backendStatePool *sync.Pool
}

func NewStateSerializer(em *ecs.EntityManager) *StateSerializer {
	return &StateSerializer{em: em, backendStatePool: &sync.Pool{
		New: func() interface{} {
			return &types.BackendGameState{
				Players:    make(map[uuid.UUID]*types.PlayerState),
				Items:      make([]uuid.UUID, 0),
				Doors:      make([]*types.DoorState, 0),
				Walls:      make([]*types.WallState, 0),
				Containers: make([]*types.ContainerState, 0),
				EscapeDoor: make([]*types.EscapeDoorState, 0),
				Switch:     make([]*types.SwitchState, 0),
			}
		},
	}}
}

func (s *StateSerializer) SerializeBackendState(ctx context.Context, sessionID uuid.UUID, entities []*ecs.Entity) (*types.BackendGameState, error) {
	backendState := s.backendStatePool.Get().(*types.BackendGameState)
	s.RestBackendStatePool(backendState)
	backendState.SessionID = sessionID

	for _, entity := range entities {
		// --- Player ---
		pc, isPlayer := entity.GetComponent(ecs.ComponentTypePlayer)

		if isPlayer {
			// -- get all player components --
			player := pc.(*components.PlayerComponent)
			if player.Escape {
				backendState.EscapedCount++
				continue
			}
			tc, _ := entity.GetComponent(ecs.ComponentTypeTransform)
			transform := tc.(*components.TransformComponent)
			vc, _ := entity.GetComponent(ecs.ComponentTypeVelocity)
			velocity := vc.(*components.VelocityComponent)
			// get player's inventory
			inventory := []*types.ItemState{}
			itemIDListC, _ := entity.GetComponent(ecs.ComponentTypeItemIDList)
			itemIDList := itemIDListC.(*components.ItemIDListComponent)

			equipmentC, ok := entity.GetComponent(ecs.ComponentTypeEquipment)
			equipmentState := &types.EquipmentState{}
			if ok {
				equipment, _ := equipmentC.(*components.EquipmentComponent)

				loadoutEntityIDs := map[string]*uuid.UUID{
					"Weapon":      equipment.WeaponSlot,
					"Head":        equipment.HeadSlot,
					"Chest":       equipment.ChestSlot,
					"Legs":        equipment.LegsSlot,
					"Gloves":      equipment.GlovesSlot,
					"Ring1":       equipment.Ring1Slot,
					"Ring2":       equipment.Ring2Slot,
					"Consumable1": equipment.Consumable1,
					"Consumable2": equipment.Consumable2,
					"Consumable3": equipment.Consumable3,
				}

				loadouts := map[string]*types.ItemState{}

				for key, loadoutID := range loadoutEntityIDs {
					if loadoutID == nil {
						continue
					}
					itemEntity, exists := s.em.GetEntity(*loadoutID)
					if !exists {
						continue
					}
					itemC, ok := itemEntity.GetComponent(ecs.ComponentTypeItem)
					if !ok {
						continue
					}
					item := itemC.(*components.ItemComponent)

					loadouts[key] = &types.ItemState{
						ItemID:        item.TemplateID,
						EntityID:      *loadoutID,
						Name:          item.Name,
						Quantity:      1, // ItemComponent 無此欄位，依需求 hardcode
						AttackPower:   int32(item.AttackPower),
						CriticalRate:  float32(item.CriticalRate),
						WeaponType:    item.WeaponType,
						DefenseRating: int32(item.DefenseRating),
						ArmorSlot:     item.ArmorSlot,
						HealingAmount: int32(item.HealingAmount),
						ManaAmount:    int32(item.ManaAmount),
						Description:   item.Description,
					}

				}
				equipmentState = &types.EquipmentState{
					Weapon:      loadouts["Weapon"],
					Head:        loadouts["Head"],
					Chest:       loadouts["Chest"],
					Gloves:      loadouts["Gloves"],
					Legs:        loadouts["Legs"],
					Ring1:       loadouts["Ring1"],
					Ring2:       loadouts["Ring2"],
					Consumable1: loadouts["Consumable1"],
					Consumable2: loadouts["Consumable2"],
					Consumable3: loadouts["Consumable3"],
				}

			}

			for _, itemID := range itemIDList.ItemIDs {
				itemEntity, exists := s.em.GetEntity(itemID)
				if exists {
					itemC, _ := itemEntity.GetComponent(ecs.ComponentTypeItem)
					item := itemC.(*components.ItemComponent)

					itemState := s.getItemState(item.TemplateID, itemID, item)

					inventory = append(inventory, itemState)
				}
			}
			playerState := &types.PlayerState{
				ID:       player.MemberID,
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
				Equipment: equipmentState,
				Escape:    player.Escape,
			}

			// Check if this is the recipient player
			backendState.Players[player.MemberID] = playerState
		}

		// --- Interactables ---

		// -- Doors --
		doorC, isDoor := entity.GetComponent(ecs.ComponentTypeDoor)
		if isDoor {
			tc, hasTransform := entity.GetComponent(ecs.ComponentTypeTransform)
			if hasTransform {
				transform := tc.(*components.TransformComponent)
				door := doorC.(*components.DoorComponent)

				isOpen := false
				openableC, hasOpenable := entity.GetComponent(ecs.ComponentTypeOpenable)
				if hasOpenable {
					openable := openableC.(*components.OpenableComponent)
					isOpen = openable.IsOpen
				}

				doorState := &types.DoorState{
					EntityID: entity.ID,
					Position: &types.Position{
						X: transform.X,
						Y: transform.Y,
					},
					Width:  door.Width,
					Height: door.Height,
					IsOpen: isOpen,
				}
				backendState.Doors = append(backendState.Doors, doorState)
			}
		}

		// -- Escape Doors --
		_, isEscapeDoor := entity.GetComponent(ecs.ComponentTypeEscapeDoor)
		if isEscapeDoor {
			tc, hasTransform := entity.GetComponent(ecs.ComponentTypeTransform)
			if !hasTransform {
				continue
			}
			transform := tc.(*components.TransformComponent)

			isOpen := false
			openableC, hasOpenable := entity.GetComponent(ecs.ComponentTypeOpenable)
			if hasOpenable {
				openable := openableC.(*components.OpenableComponent)
				isOpen = openable.IsOpen
			}

			isLocked := true
			lockableC, hasLockable := entity.GetComponent(ecs.ComponentTypeLockable)
			if hasLockable {
				lockable := lockableC.(*components.LockableComponents)
				isLocked = lockable.IsLocked
			}

			escapeDoorState := &types.EscapeDoorState{
				EntityID: entity.ID,
				Position: &types.Position{
					X: transform.X,
					Y: transform.Y,
				},
				IsOpen:   isOpen,
				IsLocked: isLocked,
			}
			backendState.EscapeDoor = append(backendState.EscapeDoor, escapeDoorState)
		}

		// -- Switches --
		switchComp, isSwitch := entity.GetComponent(ecs.ComponentTypeSwitch)
		if isSwitch {
			tc, hasTransform := entity.GetComponent(ecs.ComponentTypeTransform)
			if !hasTransform {
				continue
			}
			transform := tc.(*components.TransformComponent)
			switchComponent := switchComp.(*components.SwitchComponent)

			switchState := &types.SwitchState{
				EntityID: entity.ID,
				Position: &types.Position{
					X: transform.X,
					Y: transform.Y,
				},
				SwitchID:    switchComponent.SwitchID,
				IsActivated: switchComponent.IsActivated,
			}
			backendState.Switch = append(backendState.Switch, switchState)
		}

		// -- Containers --
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

							itemState := s.getItemState(item.TemplateID, itemID, item)

							items = append(items, itemState)
						}
					}
				}
			}

			containerState := &types.ContainerState{
				ContainerID: container.ContainerID,
				EntityID:    entity.ID,
				Position: &types.Position{
					X: transform.X,
					Y: transform.Y,
				},
				IsOpen: isOpen,
				Items:  items,
			}
			backendState.Containers = append(backendState.Containers, containerState)
		}

		// --- Walls ---
		wallComp, isWall := entity.GetComponent(ecs.ComponentTypeWall)
		if isWall {
			wall := wallComp.(*components.WallComponent)
			tc, hasTransform := entity.GetComponent(ecs.ComponentTypeTransform)
			if hasTransform {
				transform := tc.(*components.TransformComponent)
				wallState := &types.WallState{
					HouseID:  wall.HouseID,
					EntityID: entity.ID,
					Position: &types.Position{
						X: transform.X,
						Y: transform.Y,
					},
					Width:  wall.Width,
					Height: wall.Height,
				}
				backendState.Walls = append(backendState.Walls, wallState)
			}
		}

		// --- Items ---
		// itemComp, hasItem := entity.GetComponent(ecs.ComponentTypeItem)
		//
		// if hasItem {
		// 	item := itemComp.(*components.ItemComponent)
		//
		// }
	}

	return backendState, nil
}

func (s *StateSerializer) FormatStateToClientState(backendState *types.BackendGameState, playerID uuid.UUID) *types.ClientGameState {
	playerCap := len(backendState.Players) - 1
	if playerCap < 0 {
		playerCap = 0
	}

	otherPlayers := make([]*types.PlayerState, 0, playerCap)
	for id, playerState := range backendState.Players {
		if id != playerID {
			otherPlayers = append(otherPlayers, playerState)
		}
	}

	state := &types.ClientGameState{
		SessionID:     backendState.SessionID,
		Items:         backendState.Items,
		Doors:         backendState.Doors,
		Walls:         backendState.Walls,
		Containers:    backendState.Containers,
		CurrentPlayer: backendState.Players[playerID],
		OtherPlayers:  otherPlayers,
		EscapeDoor:    backendState.EscapeDoor,
		Equipment:     backendState.Equipment,
		Switch:        backendState.Switch,
		EscapedCount:  backendState.EscapedCount,
	}

	return state
}

func (s *StateSerializer) RestBackendStatePool(state *types.BackendGameState) {
	for k := range state.Players {
		delete(state.Players, k)
	}
	state.Items = state.Items[:0]
	state.Doors = state.Doors[:0]
	state.Walls = state.Walls[:0]
	state.Containers = state.Containers[:0]
	state.EscapeDoor = state.EscapeDoor[:0]
	state.Switch = state.Switch[:0]
	state.SessionID = uuid.Nil
	state.EscapedCount = 0
}

func (s *StateSerializer) PutBackendState(state *types.BackendGameState) {
	s.backendStatePool.Put(state)
}

// grab the item
func (s *StateSerializer) getItemState(itemID uuid.UUID, entityID uuid.UUID, item *components.ItemComponent) *types.ItemState {

	switch item.ItemType {
	case types.ItemTypeWeapon:
		return &types.ItemState{
			ItemID:       itemID,
			EntityID:     entityID,
			Name:         item.Name,
			AttackPower:  int32(item.AttackPower),
			CriticalRate: float32(item.CriticalRate),
			Description:  item.Description,
		}

	case types.ItemTypeArmor:
		return &types.ItemState{
			ItemID:        itemID,
			EntityID:      entityID,
			Name:          item.Name,
			Description:   item.Description,
			DefenseRating: int32(item.DefenseRating),
			ArmorSlot:     item.ArmorSlot,
		}

	case types.ItemTypeConsumable:
		return &types.ItemState{
			ItemID:        itemID,
			EntityID:      entityID,
			Name:          item.Name,
			Description:   item.Description,
			HealingAmount: int32(item.HealingAmount),
			ManaAmount:    int32(item.ManaAmount),
		}
	}

	return nil
}
