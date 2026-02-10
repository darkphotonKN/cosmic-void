package stats

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commontypes "github.com/darkphotonKN/cosmic-void-server/common/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockRepo struct {
	getPlayerStatsReturn *PlayerMatchStats
	getPlayerStatsErr    error

	upsertCalled               bool
	upsertParams               *UpdateStatsParams
	upsertErr                  error
	upsertPlayerRankingsCalled bool
	upsertPlayerRankingParams  *UpdatePlayerRankingsParams
}

func (m *mockRepo) GetPlayerMatchStats(ctx context.Context, memberID uuid.UUID) (*PlayerMatchStats, error) {
	return m.getPlayerStatsReturn, m.getPlayerStatsErr
}

func (m *mockRepo) UpsertPlayerMatchStats(ctx context.Context, params *UpdateStatsParams) (*PlayerMatchStats, error) {
	m.upsertCalled = true
	m.upsertParams = params
	return nil, m.upsertErr
}

func (m *mockRepo) UpsertPlayerRankingStats(ctx context.Context, params *UpdatePlayerRankingsParams) (*PlayerRankingStats, error) {
	m.upsertPlayerRankingsCalled = true
	m.upsertPlayerRankingParams = params
	return nil, m.upsertErr
}

// NOTE: update mock methods if needed
func (m *mockRepo) CreatePlayerRankingStats(ctx context.Context, stats *PlayerRankingStats) error {
	return nil
}

func (m *mockRepo) CreateMatchHistory(ctx context.Context, history *MatchHistory) error {
	return nil
}

func (m *mockRepo) GetPlayerRankingStats(ctx context.Context, memberID uuid.UUID) (*PlayerRankingStats, error) {
	return nil, nil
}

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

func TestUpdatePlayerStats_IncrementWin(t *testing.T) {
	repo := &mockRepo{}
	service := NewService(repo, &TestPublisher{})

	playerStats := &pb.PlayerMatchResult{
		MemberId:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440001").String(),
		Username:      "Obiwon2002",
		Win:           true,
		Kills:         10,
		Deaths:        1,
		FinalPosition: 1,
	}

	err := service.updatePlayerStats(context.Background(), playerStats)
	assert.NoError(t, err)

	slog.Info("end of updatePlayerStats increment win test", "repo.upsertParams.Wins", repo.upsertParams.Wins)

	assert.Equal(t, 1, repo.upsertParams.Wins)
	assert.Equal(t, 10, repo.upsertParams.Kills)
}
