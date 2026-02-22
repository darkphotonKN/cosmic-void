package game

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	grpcitems "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/items"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components/metrics"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/messaging"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/serializer"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/systems"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
)

// the session represents one game room with its own ECS world
type Session struct {
	ID                       uuid.UUID
	EntityManager            *ecs.EntityManager
	MessageCh                chan types.ClientPackage
	playerIDToEntitiesID     map[uuid.UUID]uuid.UUID
	playerEntityIDToPlayerID map[uuid.UUID]uuid.UUID
	mu                       sync.RWMutex

	movementSystem *systems.MovementSystem
	combatSystem   *systems.CombatSystem
	skillSystem    *systems.SkillSystem

	stopChan  chan struct{}
	isRunning bool

	// caching

	// [playerID] - interacted
	playerInteractedCache map[uuid.UUID]bool
	// [entityID] - interacted
	containerInteractedCache map[uuid.UUID]bool

	// TEST: testing only
	TestMessageSpy chan types.Message

	// item pool (session-level, items are removed once assigned to a container)
	itemPool            []itemTemplate
	itemPoolInitialized bool

	// dependency injections
	sender          SessionSender
	stateSerializer *serializer.StateSerializer
	eventEmitter    EventEmitter
	itemsClient     grpcitems.ItemsClient
}

type SessionSender interface {
	SendMessageToPlayer(playerID uuid.UUID, message types.Message) error
	BroadcastToPlayerList(players []uuid.UUID, msg types.Message) error
	SendStateToPlayer(playerID uuid.UUID, clientState *types.ClientGameState) error
	BroadcastStateToPlayerList(players []uuid.UUID, state *types.ClientGameState) error
}

type EventEmitter interface {
	PublishMatchComplete(ctx context.Context, data *types.RawMatchState) error
}

func NewSession(sender *messaging.MessageSender, serializer *serializer.StateSerializer, em *ecs.EntityManager, eventEmitter EventEmitter, itemsClient grpcitems.ItemsClient) *Session {
	sessionId := uuid.New()

	s := &Session{
		ID:            sessionId,
		EntityManager: em,
		// map [playerID] to entityID
		playerIDToEntitiesID:     make(map[uuid.UUID]uuid.UUID),
		playerEntityIDToPlayerID: make(map[uuid.UUID]uuid.UUID, constants.DefautMaxSessionPlayers),
		MessageCh:                make(chan types.ClientPackage, 100),

		movementSystem: systems.NewMovementSystem(),
		combatSystem:   systems.NewCombatSystem(),
		skillSystem:    systems.NewSkillSystem(),
		stopChan:       make(chan struct{}),
		isRunning:      false,

		playerInteractedCache:    make(map[uuid.UUID]bool, constants.DefautMaxSessionPlayers),
		containerInteractedCache: make(map[uuid.UUID]bool),
		sender:                   sender,
		stateSerializer:          serializer,
		eventEmitter:             eventEmitter,
		itemsClient:              itemsClient,
	}

	go s.Start()

	return s
}

/**
* Handles all inner workings inside a single game session.
* NOTE: this method should be run inside a goroutine.
**/
func (s *Session) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return
	}

	s.isRunning = true

	// managing incoming client messages
	go s.manageClientMessages()

	// start update game loop
	go s.manageGameLoop()
}

/**
* Manages all incoming messages between client and game session via the
* message hub.
**/
func (s *Session) manageClientMessages() {
	// TEST: testing only
	if s.TestMessageSpy != nil {
		for {
			select {
			case message := <-s.MessageCh:
				slog.Debug("Test message received", "message", message)

				// propogate to test
				s.TestMessageSpy <- message.Message
			case <-s.stopChan:
				return
			}
		}
	}
	// TEST: end testing

	for {
		select {
		case msg := <-s.MessageCh:
			if s.TestMessageSpy != nil {
				return
			}

			slog.Debug("Incoming message to game session", "sessionID", s.ID, "message", msg)

			switch constants.Action(msg.Message.Action) {
			case constants.ActionMove:

				slog.Debug("Action from client was move")
				// parse payload based on message action
				parsedPayload, err := msg.Message.ParsePayload()

				if err != nil {
					slog.Error("Failed to parse payload - types don't match", "payload", parsedPayload, "error", err)
					// Get playerID from payload if possible, otherwise skip sending error to specific player
					if playerIDStr, ok := msg.Message.Payload["player_id"].(string); ok {
						if playerID, parseErr := uuid.Parse(playerIDStr); parseErr == nil {
							s.sendErrorToPlayer(playerID, msg.Message.Action, "failed to parse move request")
						}
					}
					continue
				}

				movePayload := parsedPayload.(types.PlayerSessionMovePayload)

				slog.Debug("Parsed move payload", "payload", movePayload)

				// update based on action payload
				playerID, err := uuid.Parse(movePayload.PlayerID)
				if err != nil {
					slog.Error("Invalid PlayerID from session payload", "playerID", movePayload.PlayerID, "error", err)
					// Cannot send error to player since we don't have valid playerID
					continue
				}
				s.handleMove(playerID, movePayload.Vx, movePayload.Vy)

			case constants.ActionInteract:
				slog.Debug("Action from client was interact")

				parsedPayload, err := msg.Message.ParsePayload()

				if err != nil {
					slog.Error("Failed to parse interact payload", "error", err)
					if playerIDStr, ok := msg.Message.Payload["player_id"].(string); ok {
						if playerID, parseErr := uuid.Parse(playerIDStr); parseErr == nil {
							s.sendErrorToPlayer(playerID, msg.Message.Action, "failed to parse interact request")
						}
					}
					continue
				}

				interactPayload := parsedPayload.(types.PlayerSessionInteractPayload)
				slog.Debug("Parsed interact payload", "payload", interactPayload)

				playerID, err := uuid.Parse(interactPayload.PlayerID)

				if err != nil {
					slog.Error("Invalid PlayerID from session payload", "playerID", interactPayload.PlayerID, "error", err)
					continue
				}

				entityIDUUID, err := uuid.Parse(interactPayload.EntityID)

				if err != nil {
					slog.Error("Invalid EntityID from session payload", "entityID", interactPayload.EntityID, "error", err)
					s.sendErrorToPlayer(playerID, msg.Message.Action, "invalid target object")
					continue
				}

				err = s.handleInteract(playerID, entityIDUUID)

				if err != nil {
					s.sender.SendMessageToPlayer(playerID, types.Message{})
				}
			case constants.ActionLoot:

				slog.Debug("Action from client was loot")
				parsedPayload, err := msg.Message.ParsePayload()

				if err != nil {
					slog.Error("Failed to parse loot payload", "error", err)
					if playerIDStr, ok := msg.Message.Payload["player_id"].(string); ok {
						if playerID, parseErr := uuid.Parse(playerIDStr); parseErr == nil {
							s.sendErrorToPlayer(playerID, msg.Message.Action, "failed to parse loot request")
						}
					}
					continue
				}

				lootPayload := parsedPayload.(types.PlayerSessionLootPayload)
				slog.Debug("Parsed loot payload", "payload", lootPayload)

				playerID, err := uuid.Parse(lootPayload.PlayerID)

				if err != nil {
					slog.Error("Invalid PlayerID from session payload", "playerID", lootPayload.PlayerID, "error", err)
					continue
				}

				containerEntityID, err := uuid.Parse(lootPayload.ContainerEntityID)

				if err != nil {
					slog.Error("Invalid ContainerEntityID from session payload",
						"containerEntityID", lootPayload.ContainerEntityID,
						"playerID", playerID,
						"error", err)
					s.sendErrorToPlayer(playerID, msg.Message.Action, "invalid container target")
					continue
				}

				itemEntityIDUUIDs := []uuid.UUID{}
				for _, itemEntityId := range lootPayload.ItemEntityIDs {
					entityId, err := uuid.Parse(itemEntityId)
					if err != nil {
						slog.Error("Invalid itemEntityId", "entityId", entityId, "error", err)
					}
					itemEntityIDUUIDs = append(itemEntityIDUUIDs, entityId)
				}

				err = s.handleLoot(playerID, containerEntityID, itemEntityIDUUIDs)

				if err != nil {
					s.sender.SendMessageToPlayer(playerID, types.Message{})
				}
			}

		case <-s.stopChan:
			slog.Info("Game session message handler stopped", "sessionID", s.ID)
			return
		}
	}
}

/**
* manages all the game update loops.
* runs system code to update state of game x times every second.
**/
func (s *Session) manageGameLoop() {
	ticker := time.NewTicker((1 * time.Second) / time.Duration(constants.GameFrameRate))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// NOTE: keep for tracking game loop performance
			tickStart := time.Now()

			// TEST: exclude game loop for tests
			if s.TestMessageSpy != nil {
				return
			}
			// TEST: END test block

			entities := s.EntityManager.GetAllEntities()

			// movement
			movementSys := systems.MovementSystem{}
			deltaTime := 1.0 / float64(constants.GameFrameRate)
			movementSys.Update(deltaTime, entities)

			// interaction
			interactionSys := systems.InteractionSystem{}
			interactionSys.Update(entities)

			// broadcast state update to all players
			err := s.broadcastFullState()
			if err != nil {
				slog.Error("Error broadcasting state", "error", err)
				continue
			}

			// NOTE: record metrics for tick duration (skip if not initialized)
			if metrics.TickDuration != nil {
				metrics.TickDuration.Record(context.Background(), time.Since(tickStart).Seconds())
			}
			if metrics.EntityCount != nil {
				metrics.EntityCount.Record(context.Background(), int64(len(entities)))
			}

		case <-s.stopChan:
			slog.Info("Game session game loop stopped", "sessionID", s.ID)
			return
		}
	}
}

func (s *Session) AddPlayer(playerID uuid.UUID, username string) uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()

	PlayerConfig := PlayerConfig{
		MemberID:      playerID,
		Username:      username,
		X:             constants.PlayerRadius + rand.Float64()*(constants.MapWidth-2*constants.PlayerRadius),
		Y:             constants.PlayerRadius + rand.Float64()*(constants.MapHeight-2*constants.PlayerRadius),
		SkillName:     "Basic Attack",
		SkillLevel:    1,
		CurrentHealth: 100,
		MaxHealth:     100,
		ItemName:      "Health Potion",
		ItemQuantity:  3,

		Vx: 0,
		Vy: 0,

		ItemIDList: []uuid.UUID{},
	}

	// create player state entity
	entity := CreatePlayerEntity(s.EntityManager, PlayerConfig)

	// update player id to entity id map
	s.playerIDToEntitiesID[playerID] = entity.ID
	// update players map
	s.playerEntityIDToPlayerID[entity.ID] = playerID
	return entity.ID
}

func (s *Session) RemovePlayer(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	playerID, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("RemovePlayer: Invalid userID", "userID", userID, "error", err)
		return
	}
	entityID, exists := s.playerEntityIDToPlayerID[playerID]
	if !exists {
		slog.Warn("RemovePlayer: playerID not found in session", "playerID", playerID)
		return
	}
	s.EntityManager.RemoveEntity(entityID)

	delete(s.playerIDToEntitiesID, playerID)
	delete(s.playerEntityIDToPlayerID, entityID)
	slog.Info("Removed player from session", "playerID", playerID, "sessionID", s.ID)
}

func (s *Session) AddDoor(x, y float64) uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()

	doorConfig := DoorConfig{
		X: x,
		Y: y,
	}

	entity := CreateDoorEntity(s.EntityManager, doorConfig)
	return entity.ID
}

func (s *Session) AddContainer(x, y float64) uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()

	ContainerConfig := ContainerConfig{
		X: x,
		Y: y,
	}
	itemIDList := make([]uuid.UUID, 0)

	entity := CreateContainerEntity(s.EntityManager, ContainerConfig, itemIDList)
	return entity.ID
}

func (s *Session) Update(deltaTime float64) {
	// fmt.Printf("Session %s updating...\n", s.ID)
	// entities := s.EntityManager.GetAllEntities()

	// s.movementSystem.Update(deltaTime, entities)
	// s.combatSystem.Update(deltaTime, entities)
	// s.skillSystem.Update(deltaTime, entities)
}

func (s *Session) Shutdown() {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	slog.Info("Shutting down game session", "sessionID", s.ID)
	close(s.stopChan)
	close(s.MessageCh)
}

/**
* GetPlayerIDs returns all player IDs in this session
**/
func (s *Session) GetPlayerIDs() []uuid.UUID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	playerIDs := make([]uuid.UUID, 0, len(s.playerIDToEntitiesID))
	for playerID := range s.playerIDToEntitiesID {
		playerIDs = append(playerIDs, playerID)
	}
	return playerIDs
}

/**
* Broadcasts the current game state, after serialization, to all the players in the
* session. Each player receives a personalized view with their player state separated.
**/
func (s *Session) broadcastFullState() error {
	ctx := context.Background()
	entities := s.EntityManager.GetAllEntities()

	// create and send personalized state for each player
	for _, playerID := range s.playerEntityIDToPlayerID {
		clientState, err := s.stateSerializer.Serialize(ctx, s.ID, playerID, entities)
		if err != nil {
			slog.Error("Failed to serialize state for player", "playerID", playerID, "error", err)
			continue
		}

		err = s.sender.SendStateToPlayer(playerID, clientState)
		if err != nil {
			slog.Error("Failed to send state to player", "playerID", playerID, "error", err)
		}
	}

	return nil
}

/**
* --- State Updates Handlers ---
**/

/**
* updates the movement component transform based on the input provided
* by the client.
**/
func (s *Session) handleMove(playerID uuid.UUID, vx, vy float64) error {
	s.mu.RLock()
	// get specific player entity
	playerEntityID, ok := s.playerIDToEntitiesID[playerID]
	s.mu.RUnlock()

	if !ok {
		slog.Error("PlayerEntityID doesn't exist", "playerID", playerID)
		return fmt.Errorf("PlayerEntityID doesn't exist for playerID: %s", playerID)
	}

	playerEntity, ok := s.EntityManager.GetEntity(playerEntityID)

	if !ok {
		slog.Error("PlayerEntity doesn't exist", "playerEntityID", playerEntityID, "playerID", playerID)
		return fmt.Errorf("Player entity doesn't exist for id %s", playerID)
	}

	playerVelocityComponent, ok := playerEntity.GetComponent(ecs.ComponentTypeVelocity)

	if !ok {
		slog.Error("Player's Velocity Component doesn't exist", "entityID", playerEntity.ID)
		return fmt.Errorf("Players Velocity Component doesn't exist for entity ID: %s", playerEntity.ID)
	}

	component := playerVelocityComponent.(*components.VelocityComponent)

	// update velocity values
	component.VX = vx
	component.VY = vy

	return nil
}

/**
* handles player interacting with x object with target entity id.
**/
func (s *Session) handleInteract(playerID uuid.UUID, targetEntityID uuid.UUID) error {
	targetEntity, hasEntity := s.EntityManager.GetEntity(targetEntityID)

	if !hasEntity {
		slog.Error("Failed to retrieve target entity", "targetEntityID", targetEntityID)
		return fmt.Errorf("Error when attempting to retrieve target entity with entityID %s", targetEntityID)
	}

	// check container cache first before wasting resources on execution
	s.mu.RLock()
	_, exists := s.containerInteractedCache[targetEntityID]
	if exists {
		slog.Debug("Container entity still cached, not available for interaction", "targetEntityID", targetEntityID)
		return fmt.Errorf("container targeted entityID %s was still cached and not available to be interacted", targetEntityID)
	}
	s.mu.RUnlock()

	// get that entity's type and decide on the effect
	_, isDoorEntity := targetEntity.GetComponent(ecs.ComponentTypeDoor)
	_, isContainerEntity := targetEntity.GetComponent(ecs.ComponentTypeContainer)

	if !isDoorEntity && !isContainerEntity {
		slog.Debug("Entity type did not match any interactable entity", "targetEntityID", targetEntityID)
		return fmt.Errorf("entity type did not match any interactable entity")
	}

	// --- player entity ---

	// establish player's position
	s.mu.RLock()
	playerEntityID := s.playerIDToEntitiesID[playerID]
	s.mu.RUnlock()

	s.mu.RLock()
	// exit early if cached
	_, exists = s.playerInteractedCache[playerEntityID]
	s.mu.RUnlock()

	if exists {
		slog.Debug("Player interacted too soon", "playerEntityID", playerEntityID)
		return fmt.Errorf("player interacted too soon with playerEntityID %s", playerEntityID)
	}

	playerEntity, hasPlayerEntity := s.EntityManager.GetEntity(playerEntityID)

	if !hasPlayerEntity {
		slog.Error("Failed to retrieve target player entity", "playerEntityID", playerEntityID)

		return fmt.Errorf("Error when attempting to retrieve target player entity with entityID %s", targetEntityID)
	}

	playerTransformComponent, hasTransform := playerEntity.GetComponent(ecs.ComponentTypeTransform)

	if !hasTransform {
		slog.Error("Failed to retrieve player entity transform component", "playerEntityID", playerEntityID)
		return fmt.Errorf("Error when attempting to retrieve player entity transform component with entityID %s", playerEntityID)
	}

	playerTransform := playerTransformComponent.(*components.TransformComponent)

	// --- door entity ---

	if isDoorEntity {
		// get location
		doorTransformComponent, hasTransform := targetEntity.GetComponent(ecs.ComponentTypeTransform)

		if !hasTransform {
			slog.Error("Failed to retrieve door entity transform component", "targetEntityID", targetEntityID)
			return fmt.Errorf("Error when attempting to retrieve door entity transform component with entityID %s", targetEntityID)
		}

		doorTransform := doorTransformComponent.(*components.TransformComponent)
		// validate is within distance from player
		isWithinDistance := s.calcWithinDistance(playerTransform.X, playerTransform.Y, doorTransform.X, doorTransform.Y)

		if !isWithinDistance {
			slog.Debug("Door entity out of range for interaction", "targetID", targetEntityID, "playerID", playerID)
			s.sendErrorToPlayer(playerID, string(constants.ActionInteract), "too far away to interact")
			return ErrOutOfRange
		}

		// trigger doors swap in openable state via its OpenableComponent
		doorOpenableComponent, hasOpenable := targetEntity.GetComponent(ecs.ComponentTypeOpenable)

		if !hasOpenable {
			slog.Error("Failed to retrieve door entity openable component", "targetEntityID", targetEntityID)
			return fmt.Errorf("Error when attempting to retrieve door entity openable component with entityID %s", targetEntityID)
		}

		doorOpenable := doorOpenableComponent.(*components.OpenableComponent)

		// update state
		doorOpenable.IsOpen = !doorOpenable.IsOpen

		// add door to interacted to cache
		s.mu.Lock()
		s.containerInteractedCache[targetEntityID] = true
		s.mu.Unlock()

		// release cache in 100 milliseconds
		go func() {
			time.Sleep(time.Millisecond * 100)
			s.mu.Lock()
			delete(s.containerInteractedCache, targetEntityID)
			s.mu.Unlock()
		}()

		// add player to interacted cache
		s.mu.Lock()
		s.playerInteractedCache[playerEntityID] = true
		s.mu.Unlock()

		// remove them from cache after a short while
		go func() {
			time.Sleep(time.Millisecond * 100)
			s.mu.Lock()
			delete(s.playerInteractedCache, playerEntityID)
			s.mu.Unlock()
		}()
	}

	if isContainerEntity {
		// get location
		containerTransformComponent, hasTransform := targetEntity.GetComponent(ecs.ComponentTypeTransform)

		if !hasTransform {
			slog.Error("Failed to retrieve container entity transform component", "targetEntityID", targetEntityID)
			return fmt.Errorf("Error when attempting to retrieve container entity transform component with entityID %s", targetEntityID)
		}
		containerTransform := containerTransformComponent.(*components.TransformComponent)
		// validate is within distance from player
		isWithinDistance := s.calcWithinDistance(playerTransform.X, playerTransform.Y, containerTransform.X, containerTransform.Y)
		if !isWithinDistance {
			slog.Debug("Container entity out of range for interaction", "targetID", targetEntityID, "playerID", playerID)
			s.sendErrorToPlayer(playerID, string(constants.ActionInteract), "too far away to interact")
			return ErrOutOfRange
		}
		// trigger containers swap in openable state via its OpenableComponent
		containerOpenableComponent, hasOpenable := targetEntity.GetComponent(ecs.ComponentTypeOpenable)

		if !hasOpenable {
			slog.Error("Failed to retrieve container entity openable component", "targetEntityID", targetEntityID)
			return fmt.Errorf("Error when attempting to retrieve container entity openable component with entityID %s", targetEntityID)
		}

		containerOpenable := containerOpenableComponent.(*components.OpenableComponent)

		// Only open, never close (chest stays open once opened)
		containerOpenable.IsOpen = true

		// create items on first open by fetching from items-service via gRPC
		if containerOpenable.HasBeenOpened == false {
			containerOpenable.HasBeenOpened = true

			itemIDs, err := s.generateContainerItems()

			if err != nil {
				fmt.Printf("Error generating container items: %v\n", err)
				return fmt.Errorf("failed to generate container items: %w", err)
			}

			itemIDsComponent, hasItemIDs := targetEntity.GetComponent(ecs.ComponentTypeItemIDList)
			if !hasItemIDs {
				slog.Error("Failed to retrieve container entity itemIDs component", "targetEntityID", targetEntityID)
				return fmt.Errorf("Error when attempting to retrieve container entity itemIDs component with entityID %s", targetEntityID)
			}

			containerItemIDs := itemIDsComponent.(*components.ItemIDListComponent)
			containerItemIDs.ItemIDs = itemIDs
		}

		// add container to interacted to cache
		s.mu.Lock()
		s.containerInteractedCache[targetEntityID] = true
		s.mu.Unlock()

		// release cache in 100 milliseconds
		go func() {
			time.Sleep(time.Millisecond * 100)
			s.mu.Lock()
			delete(s.containerInteractedCache, targetEntityID)
			s.mu.Unlock()
		}()

		// add player to interacted cache
		s.mu.Lock()
		s.playerInteractedCache[playerEntityID] = true
		s.mu.Unlock()

		// remove them from cache after a short while
		go func() {
			time.Sleep(time.Millisecond * 100)
			s.mu.Lock()
			delete(s.playerInteractedCache, playerEntityID)
			s.mu.Unlock()
		}()
	}

	return nil
}

func (s *Session) handleLoot(playerID uuid.UUID, containerEntityID uuid.UUID, lootEntityIDs []uuid.UUID) error {
	// get player entity ID
	playerEntityID, ok := s.playerIDToEntitiesID[playerID]
	if !ok {
		return fmt.Errorf("Player %s not found", playerID)
	}

	playerEntity, ok := s.EntityManager.GetEntity(playerEntityID)
	if !ok {
		return fmt.Errorf("Player %s is not exists", playerID)
	}
	containerEntity, ok := s.EntityManager.GetEntity(containerEntityID)
	if !ok {
		slog.Error("Container entity does not exist", "containerEntityID", containerEntityID)
		return fmt.Errorf("entity %s is not exists", containerEntityID)
	}

	for _, entityID := range lootEntityIDs {
		_, ok := s.EntityManager.GetEntity(entityID)
		if !ok {
			slog.Error("Item entity does not exist", "entityID", entityID)
			_ = fmt.Errorf("Item entity %s is not exists", entityID)
		}
	}

	itemIDsComponent, _ := containerEntity.GetComponent(ecs.ComponentTypeItemIDList)
	containerItemIDs := itemIDsComponent.(*components.ItemIDListComponent)

	// store to player's inventory
	playerItemIDsComponent, _ := playerEntity.GetComponent(ecs.ComponentTypeItemIDList)
	playerItemIDs, _ := playerItemIDsComponent.(*components.ItemIDListComponent)
	playerItemIDs.ItemIDs = append(playerItemIDs.ItemIDs, lootEntityIDs...)

	// remove looted items from container
	removeItemEntityIDsMap := make(map[uuid.UUID]struct{})
	for _, removeEntityID := range lootEntityIDs {
		removeItemEntityIDsMap[removeEntityID] = struct{}{}
	}
	newItemIDs := []uuid.UUID{}
	// only keep items not in the removal list
	for _, itemID := range containerItemIDs.ItemIDs {
		if _, exists := removeItemEntityIDsMap[itemID]; !exists {
			newItemIDs = append(newItemIDs, itemID)
		}
	}
	containerItemIDs.ItemIDs = newItemIDs

	return nil
}

// itemTemplate is a unified representation of an item from items-service
type itemTemplate struct {
	Name          string
	BaseBuyPrice  int32
	BaseSellPrice int32
}

/**
* initItemPool fetches all item templates from items-service via gRPC once
* and fills the pool according to configured ratios (weapons, armors, consumables).
**/
func (s *Session) initItemPool() error {
	ctx := context.Background()
	ctx, span := gameItemPoolTracer.Start(ctx, "game.initItemPool")
	defer span.End()

	if s.itemsClient == nil {
		return fmt.Errorf("itemsClient is not initialized")
	}

	// Step 1: Fetch all templates from items-service
	weaponTemplates := []itemTemplate{}
	armorTemplates := []itemTemplate{}
	consumableTemplates := []itemTemplate{}

	// Fetch weapons
	weapons, err := s.itemsClient.ListWeaponsWithTemplate(ctx)
	if err != nil {
		fmt.Printf("Warning: failed to fetch weapons: %v\n", err)
	} else {
		for _, w := range weapons.Weapons {
			weaponTemplates = append(weaponTemplates, itemTemplate{
				Name:          w.ItemName,
				BaseBuyPrice:  w.BaseBuyPrice,
				BaseSellPrice: w.BaseSellPrice,
			})
		}
	}

	// Fetch armors
	armors, err := s.itemsClient.ListArmorsWithTemplate(ctx)
	if err != nil {
		fmt.Printf("Warning: failed to fetch armors: %v\n", err)
	} else {
		for _, a := range armors.Armors {
			armorTemplates = append(armorTemplates, itemTemplate{
				Name:          a.ItemName,
				BaseBuyPrice:  a.BaseBuyPrice,
				BaseSellPrice: a.BaseSellPrice,
			})
		}
	}

	// Fetch consumables
	consumables, err := s.itemsClient.ListConsumablesWithTemplate(ctx)
	if err != nil {
		fmt.Printf("Warning: failed to fetch consumables: %v\n", err)
	} else {
		for _, c := range consumables.Consumables {
			consumableTemplates = append(consumableTemplates, itemTemplate{
				Name:          c.ItemName,
				BaseBuyPrice:  c.BaseBuyPrice,
				BaseSellPrice: c.BaseSellPrice,
			})
		}
	}

	// Step 2: Calculate quantities based on ratios
	weaponCount := (constants.ItemPoolSize * constants.WeaponRatio) / 100
	armorCount := (constants.ItemPoolSize * constants.ArmorRatio) / 100
	consumableCount := (constants.ItemPoolSize * constants.ConsumableRatio) / 100

	// Step 3: Fill pool by ratio with random selection
	s.itemPool = make([]itemTemplate, 0, constants.ItemPoolSize)

	// Add weapons
	for i := 0; i < weaponCount && len(weaponTemplates) > 0; i++ {
		randomIndex := rand.IntN(len(weaponTemplates))
		s.itemPool = append(s.itemPool, weaponTemplates[randomIndex])
	}

	// Add armors
	for i := 0; i < armorCount && len(armorTemplates) > 0; i++ {
		randomIndex := rand.IntN(len(armorTemplates))
		s.itemPool = append(s.itemPool, armorTemplates[randomIndex])
	}

	// Add consumables
	for i := 0; i < consumableCount && len(consumableTemplates) > 0; i++ {
		randomIndex := rand.IntN(len(consumableTemplates))
		s.itemPool = append(s.itemPool, consumableTemplates[randomIndex])
	}

	// Step 4: Shuffle the entire pool (so items are mixed, not grouped by type)
	for i := len(s.itemPool) - 1; i > 0; i-- {
		j := rand.IntN(i + 1)
		s.itemPool[i], s.itemPool[j] = s.itemPool[j], s.itemPool[i]
	}

	if len(s.itemPool) == 0 {
		return fmt.Errorf("no items available from items-service")
	}

	s.itemPoolInitialized = true
	fmt.Printf("Item pool initialized with %d items (weapons: %d, armors: %d, consumables: %d)\n",
		len(s.itemPool), weaponCount, armorCount, consumableCount)
	return nil
}

/**
* generateContainerItems picks 1-4 unique items from the session item pool,
* removes them from the pool, creates item entities, and returns their IDs.
**/
func (s *Session) generateContainerItems() ([]uuid.UUID, error) {
	// Initialize pool on first call
	if !s.itemPoolInitialized {
		if err := s.initItemPool(); err != nil {
			return nil, err
		}
	}

	if len(s.itemPool) == 0 {
		return nil, fmt.Errorf("item pool is empty, no more items available")
	}

	// Randomly select 1-4 items (capped by remaining pool size)
	count := rand.IntN(4) + 1
	if count > len(s.itemPool) {
		count = len(s.itemPool)
	}

	// Fisher-Yates shuffle the pool
	for i := len(s.itemPool) - 1; i > 0; i-- {
		j := rand.IntN(i + 1)
		s.itemPool[i], s.itemPool[j] = s.itemPool[j], s.itemPool[i]
	}

	// Take from the end and shrink the pool
	selected := make([]itemTemplate, count)
	copy(selected, s.itemPool[len(s.itemPool)-count:])
	s.itemPool = s.itemPool[:len(s.itemPool)-count]

	// Create item entities from selected templates
	itemIDs := make([]uuid.UUID, 0, count)
	for _, item := range selected {
		config := ItemConfig{
			Name:     item.Name,
			ItemTool: s.itemsClient,
		}

		priceConfig := PriceConfig{
			BaseBuyPrice:  item.BaseBuyPrice,
			BaseSellPrice: item.BaseSellPrice,
		}
		itemID := s.AddItem(config, priceConfig)
		itemIDs = append(itemIDs, itemID)
	}

	return itemIDs, nil
}

/**
* sendErrorToPlayer sends a structured error message to a specific player.
* It provides user-friendly messages to the client.
**/
func (s *Session) sendErrorToPlayer(playerID uuid.UUID, action string, userMessage string) {
	s.sender.SendMessageToPlayer(playerID, types.Message{
		Action: action,
		Payload: map[string]interface{}{
			"success": false,
			"message": userMessage,
		},
	})
}

/**
* checks if a target is within 2d cartesian coordinates range of another.
**/
func (s *Session) calcWithinDistance(x, y, xTarget, yTarget float64) bool {
	// calculate range via range provided by interactable
	xDiff := math.Pow(x-xTarget, 2)
	yDiff := math.Pow(y-yTarget, 2)
	distanceBetween := math.Sqrt(xDiff + yDiff)

	// too far
	if distanceBetween > constants.DefaultInteractableRange {
		return false
	}

	return true
}

/**
* addItem creates an item entity from config and returns its ID
**/
func (s *Session) AddItem(itemConfig ItemConfig, priceConfig PriceConfig) uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()
	entity := CreateItemEntity(s.EntityManager, itemConfig, priceConfig)

	return entity.ID
}

/**
* Handles all processes at the end of a match session.
**/
func (s *Session) endSession() {
	// clean up
	s.Shutdown()

	// grab raw data for publishing
	rawMatchState := s.getRawMatchState()
	s.eventEmitter.PublishMatchComplete(context.Background(), rawMatchState)
}

/**
* Converts game specific entities into raw data for processing.
**/
func (s *Session) getRawMatchState() *types.RawMatchState {
	// TODO: update this to fixed player count once player count is fixed
	rawPlayers := make([]types.RawPlayerState, 0)

	entities := s.EntityManager.GetAllEntities()

	// --- player data ---
	for _, entity := range entities {
		playerComponent, isPlayer := entity.GetComponent(ecs.ComponentTypePlayer)

		if isPlayer {
			// assert back to component's original type
			playerState := playerComponent.(*components.PlayerComponent)

			// pull players end game stats state out of its entity
			statsComp, _ := entity.GetComponent(ecs.ComponentTypeStats)
			stats := statsComp.(*components.StatsComponent)

			rawPlayers = append(rawPlayers, types.RawPlayerState{
				MemberID: playerState.MemberID.String(),
				Username: playerState.Username,
				Kills:    int32(stats.Kills),
				Deaths:   int32(stats.Deaths),
			})
		}
	}

	return &types.RawMatchState{
		SessionID: s.ID,
		// TODO: need to add started at in session struct for tracking
		StartedAt: time.Now(),
		EndedAt:   time.Now(),
		Players:   rawPlayers,
	}
}

func (s *Session) InitialMapObjects() {
	// add container (ensure it's not cut off at edges)
	containerX := constants.ContainerWidthRadius + rand.Float64()*(constants.MapWidth-2*constants.ContainerWidthRadius)
	containerY := constants.ContainerHeightRadius + rand.Float64()*(constants.MapHeight-2*constants.ContainerHeightRadius)
	s.AddContainer(containerX, containerY)
}
