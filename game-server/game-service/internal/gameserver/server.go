package gameserver

import (
	"net/http"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
	"github.com/gorilla/websocket"
)

/**
* Represents the core game server, intializing the goroutines that
* talk to each other and coordinate all game sessions and websocket
* connections.
**/

type Server struct {

	// config
	ListenAddr string
	upgrader   websocket.Upgrader

	// active games
	games []*game.Game

	// online players
	players []*game.Player

	// injection
}

func NewServer(listenAddr string) *Server {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// TODO: Allow all connections by default for simplicity; can add more logic here
			return true
		},
	}

	return &Server{
		upgrader:   upgrader,
		ListenAddr: listenAddr,

		games:   make([]*game.Game, 0),
		players: make([]*game.Player, 0),
	}
}

func InitServer() {
	// start message hub concurrently
	messageHub := NewMessageHub()
	go messageHub.RunHub()
}
