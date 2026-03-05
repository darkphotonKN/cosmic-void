package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type MatchProgressComponent struct {
	// total players
	totalPlayers int

	// players that are dead [uuid]*ecs.Entity (Player)
	deadPlayers map[uuid.UUID]*ecs.Entity
}

func (p *MatchProgressComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeMatchProgress
}

func NewMatchProgressComponent(totalPlayers int) *MatchProgressComponent {
	return &MatchProgressComponent{
		totalPlayers: totalPlayers, deadPlayers: make(map[uuid.UUID]*ecs.Entity),
	}
}
