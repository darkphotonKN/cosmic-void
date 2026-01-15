package game

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	commontypes "github.com/darkphotonKN/cosmic-void-server/common/types"
	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
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

	// dependency injections
	sender          SessionSender
	stateSerializer *serializer.StateSerializer
	eventEmitter    EventEmitter
}

var ItemPool = []ItemConfig{
	{Name: "Health Potion", Quantity: 0},
	{Name: "Mana Potion", Quantity: 0},
	{Name: "Gold Coin", Quantity: 0},
	{Name: "Silver Coin", Quantity: 0},
	{Name: "Iron Sword", Quantity: 0},
	{Name: "Wooden Shield", Quantity: 0},
	{Name: "Magic Scroll", Quantity: 0},
	{Name: "Ancient Key", Quantity: 0},
	{Name: "Gemstone", Quantity: 0},
	{Name: "Mystery Aex", Quantity: 0},
}

type SessionSender interface {
	SendMessageToPlayer(playerID uuid.UUID, message types.Message) error
	BroadcastToPlayerList(players []uuid.UUID, msg types.Message) error
	SendStateToPlayer(playerID uuid.UUID, clientState *types.ClientGameState) error
	BroadcastStateToPlayerList(players []uuid.UUID, state *types.ClientGameState) error
}

type EventEmitter interface {
	PublishMatchComplete(ctx context.Context, data *commontypes.MatchEndState) error
}

func NewSession(sender *messaging.MessageSender, serializer *serializer.StateSerializer, em *ecs.EntityManager, eventEmitter EventEmitter) *Session {
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
				fmt.Printf("\nTest message received, %+v\n\n", message)

				// propogate to test
				s.TestMessageSpy <- message.Message
			default:
			}
		}
	}
	// TEST: end testing

	for {
		select {
		case msg := <-s.MessageCh:
			fmt.Printf("\nincoming message to game session %s:\n%v\n\n", s.ID, msg)

			switch constants.Action(msg.Message.Action) {
			case constants.ActionMove:

				fmt.Printf("Action from client was move\n")
				// parse payload based on message action
				parsedPayload, err := msg.Message.ParsePayload()

				if err != nil {
					// TODO: respond to client error
					fmt.Printf("\n attempting to parse payload from %+v from unsuccesfull as types don't match.\n\n", parsedPayload)
				}

				movePayload := parsedPayload.(types.PlayerSessionMovePayload)

				fmt.Printf("\nParsed move payload:\n%+v\n\n", movePayload)

				// update based on action payload
				playerID, err := uuid.Parse(movePayload.PlayerID)
				if err != nil {
					fmt.Printf("\nPlayerID %s from session payload was invalid.\n\n", movePayload.PlayerID)
					// TODO: respond to client error
				}
				s.handleMove(playerID, movePayload.Vx, movePayload.Vy)

			case constants.ActionInteract:
				fmt.Printf("Action from client was interact")

				parsedPayload, err := msg.Message.ParsePayload()

				if err != nil {
					// TODO respond to client error
				}

				interactPayload := parsedPayload.(types.PlayerSessionInteractPayload)
				fmt.Printf("\nParsed interact payload:\n%+v\n\n", interactPayload)

				playerID, err := uuid.Parse(interactPayload.PlayerID)

				if err != nil {
					fmt.Printf("\nPlayerID %s from session payload was invalid.\n\n", interactPayload.PlayerID)
					// TODO: respond to client error
				}

				entityIDUUID, err := uuid.Parse(interactPayload.EntityID)

				if err != nil {
					fmt.Printf("\nEntityID %s from session payload was invalid.\n\n", interactPayload.EntityID)
					// TODO: respond to client error
				}

				err = s.handleInteract(playerID, entityIDUUID)

				if err != nil {
					s.sender.SendMessageToPlayer(playerID, types.Message{})
				}
			}
		case <-s.stopChan:
			fmt.Printf("Game session %s: message handler stopped\n", s.ID)
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
				fmt.Printf("Error occured when broadcasting state: %+v\n", err)
				continue
			}
		case <-s.stopChan:
			fmt.Printf("Game session %s: game loop stopped\n", s.ID)
			return
		}
	}
}

func (s *Session) AddPlayer(playerID uuid.UUID, username string) uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()

	PlayerConfig := PlayerConfig{
		UserID:        playerID,
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
		fmt.Println("RemovePlayer: Invailid userID", playerID)
		return
	}
	entityID, exists := s.playerEntityIDToPlayerID[playerID]
	if !exists {
		fmt.Println("RemovePlayer: playerID not found in session", playerID)
		return
	}
	s.EntityManager.RemoveEntity(entityID)

	delete(s.playerIDToEntitiesID, playerID)
	delete(s.playerEntityIDToPlayerID, entityID)
	fmt.Printf("Remove player %s from session %s\n", playerID, s.ID)
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
	fmt.Printf("Shutting down game session id %s\n", s.ID)
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
	entities := s.EntityManager.GetAllEntities()

	// create and send personalized state for each player
	for _, playerID := range s.playerEntityIDToPlayerID {
		clientState, err := s.stateSerializer.Serialize(s.ID, playerID, entities)
		if err != nil {
			fmt.Printf("Error when attempting to serialize state for player %s: %+v\n", playerID, err)
			continue
		}

		err = s.sender.SendStateToPlayer(playerID, clientState)
		if err != nil {
			fmt.Printf("Error sending state to player %s: %+v\n", playerID, err)
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
		fmt.Printf("\nPlayerEntityID doesn't exist for playerID: %s\n\n", playerID)
		return fmt.Errorf("\nPlayerEntityID doesn't exist for playerID: %s\n\n", playerID)
	}

	playerEntity, ok := s.EntityManager.GetEntity(playerEntityID)

	if !ok {
		fmt.Printf("\nPlayerEntity doesn't exist for player playerEntityID %s\n\n", playerID)
		return fmt.Errorf("\nPlayer entity doens't exist for id %s\n\n", playerID)
	}

	playerVelocityComponent, ok := playerEntity.GetComponent(ecs.ComponentTypeVelocity)

	if !ok {
		fmt.Printf("\nPlayers Velocity Component doesn't exist for enntity ID: %s\n\n", playerEntity.ID)
		return fmt.Errorf("\nPlayers Velocity Component doesn't exist for enntity ID: %s\n\n", playerEntity.ID)
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
		fmt.Printf("Error when attempting to retrieve target entity with entityID %s\n", targetEntityID)
		return fmt.Errorf("Error when attempting to retrieve target entity with entityID %s", targetEntityID)
	}

	// check container cache first before wasting resources on execution
	s.mu.RLock()
	_, exists := s.containerInteractedCache[targetEntityID]
	if exists {
		fmt.Printf("container targeted entityID %s was still cached and not available to be interacted.\n", targetEntityID)
		return fmt.Errorf("container targeted entityID %s was still cached and not available to be interacted.\n", targetEntityID)
	}
	s.mu.RUnlock()

	// get that entity's type and decide on the effect
	_, isDoorEntity := targetEntity.GetComponent(ecs.ComponentTypeDoor)
	_, isContainerEntity := targetEntity.GetComponent(ecs.ComponentTypeContainer)

	if !isDoorEntity && !isContainerEntity {
		fmt.Printf("entity type did not match any interactable entity.\n")
		return fmt.Errorf("entity type did not match any interactable entity.\n")
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
		fmt.Printf("player interacted too soon with with playerEntityID %s\n", playerEntityID)
		return fmt.Errorf("player interacted too soon with with playerEntityID %s\n", playerEntityID)
	}

	playerEntity, hasPlayerEntity := s.EntityManager.GetEntity(playerEntityID)

	if !hasPlayerEntity {
		fmt.Printf("Error when attempting to retrieve target player entity with entityID %s\n", playerEntityID)

		return fmt.Errorf("Error when attempting to retrieve target player entity with entityID %s\n", targetEntityID)
	}

	playerTransformComponent, hasTransform := playerEntity.GetComponent(ecs.ComponentTypeTransform)

	if !hasTransform {
		fmt.Printf("Error when attempting to retrieve player entity transform component with entityID %s\n", playerEntityID)
		return fmt.Errorf("Error when attempting to retrieve player entity transform component with entityID %s", playerEntityID)
	}

	playerTransform := playerTransformComponent.(*components.TransformComponent)

	// --- door entity ---

	if isDoorEntity {
		// get location
		doorTransformComponent, hasTransform := targetEntity.GetComponent(ecs.ComponentTypeTransform)

		if !hasTransform {
			fmt.Printf("Error when attempting to retrieve door entity transform component with entityID %s\n", targetEntityID)
			return fmt.Errorf("Error when attempting to retrieve door entity transform component with entityID %s", targetEntityID)
		}

		doorTransform := doorTransformComponent.(*components.TransformComponent)
		// validate is within distance from player
		isWithinDistance := s.calcWithinDistance(playerTransform.X, playerTransform.Y, doorTransform.X, doorTransform.Y)

		if !isWithinDistance {
			// TODO: add return message to client
			fmt.Printf("Error when attempting to interact with door entity as it was out of range. targetID: %s, playerID: %s. \n", targetEntityID, playerID)
			return ErrOutOfRange
		}

		// trigger doors swap in openable state via its OpenableComponent
		doorOpenableComponent, hasOpenable := targetEntity.GetComponent(ecs.ComponentTypeOpenable)

		if !hasOpenable {
			fmt.Printf("Error when attempting to retrieve door entity openable component with entityID %s\n", targetEntityID)
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
			fmt.Printf("Error when attempting to retrieve container entity transform component with entityID %s\n", targetEntityID)
			return fmt.Errorf("Error when attempting to retrieve container entity transform component with entityID %s", targetEntityID)
		}
		containerTransform := containerTransformComponent.(*components.TransformComponent)
		// validate is within distance from player
		isWithinDistance := s.calcWithinDistance(playerTransform.X, playerTransform.Y, containerTransform.X, containerTransform.Y)
		if !isWithinDistance {
			// TODO: add return message to client
			fmt.Printf("Error when attempting to interact with container entity as it was out of range. targetID: %s, playerID: %s. \n", targetEntityID, playerID)
			return ErrOutOfRange
		}
		// trigger containers swap in openable state via its OpenableComponent
		containerOpenableComponent, hasOpenable := targetEntity.GetComponent(ecs.ComponentTypeOpenable)

		if !hasOpenable {
			fmt.Printf("Error when attempting to retrieve container entity openable component with entityID %s\n", targetEntityID)
			return fmt.Errorf("Error when attempting to retrieve container entity openable component with entityID %s", targetEntityID)
		}

		containerOpenable := containerOpenableComponent.(*components.OpenableComponent)

		// Only open, never close (chest stays open once opened)
		containerOpenable.IsOpen = true

		// create items on first open
		if containerOpenable.HasBeenOpened == false {
			containerOpenable.HasBeenOpened = true
			itemIDs := make([]uuid.UUID, 0)
			count := rand.IntN(4) + 1
			for i := 0; i < count; i++ {
				itemID := s.createRandomItem()
				itemIDs = append(itemIDs, itemID)
			}
			itemIDsComponent, hasItemIDs := targetEntity.GetComponent(ecs.ComponentTypeItemIDList)
			if !hasItemIDs {
				fmt.Printf("Error when attempting to retrieve container entity itemIDs component with entityID %s\n", targetEntityID)
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
* createRandomItem creates a random item entity and returns its ID
**/
func (s *Session) createRandomItem() uuid.UUID {
	rendomIndex := rand.IntN(10)
	itemOfPool := ItemPool[rendomIndex]
	quantity := rand.IntN(10) + 1
	item := ItemConfig{
		Name:     itemOfPool.Name,
		Quantity: quantity,
	}
	itemId := s.addItem(item)
	return itemId
}

/**
* addItem creates an item entity from config and returns its ID
**/
func (s *Session) addItem(itemConfig ItemConfig) uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()

	entity := CreateItemEntity(s.EntityManager, itemConfig)
	return entity.ID
}

/**
* Handles all processes at the end of a match session.
**/
func (s *Session) endSession() {
	// clean up
	s.Shutdown()

	matchEndData, err := s.formatMatchEndData()
	if err != nil {
		fmt.Printf("\nMatch end data err: %+v\n\n", err)
	}

	fmt.Printf("\nMatch end data: %+v\n\n", matchEndData)

	s.eventEmitter.PublishMatchComplete(context.Background(), matchEndData)
}

/**
* Formats the final end game data from the final game state.
**/
// TODO: WIP
func (s *Session) formatMatchEndData() (*commontypes.MatchEndState, error) {
	// entities := s.EntityManager.GetAllEntities()
	fmt.Println("Formatting data after match ended.")

	return nil, nil
}

func (s *Session) InitialMapObjects() {
	// add container (ensure it's not cut off at edges)
	containerX := constants.ContainerWidthRadius + rand.Float64()*(constants.MapWidth-2*constants.ContainerWidthRadius)
	containerY := constants.ContainerHeightRadius + rand.Float64()*(constants.MapHeight-2*constants.ContainerHeightRadius)
	s.AddContainer(containerX, containerY)
}
