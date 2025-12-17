package types

import (
	"fmt"

	"github.com/google/uuid"
)

/**
* MessageSender 提供統一的消息發送接口
* 所有組件（Hub, Session）都通過它發送消息
**/
type MessageSender struct {
	sendFunc func(playerID uuid.UUID, msg Message) error
}

func NewMessageSender(sendFunc func(uuid.UUID, Message) error) *MessageSender {
	return &MessageSender{sendFunc: sendFunc}
}

// SendToPlayer 發送給單一玩家
func (s *MessageSender) SendToPlayer(playerID uuid.UUID, action string, payload map[string]any) error {
	msg := Message{
		Action:  action,
		Payload: payload,
	}
	return s.sendFunc(playerID, msg)
}

// SendMessage 直接發送 Message（給 Hub 使用）
func (s *MessageSender) SendMessage(playerID uuid.UUID, msg Message) error {
	return s.sendFunc(playerID, msg)
}

// BroadcastToPlayers 廣播給多個玩家
func (s *MessageSender) BroadcastToPlayers(playerIDs []uuid.UUID, action string, payload map[string]any) error {
	var errs []error
	for _, pid := range playerIDs {
		if err := s.SendToPlayer(pid, action, payload); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("broadcast failed for %d players", len(errs))
	}
	return nil
}
