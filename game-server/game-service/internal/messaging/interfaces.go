package messaging

import (
	"github.com/google/uuid"
)

type MessageDispatcher interface {
	PushMessageToChannelQueue(playerID uuid.UUID, msg interface{}) error
}
