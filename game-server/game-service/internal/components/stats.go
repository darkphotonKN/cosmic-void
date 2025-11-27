package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type StatsComponent struct {
	Level        int
	Experience   int
	Strength     int
	Agility      int
	Intelligence int
}

func (s *StatsComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeStats
}

func NewStatsComponent() *StatsComponent {
	return &StatsComponent{
		Level:        1,
		Experience:   0,
		Strength:     10,
		Agility:      10,
		Intelligence: 10,
	}
}
