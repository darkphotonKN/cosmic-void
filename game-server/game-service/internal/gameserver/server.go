package gameserver

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/systems"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/**
* Represents the core game server, intializing the goroutines that
* talk to each other and coordinate all game sessions and websocket
* connections.
**/

type Server struct {
	upgrader   websocket.Upgrader
	serverChan chan types.ClientPackage

	// active game message channels
	msgChan map[*websocket.Conn]chan types.Message

	// active sessions
	// [sessionId] to active sessions
	sessions map[uuid.UUID]*game.Session

	// online players
	// [playerId] to player
	players map[uuid.UUID]*types.Player

	// websocket conn to player mapping
	// [active connections] to player
	connToPlayer map[*websocket.Conn]*types.Player

	mu sync.RWMutex

	// queue system
	queueSystem *systems.QueueSystem

	// unified message sender
	sender *types.MessageSender
}

func NewServer() *Server {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// TODO: Allow all connections by default for simplicity; can add more logic here
			return true
		},
	}

	server := &Server{
		upgrader: upgrader,

		serverChan: make(chan types.ClientPackage, 10),
		msgChan:    make(map[*websocket.Conn]chan types.Message, 10),

		sessions:     make(map[uuid.UUID]*game.Session, 10),
		players:      make(map[uuid.UUID]*types.Player, 10),
		connToPlayer: make(map[*websocket.Conn]*types.Player, 10),
	}

	// initialize message sender (inject send function)
	server.sender = types.NewMessageSender(server.sendMessageInternal)

	// initialize queue system
	server.queueSystem = systems.NewQueueSystem(2)
	server.queueSystem.Start()

	// initialize message hub
	messageHub := NewMessageHub(server, server.sender)
	go messageHub.Run()

	return server
}

/**
* exposes server chan for communication between server and client
**/
func (s *Server) GetServerChan() chan types.ClientPackage {
	return s.serverChan
}

/**
* maps a connected client to its player information
**/
func (s *Server) MapConnToPlayer(conn *websocket.Conn, player types.Player) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connToPlayer[conn] = &player
}

/**
* grabs player information from connected client's websocket connection
* information.
**/

func (s *Server) GetPlayerFromConn(conn *websocket.Conn) (*types.Player, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	player, exists := s.connToPlayer[conn]

	return player, exists

}

/**
* allows the creation of a new game session.
**/
func (s *Server) CreateGameSession(players []*types.Player) *game.Session {
	// create session with message sender
	newGameSession := game.NewSession(s.sender)

	for _, player := range players {
		newGameSession.AddPlayer(player.ID, player.Username)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[newGameSession.ID] = newGameSession
	fmt.Printf("New game session initiated, id: %s\n", newGameSession.ID)

	return newGameSession
}

/**
* allows the retrieval of an existing session.
**/
func (s *Server) GetGameSession(id uuid.UUID) (*game.Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[id]
	return session, exists
}

/**
* add player to queue (delegates to QueueSystem)
**/
func (s *Server) AddPlayerToQueue(player *types.Player) {
	s.queueSystem.AddPlayerChan(player)
}

/**
* get matched channel for listening to matched players
**/
func (s *Server) GetMatchedChan() chan []*types.Player {
	return s.queueSystem.MatchedChan
}

/**
* get queue status channel for listening to queue updates
**/
func (s *Server) GetQueueStatusChan() chan systems.QueueStatus {
	return s.queueSystem.QueueStatusChan
}

/**
* get conn from player ID
**/
func (s *Server) GetConnFromPlayer(playerID uuid.UUID) (*websocket.Conn, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for conn, player := range s.connToPlayer {
		if player.ID == playerID {
			return conn, true
		}
	}
	return nil, false
}

/**
* --- Internal Message Sending (used by MessageSender) ---
**/

// sendMessageInternal is the core function injected into MessageSender
func (s *Server) sendMessageInternal(playerID uuid.UUID, msg types.Message) error {
	conn, exists := s.GetConnFromPlayer(playerID)
	if !exists {
		return fmt.Errorf("player %s not found", playerID)
	}

	s.mu.RLock()
	ch, ok := s.msgChan[conn]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("message channel not found for player %s", playerID)
	}

	// non-blocking send to prevent slow clients from blocking
	select {
	case ch <- msg:
		return nil
	default:
		return fmt.Errorf("message channel full for player %s", playerID)
	}
}
