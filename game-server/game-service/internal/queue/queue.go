package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/common/utils/cache"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
)

/*
Player queue system - uses channel to listen for players joining matchmaking
*/

// QueueStatus used to notify queue status
type QueueStatus struct {
	Players []*types.Player
	Current int
	Total   int
}

type queueService struct {
	// how many people needed to start game
	matchSize       int
	MatchedChan     chan []*types.Player
	QueueStatusChan chan QueueStatus
	cacheService    cache.Cache
}

const queueKey = "game:matchmaking:queue"

func NewQueueService(matchSize int, cacheService cache.Cache) QueueService {
	return &queueService{
		matchSize:       matchSize,
		MatchedChan:     make(chan []*types.Player),
		QueueStatusChan: make(chan QueueStatus),
		cacheService:    cacheService,
	}
}

// Start launches queue listening
func (q *queueService) Start() {
	go q.MatchQueue()
	fmt.Println("QueueSystem started, listening for players...")
}

// AddPlayer adds player to matchmaking queue (via channel)
func (q *queueService) AddPlayer(player *types.Player) {
	q.PlayerJoinQueue(player)
}

// matchQueue checks queue once per second
func (q *queueService) MatchQueue() {
	fmt.Println("Listening for queue...")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		// send value from chan once per second
		case <-ticker.C:
			{
				ctx := context.Background()
				lockID, acquired, err := q.cacheService.AcquireLock(ctx, "matchmaking", 3*time.Second)
				if err != nil {
					slog.Error("failed to acquire lock", "error", err)
					continue
				}
				if !acquired {
					continue // 其他 instance 正在處理
				}

				// check queue
				queueLen, err := q.cacheService.LLen(ctx, queueKey)
				if err != nil {
					slog.Error("failed to get queue length", "error", err)
					q.cacheService.ReleaseLock(ctx, "matchmaking", lockID)
					continue
				}

				// player enough
				if int(queueLen) >= q.matchSize {
					matched := make([]*types.Player, 0, q.matchSize)
					for i := 0; i < q.matchSize; i++ {
						data, err := q.cacheService.LPop(ctx, queueKey)
						if err != nil {
							slog.Error("failed to pop player from queue", "error", err)
							break
						}
						var p types.Player
						if err := json.Unmarshal([]byte(data), &p); err != nil {
							slog.Error("failed to unmarshal player", "error", err)
							continue
						}
						matched = append(matched, &p)
					}

					q.cacheService.ReleaseLock(ctx, "matchmaking", lockID)

					if len(matched) == q.matchSize {
						fmt.Println("Match found!")
						q.MatchedChan <- matched
					}
					continue
				}
				// player not enough
				if queueLen > 0 {
					items, err := q.cacheService.LRange(ctx, queueKey, 0, -1)
					if err == nil {
						players := make([]*types.Player, 0, len(items))
						for _, item := range items {
							var p types.Player
							if err := json.Unmarshal([]byte(item), &p); err != nil {
								continue
							}
							players = append(players, &p)
						}
						fmt.Printf("Waiting: %d/%d\n", len(players), q.matchSize)
						go func() {
							q.QueueStatusChan <- QueueStatus{
								Players: players,
								Current: len(players),
								Total:   q.matchSize,
							}
						}()
					}
				}
				q.cacheService.ReleaseLock(ctx, "matchmaking", lockID)
			}
		}
	}
}

// handlePlayerJoinQueue handles logic for player joining queue
func (q *queueService) PlayerJoinQueue(player *types.Player) {
	ctx := context.Background()
	items, err := q.cacheService.LRange(ctx, queueKey, 0, -1)
	if err != nil {
		slog.Error("failed to check queue", "error", err)
		return
	}

	for _, item := range items {
		var p types.Player
		if err := json.Unmarshal([]byte(item), &p); err != nil {
			continue
		}
		if p.ID == player.ID {
			fmt.Println("player already exists", player.ID)
			return
		}
	}
	data, err := json.Marshal(player)
	if err != nil {
		slog.Error("failed to marshal player", "error", err)
		return
	}
	q.cacheService.RPush(ctx, queueKey, data)
	fmt.Printf("Player %s joined queue.\n", player.Username)
}

// TODO: disconnect remove player
func (q *queueService) PlayerRemoveQueue(player *types.Player) {
	ctx := context.Background()

	items, err := q.cacheService.LRange(ctx, queueKey, 0, -1)
	if err != nil {
		slog.Error("failed to get queue", "error", err)
		return
	}

	for _, item := range items {
		var p types.Player
		if err := json.Unmarshal([]byte(item), &p); err != nil {
			continue
		}
		if p.ID == player.ID {
			q.cacheService.LRem(ctx, queueKey, 1, item)
			return
		}
	}
}

func (q *queueService) GetMatchedChan() chan []*types.Player {
	return q.MatchedChan
}

func (q *queueService) GetQueueStatusChan() chan QueueStatus {
	return q.QueueStatusChan
}
