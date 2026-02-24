package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type HealthComponent struct {
	CurrentHealth int
	MaxHealth     int
	IsEliminated  bool
}

func (h *HealthComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeHealth
}

func NewHealthComponent(currentHealth, maxHealth int) *HealthComponent {
	return &HealthComponent{CurrentHealth: currentHealth, MaxHealth: maxHealth, IsEliminated: false}
}
