package messaging

import (
	"fmt"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/**
* MessageSender
* Responsible for all message formatting and message dispatching back to the client.
**/

type MessageSender struct {
	dispatcher MessageDispatcher
}

func NewMessageSender(dispatcher MessageDispatcher) *MessageSender {
	return &MessageSender{
		dispatcher: dispatcher,
	}
}

/**
* Sends to a single player after packaging the state and formatting the response to
* a consistent format.
**/
func (s *MessageSender) SendMessageToPlayer(playerID uuid.UUID, message types.Message) error {
	// fmt.Println("Sending message to player:", playerID)

	msg := types.Message{
		Action:  message.Action,
		Payload: message.Payload,
	}
	return s.dispatcher.PushMessageToChannelQueue(playerID, msg)
}

// SendMessage 直接發送 Message（沒有player Id）
func (s *MessageSender) SendMessageToConn(conn *websocket.Conn, msg types.Message) error {

	return s.dispatcher.PushMessageToConn(conn, msg)
}

// BroadcastToPlayerList 廣播給多個玩家（直接使用 Player list）
func (s *MessageSender) BroadcastToPlayerList(players []uuid.UUID, msg types.Message) error {
	var errs []error
	for _, player := range players {
		if err := s.SendMessageToPlayer(player, msg); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("broadcast failed for %d players", len(errs))
	}
	return nil
}

// state response specific helpers

func (s *MessageSender) SendStateToPlayer(playerID uuid.UUID, clientState *types.ClientGameState) error {
	return s.dispatcher.PushMessageToChannelQueue(playerID, clientState)
}

func (s *MessageSender) BroadcastStateToPlayerList(playerIds []uuid.UUID, state *types.ClientGameState) error {
	var errs []error
	for _, playerId := range playerIds {
		if err := s.SendStateToPlayer(playerId, state); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("broadcast failed for %d players", len(errs))
	}
	return nil

}
