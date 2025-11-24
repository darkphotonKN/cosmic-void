package ecs

import "github.com/google/uuid"

/**
* Entity Manager
*
* info to team:
* creates and track managers
**/

type EntityManager struct {
	entities map[uuid.UUID]*Entity
}
