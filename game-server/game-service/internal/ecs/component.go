package ecs

/**
* Components
*
* info to team:
* the component interface it tells what all components must implement.
* components are pure data, no logic
**/

type Component interface {
	// returns the component type name, like "Player".
	// all components must be able to return its own type
	Type() ComponentType
}
