package messaging

import (
	"fmt"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/serializer"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
)

/**
* MessageSender
* Responsible for all message formatting and message dispatching back to the client.
**/

type MessageSender struct {
	dispatcher MessageDispatcher
	serializer serializer.StateSerializer
}

func NewMessageSender(dispatcher MessageDispatcher) *MessageSender {
	return &MessageSender{
		dispatcher: dispatcher,
	}
}

/**
* Sends to a single player after serializing the entire game state and formatting to
* the appropriate response format.
**/
func (s *MessageSender) SendToPlayer(playerID uuid.UUID, message types.Message) error {
	fmt.Println("Sending message to player:", playerID)

	// format to single shared state for client consumption
	msg := types.Message{
		Action:  message.Action,
		Payload: message.Payload,
	}
	return s.dispatcher.PushMessageToChannelQueue(playerID, msg)
}

// SendMessage 直接發送 Message（給 Hub 使用）
func (s *MessageSender) SendMessage(playerID uuid.UUID, msg types.Message) error {
	return s.dispatcher.PushMessageToChannelQueue(playerID, msg)
}

// BroadcastToPlayerList 廣播給多個玩家（直接使用 Player list）
func (s *MessageSender) BroadcastToPlayerList(players []*types.Player, msg types.Message) error {
	var errs []error
	for _, player := range players {
		if err := s.SendToPlayer(player.ID, msg); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("broadcast failed for %d players", len(errs))
	}
	return nil
}
