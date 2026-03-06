package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type SwitchComponent struct {
	SwitchID    int
	IsActivated bool
}

func (s *SwitchComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeSwitch
}

func NewSwitchComponent(switchID int) *SwitchComponent {
	return &SwitchComponent{
		SwitchID: switchID,
		IsActivated: false,
	}
}
