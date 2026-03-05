package systems

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
)

type RulesSystem struct{}

/**
* This system is in charge of observing the game state to move the
* game towards the "game over" state.
**/
func NewRulesSystem() *RulesSystem {
	return &RulesSystem{}
}

// NOTE: this runs every game tick
func (s *RulesSystem) Update(deltaTime float64, entities []*ecs.Entity) {

	for _, entity := range entities {
		// validation for player
		_, hasPlayer := entity.GetComponent(ecs.ComponentTypePlayer)
		healthComp, hasHealth := entity.GetComponent(ecs.ComponentTypeHealth)

		if !hasHealth || !hasPlayer {
			continue
		}

		health := healthComp.(*components.HealthComponent)

		for _, entity := range entities {
		}
	}

}
