package gameserver

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/**
* Represents the core game server, intializing the goroutines that
* talk to each other and coordinate all game sessions and websocket
* connections.
**/

type Server struct {
	// config
	upgrader websocket.Upgrader

	// messages
	// main server channel
	serverChan chan ClientPackage

	// active game message channels
	msgChan map[*websocket.Conn]chan Message

	// active sessions
	// [sessionId] to active sessions
	sessions map[uuid.UUID]*game.Session

	// online players
	players map[uuid.UUID]*game.Player

	// active connections to player maps
	connToPlayer map[*websocket.Conn]*game.Player

	// other
	mu sync.Mutex
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

		serverChan: make(chan ClientPackage, 0),
		msgChan:    make(map[*websocket.Conn]chan Message, 10),

		sessions:     make(map[uuid.UUID]*game.Session, 0),
		players:      make(map[uuid.UUID]*game.Player, 0),
		connToPlayer: make(map[*websocket.Conn]*game.Player, 0),
	}

	// initialize default setup
	server.InitServerSetup()

	return server
}

/**
* default pre-server setup tasks
**/
func (s *Server) InitServerSetup() {
	// start message hub concurrently with the same server channel
	// instance
	messageHub := NewMessageHub(s.serverChan)
	go messageHub.Run()
}

/**
* maps a connected client to its player information
**/
func (s *Server) MapConnToPlayer(conn *websocket.Conn, player game.Player) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connToPlayer[conn] = &player
}

/**
* grabs player information from connected client's websocket connection
* information.
**/

func (s *Server) GetPlayerFromConn(conn *websocket.Conn) (*game.Player, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	player, exists := s.connToPlayer[conn]

	return player, exists
}

/**
* allows the creation of a new game session.
**/
func (s *Server) CreateGameSession(players []*game.Player) *game.Session {
	newSessionId := uuid.New()
	newGameSession := game.NewSession(newSessionId.String())

	for _, player := range players {
		newGameSession.AddPlayer(player.ID, player.Username)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[newSessionId] = newGameSession
	fmt.Printf("New game session initiated, id: %s\n", newSessionId)

	return newGameSession
}
