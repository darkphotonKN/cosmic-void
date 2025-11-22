package gameserver

import (
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

	// active games
	// [gameId] to active game
	games map[uuid.UUID]*game.Game

	// online players
	players map[uuid.UUID]*game.Player

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

		games:   make(map[uuid.UUID]*game.Game, 0),
		players: make(map[uuid.UUID]*game.Player, 0),
	}

	// initialize default setup
	server.InitServer()

	return server
}

func (s *Server) InitServer() {
	// start message hub concurrently with the same server channel
	// instance
	messageHub := NewMessageHub(s.serverChan)
	go messageHub.Run()
}
