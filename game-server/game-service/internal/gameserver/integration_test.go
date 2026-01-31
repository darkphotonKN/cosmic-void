package gameserver

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	itemspb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/items"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commontypes "github.com/darkphotonKN/cosmic-void-server/common/types"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	grpcauth "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/auth"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/messaging"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/queue"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

var (
	ErrPlayerNotConnected     = errors.New("player not connected")
	ErrBroadcastFailed        = errors.New("broadcast failed")
	ErrAllPlayersFailed       = errors.New("broadcast failed for all players")
	ErrPartialBroadcastFailed = errors.New("broadcast partially failed")
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

// MockEventEmitter for testing
type MockEventEmitter struct{}

func (m *MockEventEmitter) PublishMatchComplete(ctx context.Context, data *commontypes.MatchEndState) error {
	return nil
}

// MockItemsClient for testing
type MockItemsClient struct{}

func (m *MockItemsClient) CreateWeapon(ctx context.Context, req *itemspb.CreateWeaponRequest) (*itemspb.Weapon, error) {
	return &itemspb.Weapon{
		Id:           uuid.New().String(),
		TypeId:       "test-type",
		RarityId:     "common",
		AttackPower:  15,
		Durability:   100,
		CriticalRate: 0.1,
		WeaponType:   "sword",
		Description:  "A test weapon",
	}, nil
}

func (m *MockItemsClient) GetWeaponWithTemplateByID(ctx context.Context, req *itemspb.GetWeaponRequest) (*itemspb.WeaponDetail, error) {
	return &itemspb.WeaponDetail{
		Id:             req.Id,
		TypeId:         "test-type",
		RarityId:       "common",
		AttackPower:    15,
		Durability:     100,
		CriticalRate:   0.1,
		WeaponType:     "sword",
		Description:    "A test weapon",
		ItemTemplateId: "test-template",
		ItemName:       "Test Sword",
		ItemCode:       "TEST_SWORD",
		IconUrl:        "/icons/test-sword.png",
		IsTradeable:    true,
		IsDroppable:    true,
		RequiredLevel:  1,
		BaseSellPrice:  100,
		BaseBuyPrice:   200,
	}, nil
}

func (m *MockItemsClient) ListWeaponsWithTemplate(ctx context.Context) (*itemspb.ListWeaponsResponse, error) {
	return &itemspb.ListWeaponsResponse{
		Weapons: []*itemspb.WeaponDetail{},
	}, nil
}

func (m *MockItemsClient) ListArmorsWithTemplate(ctx context.Context) (*itemspb.ListArmorsResponse, error) {
	return &itemspb.ListArmorsResponse{
		Armors: []*itemspb.ArmorDetail{},
	}, nil
}

func (m *MockItemsClient) ListConsumablesWithTemplate(ctx context.Context) (*itemspb.ListConsumablesResponse, error) {
	return &itemspb.ListConsumablesResponse{
		Consumables: []*itemspb.ConsumableDetail{},
	}, nil
}

type mockQueueService struct {
	players         []*types.Player
	matchedChan     chan []*types.Player
	statusChan      chan queue.QueueStatus
	QueueStatusChan chan queue.QueueStatus
}

func NewMockQueueService() *mockQueueService {
	return &mockQueueService{
		players:         make([]*types.Player, 0),
		matchedChan:     make(chan []*types.Player),
		statusChan:      make(chan queue.QueueStatus),
		QueueStatusChan: make(chan queue.QueueStatus),
	}
}
func (m *mockQueueService) JoinQueue()                    {}
func (m *mockQueueService) PlayerJoinQueue(*types.Player) {}

func (m *mockQueueService) GetQueueStatusChan() chan queue.QueueStatus {
	return m.QueueStatusChan
}
func (m *mockQueueService) AddPlayerChan(player *types.Player) {

}

func (m *mockQueueService) PlayerRemoveQueue(player *types.Player) {}
func (m *mockQueueService) MatchQueue()                            {}
func (m *mockQueueService) Start() {
	// 測試時不需要真的啟動
}

func (m *mockQueueService) GetMatchedChan() chan []*types.Player {
	return m.matchedChan
}

/**
* testing cross module functionality and behaviors.
**/

// TODO: doesn't work atm, FIX
// TestServerHubSessionIntegration tests the full flow
// message sent to client sends to Server →  Hub routes →  Session receives
// func TestServerHubSessionIntegration(t *testing.T) {
// 	mockAuthClient := &MockAuthClient{}
// 	mockQueue := NewMockQueueService()
// 	mockEventEmitter := &MockEventEmitter{}
// 	mockItemsClient := &MockItemsClient{}
// 	server := NewServer(mockAuthClient, mockQueue, mockEventEmitter, mockItemsClient)
//
// 	// create test players
// 	player1 := &types.Player{
// 		ID:       uuid.New(),
// 		Username: "TestPlayer1",
// 	}
//
// 	player2 := &types.Player{
// 		ID:       uuid.New(),
// 		Username: "TestPlayer2",
// 	}
//
// 	testPlayers := []*types.Player{player1, player2}
//
// 	// create game session through server
// 	session := server.CreateGameSession(testPlayers)
// 	session.TestMessageSpy = make(chan types.Message)
//
// 	require.NotNil(t, session, "Session should be created")
//
// 	// give goroutines time to start
// 	time.Sleep(100 * time.Millisecond)
//
// 	// clean up at end
// 	defer session.Shutdown()
//
// 	// send a game action message that should be routed to session
// 	clientMsg := types.Message{
// 		Action: string(constants.ActionMove),
// 		Payload: map[string]interface{}{
// 			"session_id": session.ID.String(),
// 			"player_id":  player1.ID.String(),
// 			"vx":         1.0,
// 			"vy":         0.5,
// 		},
// 	}
//
// 	clientPackage := types.ClientPackage{
// 		Message: clientMsg,
// 		Conn:    nil, // no real connection needed for this test
// 	}
//
// 	// simulating websocket server, send to servers channel
// 	// hub should received it at this point
//
// 	server.serverChan <- clientPackage
//
// 	// the game Session should receive the message after hub reroutes it
// 	select {
// 	case receivedPackage := <-session.MessageCh:
// 		assert.Equal(t, string(constants.ActionMove), receivedPackage.Message.Action)
// 		assert.Equal(t, session.ID.String(), receivedPackage.Message.Payload["session_id"])
// 		assert.Equal(t, 1.0, receivedPackage.Message.Payload["vx"])
// 		fmt.Printf("Session received message: %+v\n", receivedPackage.Message)
// 	case <-time.After(2 * time.Second):
// 		t.Fatal("Message was not routed to session within timeout")
// 	}
// }

func registerTestConn(s *Server, conn *websocket.Conn, player *types.Player) chan interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connToPlayer[conn] = player
	s.players[player.ID] = player

	msgCh := make(chan interface{}, 10)
	s.msgChan[conn] = msgCh
	return msgCh
}

func TestQueueFindGameFlow(t *testing.T) {
	mockAuthClient := &MockAuthClient{}
	mockQueue := NewMockQueueService()
	mockEventEmitter := &MockEventEmitter{}
	mockItemsClient := &MockItemsClient{}
	server := NewServer(mockAuthClient, mockQueue, mockEventEmitter, mockItemsClient)

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
					msgMessage, ok := msg.(types.Message)
					if !ok {
						continue
					}
					if msgMessage.Action == "game_found" {
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

type Conn struct{}

func (c *Conn) WriteJSON(v interface{}) error {
	fmt.Println("Mock WriteJSON")
	return nil
}

func TestSenderToBroadcastToPlayerList(t *testing.T) {
	testCases := []struct {
		name          string                    // 測試案例名稱
		setupPlayers  func(*Server) []uuid.UUID // 設置玩家的函數
		action        constants.Action          // 要測試的 action
		payload       map[string]interface{}    // 訊息內容
		expectedError error                     // 預期的錯誤 (使用 errors.New() 定義的)
		errorContains string                    // 錯誤訊息應該包含的文字 (用於更詳細的檢查)
	}{
		{
			name: "broadcast to all connected players - success",
			setupPlayers: func(s *Server) []uuid.UUID {
				// 情境:所有玩家都有連線
				player1 := &types.Player{ID: uuid.New(), Username: "Player1"}
				player2 := &types.Player{ID: uuid.New(), Username: "Player2"}
				player3 := &types.Player{ID: uuid.New(), Username: "Player3"}

				// 模擬建立連線和 message channel
				conn1 := &websocket.Conn{} // 這裡需要實際的 mock connection
				conn2 := &websocket.Conn{}
				conn3 := &websocket.Conn{}

				s.mu.Lock()
				s.connToPlayer[conn1] = player1
				s.connToPlayer[conn2] = player2
				s.connToPlayer[conn3] = player3
				s.msgChan[conn1] = make(chan interface{}, 10)
				s.msgChan[conn2] = make(chan interface{}, 10)
				s.msgChan[conn3] = make(chan interface{}, 10)
				s.mu.Unlock()

				return []uuid.UUID{player1.ID, player2.ID, player3.ID}
			},
			action: constants.ActionFindGame,
			payload: map[string]interface{}{
				"info": "Game found, starting match",
			},
			expectedError: nil, // 不應該有錯誤
		},
		{
			name: "broadcast to disconnected players - should error",
			setupPlayers: func(s *Server) []uuid.UUID {
				// 情境:玩家已經創建但沒有連線
				player1 := uuid.New()
				player2 := uuid.New()
				player3 := uuid.New()

				// 不設置連線,模擬玩家已斷線
				return []uuid.UUID{player1, player2, player3}
			},
			action: constants.ActionFindGame,
			payload: map[string]interface{}{
				"info": "Test broadcast to disconnected players",
			},
			expectedError: ErrAllPlayersFailed, // 使用預定義的錯誤常量
			errorContains: "broadcast failed for 3 players",
		},
		{
			name: "broadcast to mixed connected/disconnected players",
			setupPlayers: func(s *Server) []uuid.UUID {
				// 情境:部分玩家有連線,部分沒有
				player1 := &types.Player{ID: uuid.New(), Username: "ConnectedPlayer"}
				player2 := uuid.New() // 沒有連線的玩家
				player3 := uuid.New() // 沒有連線的玩家

				conn1 := &websocket.Conn{}
				s.mu.Lock()
				s.connToPlayer[conn1] = player1
				s.msgChan[conn1] = make(chan interface{}, 10)
				s.mu.Unlock()

				return []uuid.UUID{player1.ID, player2, player3}
			},
			action: constants.ActionQueue,
			payload: map[string]interface{}{
				"current": 2,
				"total":   3,
			},
			expectedError: ErrPartialBroadcastFailed, // 使用預定義的錯誤常量
			errorContains: "broadcast failed for 2 players",
		},
		{
			name: "broadcast with empty player list",
			setupPlayers: func(s *Server) []uuid.UUID {
				// 情境:空的玩家列表
				return []uuid.UUID{}
			},
			action: constants.ActionFindGame,
			payload: map[string]interface{}{
				"info": "No players to broadcast to",
			},
			expectedError: nil, // 空列表不應該報錯,只是什麼都不做
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serviceName := "game"
			consulAddr := commonhelpers.GetEnvString("CONSUL_ADDR", "localhost:8510")
			registry, _ := consul.NewRegistry(consulAddr, serviceName)
			authClient := grpcauth.NewClient(registry)
			mockQueue := NewMockQueueService()
			mockEventEmitter := &MockEventEmitter{}
			mockItemsClient := &MockItemsClient{}
			server := NewServer(authClient, mockQueue, mockEventEmitter, mockItemsClient)
			playerIDs := tc.setupPlayers(server)

			newSender := messaging.NewMessageSender(server)
			err := newSender.BroadcastToPlayerList(playerIDs, types.Message{
				Action:  string(tc.action),
				Payload: tc.payload,
			})

			if tc.expectedError != nil {
				assert.Error(t, err, "Expected error for test case: %s", tc.name)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains,
						"Error message should contain: %s", tc.errorContains)
				}
			} else {
				assert.NoError(t, err, "Should not return error for test case: %s", tc.name)
			}
		})
	}
}

func TestActionConstants(t *testing.T) {
	testCases := []struct {
		action   constants.Action
		expected string
	}{
		{constants.ActionAttack, "attack"},
		{constants.ActionMove, "move"},
		{constants.ActionInteract, "interact"},
		{constants.ActionPickup, "pickup"},
		{constants.ActionDropItem, "drop"},
		{constants.ActionUseItem, "use_item"},
		{constants.ActionChat, "chat"},
		{constants.ActionFindGame, "find_game"},
		{constants.ActionLeaveQueue, "leave_game"},
		{constants.ActionQueue, "queue_status"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.action), func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.action),
				"Action constant %s should equal %s", tc.action, tc.expected)
		})
	}
}
