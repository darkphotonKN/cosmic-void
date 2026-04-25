package queue

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
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

	mu      sync.Mutex
	players []*types.Player
}

func NewQueueService(matchSize int) *queueService {
	return &queueService{
		matchSize:       matchSize,
		MatchedChan:     make(chan []*types.Player),
		QueueStatusChan: make(chan QueueStatus),
		players:         make([]*types.Player, 0, matchSize),
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
				q.mu.Lock()

				// player enough
				if len(q.players) >= q.matchSize {
					matched := make([]*types.Player, q.matchSize)
					copy(matched, q.players[:q.matchSize])
					q.players = q.players[q.matchSize:]

					q.mu.Unlock()

					fmt.Println("Match found!")
					q.MatchedChan <- matched
					continue
				}
				// player not enough
				if len(q.players) > 0 {
					players := make([]*types.Player, len(q.players))
					copy(players, q.players)

					q.mu.Unlock()

					fmt.Printf("Waiting: %d/%d\n", len(players), q.matchSize)
					go func() {
						q.QueueStatusChan <- QueueStatus{
							Players: players,
							Current: len(players),
							Total:   q.matchSize,
						}
					}()
					continue
				}

				q.mu.Unlock()
			}
		}
	}
}

// handlePlayerJoinQueue handles logic for player joining queue
func (q *queueService) PlayerJoinQueue(player *types.Player) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, p := range q.players {
		if p.ID == player.ID {
			slog.Error("player already exists", player.ID)
			return game.ErrPlayerAlreadyInQueue
		}
	}
	q.players = append(q.players, player)
	slog.Debug("Player joined queue.",
		"player_username", player.Username,
	)
	return nil
}

// TODO: disconnect remove player
func (q *queueService) PlayerRemoveQueue(player *types.Player) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, p := range q.players {
		if p.ID == player.ID {
			q.players = append(q.players[:i], q.players[i+1:]...)
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
