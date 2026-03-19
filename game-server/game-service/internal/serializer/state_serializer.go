package serializer

import (
	"context"
	"log/slog"

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

func (s *StateSerializer) SerializeBackendState(ctx context.Context, sessionID uuid.UUID, entities []*ecs.Entity) (*types.BackendGameState, error) {
	backendState := &types.BackendGameState{
		SessionID:  sessionID,
		Players:    make(map[uuid.UUID]*types.PlayerState, 0),
		Items:      make([]uuid.UUID, 0),
		Doors:      make([]*types.DoorState, 0),
		Containers: make([]*types.ContainerState, 0),
		EscapeDoor: make([]*types.EscapeDoorState, 0),
		Switch:     make([]*types.SwitchState, 0),
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

					slog.Debug("Item found before serialization.",
						"item", item,
					)

					itemState := &types.ItemState{
						ItemID:   itemID,
						EntityID: itemEntity.ID,
						Name:     item.Name,
						Quantity: 1,
					}

					populateItemDetails(ctx, item, itemState)

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
			}

			// Check if this is the recipient player
			backendState.Players[player.MemberID] = playerState
		}

		// --- Interactables ---

		// -- Doors --

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

							itemState := &types.ItemState{
								ItemID:   itemID,
								EntityID: itemEntity.ID,
								Name:     item.Name,
								Quantity: 1,
							}

							populateItemDetails(ctx, item, itemState)

							items = append(items, itemState)
						}
					}
				}
			}

			slog.Debug("items after extracting and formatting from entity", "items", items)

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
		// --- Items ---
		// TODO: add this after item entity is added
		itemComp, hasItem := entity.GetComponent(ecs.ComponentTypeItem)

		if hasItem {
			item := itemComp.(*components.ItemComponent)

			slog.Debug("item component found in game",
				"item", item,
			)
		}
	}

	return backendState, nil
}

func (s *StateSerializer) FormatStateToClientState(backendState *types.BackendGameState, playerID uuid.UUID) *types.ClientGameState {
	otherPlayers := make([]*types.PlayerState, 0, len(backendState.Players)-1)
	for id, playerState := range backendState.Players {
		if id != playerID {
			otherPlayers = append(otherPlayers, playerState)
		}
	}

	state := &types.ClientGameState{
		SessionID:     backendState.SessionID,
		Items:         backendState.Items,
		Doors:         backendState.Doors,
		Containers:    backendState.Containers,
		CurrentPlayer: backendState.Players[playerID],
		OtherPlayers:  otherPlayers,
		EscapeDoor:    backendState.EscapeDoor,
		Switch:        backendState.Switch,
	}

	return state
}

type Tester struct {
	value string
}

func (s *StateSerializer) Serialize(ctx context.Context, sessionID uuid.UUID, recipientPlayerID uuid.UUID, entities []*ecs.Entity) (*types.ClientGameState, error) {

	state := &types.ClientGameState{
		SessionID:     sessionID,
		CurrentPlayer: nil,
		OtherPlayers:  make([]*types.PlayerState, 0),
		Items:         make([]uuid.UUID, 0),
		Doors:         make([]*types.DoorState, 0),
		Containers:    make([]*types.ContainerState, 0),
		EscapeDoor:    make([]*types.EscapeDoorState, 0),
		Switch:        make([]*types.SwitchState, 0),
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

					itemState := &types.ItemState{
						ItemID:   itemID,
						EntityID: itemEntity.ID,
						Name:     item.Name,
						Quantity: 1,
					}

					populateItemDetails(ctx, item, itemState)

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
			}

			// Check if this is the recipient player
			if player.MemberID == recipientPlayerID {
				state.CurrentPlayer = playerState
			} else {
				state.OtherPlayers = append(state.OtherPlayers, playerState)
			}
		}

		// --- Interactables ---

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
			state.EscapeDoor = append(state.EscapeDoor, escapeDoorState)
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
			state.Switch = append(state.Switch, switchState)
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

							itemState := &types.ItemState{
								ItemID:   itemID,
								EntityID: itemEntity.ID,
								Name:     item.Name,
								Quantity: 1,
							}

							populateItemDetails(ctx, item, itemState)

							items = append(items, itemState)
						}
					}
				}
			}

			slog.Debug("items after extracting and formatting from entity", "items", items)

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
			state.Containers = append(state.Containers, containerState)
		}

		// --- Items ---
		// TODO: add this after item entity is added

	}

	return state, nil
}

// populateItemDetails reads item details directly from the ItemComponent
// which are populated at item creation time, avoiding per-tick gRPC calls.
func populateItemDetails(ctx context.Context, item *components.ItemComponent, itemState *types.ItemState) {
	itemState.AttackPower = int32(item.AttackPower)
	itemState.Durability = int32(item.Durability)
	itemState.CriticalRate = float32(item.CriticalRate)
	itemState.WeaponType = item.WeaponType
	itemState.DefenseRating = int32(item.DefenseRating)
	itemState.ArmorSlot = item.ArmorSlot
	itemState.HealingAmount = int32(item.HealingAmount)
	itemState.ManaAmount = int32(item.ManaAmount)
	itemState.Description = item.Description
}

// populateItemDetailsViaGRPC is the original implementation that fetches item
// details from items-service via gRPC on every call. Kept for reference.
// Deprecated: use populateItemDetails instead, which reads from pre-populated ItemComponent.
//
// func populateItemDetailsViaGRPC(ctx context.Context, item *components.ItemComponent, itemState *types.ItemState) {
// 	if item.ItemTool == nil {
// 		return
// 	}
//
// 	// Try weapons
// 	weaponResponse, err := item.ItemTool.ListWeaponsWithTemplate(ctx)
// 	if err != nil {
// 		log.Printf("Failed to fetch weapons for item %s: %v", item.ItemName, err)
// 	} else if weaponResponse != nil {
// 		for _, weapon := range weaponResponse.Weapons {
// 			if weapon.ItemName == item.ItemName {
// 				itemState.AttackPower = weapon.AttackPower
// 				itemState.Durability = weapon.Durability
// 				itemState.CriticalRate = weapon.CriticalRate
// 				itemState.WeaponType = weapon.WeaponType
// 				itemState.Description = weapon.Description
// 				return
// 			}
// 		}
// 	}
//
// 	// Try armors
// 	armorResponse, err := item.ItemTool.ListArmorsWithTemplate(ctx)
// 	if err != nil {
// 		log.Printf("Failed to fetch armors for item %s: %v", item.ItemName, err)
// 	} else if armorResponse != nil {
// 		for _, armor := range armorResponse.Armors {
// 			if armor.ItemName == item.ItemName {
// 				itemState.DefenseRating = armor.DefenseRating
// 				itemState.ArmorSlot = armor.ArmorSlot
// 				itemState.Description = armor.Description
// 				return
// 			}
// 		}
// 	}
//
// 	// Try consumables
// 	consumableResponse, err := item.ItemTool.ListConsumablesWithTemplate(ctx)
// 	if err != nil {
// 		log.Printf("Failed to fetch consumables for item %s: %v", item.ItemName, err)
// 	} else if consumableResponse != nil {
// 		for _, consumable := range consumableResponse.Consumables {
// 			if consumable.ItemName == item.ItemName {
// 				itemState.HealingAmount = consumable.HealingAmount
// 				itemState.ManaAmount = consumable.ManaAmount
// 				itemState.Description = consumable.Description
// 				return
// 			}
// 		}
// 	}
//
// 	log.Printf("No matching item details found for: %s", item.ItemName)
// }
