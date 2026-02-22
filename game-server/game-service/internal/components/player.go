package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type PlayerComponent struct {
	MemberID uuid.UUID
	Username string
}

func (p *PlayerComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypePlayer
}

func NewPlayerComponent(memberID uuid.UUID, username string) *PlayerComponent {
	return &PlayerComponent{MemberID: memberID, Username: username}
}
