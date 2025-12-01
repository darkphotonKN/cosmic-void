package systems

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"


/*
 1. 條件判斷 (能不能攻擊)
 2. 目標選擇 (打誰)
 3. 傷害計算 (打多少)
 4. 結果應用 (扣血) - 使用 DamageCalculator 計算傷害
 5. 狀態更新 (冷卻、動畫等)
*/
type CombatSystem struct{}

func NewCombatSystem() *CombatSystem {
	return &CombatSystem{}
}

// NOTE: this runs every game tick
func (s *CombatSystem) Update(deltaTime float64, entities []*ecs.Entity) {
	// Combat logic to be implemented

}
