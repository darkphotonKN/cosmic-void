package game

import (
	"sync"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

// the session represents one game room with its own ECS world
type Session struct {
	ID             string
	EntityManager  *ecs.EntityManager
	playerEntities map[uuid.UUID]uuid.UUID
	messageCh      chan types.Message
	mu             sync.RWMutex
}

func NewSession(roomID string) *Session {
	return &Session{
		ID:             roomID,
		EntityManager:  ecs.NewEntityManager(),
		playerEntities: make(map[uuid.UUID]uuid.UUID),
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
		Strength:      10,
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
