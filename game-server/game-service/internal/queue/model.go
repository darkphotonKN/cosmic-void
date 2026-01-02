package queue

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
)

type QueueService interface {
	Start()
	AddPlayerChan(player *types.Player)
	JoinQueue()
	MatchQueue()
	PlayerJoinQueue(player *types.Player)
	PlayerRemoveQueue(player *types.Player)
	GetMatchedChan() chan []*types.Player
	GetQueueStatusChan() chan QueueStatus
}
