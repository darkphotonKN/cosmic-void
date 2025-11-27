package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type SkillComponent struct {
	SkillName string
	Level     int
}

func (s *SkillComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeSkill
}

func NewSkillComponent(skillName string, level int) *SkillComponent {
	return &SkillComponent{SkillName: skillName, Level: level}
}
