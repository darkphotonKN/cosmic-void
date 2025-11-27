package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type AttackComponent struct {
	Strength int
}

func (a *AttackComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeAttack
}

func NewAttackComponent(strength int) *AttackComponent {
	return &AttackComponent{Strength: strength}
}
