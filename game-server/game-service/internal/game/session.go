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
	playerEntities map[string]uuid.UUID
	mu             sync.RWMutex
}

func NewSession(roomID string) *Session {
	return &Session{
		ID:             roomID,
		EntityManager:  ecs.NewEntityManager(),
		playerEntities: make(map[string]uuid.UUID),
	}
}

func (s *Session) AddPlayer(userID, username string) uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()

	entity := CreatePlayerEntity(s.EntityManager, userID, username, 0, 0)
	s.playerEntities[userID] = entity.ID
	return entity.ID
}
