package systems

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"

/*
	pure math calculation
*/

type DamageCalculator struct{}

func NewDamageCalculator() *DamageCalculator {
	return &DamageCalculator{}
}

func (dc *DamageCalculator) CalculatePhysicalDamage(attackerStats *components.StatsComponent, defenderStats *components.StatsComponent) int {
	baseDamage := attackerStats.Strength * 2
	defense := defenderStats.Agility / 2
	finalDamage := baseDamage - defense

	if finalDamage < 1 {
		return 1
	}
	return finalDamage
}

func (dc *DamageCalculator) CalculateMagicalDamage(attackerStats *components.StatsComponent,
	skillLevel int) int {
	return attackerStats.Intelligence * skillLevel * 3
}
