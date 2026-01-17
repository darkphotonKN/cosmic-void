package gameserver

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	grpcauth "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/auth"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/messaging"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/queue"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/serializer"
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
	msgChan map[*websocket.Conn]chan interface{}

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

	queue queue.QueueService
	// auth client for gRPC calls
	authClient grpcauth.AuthClient

	// message broker communication channel
	eventEmitter game.EventEmitter
}

type MessageSender interface {
	BroadcastToPlayerList(players []*types.Player, msg types.Message) error
}

func NewServer(authClient grpcauth.AuthClient, queueService queue.QueueService, eventEmitter game.EventEmitter) *Server {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// TODO: Allow all connections by default for simplicity; can add more logic here
			return true
		},
	}

	server := &Server{
		upgrader: upgrader,

		serverChan: make(chan types.ClientPackage, 10),
		msgChan:    make(map[*websocket.Conn]chan interface{}, constants.MaxMsgChanBuffer),

		sessions:     make(map[uuid.UUID]*game.Session, 10),
		players:      make(map[uuid.UUID]*types.Player, 10),
		connToPlayer: make(map[*websocket.Conn]*types.Player, 10),

		queue:        queueService,
		authClient:   authClient,
		eventEmitter: eventEmitter,
	}

	// initialize message sender
	newSender := messaging.NewMessageSender(server)

	server.queue.Start()

	// initialize message hub
	messageHub := NewMessageHub(server, newSender)
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

	// 檢查是否有相同 player ID 的舊連線，如果有則清除
	for oldConn, existingPlayer := range s.connToPlayer {
		if existingPlayer.ID == player.ID && oldConn != conn {
			fmt.Printf("Player %s reconnected, cleaning up old connection\n", player.Username)
			// 關閉舊的 msgChan
			if ch, exists := s.msgChan[oldConn]; exists {
				close(ch)
				delete(s.msgChan, oldConn)
			}
			// 移除舊的 conn -> player 映射
			delete(s.connToPlayer, oldConn)
			break
		}
	}

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
	// create entity manager first so it can be shared
	entityManager := ecs.NewEntityManager()
	stateSerializer := serializer.NewStateSerializer(entityManager)

	// create session with message sender
	newGameSession := game.NewSession(messaging.NewMessageSender(s), stateSerializer, entityManager, s.eventEmitter)

	newGameSession.InitialMapObjects()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, player := range players {
		// 將玩家加入 session
		newGameSession.AddPlayer(player.ID, player.Username)
		connected := constants.Connected
		// 更新玩家的 SessionId
		player.CurrentGameSessionId = newGameSession.ID
		player.ConnectState = &connected

		// 同時更新 server 的 players map 中的玩家資訊 (如果存在)
		if existingPlayer, exists := s.players[player.ID]; exists {
			existingPlayer.CurrentGameSessionId = newGameSession.ID
			existingPlayer.ConnectState = &connected
		}
	}

	s.sessions[newGameSession.ID] = newGameSession
	fmt.Printf("New game session initiated, id: %s, players: %d\n", newGameSession.ID, len(players))

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
	s.queue.AddPlayerChan(player)
}

/**
* remove player from queue (delegates to QueueSystem)
**/
func (s *Server) RemovePlayerFromQueue(player *types.Player) {
	s.queue.PlayerRemoveQueue(player)
}

/**
* get matched channel for listening to matched players
**/
func (s *Server) GetMatchedChan() chan []*types.Player {
	return s.queue.GetMatchedChan()
}

/**
* get queue status channel for listening to queue updates
**/
func (s *Server) GetQueueStatusChan() chan queue.QueueStatus {
	return s.queue.GetQueueStatusChan()
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

/**
* PushMessageToChannelQueue
* Allows the server to sequentially pipe multiple messages into a single channel for sequential writes back to the client due to gorilla websockets constraint of max one concurrent writer with conn.
**/
func (s *Server) PushMessageToChannelQueue(playerID uuid.UUID, msg interface{}) error {
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

func (s *Server) PushMessageToConn(conn *websocket.Conn, msg interface{}) error {
	typeMsg, ok := msg.(types.Message)
	if !ok {
		return fmt.Errorf("invalid message type")
	}
	if conn == nil {
		fmt.Println("Warning: nil connection, skipping send")
		return nil
	}
	s.mu.RLock()
	ch, ok := s.msgChan[conn]
	s.mu.RUnlock()

	if !ok {
		fmt.Println("Warning: message channel not found for connection")
		return nil
	}

	select {
	case ch <- typeMsg:
		return nil
	default:
		return fmt.Errorf("message channel full for connection")
	}

}

/**
* returns the auth client for gRPC calls
**/
func (s *Server) GetAuthClient() grpcauth.AuthClient {
	return s.authClient
}
