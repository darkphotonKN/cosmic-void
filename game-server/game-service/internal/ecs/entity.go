package ecs

import (
	"sync"

	"github.com/google/uuid"
)

/**
* Entities
*
* extra team info: think of entities like a container that holds components.
*
	- AddComponent(component Component)
  - GetComponent(componentType string) (Component, bool)
  - HasComponent(componentType string) bool
  - RemoveComponent(componentType string)
**/

type ComponentType string

const (
	ComponentTypePlayer    ComponentType = "Player"
	ComponentTypeTransform ComponentType = "Transform"
)

type Entity struct {
	ID         uuid.UUID
	components map[ComponentType]Component
	mu         sync.RWMutex
}

func (e *Entity) AddComponent(component Component) {
	e.mu.Lock()
	defer e.mu.Unlock()

	componentType := component.Type()

	e.components[componentType] = component
}

func (e *Entity) GetComponent(componentType ComponentType) (Component, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	component, exists := e.components[componentType]

	return component, exists
}
