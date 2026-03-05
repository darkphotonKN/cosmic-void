package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type MatchProgressComponent struct {
	// total players
	totalPlayers int

	// players that are dead
	// [uuid]Player
	deadPlayers map[uuid.UUID]*ecs.Entity
}

func (p *MatchProgressComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeMatchProgress
}

func NewMatchProgressComponent(players *ecs.Entity) *MatchProgressComponent {
	return &MatchProgressComponent{MemberID: memberID, Username: username}
}
