package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type MatchProgressComponent struct {
	// total players alive
	TotalAlivePlayers int

	// players that are dead [uuid]*ecs.Entity (Player)
	DeadPlayers map[uuid.UUID]ecs.Component
}

func (p *MatchProgressComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeMatchProgress
}

func NewMatchProgressComponent(totalAlivePlayers int) *MatchProgressComponent {
	return &MatchProgressComponent{
		TotalAlivePlayers: totalAlivePlayers, DeadPlayers: make(map[uuid.UUID]ecs.Component),
	}
}
