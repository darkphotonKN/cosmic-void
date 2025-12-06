package gameserver

import (
	"fmt"
	"testing"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
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

	require.NotNil(t, session, "Session should be created")

	// give goroutines time to start
	time.Sleep(100 * time.Millisecond)

	// clean up at end
	defer session.Shutdown()

	// send a game action message that should be routed to session
	clientMsg := types.Message{
		Action: string(constants.ActionMove),
		Payload: map[string]interface{}{
			"sessionId": session.ID.String(),
			"playerId":  player1.ID.String(),
			"vx":        1.0,
			"vy":        0.0,
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

		fmt.Printf("\npayload was: %+v\n\n", payload)

	case <-time.After(2 * time.Second):
		t.Fatal("Message was not routed to session within timeout")
	}
}
