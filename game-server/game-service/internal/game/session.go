package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/systems"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
)

// the session represents one game room with its own ECS world
type Session struct {
	ID             uuid.UUID
	EntityManager  *ecs.EntityManager
	MessageCh      chan types.Message
	playerEntities map[uuid.UUID]uuid.UUID
	mu             sync.RWMutex

	movementSystem *systems.MovementSystem
	combatSystem   *systems.CombatSystem
	skillSystem    *systems.SkillSystem

	stopChan  chan struct{}
	isRunning bool

	// TEST: testing only
	TestMessageSpy chan types.Message
}

func NewSession() *Session {
	sessionId := uuid.New()

	s := &Session{
		ID:             sessionId,
		EntityManager:  ecs.NewEntityManager(),
		playerEntities: make(map[uuid.UUID]uuid.UUID),
		MessageCh:      make(chan types.Message),

		movementSystem: systems.NewMovementSystem(),
		combatSystem:   systems.NewCombatSystem(),
		skillSystem:    systems.NewSkillSystem(),
		stopChan:       make(chan struct{}),
		isRunning:      false,
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
				s.TestMessageSpy <- message
			default:
			}
		}
	}
	// TEST: end testing

	for {
		select {
		case message := <-s.MessageCh:
			fmt.Printf("\nincoming message to game session %s:\n%v\n\n", s.ID, message)

			switch constants.Action(message.Action) {
			case constants.ActionMove:
				fmt.Printf("Action from client was move\n")

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
	// TODO: update from once per second to 30 times a second
	ticker := time.NewTicker((1 * time.Second) / framerate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			// --- update game loop ---

			entities := s.EntityManager.GetAllEntities()
			movementSys := systems.MovementSystem{}
			movementSys.Update(float64(1), entities)
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

		Vx:    0,
		Vy:    0,
		Speed: 5,
	}

	entity := CreatePlayerEntity(s.EntityManager, PlayerConfig)
	s.playerEntities[userID] = entity.ID
	return entity.ID
}

func (s *Session) RemovePlayer(userID string) {
	// 選項 1: 直接移除
	// 選項 2: 標記為死亡，等待復活
}

func (s *Session) Update(deltaTime float64) {
	fmt.Printf("Session %s updating...\n", s.ID)
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
