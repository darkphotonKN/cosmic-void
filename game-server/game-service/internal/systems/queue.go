package systems

import (
	"fmt"
	"sync"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
)

/*
Player queue system - 使用 channel 監聽玩家加入配對
*/

// QueueStatus 用於通知排隊狀態
type QueueStatus struct {
	Players []*types.Player
	Current int
	Total   int
}

type QueueSystem struct {
	// 接收要加入配對的玩家
	playerChan chan *types.Player
	queue      []*types.Player
	// 需要多少人才能開始遊戲
	matchSize int

	mu sync.RWMutex

	MatchedChan     chan []*types.Player
	QueueStatusChan chan QueueStatus
}

func NewQueueSystem(matchSize int) *QueueSystem {
	return &QueueSystem{
		playerChan:      make(chan *types.Player),
		matchSize:       matchSize,
		queue:           make([]*types.Player, 0),
		MatchedChan:     make(chan []*types.Player),
		QueueStatusChan: make(chan QueueStatus),
	}
}

// Start 啟動 queue 監聽
func (q *QueueSystem) Start() {
	go q.matchQueue()
	go q.JoinQueue()
	fmt.Println("QueueSystem started, listening for players...")
}

// AddPlayer 將玩家加入配對 queue（透過 channel）
func (q *QueueSystem) AddPlayerChan(player *types.Player) {
	q.playerChan <- player
}

func (q *QueueSystem) JoinQueue() {
	for {
		select {
		case player := <-q.playerChan:
			q.PlayerJoinQueue(player)
		}
	}
}

// matchQueue 每秒檢查一次 queue
func (q *QueueSystem) matchQueue() {
	fmt.Println("Listening for queue...")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		// fmt.Println("match queue")

		select {
		// 每秒從chan送一次值
		case <-ticker.C:
			q.mu.Lock()
			sizeLen := len(q.queue) >= q.matchSize
			q.mu.Unlock()
			defer q.mu.Unlock()
			// 人數滿了
			if sizeLen {
				matched := make([]*types.Player, q.matchSize)
				q.mu.Lock()
				// 取前兩個
				copy(matched, q.queue[:q.matchSize])
				// 移除前兩個
				q.queue = q.queue[q.matchSize:]
				q.mu.Unlock()
				fmt.Println("Match found!")
				q.MatchedChan <- matched
				continue
			}
			// 人數不足，通知玩家目前排隊人數
			if len(q.queue) > 0 {
				fmt.Printf("Waiting: %d/%d\n", len(q.queue), q.matchSize)
				// 複製一份 queue 發送狀態
				q.mu.Lock()
				playersCopy := make([]*types.Player, len(q.queue))
				copy(playersCopy, q.queue)
				q.mu.Unlock()

				// 發送到 QueueStatusChan（用 goroutine 避免阻塞）
				go func() {
					q.QueueStatusChan <- QueueStatus{
						Players: playersCopy,
						Current: len(playersCopy),
						Total:   q.matchSize,
					}
				}()
				continue
			}

		}
	}
}

// handlePlayerJoinQueue 處理玩家加入 queue 的邏輯
func (q *QueueSystem) PlayerJoinQueue(player *types.Player) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// check if player in queue
	for _, p := range q.queue {
		if p.ID == player.ID {
			fmt.Println("player already exists", player.ID)
			return
		}
	}

	// 加入 queue
	q.queue = append(q.queue, player)
	fmt.Printf("Player %s joined queue. Waiting: %d/%d\n", player.Username, len(q.queue), q.matchSize)
}

// TODO: discconnect remove player
func (q *QueueSystem) PlayerRemoveQueue(player *types.Player) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, queue := range q.queue {
		if queue.ID == player.ID {
			q.queue = append(q.queue[:i], q.queue[i+1:]...)
			return
		}
	}
}
