package messaging

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
)

type MessageDispatcher interface {
	PushMessageToChannelQueue(playerID uuid.UUID, msg types.Message) error
}
