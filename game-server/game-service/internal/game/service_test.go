package game

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commontypes "github.com/darkphotonKN/cosmic-void-server/common/types"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	serviceName = "stats"
	grpcAddr    = commonhelpers.GetEnvString("GRPC_STATS_ADDR", "7011")
	consulAddr  = commonhelpers.GetEnvString("CONSUL_ADDR", "localhost:8510")

	amqpUser     = commonhelpers.GetEnvString("RABBITMQ_USER", "guest")
	amqpPassword = commonhelpers.GetEnvString("RABBITMQ_PASS", "guest")
	amqpHost     = commonhelpers.GetEnvString("RABBITMQ_HOST", "localhost")
	amqpPort     = commonhelpers.GetEnvString("RABBITMQ_PORT", "5672")
)

// test publisher just for testing
type TestPublisher struct {
}

func (p *TestPublisher) PublishWithContext(_ context.Context, exchange, key string, msg commonbroker.Message) error {
	var matchEndData commontypes.MatchEndState

	err := json.Unmarshal(msg.Body, &matchEndData)

	if err != nil {
		slog.Info("Error when unmarshalling published message, message type could not not be unmarshaled to expected type MatchEndState")
		return err
	}

	slog.Info("Worked!", "matchEndData", matchEndData)
	return nil
}

type TestFramework struct {
	sample  string
	success string
}

func TestPublishMatchComplete_DataStructure(t *testing.T) {

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

	service := NewService(&TestPublisher{})

	jsonBinary, err := json.Marshal(matchEndData)

	if err != nil {
		slog.Info("Error when unmarshalling published message.")
		assert.NoError(t, err)
	}

	err = service.publishCh.PublishWithContext(context.Background(), "test", "", commonbroker.Message{
		Body: jsonBinary,
	})

	// check no error occured
	assert.NoError(t, err)
}
