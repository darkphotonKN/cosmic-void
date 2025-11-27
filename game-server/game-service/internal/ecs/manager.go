package ecs

import (
	"sync"

	"github.com/google/uuid"
)

/**
* Entity Manager
*
* info to team:
* creates and track managers
**/

type EntityManager struct {
	entities map[uuid.UUID]*Entity
	mu       sync.RWMutex
}

func NewEntityManager() *EntityManager {
	return &EntityManager{
		entities: make(map[uuid.UUID]*Entity),
	}
}

func (m *EntityManager) CreateEntity() *Entity {
	newEntity := NewEntity()

	m.mu.Lock()
	defer m.mu.Unlock()

	// save it in entity manager
	m.entities[newEntity.ID] = newEntity

	return newEntity
}

func (m *EntityManager) GetEntity(id uuid.UUID) (*Entity, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entity, exists := m.entities[id]
	return entity, exists
}

func (m *EntityManager) RemoveEntity(id uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.entities, id)
}

/**
* here we want to return a slice copy to prevent direct access to the map
* which would allow the caller to use update or delete without locks
**/
func (m *EntityManager) GetAllEntities() []*Entity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entityList := make([]*Entity, len(m.entities))

	index := 0
	for _, entity := range m.entities {
		entityList[index] = entity

		index += 1
	}

	return entityList
}
