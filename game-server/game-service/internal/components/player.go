package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type PlayerComponent struct {
	UserID   string
	Username string
}

func (p *PlayerComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypePlayer
}

func NewPlayerComponent(userID, username string) *PlayerComponent {
	return &PlayerComponent{UserID: userID, Username: username}
}
