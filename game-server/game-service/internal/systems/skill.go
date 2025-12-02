package systems

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

/*
 1. 檢查是否要施放技能
 2. 檢查技能施放條件
    - 技能冷卻是否完成？
    - 是否有足夠的資源（MP, 能量等，如果你有）
3. 根據技能類型執行不同邏輯
   switch skillName {
   case "Fireball":
        - 計算技能傷害（通常基於 Intelligence）
        - 選擇目標
        - 應用傷害
        - 可能有 AOE 範圍傷害
    case "Heal":
        - 不是傷害，是加血
        - 選擇友方目標
        - 恢復 Health
    case "Shield":
        - 不是傷害，是加 Buff
        - 給目標添加 Buff component
    case "Poison":
        - 添加 Debuff component（持續傷害）
4. 更新技能冷卻時間
5. 消耗資源（如果有 MP 系統）
*/

type SkillSystem struct{}

func NewSkillSystem() *SkillSystem {
	return &SkillSystem{}
}

func (s *SkillSystem) Update(deltaTime float64, entities []*ecs.Entity) {
	// Skill logic to be implemented

}
