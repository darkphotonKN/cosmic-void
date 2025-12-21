package game

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/systems"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
)

// the session represents one game room with its own ECS world
type Session struct {
	ID             uuid.UUID
	EntityManager  *ecs.EntityManager
	MessageCh      chan types.ClientPackage
	playerEntities map[uuid.UUID]uuid.UUID
	mu             sync.RWMutex

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

	sender *types.MessageSender
}

func NewSession(sender *types.MessageSender) *Session {
	sessionId := uuid.New()

	s := &Session{
		ID:            sessionId,
		EntityManager: ecs.NewEntityManager(),
		// map [playerID] to entityID
		playerEntities: make(map[uuid.UUID]uuid.UUID),
		MessageCh:      make(chan types.ClientPackage, 100),

		movementSystem: systems.NewMovementSystem(),
		combatSystem:   systems.NewCombatSystem(),
		skillSystem:    systems.NewSkillSystem(),
		stopChan:       make(chan struct{}),
		isRunning:      false,

		playerInteractedCache: make(map[uuid.UUID]bool),
		sender:                sender,
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

				s.handleInteract(playerID, entityIDUUID)
			}
		}
	}
}

const framerate = 1

/**
* manages all the game update loops.
* runs system code to update state of game x times every second.
**/
func (s *Session) manageGameLoop() {
	// TODO: update from once per second to 30 / 60 times a second
	ticker := time.NewTicker((1 * time.Second) / framerate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			entities := s.EntityManager.GetAllEntities()

			// movement
			movementSys := systems.MovementSystem{}
			movementSys.Update(float64(1), entities)

			// interaction
			interactionSys := systems.InteractionSystem{}
			interactionSys.Update(entities)
		}
	}
}

func (s *Session) AddPlayer(userID uuid.UUID, username string) uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()

	PlayerConfig := PlayerConfig{
		UserID:        userID,
		Username:      username,
		X:             0,
		Y:             0,
		SkillName:     "Basic Attack",
		SkillLevel:    1,
		CurrentHealth: 100,
		MaxHealth:     100,
		ItemName:      "Health Potion",
		ItemQuantity:  3,

		Vx: 0,
		Vy: 0,
	}

	entity := CreatePlayerEntity(s.EntityManager, PlayerConfig)

	fmt.Printf("\nupdating playerEntities map at KEY playerID:\n%s\n\n", userID)
	fmt.Printf("\nupdating playerEntities map with player entity id VALUE \n\n", entity.ID)
	s.playerEntities[userID] = entity.ID

	return entity.ID
}

func (s *Session) RemovePlayer(userID string) {
	// 選項 1: 直接移除
	// 選項 2: 標記為死亡，等待復活
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

	playerIDs := make([]uuid.UUID, 0, len(s.playerEntities))
	for playerID := range s.playerEntities {
		playerIDs = append(playerIDs, playerID)
	}
	return playerIDs
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
	playerEntityID, ok := s.playerEntities[playerID]
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
	// TEST: pulling out all entities for debugging
	fmt.Printf("\nALL player entities stored in game session: %+v\n\n", s.playerEntities)
	fmt.Printf("\nattempting to key in playerEntities map with playerID: %s\n\n", playerID)
	playerEntityID := s.playerEntities[playerID]
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
			time.Sleep(time.Second * 1)
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
