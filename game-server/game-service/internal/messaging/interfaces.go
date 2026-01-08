package messaging

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageDispatcher interface {
	PushMessageToChannelQueue(playerID uuid.UUID, msg interface{}) error
	PushMessageToConn(conn *websocket.Conn, msg interface{}) error
}
