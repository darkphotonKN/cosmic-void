package gameserver

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/**
* testing cross module functionality and behaviors.
**/

// TestServerHubSessionIntegration tests the full flow
// message sent to client sends to Server →  Hub routes →  Session receives
func TestServerHubSessionIntegration(t *testing.T) {
	server := NewServer()

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
			"vy":         0.0,
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
	case receivedMsg := <-session.MessageCh:
		assert.Equal(t, string(constants.ActionMove), receivedMsg.Action, "Action should match")
		payload := receivedMsg.Payload

		fmt.Printf("\nclient package payload in test was: %+v\n\n", payload)

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
	server := NewServer()

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
					Action:  string(constants.ActionFindGame),
					Payload: player,
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
