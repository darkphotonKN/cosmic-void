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
**/

type ComponentType string

const (
	ComponentTypePlayer ComponentType = "Player"
	ComponentTypeNPC    ComponentType = "NPC"
	ComponentTypeEnemy  ComponentType = "Enemy"

	ComponentTypeItem ComponentType = "Item"
	ComponentTypeDoor ComponentType = "Door"

	ComponentTypeTransform ComponentType = "Transform"
	ComponentTypeVelocity  ComponentType = "Velocity"

	ComponentTypeHealth ComponentType = "Health"
	ComponentTypeAttack ComponentType = "Attack"
	ComponentTypeBuff   ComponentType = "Buff"
	ComponentTypeDebuff ComponentType = "Debuff"
	ComponentTypeSkill  ComponentType = "Skill"

	ComponentTypeStats      ComponentType = "Stats"
	ComponentTypeLevel      ComponentType = "Level"
	ComponentTypeExperience ComponentType = "Experience"
	ComponentTypeInventory  ComponentType = "Inventory"
	ComponentTypeEquipment  ComponentType = "Equipment"

	ComponentTypeInteractable ComponentType = "Interactable"
	ComponentTypeDialogue     ComponentType = "Dialogue"
)

type Entity struct {
	ID         uuid.UUID
	components map[ComponentType]Component
	mu         sync.RWMutex
}

func NewEntity() *Entity {
	return &Entity{
		ID:         uuid.New(),
		components: make(map[ComponentType]Component, 0),
	}
}

func (e *Entity) AddComponent(component Component) {
	e.mu.Lock()
	defer e.mu.Unlock()

	componentType := component.Type()

	e.components[componentType] = component
}

func (e *Entity) RemoveComponent(componentType ComponentType) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.components, componentType)
}

func (e *Entity) HasComponent(componentType ComponentType) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, exists := e.components[componentType]

	return exists
}

func (e *Entity) GetComponent(componentType ComponentType) (Component, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	component, exists := e.components[componentType]

	return component, exists
}
