package game

import (
	"fmt"
	"sync"
	"time"

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
}

func NewSession() *Session {
	sessionId := uuid.New()

	s := &Session{
		ID:             sessionId,
		EntityManager:  ecs.NewEntityManager(),
		playerEntities: make(map[uuid.UUID]uuid.UUID),

		movementSystem: systems.NewMovementSystem(),
		combatSystem:   systems.NewCombatSystem(),
		skillSystem:    systems.NewSkillSystem(),
		stopChan:       make(chan struct{}),
		isRunning:      false,
	}

	go s.Start()

	return s
}

const framerate = 1

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

	// update game loop
	ticker := time.NewTicker((1 * time.Second) / framerate)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:

				// --- update game loop ---
				fmt.Println("update game loop")
				entities := s.EntityManager.GetAllEntities()
				movementSys := systems.MovementSystem{}
				movementSys.Update(float64(1), entities)
			}
		}
	}()
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
	fmt.Printf("Shutting down session %s\n", s.ID)
	close(s.stopChan)
}
