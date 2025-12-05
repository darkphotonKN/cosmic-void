package systems

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
)

/*
	player queue system
*/

type QueueSystem struct{}

func NewQueueSystem() *QueueSystem {
	return &QueueSystem{}
}

func (q *QueueSystem) AddPlayerToQueue(player *types.Player) {

}
