package systems

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

/*
1. condition check (can attack or not)
2. target selection (who to hit)
3. damage calculation (how much damage)
4. result application (deduct health) - use DamageCalculator to calculate damage
5. state update (cooldown, animation, etc)
*/
type CombatSystem struct{}

func NewCombatSystem() *CombatSystem {
	return &CombatSystem{}
}

// NOTE: this runs every game tick
func (s *CombatSystem) Update(deltaTime float64, entities []*ecs.Entity) {
	// Combat logic to be implemented

}
