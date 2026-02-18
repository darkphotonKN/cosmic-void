package queue

import (
	"fmt"
	"sync"
	"time"

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
	// receive players to join matchmaking
	playerChan chan *types.Player
	queue      []*types.Player
	// how many people needed to start game
	matchSize int

	mu sync.RWMutex

	MatchedChan     chan []*types.Player
	QueueStatusChan chan QueueStatus
}

func NewQueueService(matchSize int) QueueService {
	return &queueService{
		playerChan:      make(chan *types.Player),
		matchSize:       matchSize,
		queue:           make([]*types.Player, 0),
		MatchedChan:     make(chan []*types.Player),
		QueueStatusChan: make(chan QueueStatus),
	}
}

// Start launches queue listening
func (q *queueService) Start() {
	go q.MatchQueue()
	go q.JoinQueue()
	fmt.Println("QueueSystem started, listening for players...")
}

// AddPlayer adds player to matchmaking queue (via channel)
func (q *queueService) AddPlayerChan(player *types.Player) {
	q.playerChan <- player
}

func (q *queueService) JoinQueue() {
	for {
		select {
		case player := <-q.playerChan:
			q.PlayerJoinQueue(player)
		}
	}
}

// matchQueue checks queue once per second
func (q *queueService) MatchQueue() {
	fmt.Println("Listening for queue...")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		// fmt.Println("match queue")

		select {
		// send value from chan once per second
		case <-ticker.C:
			q.mu.Lock()
			sizeLen := len(q.queue) >= q.matchSize
			q.mu.Unlock()
			defer q.mu.Unlock()
			// enough people
			if sizeLen {
				matched := make([]*types.Player, q.matchSize)
				q.mu.Lock()
				// take first two
				copy(matched, q.queue[:q.matchSize])
				// remove first two
				q.queue = q.queue[q.matchSize:]
				q.mu.Unlock()
				fmt.Println("Match found!")
				q.MatchedChan <- matched
				continue
			}
			// not enough people, notify players of current queue count
			if len(q.queue) > 0 {
				fmt.Printf("Waiting: %d/%d\n", len(q.queue), q.matchSize)
				// copy queue to send status
				q.mu.Lock()
				playersCopy := make([]*types.Player, len(q.queue))
				copy(playersCopy, q.queue)
				q.mu.Unlock()

				// send to QueueStatusChan (use goroutine to avoid blocking)
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

// handlePlayerJoinQueue handles logic for player joining queue
func (q *queueService) PlayerJoinQueue(player *types.Player) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// check if player in queue
	for _, p := range q.queue {
		if p.ID == player.ID {
			fmt.Println("player already exists", player.ID)
			return
		}
	}

	// join queue
	q.queue = append(q.queue, player)
	fmt.Printf("Player %s joined queue. Waiting: %d/%d\n", player.Username, len(q.queue), q.matchSize)
}

// TODO: disconnect remove player
func (q *queueService) PlayerRemoveQueue(player *types.Player) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, queue := range q.queue {
		if queue.ID == player.ID {
			q.queue = append(q.queue[:i], q.queue[i+1:]...)
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
