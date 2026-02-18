package systems

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

/*
 1. check if skill should be cast
 2. check skill casting conditions
    - is skill cooldown complete?
    - are there enough resources (MP, energy, etc, if you have them)
3. execute different logic based on skill type
   switch skillName {
   case "Fireball":
        - calculate skill damage (usually based on Intelligence)
        - select target
        - apply damage
        - might have AOE range damage
    case "Heal":
        - not damage, is healing
        - select friendly target
        - restore Health
    case "Shield":
        - not damage, is adding Buff
        - add Buff component to target
    case "Poison":
        - add Debuff component (damage over time)
4. update skill cooldown time
5. consume resources (if there's MP system)
*/

type SkillSystem struct{}

func NewSkillSystem() *SkillSystem {
	return &SkillSystem{}
}

func (s *SkillSystem) Update(deltaTime float64, entities []*ecs.Entity) {
	// Skill logic to be implemented

}
