package game

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"

	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/messaging"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/serializer"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	"google.golang.org/protobuf/proto"
)

// test velocity updates transform of player entity after handle move and system update cycle
type MessageSender struct{}

func (m *MessageSender) PushMessageToChannelQueue(
	playerID uuid.UUID,
	msg interface{},
) error {
	slog.Debug("Pushing message to channel queue TEST IMPLEMENTATION")
	return nil
}

func (m *MessageSender) PushMessageToConn(
	conn *websocket.Conn,
	msg interface{},
) error {
	slog.Debug("Push Messagae to Conn")
	return nil
}

// Mock EventEmitter for testing
type mockEventEmitter struct{}

func (m *mockEventEmitter) PublishMatchComplete(ctx context.Context, data *types.RawMatchState) error {
	return nil
}

func TestHandleMoveUpdatesPositionIntegration(t *testing.T) {
	mockMessageSender := MessageSender{}
	sender := messaging.NewMessageSender(&mockMessageSender)
	em := ecs.NewEntityManager()
	stateSerializer := serializer.NewStateSerializer(em)
	mockEmitter := &mockEventEmitter{}
	session := NewSession(sender, stateSerializer, em, mockEmitter, nil)

	player1ID := uuid.New()
	username := "Player1"
	playerEntityID := session.AddPlayer(player1ID, username)

	// check player initial position
	playerEntity, ok := session.EntityManager.GetEntity(playerEntityID)

	if !ok {
		slog.Error("PlayerEntity doesn't exist", "playerEntityID", playerEntityID)
	}
	assert.Equal(t, true, ok)

	playerTransformComponent, ok := playerEntity.GetComponent(ecs.ComponentTypeTransform)

	if !ok {
		slog.Error("Player's Velocity Component doesn't exist", "entityID", playerEntity.ID)
	}

	assert.NotNil(t, playerTransformComponent)

	component := playerTransformComponent.(*components.TransformComponent)
	slog.Debug("Player transform coordinates initial", "coordinates", component)

	assert.Equal(t, float64(0), component.X)
	assert.Equal(t, float64(0), component.Y)

	// player speed moves with speed speedX and speedY
	speedX := 0.81
	speedY := 0.81
	session.handleMove(player1ID, speedX, speedY)

	// account for system game loop refresh rate, but only time for 1 move
	time.Sleep(time.Millisecond * 1200)

	slog.Debug("Player transform coordinates after update", "coordinates", component)
	assert.Equal(t, float64(0.81), component.X)
	assert.Equal(t, float64(0.81), component.Y)
}

/**
* test integration between match publish and event
**/
func TestPublishMatchCompleteIntegration(t *testing.T) {
	// Test member IDs from the actual registered members
	testMemberIDOne := "7f12d971-5879-4057-84c5-408a36de913c" // feb19
	testMemberIDTwo := "0760888e-f489-4a68-a83f-c1abddc64f10" // feb20
	testMemberIDThree := "61363b86-5eef-4ddd-b944-3c0869b99182"
	testMemberIDFour := "f5535dc6-1d6d-4f0b-b003-65db9bbf24f0"
	testMemberIDFive := "49fd6267-d7ec-4963-948a-832cc7140c9c"
	testMemberIDSix := "b5338eba-fffb-443c-bea4-51629d60be7c"

	// Create 3 different match scenarios using RawMatchState
	matchScenarios := []struct {
		sessionID      string
		winnerID       string
		winnerUsername string
		rawMatchState  *types.RawMatchState
	}{
		// Match 1: testMemberIDThree wins (highest kills, no deaths)
		{
			sessionID:      "550e8400-e29b-41d4-a716-446655440001",
			winnerID:       testMemberIDThree,
			winnerUsername: "player_three",
			rawMatchState: &types.RawMatchState{
				SessionID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
				Players: []types.RawPlayerState{
					{
						MemberID: testMemberIDThree,
						Username: "player_three",
						Kills:    7,
						Deaths:   0,
					},
					{
						MemberID: testMemberIDOne,
						Username: "test1feb19",
						Kills:    4,
						Deaths:   1,
					},
					{
						MemberID: testMemberIDTwo,
						Username: "test1feb20",
						Kills:    3,
						Deaths:   1,
					},
					{
						MemberID: testMemberIDFour,
						Username: "player_four",
						Kills:    2,
						Deaths:   2,
					},
					{
						MemberID: testMemberIDFive,
						Username: "player_five",
						Kills:    1,
						Deaths:   1,
					},
					{
						MemberID: testMemberIDSix,
						Username: "player_six",
						Kills:    0,
						Deaths:   2,
					},
				},
			},
		},
		// Match 2: testMemberIDOne wins
		{
			sessionID:      "550e8400-e29b-41d4-a716-446655440002",
			winnerID:       testMemberIDOne,
			winnerUsername: "test1feb19",
			rawMatchState: &types.RawMatchState{
				SessionID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
				Players: []types.RawPlayerState{
					{
						MemberID: testMemberIDOne,
						Username: "test1feb19",
						Kills:    8,
						Deaths:   0,
					},
					{
						MemberID: testMemberIDFour,
						Username: "player_four",
						Kills:    5,
						Deaths:   1,
					},
					{
						MemberID: testMemberIDThree,
						Username: "player_three",
						Kills:    3,
						Deaths:   1,
					},
					{
						MemberID: testMemberIDFive,
						Username: "player_five",
						Kills:    2,
						Deaths:   1,
					},
					{
						MemberID: testMemberIDTwo,
						Username: "test1feb20",
						Kills:    1,
						Deaths:   2,
					},
					{
						MemberID: testMemberIDSix,
						Username: "player_six",
						Kills:    1,
						Deaths:   2,
					},
				},
			},
		},
		// Match 3: testMemberIDFive wins
		{
			sessionID:      "550e8400-e29b-41d4-a716-446655440003",
			winnerID:       testMemberIDFive,
			winnerUsername: "player_five",
			rawMatchState: &types.RawMatchState{
				SessionID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
				Players: []types.RawPlayerState{
					{
						MemberID: testMemberIDFive,
						Username: "player_five",
						Kills:    6,
						Deaths:   0,
					},
					{
						MemberID: testMemberIDTwo,
						Username: "test1feb20",
						Kills:    4,
						Deaths:   1,
					},
					{
						MemberID: testMemberIDSix,
						Username: "player_six",
						Kills:    3,
						Deaths:   0,
					},
					{
						MemberID: testMemberIDOne,
						Username: "test1feb19",
						Kills:    2,
						Deaths:   1,
					},
					{
						MemberID: testMemberIDThree,
						Username: "player_three",
						Kills:    2,
						Deaths:   2,
					},
					{
						MemberID: testMemberIDFour,
						Username: "player_four",
						Kills:    0,
						Deaths:   1,
					},
				},
			},
		},
	}

	ch, close := broker.Connect(amqpUser, amqpPassword, amqpHost, amqpPort)

	broker.DeclareExchange(ch, commonconstants.GameEventsExchange, "topic")

	defer func() {
		close()
		ch.Close()
	}()

	testQueue := fmt.Sprintf("%s.test", commonconstants.StatsGameMatchEndedQueue)

	// queue setup
	_, err := ch.QueueDeclare(
		testQueue, // queue name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		slog.Error("Failed to declare queue", "error", err)
		assert.NoError(t, err)
	}

	// bind the queue to the exchange
	err = ch.QueueBind(
		testQueue,                          // queue name
		commonconstants.GameMatchEnded,     // routing key
		commonconstants.GameEventsExchange, // exchange
		false,                              // no-wait
		nil,                                // args
	)
	if err != nil {
		slog.Error("Failed to bind queue to exchange", "error", err)
		assert.NoError(t, err)
	}

	// service setup
	publishCh := commonbroker.NewAmqpPublisher(ch) // use adapter
	service := NewService(publishCh)

	// Publish all 3 match scenarios
	for i, matchData := range matchScenarios {
		slog.Info(fmt.Sprintf("Publishing match %d", i+1), "winnerID", matchData.winnerID)

		// Add time spacing between matches (as if they happened at different times)
		matchStartTime := time.Now().Add(time.Duration(-(45 - i*15)) * time.Minute)
		matchEndTime := matchStartTime.Add(15 * time.Minute)

		// Set the timestamps for the raw match state
		matchData.rawMatchState.StartedAt = matchStartTime
		matchData.rawMatchState.EndedAt = matchEndTime

		slog.Info("Match raw data", "matchNumber", i+1, "data", matchData.rawMatchState)

		err := service.PublishMatchComplete(context.Background(), matchData.rawMatchState)
		if err != nil {
			slog.Error("Failed to publish match complete", "error", err, "matchNumber", i+1)
			assert.NoError(t, err)
		}

		// Small delay between publishes to ensure proper processing
		time.Sleep(100 * time.Millisecond)
	}

	msgs, err := ch.Consume(testQueue, "", false, false, false, false, nil)
	assert.NoError(t, err)

	// Consume and verify all 3 messages
	matchesReceived := 0
	timeout := time.After(time.Second * 10)

	for matchesReceived < 3 {
		select {
		case msg := <-msgs:
			var data pb.MatchEndedEvent

			if err := proto.Unmarshal(msg.Body, &data); err != nil {
				slog.Error("Failed to parse match completed event", "error", err)
				msg.Nack(false, false)
				assert.NoError(t, err)
			}

			matchesReceived++
			slog.Info(fmt.Sprintf("Received match %d", matchesReceived), "sessionID", data.SessionId)

			// Acknowledge message
			msg.Ack(false)

			// Verify data integrity - check that raw stats are passed correctly
			// Note: The service now just passes raw data, it doesn't determine win/loss
			if matchesReceived == 1 {
				// First match: verify testMemberIDThree has highest kills and no deaths
				for _, player := range data.Players {
					if player.MemberId == testMemberIDThree {
						assert.Equal(t, int32(7), player.Kills)
						assert.Equal(t, int32(0), player.Deaths)
					}
					if player.MemberId == testMemberIDOne {
						assert.Equal(t, int32(4), player.Kills)
						assert.Equal(t, int32(1), player.Deaths)
					}
				}
			} else if matchesReceived == 2 {
				// Second match: verify testMemberIDOne has highest kills
				for _, player := range data.Players {
					if player.MemberId == testMemberIDOne {
						assert.Equal(t, int32(8), player.Kills)
						assert.Equal(t, int32(0), player.Deaths)
					}
					if player.MemberId == testMemberIDFour {
						assert.Equal(t, int32(5), player.Kills)
						assert.Equal(t, int32(1), player.Deaths)
					}
				}
			} else if matchesReceived == 3 {
				// Third match: verify testMemberIDFive has good stats
				for _, player := range data.Players {
					if player.MemberId == testMemberIDFive {
						assert.Equal(t, int32(6), player.Kills)
						assert.Equal(t, int32(0), player.Deaths)
					}
					if player.MemberId == testMemberIDTwo {
						assert.Equal(t, int32(4), player.Kills)
						assert.Equal(t, int32(1), player.Deaths)
					}
				}
			}

		case <-timeout:
			t.Fatalf("Timed out when waiting for consuming messages. Received %d of 3 expected messages", matchesReceived)
		}
	}

	slog.Info("Successfully published and consumed all 3 match scenarios")
}
