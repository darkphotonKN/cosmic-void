package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development
		return true
	},
}

type Client struct {
	id     uuid.UUID
	userID uuid.UUID
	roomID uuid.UUID
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
}

func NewClient(hub *Hub, conn *websocket.Conn, userID, roomID uuid.UUID) *Client {
	return &Client{
		id:     uuid.New(),
		userID: userID,
		roomID: roomID,
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse and handle the message
		c.handleMessage(messageBytes)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(messageBytes []byte) {
	var msg Message
	if err := json.Unmarshal(messageBytes, &msg); err != nil {
		log.Printf("Error parsing message: %v", err)
		return
	}

	// Set the sender info
	msg.SenderID = c.userID.String()
	msg.RoomID = c.roomID.String()

	switch msg.Type {
	case "player_move":
		c.handlePlayerMove(msg)
	case "player_action":
		c.handlePlayerAction(msg)
	case "chat_message":
		c.handleChatMessage(msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (c *Client) handlePlayerMove(msg Message) {
	// Broadcast move to all clients in the same room
	if msgBytes, err := json.Marshal(msg); err == nil {
		c.hub.BroadcastToRoom(c.roomID, msgBytes)
	}
}

func (c *Client) handlePlayerAction(msg Message) {
	// Broadcast action to all clients in the same room
	if msgBytes, err := json.Marshal(msg); err == nil {
		c.hub.BroadcastToRoom(c.roomID, msgBytes)
	}
}

func (c *Client) handleChatMessage(msg Message) {
	// Add timestamp to chat message
	msg.Timestamp = time.Now().Format(time.RFC3339)

	// Broadcast chat to all clients in the same room
	if msgBytes, err := json.Marshal(msg); err == nil {
		c.hub.BroadcastToRoom(c.roomID, msgBytes)
	}
}