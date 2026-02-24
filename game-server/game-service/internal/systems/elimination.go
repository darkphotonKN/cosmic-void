package systems

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
)

/**
* Tracks players health levels and processes eliminiations once their
* health levels pass a certain threshold.
**/
type EliminationSystem struct {
}

func NewEliminationSystem() *EliminationSystem {
	return &EliminationSystem{}
}

func (s *EliminationSystem) Update(deltaTime float64, entities []*ecs.Entity, sessionID uuid.UUID, eliminationCh chan *types.Player) {
	for _, entity := range entities {
		playerComp, isPlayer := entity.GetComponent(ecs.ComponentTypePlayer)

		if !isPlayer {
			continue
		}

		healthComp, hasHealth := entity.GetComponent(ecs.ComponentTypeHealth)

		if !hasHealth {
			continue
		}

		player := playerComp.(*components.PlayerComponent)
		health := healthComp.(*components.HealthComponent)

		// track eliminated players and send them for elimination processing
		if health.CurrentHealth <= 0 && !health.IsEliminated {
			eliminationCh <- &types.Player{
				ID:                   player.MemberID,
				Username:             player.Username,
				CurrentGameSessionId: sessionID,
			}

			// eliminate player
			health.IsEliminated = true
		}
	}
}
