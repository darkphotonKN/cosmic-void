package gameserver

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	grpcauth "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/auth"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAuthClient for testing
type MockAuthClient struct{}

func (m *MockAuthClient) GetMember(ctx context.Context, req *pb.GetMemberRequest) (*pb.Member, error) {
	return &pb.Member{
		Id:    req.Id,
		Name:  "TestUser",
		Email: "test@test.com",
	}, nil
}

func (m *MockAuthClient) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	return &pb.ValidateTokenResponse{
		Valid:    true,
		MemberId: uuid.New().String(),
	}, nil
}

/**
* testing cross module functionality and behaviors.
**/

// TestServerHubSessionIntegration tests the full flow
// message sent to client sends to Server →  Hub routes →  Session receives
func TestServerHubSessionIntegration(t *testing.T) {
	mockAuthClient := &MockAuthClient{}
	server := NewServer(mockAuthClient)

	// create test players
	player1 := &types.Player{
		ID:       uuid.New(),
		Username: "TestPlayer1",
	}
	player2 := &types.Player{
		ID:       uuid.New(),
		Username: "TestPlayer2",
	}

	testPlayers := []*types.Player{player1, player2}

	// create game session through server
	session := server.CreateGameSession(testPlayers)
	session.TestMessageSpy = make(chan types.Message)

	require.NotNil(t, session, "Session should be created")

	// give goroutines time to start
	time.Sleep(100 * time.Millisecond)

	// clean up at end
	defer session.Shutdown()

	// send a game action message that should be routed to session
	clientMsg := types.Message{
		Action: string(constants.ActionMove),
		Payload: map[string]interface{}{
			"session_id": session.ID.String(),
			"player_id":  player1.ID.String(),
			"vx":         1.0,
			"vy":         0.5,
		},
	}

	clientPackage := types.ClientPackage{
		Message: clientMsg,
		Conn:    nil, // no real connection needed for this test
	}

	// simulating websocket server, send to servers channel
	// hub should received it at this point

	server.serverChan <- clientPackage

	// the game Session should receive the message after hub reroutes it
	select {
	case receivedPackage := <-session.MessageCh:
		assert.Equal(t, string(constants.ActionMove), receivedPackage.Message.Action)
		assert.Equal(t, session.ID.String(), receivedPackage.Message.Payload["session_id"])
		assert.Equal(t, 1.0, receivedPackage.Message.Payload["vx"])
		fmt.Printf("✅ Session received message: %+v\n", receivedPackage.Message)
	case <-time.After(2 * time.Second):
		t.Fatal("Message was not routed to session within timeout")
	}
}

func registerTestConn(s *Server, conn *websocket.Conn, player *types.Player) chan types.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connToPlayer[conn] = player
	s.players[player.ID] = player

	msgCh := make(chan types.Message, 10)
	s.msgChan[conn] = msgCh
	return msgCh
}

func TestQueueFindGameFlow(t *testing.T) {
	mockAuthClient := &MockAuthClient{}
	server := NewServer(mockAuthClient)

	playerCount := 10
	var wg sync.WaitGroup
	wg.Add(playerCount)

	for i := 1; i <= playerCount; i++ {
		time.Sleep(3 * time.Second)
		go func(idx int) {
			defer wg.Done()

			fakeConn := &websocket.Conn{}
			player := &types.Player{ID: uuid.New(), Username: fmt.Sprintf("Player%d", idx)}
			msgCh := registerTestConn(server, fakeConn, player)

			server.serverChan <- types.ClientPackage{
				Message: types.Message{
					Action: string(constants.ActionFindGame),
					Payload: map[string]interface{}{
						"ID":       player.ID,
						"Username": player.Username,
					},
				},
				Conn: fakeConn,
			}

			timeout := time.After(10 * time.Second)
			for {
				select {
				case msg := <-msgCh:
					if msg.Action == "game_found" {
						server.mu.RLock()
						currentSessions := len(server.sessions)
						server.mu.RUnlock()
						fmt.Printf("✅ Player%d 收到 game_found，目前 session 数量: %d\n", idx, currentSessions)
						return
					}
					// queue_status 继续等待
				case <-timeout:
					fmt.Printf("❌ Player%d 沒收到 game_found\n", idx)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(3 * time.Second)

	server.mu.RLock()
	sessionCount := len(server.sessions)
	server.mu.RUnlock()

	expectedSessions := playerCount / 2
	assert.Equal(t, expectedSessions, sessionCount)
	fmt.Println("總共創建遊戲數量", expectedSessions)
}
func TestResponseBuilderIntegration(t *testing.T) {
	mockAuthClient := &MockAuthClient{}
	server := NewServer(mockAuthClient)
	// create test players
	player1 := &types.Player{
		ID:       uuid.New(),
		Username: "TestPlayer1",
	}
	player2 := &types.Player{
		ID:       uuid.New(),
		Username: "TestPlayer2",
	}

	testPlayers := []*types.Player{player1, player2}

	// create game session through server
	session := server.CreateGameSession(testPlayers)

	clientMsg := types.Message{
		Action: string(constants.ActionMove),
		Payload: map[string]interface{}{
			"session_id": session.ID.String(),
			"player_id":  player1.ID.String(),
			"vx":         1.0,
			"vy":         0.0,
		},
	}
	clientPackage := types.ClientPackage{
		Message: clientMsg,
		Conn:    nil, // no real connection needed for this test
	}

	response := types.NewResponseBuilder()

	testConn := &Conn{}
	responseSuccess := response.Success(
		testConn,
		clientPackage.Message.Action,
		map[string]interface{}{"status": "ok"},
	)

	responseErr := response.Error(
		testConn,
		clientPackage.Message.Action,
		constants.ErrorInvalidSessionID,
		"Invalid session ID",
	)

	assert.Nil(t, responseSuccess, "Response Error method should not return error")
	assert.Nil(t, responseErr, "Response Error method should not return error")

}

type Conn struct{}

func (c *Conn) WriteJSON(v interface{}) error {
	fmt.Println("Mock WriteJSON")
	return nil
}

func TestSenderToBroadcastToPlayerList(t *testing.T) {
	serviceName := "game"
	consulAddr := commonhelpers.GetEnvString("CONSUL_ADDR", "localhost:8510")
	registry, err := consul.NewRegistry(consulAddr, serviceName)
	authClient := grpcauth.NewClient(registry)
	server := NewServer(authClient)

	newSender := types.NewMessageSender(server)
	// create test players
	player1 := &types.Player{
		ID:       uuid.New(),
		Username: "TestPlayer1",
	}
	player2 := &types.Player{
		ID:       uuid.New(),
		Username: "TestPlayer2",
	}
	player3 := &types.Player{
		ID:       uuid.New(),
		Username: "TestPlayer3",
	}

	testPlayers := []*types.Player{player1, player2, player3}
	newSender.BroadcastToPlayerList(testPlayers, types.Message{
		Action: string(constants.ActionFindGame),
		Payload: map[string]interface{}{
			"info": "This is a test broadcast message",
		},
	})

	assert.NotNil(t, err, "Broadcast should return error for missing player connection")
}
