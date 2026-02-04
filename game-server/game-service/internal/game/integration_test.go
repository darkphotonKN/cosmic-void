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
	commontypes "github.com/darkphotonKN/cosmic-void-server/common/types"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/messaging"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/serializer"
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

func (m *mockEventEmitter) PublishMatchComplete(ctx context.Context, data *commontypes.MatchEndState) error {
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
	testMemberIDOne := "213b277a-68b8-4da2-ab6e-adb4f28e7b0d"
	testMemberIDTwo := "4bbd9306-f06e-440e-a870-a2db4e07a7a6"

	// create test data player match results
	matchEndData := &commontypes.MatchEndState{
		SessionID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		MatchStartedAt: time.Now().Add(-15 * time.Minute), // Match lasted 15 minutes
		MatchEndedAt:   time.Now(),
		PlayerMatchResults: []*commontypes.PlayerMatchResults{
			{
				MemberID:      testMemberIDOne,
				Username:      "testplayer1",
				Win:           false,
				FinalPosition: 2,
				Kills:         3,
				Deaths:        1,
			},
			{
				MemberID:      testMemberIDTwo,
				Username:      "test2",
				Win:           true,
				FinalPosition: 1,
				Kills:         10,
				Deaths:        0,
			},
		},
	}

	slog.Info("Match end data", "data", matchEndData)

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

	service.PublishMatchComplete(context.Background(), matchEndData)

	msgs, err := ch.Consume(testQueue, "", false, false, false, false, nil)

	assert.NoError(t, err)

	select {
	case msg := <-msgs:
		var data pb.MatchEndedEvent

		if err := proto.Unmarshal(msg.Body, &data); err != nil {
			slog.Error("Failed to parse match completed event", "error", err)

			msg.Nack(false, false)

			assert.NoError(t, err)
		}

		// check each player from consumed results
		for _, player := range data.Players {
			if player.MemberId == testMemberIDOne {
				assert.Equal(t, player.Win, false)
			}
			if player.MemberId == testMemberIDTwo {
				assert.Equal(t, player.Win, true)
			}
		}

	// for timeout
	case <-time.After(time.Second * 5):
		t.Fatal("Timed out when waiting for consuming message for testing publish match complete.")
	}
}
