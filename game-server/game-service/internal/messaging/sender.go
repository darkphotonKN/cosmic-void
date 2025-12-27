package messaging

import (
	"fmt"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/serializer"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
)

/**
* MessageSender 提供統一的消息發送接口
* 所有組件（Hub, Session）都通過它發送消息
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

// SendToPlayer 發送給單一玩家
func (s *MessageSender) SendToPlayer(playerID uuid.UUID, message types.Message) error {
	fmt.Println("Sending message to player:", playerID)
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
