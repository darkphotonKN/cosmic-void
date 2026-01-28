package game

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"

	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	commontypes "github.com/darkphotonKN/cosmic-void-server/common/types"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/messaging"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/serializer"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"

	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
)

// test velocity updates transform of player entity after handle move and system update cycle
type MessageSender struct{}

func (m *MessageSender) PushMessageToChannelQueue(
	playerID uuid.UUID,
	msg interface{},
) error {
	return nil
}

func (m *MessageSender) PushMessageToConn(
	conn *websocket.Conn,
	msg interface{},
) error {
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

	playerTransformComponent, ok := playerEntity.GetComponent(ecs.ComponentTypeTransform)

	if !ok {
		slog.Error("Player's Velocity Component doesn't exist", "entityID", playerEntity.ID)
	}

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
func TestPublishMatchComplete(t *testing.T) {
	// create test data player match results
	matchEndData := &commontypes.MatchEndState{
		SessionID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		MatchStartedAt: time.Now().Add(-15 * time.Minute), // Match lasted 15 minutes
		MatchEndedAt:   time.Now(),
		PlayerMatchResults: []*commontypes.PlayerMatchResults{
			{
				MemberID:      "213b277a-68b8-4da2-ab6e-adb4f28e7b0d",
				Username:      "testplayer1",
				Win:           true,
				FinalPosition: 1,
				Kills:         8,
				Deaths:        2,
			},
			{
				MemberID:      "4bbd9306-f06e-440e-a870-a2db4e07a7a6",
				Username:      "test2",
				Win:           false,
				FinalPosition: 2,
				Kills:         6,
				Deaths:        3,
			},
		},
	}

	slog.Info("Match end data", "data", matchEndData)

	ch, close := broker.Connect(amqpUser, amqpPassword, amqpHost, amqpPort)

	broker.DeclareExchange(ch, commonconstants.GameMatchEndedEvent, "fanout")

	defer func() {
		close()
		ch.Close()
	}()

	// service setup
	publishCh := commonbroker.NewAmqpPublisher(ch) // use adapter
	service := NewService(publishCh)

	dataJSON, err := json.Marshal(matchEndData)

	assert.NoError(t, err)

	service.publishCh.PublishWithContext(context.Background(), commonconstants.GameMatchEndedEvent, "fanout", commonbroker.Message{
		Body:         dataJSON,
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
	})

	// TODO: consume for testing

}
