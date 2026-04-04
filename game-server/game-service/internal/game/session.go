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
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/systems"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/utils"
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
	wg                       sync.WaitGroup

	stopChan  chan struct{}
	isRunning bool

	// caching

	// [playerID] - interacted
	playerInteractedCache map[uuid.UUID]bool

	// [entityID] - interacted
	containerInteractedCache map[uuid.UUID]bool

	// elimination tracking
	eliminationCh chan *types.Player
	// [memberID] finish position
	eliminations map[uuid.UUID]int

	// end session signal
	endSessionCh chan bool

	// TEST: testing only
	TestMessageSpy chan types.Message

	// item pool (session level, items are removed once assigned to a container)
	itemPool            types.ItemPool
	itemPoolInitialized bool

	// dependency injections
	sessionCloser   SessionCloser
	sender          SessionSender
	stateSerializer StateSerializer
	eventEmitter    EventEmitter
	itemsClient     grpcitems.ItemsClient

	movementSystem *systems.MovementSystem
	combatSystem   *systems.CombatSystem
	skillSystem    *systems.SkillSystem

	// escape
	switchEntityIDs  []uuid.UUID
	exitDoorEntityID uuid.UUID
	escapeSuccess    bool
}

type SessionCloser interface {
	CloseSession(sessionID uuid.UUID) error
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

type StateSerializer interface {
	PutBackendState(backendState *types.BackendGameState)
	SerializeBackendState(ctx context.Context, sessionID uuid.UUID, entities []*ecs.Entity) (*types.BackendGameState, error)
	FormatStateToClientState(backendState *types.BackendGameState, playerID uuid.UUID) *types.ClientGameState
}

func NewSession(sessionCloser SessionCloser, sender *messaging.MessageSender, serializer StateSerializer, em *ecs.EntityManager, eventEmitter EventEmitter, itemsClient grpcitems.ItemsClient) *Session {
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

		eliminationCh: make(chan *types.Player),
		eliminations:  make(map[uuid.UUID]int),

		endSessionCh: make(chan bool),

		sessionCloser:   sessionCloser,
		sender:          sender,
		stateSerializer: serializer,
		eventEmitter:    eventEmitter,
		itemsClient:     itemsClient,
	}

	go s.Start()

	return s
}

/**
* Handles all inner workings inside a single game session.
* NOTE: this method needs to be run inside a goroutine.
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

	// game loop processing
	go s.manageGameLoop()

	// eliminations processing
	go s.manageEliminations()

	// end game processing
	go s.manageEndSession()
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

			slog.Debug("Incoming message to game session",
				"sessionID", s.ID,
				"message", msg,
				"action", msg.Message.Action,
			)

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
					slog.Error("handleInteract failed to process on entity.",
						"player_id", playerID,
						"entity_id", entityIDUUID,
						"error", err)
					s.sender.SendMessageToPlayer(playerID, types.Message{})
				}
			case constants.ActionAttack:
				slog.Debug("Action from client was attack")
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
				attackPayload := parsedPayload.(types.PlayerSectionAttackPayload)
				slog.Debug("Parse attack payload",
					"attackPayload", attackPayload)

				playerID, err := uuid.Parse(attackPayload.PlayerID)

				if err != nil {
					slog.Error("Invalid PlayerID from session payload", "playerID", attackPayload.PlayerID, "error", err)
					continue
				}

				enemyEntityID, err := uuid.Parse(attackPayload.EnemyEntityID)

				if err != nil {
					slog.Error("Invalid ContainerEntityID from session payload",
						"containerEntityID", attackPayload.EnemyEntityID,
						"playerID", playerID,
						"error", err)
					s.sendErrorToPlayer(playerID, msg.Message.Action, "invalid container target")
					continue
				}

				err = s.handleAttack(playerID, enemyEntityID)

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

	// --- debugging ---

	slog.Debug("Showing all item entities that was created at the start of the game session.",
		"itemsState", s.itemPool,
	)

	// --- end debugging ---

	// --- core game loop --
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

			// elimination
			eliminationSys := systems.EliminationSystem{}
			eliminationSys.Update(deltaTime, entities, s.ID, s.eliminationCh)

			// rules
			rulesSys := systems.RulesSystem{}
			rulesSys.Update(deltaTime, entities, s.endSessionCh)

			// broadcast state update to all players
			err := s.broadcastFullState(entities)
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

/**
* Tracks end game status sent over by the rules system.
**/

func (s *Session) manageEndSession() {
	for endSession := range s.endSessionCh {
		if endSession {
			s.endSession()
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
		Escape:     false,
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

func (s *Session) AddDoor(x, y, width, height float64) uuid.UUID {
	doorConfig := DoorConfig{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
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

func (s *Session) AddBuilding(bx, by, bw, bh, wallThickness, doorWidth float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	doorOffset := (bw - doorWidth) / 2

	walls := []WallConfig{
		// 上牆
		{X: bx, Y: by, Width: bw, Height: wallThickness},
		// 左牆
		{X: bx, Y: by, Width: wallThickness, Height: bh},
		// 右牆
		{X: bx + bw - wallThickness, Y: by, Width: wallThickness, Height: bh},
		// 下左段
		{X: bx, Y: by + bh - wallThickness, Width: doorOffset, Height: wallThickness},
		// 下右段
		{X: bx + doorOffset + doorWidth, Y: by + bh - wallThickness, Width: doorOffset, Height: wallThickness},
	}

	for _, wallConfig := range walls {
		CreateWallEntity(s.EntityManager, wallConfig)
	}

	// 門（下牆缺口處，比牆壁薄）
	doorX := bx + doorOffset
	doorY := by + bh - wallThickness
	s.AddDoor(doorX, doorY, doorWidth, wallThickness)
}

func (s *Session) AddEscape(x, y float64) uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()

	Config := EscapeConfig{
		X: x,
		Y: y,
	}
	entity := CreateEscapeDoorEntity(s.EntityManager, Config)
	return entity.ID
}

func (s *Session) Shutdown() {
	s.mu.Lock()

	if !s.isRunning {
		s.mu.Unlock()
		return
	}

	slog.Info("Shutting down game session", "sessionID", s.ID)

	// remove session from server
	s.sessionCloser.CloseSession(s.ID)

	// clean up channels
	close(s.stopChan)
	close(s.MessageCh)
	close(s.eliminationCh)
	close(s.endSessionCh)

	s.mu.Unlock()

	// NOTE: wait for all dependencies of channels to clean up before exiting
	// most importantly for eliminationCh to finish as we need the eliminations to
	// be complete before calculating raw match state after a game session ends
	s.wg.Wait()
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
func (s *Session) broadcastFullState(entities []*ecs.Entity) error {
	ctx := context.Background()
	backendState, err := s.stateSerializer.SerializeBackendState(ctx, s.ID, entities)
	if err != nil {
		slog.Error("Failed to serialize state", "error", err)
		return err
	}

	clientStates := make(map[uuid.UUID]*types.ClientGameState)
	for _, playerID := range s.playerEntityIDToPlayerID {
		clientStates[playerID] = s.stateSerializer.FormatStateToClientState(backendState, playerID)
	}

	s.stateSerializer.PutBackendState(backendState)

	for playerID, clientState := range clientStates {
		go func(pID uuid.UUID, cState *types.ClientGameState) {
			s.sender.SendStateToPlayer(pID, cState)
		}(playerID, clientState)
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

	slog.Debug("Entity being interacted on in handleInteract.",
		"player_id", playerID,
		"target_entity_id", targetEntityID)

	if !hasEntity {
		slog.Error("Failed to retrieve target entity", "targetEntityID", targetEntityID)
		return fmt.Errorf("Error when attempting to retrieve target entity with entityID %s", targetEntityID)
	}

	// rate limiting cache
	s.mu.Lock() // hold lock to prevent check then act races here
	// check
	_, exists := s.containerInteractedCache[targetEntityID]
	if exists {
		slog.Debug("Container entity still cached, not available for interaction", "targetEntityID", targetEntityID)
		return fmt.Errorf("container targeted entityID %s was still cached and not available to be interacted", targetEntityID)
	} else {
		// act
		s.containerInteractedCache[targetEntityID] = true
	}
	// release lock after act
	s.mu.Unlock()

	// get that entity's type and decide on the effect
	_, isDoorEntity := targetEntity.GetComponent(ecs.ComponentTypeDoor)
	_, isContainerEntity := targetEntity.GetComponent(ecs.ComponentTypeContainer)
	itemComp, isItemEntity := targetEntity.GetComponent(ecs.ComponentTypeItem)
	switchComp, isSwitch := targetEntity.GetComponent(ecs.ComponentTypeSwitch)
	_, isEscapeDoor := targetEntity.GetComponent(ecs.ComponentTypeEscapeDoor)

	if !isDoorEntity && !isContainerEntity && !isSwitch && !isEscapeDoor && !isItemEntity {
		slog.Debug("Entity type did not match any interactable entity", "targetEntityID", targetEntityID)
		return fmt.Errorf("entity type did not match any interactable entity")
	}

	// --- player entity ---

	// establish player's position
	playerEntityID := s.playerIDToEntitiesID[playerID]

	// exit early if cached
	_, exists = s.playerInteractedCache[playerEntityID]

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

		slog.Debug("target entity isContainerEntity",
			"entity_id", targetEntityID,
			"entity_transform",
			struct {
				x float64
				y float64
			}{
				x: containerTransform.X,
				y: containerTransform.Y},
		)

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

		// only open, never close (chest stays open once opened)
		containerOpenable.IsOpen = true

		// create items on first open by using seeded itemPool
		if containerOpenable.HasBeenOpened == false {
			containerOpenable.HasBeenOpened = true

			// creates RANDOM items on the spot
			itemIDs, err := s.generateItems()

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
			// relate the container with the newly generated items
			containerItemIDs.ItemIDs = itemIDs
		}

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

	// --- item entity ---
	// when directly acting to an item
	if isItemEntity {
		slog.Debug("target entity is item entity",
			"entity_id", targetEntityID,
		)
		// lock to prevent races
		s.mu.Lock()
		defer s.mu.Unlock()
		targetItem, ok := itemComp.(*components.ItemComponent)

		if !ok {
			return fmt.Errorf("Error when asserting component to item.")
		}

		slog.Debug("Entity interacting with is an item entity.")

		// add item to player. playerEntity already validated during transformComp extraction
		playerItemIDListComp, exists := playerEntity.GetComponent(ecs.ComponentTypeItemIDList)

		if !exists {
			return fmt.Errorf("ItemIDList component doesnt exist in player's entity.")
		}

		playerItemIDList, ok := playerItemIDListComp.(*components.ItemIDListComponent)

		if !ok {
			return fmt.Errorf("ItemIDList component could not be asserted into its conerete type.")
		}

		// for edge cases
		for _, itemID := range playerItemIDList.ItemIDs {
			if itemID == targetItem.TemplateID {
				return fmt.Errorf("Item duplicate, attempting to add item that already exists into player's inventory.")
			}
		}

		// add to player item list
		playerItemIDList.ItemIDs = append(playerItemIDList.ItemIDs, targetEntityID)

		// TEST: for debugging
		slog.Debug("Retrieved player inventory list. Adding item.",
			"player_item_list", playerItemIDList,
			"target_item_name", targetItem.Name,
		)
		// TEST: end debugging

		// find the respecitive container and remove it from the container
		for _, entity := range s.EntityManager.GetAllEntities() {
			isContainer := entity.HasComponent(ecs.ComponentTypeContainer)
			if !isContainer {
				continue
			}

			containerItemIDListComp, exists := entity.GetComponent(ecs.ComponentTypeItemIDList)
			if !exists {
				slog.Warn("Couldnt find itemIDList component in container when attempting to remove item from container entity",
					"container_id", entity.ID,
				)
				continue
			}

			containerItemIDList, ok := containerItemIDListComp.(*components.ItemIDListComponent)
			if !ok {
				return fmt.Errorf("Failed to assert container item id list type when attempting to removing item from container entity")
			}

			updatedList := make([]uuid.UUID, 0, len(containerItemIDList.ItemIDs)-1)

			for _, itemID := range containerItemIDList.ItemIDs {
				if itemID == targetEntityID {
					continue
				}
				updatedList = append(updatedList, itemID)
			}

			containerItemIDList.ItemIDs = updatedList
			break
		}

		// unlocks here too
		return nil
	}

	// Check if the switch is on so we can open the emergency exit
	// -- switch entity --
	if isSwitch {
		// get location
		switchTransformComponent, hasTransform := targetEntity.GetComponent(ecs.ComponentTypeTransform)

		if !hasTransform {
			slog.Error("Failed to retrieve switch entity transform component", "targetEntityID", targetEntityID)
			return fmt.Errorf("Error when attempting to retrieve switch entity transform component with entityID %s", targetEntityID)
		}

		switchTransform := switchTransformComponent.(*components.TransformComponent)
		// validate is within distance from player
		isWithinDistance := s.calcWithinDistance(playerTransform.X, playerTransform.Y, switchTransform.X, switchTransform.Y)

		if !isWithinDistance {
			slog.Debug("Switch entity out of range for interaction", "targetID", targetEntityID, "playerID", playerID)
			s.sendErrorToPlayer(playerID, string(constants.ActionInteract), "too far away to interact")
			return ErrOutOfRange
		}

		switchComponent := switchComp.(*components.SwitchComponent)
		if switchComponent.IsActivated {
			return fmt.Errorf("switch already activated")
		}
		switchComponent.IsActivated = true
		slog.Info("Switch activated!", "playerID", playerID)

		exitDoor, exists := s.EntityManager.GetEntity(s.exitDoorEntityID)
		if exists {
			lockableComp, hasLockable := exitDoor.GetComponent(ecs.ComponentTypeLockable)
			if hasLockable {
				lockable := lockableComp.(*components.LockableComponents)
				lockable.IsLocked = false

				slog.Info("Exit door unlocked!")
			}
		}

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

	// -- escape door entity --
	if isEscapeDoor {
		// get location
		escapeDoorTransformComponent, hasTransform := targetEntity.GetComponent(ecs.ComponentTypeTransform)

		if !hasTransform {
			slog.Error("Failed to retrieve escape door entity transform component", "targetEntityID", targetEntityID)
			return fmt.Errorf("Error when attempting to retrieve escape door entity transform component with entityID %s", targetEntityID)
		}

		escapeDoorTransform := escapeDoorTransformComponent.(*components.TransformComponent)
		// validate is within distance from player
		isWithinDistance := s.calcWithinDistance(playerTransform.X, playerTransform.Y, escapeDoorTransform.X, escapeDoorTransform.Y)

		if !isWithinDistance {
			slog.Debug("Escape door entity out of range for interaction", "targetID", targetEntityID, "playerID", playerID)
			s.sendErrorToPlayer(playerID, string(constants.ActionInteract), "too far away to interact")
			return ErrOutOfRange
		}

		// check if door is locked
		lockableComp, hasLockable := targetEntity.GetComponent(ecs.ComponentTypeLockable)
		if !hasLockable {
			slog.Error("Escape door does not have lockable component", "targetEntityID", targetEntityID)
			return fmt.Errorf("escape door does not have lockable component")
		}

		lockable := lockableComp.(*components.LockableComponents)
		if lockable.IsLocked {
			slog.Debug("Escape door is still locked", "targetID", targetEntityID, "playerID", playerID)

			return fmt.Errorf("escape door is locked")
		}

		// door is unlocked, open it and let player escape!
		openableComp, hasOpenable := targetEntity.GetComponent(ecs.ComponentTypeOpenable)
		if hasOpenable {
			openable := openableComp.(*components.OpenableComponent)
			openable.IsOpen = true
			slog.Info("Escape door opened!", "playerID", playerID)
		}

		// trigger escape after a short delay to allow door animation
		slog.Info("Player is escaping through the door!", "playerID", playerID)
		s.handlePlayerEscape(playerID)
	}

	return nil
}

func (s *Session) handlePlayerEscape(playerID uuid.UUID) {
	s.mu.Lock()
	s.escapeSuccess = true
	s.mu.Unlock()

	s.mu.Lock()
	playerEntityID, ok := s.playerIDToEntitiesID[playerID]
	s.mu.Unlock()

	if !ok {
		slog.Error("Player entity ID not found", "playerID", playerID)
		return
	}

	playerEntity, exists := s.EntityManager.GetEntity(playerEntityID)
	if !exists {
		slog.Error("Player entity not found", "playerEntityID", playerEntityID)
		return
	}

	playerComp, hasPlayer := playerEntity.GetComponent(ecs.ComponentTypePlayer)
	if !hasPlayer {
		slog.Error("Player component not found", "playerEntityID", playerEntityID)
		return
	}
	player := playerComp.(*components.PlayerComponent)
	player.Escape = true
	slog.Info("Player escaped!", "playerID", playerID, "username", player.Username)

}

func (s *Session) handleAttack(playerID uuid.UUID, enemyEntityID uuid.UUID) error {
	playerEntityID, ok := s.playerIDToEntitiesID[playerID]
	if !ok {
		return fmt.Errorf("Player %s not found", playerID)
	}

	playerEntity, ok := s.EntityManager.GetEntity(playerEntityID)
	if !ok {
		return fmt.Errorf("Player %s is not exists", playerID)
	}
	// enemyEntity, ok := s.EntityManager.GetEntity(enemyEntityID)
	// if !ok {
	// 	slog.Error("Enemy entity does not exist", "enemyEntityID", enemyEntityID)
	// 	return fmt.Errorf("entity %s is not exists", enemyEntityID)
	// }
	// 確認目標存在
	_, enemyExists := s.EntityManager.GetEntity(enemyEntityID)
	if !enemyExists {
		return fmt.Errorf("Enemy entity %s does not exist", enemyEntityID)
	}

	playerC, hasPlayer := playerEntity.GetComponent(ecs.ComponentTypePlayer)
	if !hasPlayer {
		return fmt.Errorf("Player does not have player component")
	}
	player := playerC.(*components.PlayerComponent)
	player.HasHit = true
	player.AttackActive = true
	player.AttackTargetEntityID = enemyEntityID
	return nil
}

const (
	weaponDropRate     float64 = 0.2
	armorDropRate      float64 = 0.3
	consumableDropRate float64 = 0.5
)

/**
* generateItems generates new random items from the session itemPool base,
* creates item entities, and returns their IDs.
**/
func (s *Session) generateItems() ([]uuid.UUID, error) {
	slog.Debug("generating items from itemPool when opening container",
		"s.itemPool", s.itemPool)

	// decide on RNG values
	numberOfItems := utils.GenRandomBetween(2, 4)

	// distribution of items
	var numberOfWeapons int
	var numberOfArmor int
	var numberOfConsumables int

	var rollRangeStart float64 = 1
	var rollRangeEnd float64 = 10

	for i := 0; i < numberOfItems; i++ {
		// roll to determine which type to get
		roll := utils.GenRandomBetween(int(rollRangeStart), int(rollRangeEnd))
		slog.Debug("Rolled based on range.",
			"roll", roll,
			"rollRangeStart", rollRangeStart,
			"rollRangeEnd", rollRangeEnd)

		weaponWeight := math.RoundToEven(float64((rollRangeEnd - rollRangeStart + 1) * weaponDropRate))

		slog.Debug("Calculated weaponWeight.",
			"weaponWeight", weaponWeight)
		armorWeight := math.RoundToEven(float64((rollRangeEnd - rollRangeStart + 1) * armorDropRate))
		slog.Debug("Calculated armorWeight.",
			"armorWeight", armorWeight)

		if roll <= int(weaponWeight) {
			numberOfWeapons++
			continue
		}

		if roll > int(weaponWeight) && roll <= int(armorWeight+weaponWeight) {
			numberOfArmor++
			continue
		}

		numberOfConsumables++
	}

	// validate item pool correctly generated items
	if s.itemPool.Count <= 0 {
		return nil, fmt.Errorf("Item pool was empty.")
	}

	newItemEntityIDs := make([]uuid.UUID, 0, numberOfArmor+numberOfWeapons+numberOfConsumables)

	// pull from item pool based on the number of random items
	if numberOfWeapons != 0 {
		for i := 0; i < numberOfWeapons; i++ {
			// find weapon
			itemConfig, err := s.findSingleItemBase(types.ItemTypeWeapon)
			if err != nil {
				return nil, err
			}

			// create entity
			id := s.AddItem(*itemConfig)
			newItemEntityIDs = append(newItemEntityIDs, id)
		}
	}

	if numberOfArmor != 0 {
		for i := numberOfWeapons; i < numberOfArmor; i++ {
			itemConfig, err := s.findSingleItemBase(types.ItemTypeArmor)
			if err != nil {
				return nil, err
			}

			// create entity
			id := s.AddItem(*itemConfig)
			newItemEntityIDs = append(newItemEntityIDs, id)
		}
	}

	if numberOfConsumables != 0 {
		for i := numberOfWeapons + numberOfArmor; i < numberOfConsumables; i++ {
			itemConfig, err := s.findSingleItemBase(types.ItemTypeConsumable)
			if err != nil {
				return nil, err
			}

			// create entity
			id := s.AddItem(*itemConfig)
			newItemEntityIDs = append(newItemEntityIDs, id)
		}
	}

	return newItemEntityIDs, nil
}

func (s *Session) findSingleItemBase(itemType types.ItemType) (*types.ItemConfig, error) {
	slog.Info("findSingleBaseItem check all itemPool contents before finding item",
		"armor_count", len(s.itemPool.Armor),
		"weapon_count", len(s.itemPool.Weapons),
		"consumable_count", len(s.itemPool.Consumables),
	)
	switch itemType {
	case types.ItemTypeWeapon:
		randCount := utils.GenRandomBetween(0, len(s.itemPool.Weapons)-1)
		item := *s.itemPool.Weapons[randCount]
		return &item, nil
	case types.ItemTypeArmor:
		randCount := utils.GenRandomBetween(0, len(s.itemPool.Armor)-1)
		slog.Info("randomCount rolled for Armor",
			"randCount", randCount,
		)
		item := *s.itemPool.Armor[randCount]
		return &item, nil
	case types.ItemTypeConsumable:
		randCount := utils.GenRandomBetween(0, len(s.itemPool.Consumables)-1)
		item := *s.itemPool.Consumables[randCount]
		slog.Info("randomCount rolled for Consumable",
			"randCount", randCount,
		)
		return &item, nil
	default:
		return nil, fmt.Errorf("No items matched.")
	}
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
func (s *Session) AddItem(itemConfig types.ItemConfig) uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()
	entity := CreateItemEntity(s.EntityManager, itemConfig)
	return entity.ID
}

/**
* Manages eliminations outside primary game loop to claculate end game results.
**/
func (s *Session) manageEliminations() {
	s.wg.Add(1)
	defer s.wg.Done()

	for player := range s.eliminationCh {
		slog.Debug("Player eliminated",
			"ID", player.ID,
			"Username", player.Username,
			"SessionID", player.CurrentGameSessionId)

		// store in elimination for processing
		s.mu.Lock()
		s.eliminations[player.ID] = len(s.eliminations)
		s.mu.Unlock()
	}
}

/**
* Handles all processes at the end of a match session.
**/
func (s *Session) endSession() {
	// clean up
	s.Shutdown()

	// grab raw data for publishing end game stats
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
		// escapeDoorComp, _ := entity.GetComponent(ecs.ComponentTypeEscapeDoor)

		if isPlayer {
			// assert back to component's original type
			playerState := playerComponent.(*components.PlayerComponent)
			statsComp, _ := entity.GetComponent(ecs.ComponentTypeStats)
			stats := statsComp.(*components.StatsComponent)

			rawPlayers = append(rawPlayers, types.RawPlayerState{
				MemberID: playerState.MemberID.String(),
				Username: playerState.Username,
				Kills:    int32(stats.Kills),
				Deaths:   int32(stats.Deaths),
				Escape:   playerState.Escape,
			})
		}
	}

	return &types.RawMatchState{
		SessionID: s.ID,
		// TODO: need to add started at in session struct for tracking
		StartedAt:        time.Now(),
		EndedAt:          time.Now(),
		Players:          rawPlayers,
		EliminationOrder: s.eliminations,
	}
}

func (s *Session) InitialSystems() {
	// creates and setsup match progress entity within the entity manager
	CreateMatchProgressEntity(s.EntityManager)
}

func (s *Session) InitialMapObjects() {
	// add container (ensure it's not cut off at edges)
	containerX := constants.ContainerWidthRadius + rand.Float64()*(constants.MapWidth-2*constants.ContainerWidthRadius)
	containerY := constants.ContainerHeightRadius + rand.Float64()*(constants.MapHeight-2*constants.ContainerHeightRadius)
	s.AddContainer(containerX, containerY)

	// add building (300x240, wall thickness 20, door width 50, door on bottom)
	s.AddBuilding(400, 300, 300, 240, 20, 50)

	// add EscapeDoor
	exitDoorX := constants.ContainerWidthRadius + rand.Float64()*(constants.MapWidth-2*constants.ContainerWidthRadius)
	exitDoorY := constants.ContainerHeightRadius + rand.Float64()*(constants.MapHeight-2*constants.ContainerHeightRadius)
	exitDoor := CreateEscapeDoorEntity(s.EntityManager, EscapeConfig{
		X: exitDoorX,
		Y: exitDoorY,
	})
	s.exitDoorEntityID = exitDoor.ID

	// add Switch
	switchX := constants.ContainerWidthRadius + rand.Float64()*(constants.MapWidth-2*constants.ContainerWidthRadius)
	switchY := constants.ContainerHeightRadius + rand.Float64()*(constants.MapHeight-2*constants.ContainerHeightRadius)
	switchEntity := CreateSwitchEntity(s.EntityManager, SwitchConfig{
		X:        switchX,
		Y:        switchY,
		SwitchID: 1,
	})

	s.switchEntityIDs = []uuid.UUID{switchEntity.ID}

	// --- Create Items ---

	ctx := context.Background()

	s.InitializeItems(ctx)
}

func (s *Session) InitializeItems(ctx context.Context) error {
	data, err := s.itemsClient.ListItemTemplates(ctx)

	if err != nil {
		slog.Error("Error when attempting to get list of base armors for game creation.",
			"error", err,
		)
	}

	if data.Items == nil {
		slog.Error("No items returned from ListItemTemplates")
		return fmt.Errorf("No items returned from ListItemTemplates")
	}

	for _, item := range data.Items {
		templateId, err := uuid.Parse(item.Id)

		if err != nil {
			slog.Error("Error when attempting to parse template id as uuid during game creation.",
				"error", err,
				"armor.ItemId", item.Id,
			)
			return err
		}

		itemType := types.ItemType(item.ItemType)

		var newItemConfig types.ItemConfig

		slog.Info("Creating item",
			"itemType", itemType,
			"item", item,
		)

		switch itemType {
		case types.ItemTypeWeapon:
			newItemConfig = types.ItemConfig{
				TemplateID:  templateId,
				ItemType:    itemType,
				Name:        item.ItemName,
				Description: item.Description,
				BuyPrice:    int(item.BaseBuyPrice),
				SellPrice:   int(item.BaseSellPrice),

				DefenseRating:   int(item.DefenseRating),
				MagicResistance: int(item.MagicResistance),
				ArmorSlot:       item.ArmorSlot,
			}

			s.itemPool.Weapons = append(s.itemPool.Weapons, &newItemConfig)
			s.itemPool.Count++

		case types.ItemTypeArmor:
			newItemConfig = types.ItemConfig{
				TemplateID:  templateId,
				ItemType:    itemType,
				Name:        item.ItemName,
				Description: item.Description,
				BuyPrice:    int(item.BaseBuyPrice),
				SellPrice:   int(item.BaseSellPrice),

				WeaponType:   item.WeaponType,
				AttackPower:  int(item.AttackPower),
				CriticalRate: float64(item.CriticalRate),
			}

			s.itemPool.Armor = append(s.itemPool.Armor, &newItemConfig)
			s.itemPool.Count++

		case types.ItemTypeConsumable:
			newItemConfig = types.ItemConfig{
				TemplateID:  templateId,
				ItemType:    itemType,
				Name:        item.ItemName,
				Description: item.Description,
				BuyPrice:    int(item.BaseBuyPrice),
				SellPrice:   int(item.BaseSellPrice),

				HealingAmount: int(item.HealingAmount),
				ManaAmount:    int(item.ManaAmount),
				BuffDuration:  int(item.BuffDuration),
			}

			// add to sessions internal state
			s.itemPool.Consumables = append(s.itemPool.Consumables, &newItemConfig)
			s.itemPool.Count++

		default:
			slog.Error("No valid item types match the ItemType that was read from the data",
				"item.ItemType", item.ItemType,
			)
			return fmt.Errorf("No valid item types match the ItemType that was read from the data")
		}

		slog.Info("Created item",
			"current_itemPool", s.itemPool,
		)

	}

	return nil
}
