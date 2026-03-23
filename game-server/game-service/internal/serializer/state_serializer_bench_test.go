package serializer

import (
	"context"
	"fmt"
	"testing"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
)

// setupTestEntities 創建一個完整的測試場景，包含玩家、容器、道具等
// 參數：
//   - b: testing.B 用於報告錯誤
//   - playerCount: 要創建的玩家數量
//   - itemsPerPlayer: 每個玩家庫存中的道具數量
//   - containerCount: 要創建的容器數量
//   - itemsPerContainer: 每個容器中的道具數量
func setupTestEntities(b *testing.B, playerCount, itemsPerPlayer, containerCount, itemsPerContainer int) (*ecs.EntityManager, []*ecs.Entity) {
	em := ecs.NewEntityManager()
	entities := make([]*ecs.Entity, 0)

	// ==========================================
	// 第一部分：創建玩家實體 (Player Entities)
	// ==========================================
	for i := 0; i < playerCount; i++ {
		// 1. 創建玩家實體
		playerEntity := em.CreateEntity()

		// 2. 添加 PlayerComponent（玩家核心數據）
		playerComp := &components.PlayerComponent{
			MemberID:             uuid.New(),
			Username:             fmt.Sprintf("TestPlayer_%d", i),
			HasHit:               false,
			AttackActive:         false,
			AttackCooldown:       0.0,
			AttackTargetEntityID: uuid.Nil,
			Escape:               false,
		}
		playerEntity.AddComponent(playerComp)

		// 3. 添加 TransformComponent（位置信息）
		// 把玩家散布在地圖上，避免重疊
		transformComp := &components.TransformComponent{
			X: float64(i*150 + 100), // 橫向排列，間距 150
			Y: float64(i*150 + 100),
		}
		playerEntity.AddComponent(transformComp)

		// 4. 添加 VelocityComponent（移動速度）
		velocityComp := &components.VelocityComponent{
			VX:    0.0, // 初始靜止
			VY:    0.0,
			Speed: 5.0 + float64(i)*0.5, // 每個玩家速度略有不同
		}
		playerEntity.AddComponent(velocityComp)

		// 5. 創建玩家的庫存道具
		itemIDs := make([]uuid.UUID, 0, itemsPerPlayer)

		for j := 0; j < itemsPerPlayer; j++ {
			// 創建道具實體
			itemEntity := em.CreateEntity()

			// 決定道具類型（輪流創建不同類型）
			var itemType types.ItemType
			var itemName string
			var itemComp *components.ItemComponent

			switch j % 3 {
			case 0: // 武器
				itemType = types.ItemTypeWeapon
				itemName = fmt.Sprintf("Sword_%d", j)
				itemComp = &components.ItemComponent{
					TemplateID:   uuid.New(),
					ItemType:     itemType,
					Name:         itemName,
					AttackPower:  10 + j*2, // 攻擊力遞增
					CriticalRate: 0.15 + float64(j)*0.05,
					WeaponType:   "melee",
					BuyPrice:     100 + j*10,
					SellPrice:    50 + j*5,
					Description:  fmt.Sprintf("A powerful sword #%d", j),
				}

			case 1: // 防具
				itemType = types.ItemTypeArmor
				itemName = fmt.Sprintf("Armor_%d", j)
				itemComp = &components.ItemComponent{
					TemplateID:      uuid.New(),
					ItemType:        itemType,
					Name:            itemName,
					DefenseRating:   15 + j*3,
					MagicResistance: 10 + j*2,
					ArmorSlot:       "chest",
					BuyPrice:        150 + j*15,
					SellPrice:       75 + j*7,
					Description:     fmt.Sprintf("Sturdy armor #%d", j),
				}

			case 2: // 消耗品
				itemType = types.ItemTypeConsumable
				itemName = fmt.Sprintf("Potion_%d", j)
				itemComp = &components.ItemComponent{
					TemplateID:    uuid.New(),
					ItemType:      itemType,
					Name:          itemName,
					HealingAmount: 50 + j*10,
					ManaAmount:    30 + j*5,
					BuffDuration:  5 + j,
					BuyPrice:      20 + j*2,
					SellPrice:     10 + j,
					Description:   fmt.Sprintf("Healing potion #%d", j),
				}
			}

			itemEntity.AddComponent(itemComp)
			itemIDs = append(itemIDs, itemEntity.ID)
			entities = append(entities, itemEntity)
		}

		// 6. 添加 ItemIDListComponent（玩家的庫存列表）
		itemIDListComp := &components.ItemIDListComponent{
			ItemIDs: itemIDs,
		}
		playerEntity.AddComponent(itemIDListComp)

		// 把玩家實體加入列表
		entities = append(entities, playerEntity)
	}

	// ==========================================
	// 第二部分：創建容器實體 (Container Entities)
	// ==========================================
	for i := 0; i < containerCount; i++ {
		// 1. 創建容器實體
		containerEntity := em.CreateEntity()

		// 2. 添加 ContainerComponent
		containerComp := &components.ContainerComponent{
			ContainerID: uuid.New(),
		}
		containerEntity.AddComponent(containerComp)

		// 3. 添加 TransformComponent（容器位置）
		// 放置在地圖的其他位置
		containerTransformComp := &components.TransformComponent{
			X: float64(i*200 + 500), // 橫向排列，從 x=500 開始
			Y: float64(i*200 + 500),
		}
		containerEntity.AddComponent(containerTransformComp)

		// 4. 添加 OpenableComponent（可開啟狀態）
		openableComp := &components.OpenableComponent{
			IsOpen:        i%2 == 0, // 一半開啟，一半關閉
			HasBeenOpened: i%2 == 0, // 開啟的已經被打開過
		}
		containerEntity.AddComponent(openableComp)

		// 5. 創建容器內的道具
		containerItemIDs := make([]uuid.UUID, 0, itemsPerContainer)

		for j := 0; j < itemsPerContainer; j++ {
			// 創建容器內的道具實體
			containerItemEntity := em.CreateEntity()

			// 創建不同類型的道具
			itemType := types.ItemTypeWeapon
			if j%2 == 0 {
				itemType = types.ItemTypeConsumable
			}

			containerItemComp := &components.ItemComponent{
				TemplateID:    uuid.New(),
				ItemType:      itemType,
				Name:          fmt.Sprintf("ContainerItem_C%d_I%d", i, j),
				AttackPower:   5 + j,
				HealingAmount: 25 + j*5,
				BuyPrice:      50 + j*5,
				SellPrice:     25 + j*2,
				Description:   fmt.Sprintf("Item from container %d, slot %d", i, j),
			}
			containerItemEntity.AddComponent(containerItemComp)

			containerItemIDs = append(containerItemIDs, containerItemEntity.ID)
			entities = append(entities, containerItemEntity)
		}

		// 6. 添加 ItemIDListComponent（容器的道具列表）
		containerItemIDListComp := &components.ItemIDListComponent{
			ItemIDs: containerItemIDs,
		}
		containerEntity.AddComponent(containerItemIDListComp)

		// 把容器實體加入列表
		entities = append(entities, containerEntity)
	}

	// ==========================================
	// 第三部分：創建逃生門實體 (Escape Door Entity)
	// ==========================================
	escapeDoorEntity := em.CreateEntity()

	// 1. 添加 EscapeDoorComponent
	escapeDoorComp := &components.EscapeDoorComponent{}
	escapeDoorEntity.AddComponent(escapeDoorComp)

	// 2. 添加 TransformComponent
	escapeDoorTransformComp := &components.TransformComponent{
		X: 1000.0, // 放在地圖遠端
		Y: 1000.0,
	}
	escapeDoorEntity.AddComponent(escapeDoorTransformComp)

	// 3. 添加 OpenableComponent
	escapeDoorOpenableComp := &components.OpenableComponent{
		IsOpen:        false, // 初始關閉
		HasBeenOpened: false,
	}
	escapeDoorEntity.AddComponent(escapeDoorOpenableComp)

	// 4. 添加 LockableComponent
	escapeDoorLockableComp := &components.LockableComponents{
		IsLocked: true, // 初始鎖定
	}
	escapeDoorEntity.AddComponent(escapeDoorLockableComp)

	entities = append(entities, escapeDoorEntity)

	// ==========================================
	// 第四部分：創建開關實體 (Switch Entities)
	// ==========================================
	switchCount := 3 // 創建 3 個開關
	for i := 0; i < switchCount; i++ {
		switchEntity := em.CreateEntity()

		// 1. 添加 SwitchComponent
		switchComp := &components.SwitchComponent{
			SwitchID:    i + 1,
			IsActivated: i == 0, // 第一個開關已激活
		}
		switchEntity.AddComponent(switchComp)

		// 2. 添加 TransformComponent
		switchTransformComp := &components.TransformComponent{
			X: float64(i*100 + 800),
			Y: 500.0,
		}
		switchEntity.AddComponent(switchTransformComp)

		entities = append(entities, switchEntity)
	}

	// ==========================================
	// 第五部分：創建一些獨立的道具實體（地板掉落物）
	// ==========================================
	groundItemCount := 5
	for i := 0; i < groundItemCount; i++ {
		groundItemEntity := em.CreateEntity()

		groundItemComp := &components.ItemComponent{
			TemplateID:    uuid.New(),
			ItemType:      types.ItemTypeConsumable,
			Name:          fmt.Sprintf("GroundItem_%d", i),
			HealingAmount: 30,
			BuyPrice:      15,
			SellPrice:     7,
			Description:   "An item found on the ground",
		}
		groundItemEntity.AddComponent(groundItemComp)

		// 給地板道具也添加位置
		groundItemTransformComp := &components.TransformComponent{
			X: float64(i*80 + 300),
			Y: 200.0,
		}
		groundItemEntity.AddComponent(groundItemTransformComp)

		entities = append(entities, groundItemEntity)
	}

	return em, entities
}

func BenchmarkSerializeBackendState_NoPool(b *testing.B) {
	em, entities := setupTestEntities(b, 10, 5, 5, 3)
	serializer := NewStateSerializer(em)
	ctx := context.Background()
	sessionID := uuid.New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := serializer.SerializeBackendState(ctx, sessionID, entities)
		if err != nil {
			b.Fatal(err)
		}
	}

}

func BenchmarkSerializeBackendState_WithPool(b *testing.B) {
	em, entities := setupTestEntities(b, 7, 6, 5, 6)
	serializer := NewStateSerializer(em)
	ctx := context.Background()
	sessionID := uuid.New()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		state, err := serializer.SerializeBackendState(ctx, sessionID, entities)
		if err != nil {
			b.Fatal(err)
		}

		serializer.PutBackendState(state)
	}
}
