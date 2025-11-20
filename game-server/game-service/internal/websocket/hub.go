package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	clients    map[*Client]bool
	rooms      map[uuid.UUID]map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[uuid.UUID]map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add to global clients
	h.clients[client] = true

	// Add to room
	if client.roomID != uuid.Nil {
		if h.rooms[client.roomID] == nil {
			h.rooms[client.roomID] = make(map[*Client]bool)
		}
		h.rooms[client.roomID][client] = true
	}

	log.Printf("Client %s joined room %s", client.id, client.roomID)

	// Send welcome message
	welcomeMsg := Message{
		Type:    "welcome",
		Payload: map[string]string{"message": "Connected to game server"},
		RoomID:  client.roomID.String(),
	}

	if msgBytes, err := json.Marshal(welcomeMsg); err == nil {
		select {
		case client.send <- msgBytes:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		// Remove from global clients
		delete(h.clients, client)
		close(client.send)

		// Remove from room
		if client.roomID != uuid.Nil {
			if roomClients, exists := h.rooms[client.roomID]; exists {
				delete(roomClients, client)

				// Clean up empty room
				if len(roomClients) == 0 {
					delete(h.rooms, client.roomID)
				}
			}
		}

		log.Printf("Client %s left room %s", client.id, client.roomID)
	}
}

func (h *Hub) broadcastMessage(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) BroadcastToRoom(roomID uuid.UUID, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if roomClients, exists := h.rooms[roomID]; exists {
		for client := range roomClients {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
				delete(roomClients, client)
			}
		}
	}
}

func (h *Hub) GetRoomClientCount(roomID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if roomClients, exists := h.rooms[roomID]; exists {
		return len(roomClients)
	}
	return 0
}

func (h *Hub) GetClientByUserID(userID uuid.UUID) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.userID == userID {
			return client
		}
	}
	return nil
}

func (h *Hub) SendToUser(userID uuid.UUID, message []byte) bool {
	client := h.GetClientByUserID(userID)
	if client == nil {
		return false
	}

	select {
	case client.send <- message:
		return true
	default:
		return false
	}
}