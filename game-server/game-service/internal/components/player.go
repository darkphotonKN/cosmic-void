package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type PlayerComponent struct {
	MemberID             uuid.UUID
	Username             string
	HasHit               bool
	AttackActive         bool
	AttackCooldown       float64
	AttackTargetEntityID uuid.UUID
	Escape               bool
}

func (p *PlayerComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypePlayer
}

func NewPlayerComponent(memberID uuid.UUID, username string, hasHit, attackActive bool, escape bool) *PlayerComponent {
	return &PlayerComponent{MemberID: memberID, Username: username, HasHit: hasHit, AttackActive: attackActive, Escape: escape}
}
