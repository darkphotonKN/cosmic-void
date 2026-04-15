package systems

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
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
func (s *RulesSystem) Update(deltaTime float64, entities []*ecs.Entity, endSessionCh chan bool) {
	var matchProgressFound bool
	var matchProgressComp ecs.Component
	deadPlayers := make(map[uuid.UUID]ecs.Component)
	escapedCount := 0

	for _, entity := range entities {
		if !matchProgressFound {
			// find the match progress in the same loop
			matchProgressC, hasMatchProgress := entity.GetComponent(ecs.ComponentTypeMatchProgress)

			if hasMatchProgress {
				matchProgressComp = matchProgressC
				// find match progress component once
				matchProgressFound = true
			}
		}

		// validation for player
		playerComp, hasPlayer := entity.GetComponent(ecs.ComponentTypePlayer)
		healthComp, hasHealth := entity.GetComponent(ecs.ComponentTypeHealth)

		if !hasHealth || !hasPlayer {
			continue
		}

		health := healthComp.(*components.HealthComponent)
		player := playerComp.(*components.PlayerComponent)

		if health.IsEliminated {
			// add to the dead players list
			deadPlayers[player.MemberID] = player
		}
		if player.Escape {
			escapedCount++
		}
	}

	// compare totalPlayers with number that is eliminated
	if matchProgressComp == nil {
		return
	}

	matchProgress := matchProgressComp.(*components.MatchProgressComponent)

	// update match progress dead players state in one go
	for _, p := range deadPlayers {
		player := p.(*components.PlayerComponent)
		if _, ok := matchProgress.DeadPlayers[player.MemberID]; !ok {
			matchProgress.DeadPlayers[player.MemberID] = player
		}
	}

	// matchProgress.TotalAlivePlayers = matchProgress.TotalAlivePlayers - len(deadPlayers)
	activePlayers := matchProgress.TotalAlivePlayers - len(deadPlayers) - escapedCount

	if activePlayers <= 1 {
		// signal end game
		endSessionCh <- true
	}
}
