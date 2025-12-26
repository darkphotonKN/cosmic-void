package types

import (
	"fmt"

	"github.com/google/uuid"
)

/**
* MessageSender 提供統一的消息發送接口
* 所有組件（Hub, Session）都通過它發送消息
**/

type SenderInterface interface {
	SendMessageInternal(playerID uuid.UUID, msg Message) error
}

type MessageSender struct {
	sender SenderInterface
}

func NewMessageSender(server SenderInterface) *MessageSender {
	return &MessageSender{
		sender: server,
	}
}

// SendToPlayer 發送給單一玩家
func (s *MessageSender) SendToPlayer(playerID uuid.UUID, message Message) error {
	fmt.Println("Sending message to player:", playerID)
	msg := Message{
		Action:  message.Action,
		Payload: message.Payload,
	}
	return s.sender.SendMessageInternal(playerID, msg)
}

// SendMessage 直接發送 Message（給 Hub 使用）
func (s *MessageSender) SendMessage(playerID uuid.UUID, msg Message) error {
	return s.sender.SendMessageInternal(playerID, msg)
}

// BroadcastToPlayerList 廣播給多個玩家（直接使用 Player list）
func (s *MessageSender) BroadcastToPlayerList(players []*Player, msg Message) error {
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
