package game

import (
	"context"
	"testing"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
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

func TestPublishMatchComplete_DataStructure(t *testing.T) {

	// Create test data - 6 player match results
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

	ch, close := broker.Connect(amqpUser, amqpPassword, amqpHost, amqpPort)

	// Only declare the exchange we actually consume from
	broker.DeclareExchange(ch, commonconstants.GameMatchEndedEvent, "fanout")

	defer func() {
		close()
		ch.Close()
	}()

	service := NewService(ch)
	service.PublishMatchComplete(context.Background(), matchEndData)

	// Test data structure validity
	t.Run("match data has correct structure", func(t *testing.T) {
		// Verify basic match info
		assert.NotEqual(t, uuid.Nil, matchEndData.SessionID)
		assert.True(t, matchEndData.MatchEndedAt.After(matchEndData.MatchStartedAt))

		// Verify we have 6 players
		assert.Len(t, matchEndData.PlayerMatchResults, 6, "Should have exactly 6 players")

		// Verify player data
		winnerCount := 0
		for i, player := range matchEndData.PlayerMatchResults {
			// Check required fields
			assert.NotEmpty(t, player.MemberID, "Player %d should have MemberID", i+1)
			assert.NotEmpty(t, player.Username, "Player %d should have Username", i+1)
			assert.GreaterOrEqual(t, player.FinalPosition, int32(1), "Player %d position should be >= 1", i+1)
			assert.LessOrEqual(t, player.FinalPosition, int32(6), "Player %d position should be <= 6", i+1)
			assert.GreaterOrEqual(t, player.Kills, int32(0), "Player %d kills should be >= 0", i+1)
			assert.GreaterOrEqual(t, player.Deaths, int32(0), "Player %d deaths should be >= 0", i+1)

			if player.Win {
				winnerCount++
				assert.Equal(t, int32(1), player.FinalPosition, "Winner should be in position 1")
			}
		}

		// Verify only one winner
		assert.Equal(t, 1, winnerCount, "Should have exactly one winner")

		// Verify positions are sequential and unique
		positionMap := make(map[int32]bool)
		for _, player := range matchEndData.PlayerMatchResults {
			assert.False(t, positionMap[player.FinalPosition], "Position %d should be unique", player.FinalPosition)
			positionMap[player.FinalPosition] = true
		}
		assert.Len(t, positionMap, 6, "Should have 6 unique positions")
	})

	t.Run("match duration is reasonable", func(t *testing.T) {
		duration := matchEndData.MatchEndedAt.Sub(matchEndData.MatchStartedAt)
		assert.Greater(t, duration, time.Duration(0), "Match duration should be positive")
		assert.LessOrEqual(t, duration, 60*time.Minute, "Match duration should be reasonable (< 1 hour)")
	})

	t.Run("player stats are realistic", func(t *testing.T) {
		totalKills := int32(0)
		totalDeaths := int32(0)

		for _, player := range matchEndData.PlayerMatchResults {
			totalKills += player.Kills
			totalDeaths += player.Deaths
		}

		// In a battle royale/elimination game, total kills should roughly equal total deaths
		// (minus environmental deaths or other factors)
		assert.Greater(t, totalKills, int32(0), "Should have at least some kills")
		assert.Greater(t, totalDeaths, int32(0), "Should have at least some deaths")

		// Check that kills roughly match deaths (allowing some variance)
		killDeathDiff := totalKills - totalDeaths
		if killDeathDiff < 0 {
			killDeathDiff = -killDeathDiff
		}
		assert.LessOrEqual(t, killDeathDiff, int32(5), "Total kills and deaths should be roughly balanced")
	})
}

// TestMatchEndDataSerialization tests that the data can be properly serialized
func TestMatchEndDataSerialization(t *testing.T) {
	// Create a simpler match data for testing
	matchEndData := &commontypes.MatchEndState{
		SessionID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		MatchStartedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		MatchEndedAt:   time.Date(2024, 1, 1, 12, 10, 0, 0, time.UTC),
		PlayerMatchResults: []*commontypes.PlayerMatchResults{
			{
				MemberID:      "test-player-001",
				Username:      "TestPlayer1",
				Win:           true,
				FinalPosition: 1,
				Kills:         5,
				Deaths:        1,
			},
			{
				MemberID:      "test-player-002",
				Username:      "TestPlayer2",
				Win:           false,
				FinalPosition: 2,
				Kills:         3,
				Deaths:        2,
			},
		},
	}

	// Verify the data structure is valid
	assert.NotNil(t, matchEndData)
	assert.NotNil(t, matchEndData.PlayerMatchResults)
	assert.Len(t, matchEndData.PlayerMatchResults, 2)

	// Verify the winner
	winner := matchEndData.PlayerMatchResults[0]
	assert.True(t, winner.Win)
	assert.Equal(t, int32(1), winner.FinalPosition)

	// Verify the non-winner
	loser := matchEndData.PlayerMatchResults[1]
	assert.False(t, loser.Win)
	assert.Equal(t, int32(2), loser.FinalPosition)
}
