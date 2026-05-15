package game

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonoutbox "github.com/darkphotonKN/cosmic-void-server/common/outbox"
	commontypes "github.com/darkphotonKN/cosmic-void-server/common/types"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// mockOutboxPublisher implements commonoutbox.OutboxPublisher.
// formatMatchData doesn't actually call into it — this exists only so
// NewService can be constructed for unit tests.
type mockOutboxPublisher struct{}

func (m *mockOutboxPublisher) CreateOutbox(_ context.Context, _ commonoutbox.OutboxParams) error {
	return nil
}

func (m *mockOutboxPublisher) CreateOutboxTx(_ context.Context, _ *sqlx.Tx, _ commonoutbox.OutboxParams) error {
	return nil
}

func TestFormatMatchData(t *testing.T) {
	sessionID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	startedAt := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	endedAt := startedAt.Add(15 * time.Minute)

	cosmicSword := &types.ExtractedItem{
		TemplateID:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ItemType:     "weapon",
		Name:         "Cosmic Sword",
		AttackPower:  42,
		CriticalRate: 0.15,
		WeaponType:   "sword",
	}

	ironHelmet := &types.ExtractedItem{
		TemplateID:    uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		ItemType:      "armor",
		Name:          "Iron Helmet",
		DefenseRating: 10,
		ArmorSlot:     "head",
	}
	healthPotion := &types.ExtractedItem{
		TemplateID:    uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		ItemType:      "consumable",
		Name:          "Health Potion",
		HealingAmount: 50,
	}

	cases := []struct {
		name    string
		players []types.RankedPlayerState
	}{
		{
			name: "two players: alice wins with weapon and helmet, bob loses empty-handed",
			players: []types.RankedPlayerState{
				{
					MemberID:      "213b277a-68b8-4da2-ab6e-adb4f28e7b0d",
					Username:      "alice",
					Kills:         5,
					Deaths:        1,
					FinalPosition: 1,
					Win:           true,
					Equipment: types.ExtractedEquipment{
						WeaponSlot: cosmicSword,
						HeadSlot:   ironHelmet,
					},
					Inventory: []*types.ExtractedItem{healthPotion},
				},
				{
					MemberID:      "4bbd9306-f06e-440e-a870-a2db4e07a7a6",
					Username:      "bob",
					Kills:         2,
					Deaths:        5,
					FinalPosition: 2,
					Win:           false,
					Equipment:     types.ExtractedEquipment{},
					Inventory:     nil,
				},
			},
		},
		{
			name: "three players: charlie wins, diana escapes with two potions, evan loses",
			players: []types.RankedPlayerState{
				{
					MemberID:      "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
					Username:      "charlie",
					Kills:         3,
					Deaths:        0,
					FinalPosition: 1,
					Win:           true,
					Equipment: types.ExtractedEquipment{
						WeaponSlot: cosmicSword,
					},
				},
				{
					MemberID:      "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
					Username:      "diana",
					Kills:         1,
					Deaths:        1,
					FinalPosition: 2,
					Escape:        true,
					Inventory:     []*types.ExtractedItem{healthPotion, healthPotion},
				},
				{
					MemberID:      "cccccccc-cccc-cccc-cccc-cccccccccccc",
					Username:      "evan",
					Kills:         0,
					Deaths:        3,
					FinalPosition: 3,
				},
			},
		},
	}

	svc := NewService(&mockOutboxPublisher{})

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := svc.formatMatchData(sessionID, startedAt, endedAt, tc.players)
			assert.NoError(t, err)
			assert.NotNil(t, got)

			// --- MatchEndedEvent: unmarshal back into proto and verify ---
			var matchEnded pb.MatchEndedEvent
			err = proto.Unmarshal(got.MatchEndedEvent, &matchEnded)
			assert.NoError(t, err, "MatchEndedEvent payload should round trip through protobuf")

			assert.Equal(t, sessionID.String(), matchEnded.SessionId)
			assert.True(t, startedAt.Equal(matchEnded.MatchStartedAt.AsTime()))
			assert.True(t, endedAt.Equal(matchEnded.MatchEndedAt.AsTime()))
			assert.Len(t, matchEnded.Players, len(tc.players))

			for i, p := range tc.players {
				assert.Equal(t, p.MemberID, matchEnded.Players[i].MemberId)
				assert.Equal(t, p.Username, matchEnded.Players[i].Username)
				assert.Equal(t, p.Kills, matchEnded.Players[i].Kills)
				assert.Equal(t, p.Deaths, matchEnded.Players[i].Deaths)
				assert.Equal(t, p.FinalPosition, matchEnded.Players[i].FinalPosition)
				assert.Equal(t, p.Win, matchEnded.Players[i].Win)
				assert.Equal(t, p.Escape, matchEnded.Players[i].Escape)
			}

			// --- ItemsExtractedEvent: unmarshal back into proto and verify ---
			var itemsExtracted pb.ItemsExtractedEvent
			err = proto.Unmarshal(got.ItemsExtractedEvent, &itemsExtracted)
			assert.NoError(t, err, "ItemsExtractedEvent payload should round-trip through protobuf")

			assert.Equal(t, sessionID.String(), itemsExtracted.SessionId)
			assert.NotEmpty(t, itemsExtracted.EventId, "EventId should be generated for idempotency")
			assert.Len(t, itemsExtracted.PlayerItems, len(tc.players))

			for i, p := range tc.players {
				pi := itemsExtracted.PlayerItems[i]
				assert.Equal(t, p.MemberID, pi.MemberId)

				// weapon slot
				if p.Equipment.WeaponSlot != nil {
					assert.NotNil(t, pi.Equipment.Weapon, "weapon should be set for %s", p.Username)
					assert.Equal(t, p.Equipment.WeaponSlot.Name, pi.Equipment.Weapon.Name)
					assert.Equal(t, int32(p.Equipment.WeaponSlot.AttackPower), pi.Equipment.Weapon.AttackPower)
				} else {
					assert.Nil(t, pi.Equipment.Weapon, "weapon should be nil for %s", p.Username)
				}

				// head slot
				if p.Equipment.HeadSlot != nil {
					assert.NotNil(t, pi.Equipment.Head, "head should be set for %s", p.Username)
					assert.Equal(t, p.Equipment.HeadSlot.Name, pi.Equipment.Head.Name)
				} else {
					assert.Nil(t, pi.Equipment.Head, "head should be nil for %s", p.Username)
				}

				// inventory
				assert.Len(t, pi.Inventory, len(p.Inventory), "inventory length for %s", p.Username)
				for j, inv := range p.Inventory {
					assert.Equal(t, inv.Name, pi.Inventory[j].Name)
				}
			}
		})
	}
}

var (
	serviceName = "stats"
	grpcAddr    = commonhelpers.GetEnvString("GRPC_STATS_ADDR", "7011")

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

	publisher := &TestPublisher{}

	jsonBinary, err := json.Marshal(matchEndData)

	if err != nil {
		slog.Info("Error when unmarshalling published message.")
		assert.NoError(t, err)
	}

	err = publisher.PublishWithContext(context.Background(), "test", "", commonbroker.Message{
		Body: jsonBinary,
	})

	// check no error occured
	assert.NoError(t, err)
}
