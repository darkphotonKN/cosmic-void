package game

import (
	"context"
	"log/slog"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	commonoutbox "github.com/darkphotonKN/cosmic-void-server/common/outbox"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service struct {
	outboxPublisher commonoutbox.OutboxPublisher
}

func NewService(outboxPublisher commonoutbox.OutboxPublisher) *service {
	return &service{
		outboxPublisher: outboxPublisher,
	}
}

func (s *service) PublishMatchComplete(ctx context.Context, data *types.RawMatchState) {
	slog.Debug("service publishingMatchComplete")
	// calculate winner and determine final positions
	rankedPlayers := s.rankPlayers(data.Players, data.EliminationOrder)

	// proto marshal
	protoData, err := s.formatMatchData(data.SessionID, data.StartedAt, data.EndedAt, rankedPlayers)

	if err != nil {
		slog.Error("Error formatting game match end event", "error", err)
		return
	}

	err = s.outboxPublisher.CreateOutbox(ctx, commonoutbox.OutboxParams{
		RoutingKey: commonconstants.ItemsExtracted,
		Exchange:   commonconstants.GameEventsExchange,
		Payload:    protoData.ItemsExtractedEvent,
	})

	if err != nil {
		slog.Error("Error publishing items extracted event", "error", err)
		return
	}

	// send to outbox for publishing
	err = s.outboxPublisher.CreateOutbox(ctx, commonoutbox.OutboxParams{
		RoutingKey: commonconstants.GameMatchEnded,
		Exchange:   commonconstants.GameEventsExchange,
		Payload:    protoData.MatchEndedEvent,
	})

	if err != nil {
		slog.Error("Error publishing game match end event", "error", err)
		return
	}

	slog.Debug("Successfully created outbox item for two events",
		"event_one_name", commonconstants.GameMatchEnded,
		"event_two_name", commonconstants.ItemsExtracted,
	)
}

/**
* Formats from raw game state to match end state.
**/
func (s *service) formatMatchData(sessionID uuid.UUID, startedAt time.Time, endedAt time.Time, players []types.RankedPlayerState) (*types.FormattedMatchData, error) {

	// format data for marshalling as protobuf
	playerMatchRes := make([]*pb.PlayerMatchResult, len(players))

	for i, player := range players {
		playerMatchRes[i] = &pb.PlayerMatchResult{
			MemberId:      player.MemberID,
			Username:      player.Username,
			Kills:         player.Kills,
			Deaths:        player.Deaths,
			FinalPosition: player.FinalPosition,
			Win:           player.Win,
			Escape:        player.Escape,
		}
	}

	matchEndedEvent := pb.MatchEndedEvent{
		SessionId:      string(sessionID.String()),
		MatchStartedAt: timestamppb.New(startedAt),
		MatchEndedAt:   timestamppb.New(endedAt),
		Players:        playerMatchRes,
	}

	slog.Debug("matchEndedEvent in formatMatchData before marshalling into protobuf",
		"match_ended_event", matchEndedEvent,
	)

	// marshal to protobuf
	protoData, matchEndedErr := proto.Marshal(&matchEndedEvent)

	if matchEndedErr != nil {
		slog.Error("could not marshal end match data to MatchEndedEvent proto",
			"session_id", sessionID,
			"error", matchEndedErr,
		)
	}

	playerItems := make([]*pb.PlayerItems, len(players))

	for idx, player := range players {
		playerItems[idx] = &pb.PlayerItems{
			MemberId: player.MemberID,
			Equipment: &pb.Equipment{
				Weapon:       extractedItemToPb(player.Equipment.WeaponSlot),
				Head:         extractedItemToPb(player.Equipment.HeadSlot),
				Chest:        extractedItemToPb(player.Equipment.ChestSlot),
				Gloves:       extractedItemToPb(player.Equipment.GlovesSlot),
				Legs:         extractedItemToPb(player.Equipment.LegsSlot),
				Ring_1:       extractedItemToPb(player.Equipment.Ring1Slot),
				Ring_2:       extractedItemToPb(player.Equipment.Ring2Slot),
				Consumable_1: extractedItemToPb(player.Equipment.Consumable1),
				Consumable_2: extractedItemToPb(player.Equipment.Consumable2),
				Consumable_3: extractedItemToPb(player.Equipment.Consumable3),
			},
			Inventory: extractedInventoryToPb(player.Inventory),
		}
	}

	// generate eventId for idemptotency deduplication
	eventId := uuid.NewString()

	itemsExtractedEvent := pb.ItemsExtractedEvent{
		SessionId:   string(sessionID.String()),
		EventId:     eventId,
		PlayerItems: playerItems,
	}

	slog.Debug("itemsExtractedEvent in formatMatchData before marshalling into protobuf",
		"items_extracted_event", itemsExtractedEvent,
	)

	itemsExtractedProtoData, itemsExtractErr := proto.Marshal(&itemsExtractedEvent)

	if itemsExtractErr != nil {
		slog.Error("could not marshal items extracted event to ItemExtractedEvent proto",
			"session_id", sessionID,
			"error", itemsExtractErr,
		)
	}

	if matchEndedErr != nil && itemsExtractErr != nil {
		return nil, matchEndedErr
	}

	data := &types.FormattedMatchData{
		MatchEndedEvent:     protoData,
		ItemsExtractedEvent: itemsExtractedProtoData,
	}

	return data, nil
}

/**
* Ranks players with a final position plus determine and mark winner.
**/
func (s *service) rankPlayers(players []types.RawPlayerState, eliminationOrder map[uuid.UUID]int) []types.RankedPlayerState {
	// determine winner and other business logic
	rankedPlayers := make([]types.RankedPlayerState, len(players))

	for idx, player := range players {
		memberID, err := uuid.Parse(player.MemberID)

		if err != nil {
			slog.Debug("Player memberID not a uuid, couldn't parse",
				"memberID", memberID)
			continue
		}

		position, exists := eliminationOrder[memberID]

		rankedPlayer := types.RankedPlayerState{
			MemberID: player.MemberID,
			Username: player.Username,
			Kills:    player.Kills,
			Deaths:   player.Deaths,
			Escape:   player.Escape,
		}

		// determining final positions
		// total number of players minus their stored elimination order position
		rankedPlayer.FinalPosition = int32(len(players) - position)

		// determine winner. someone who doesnt exist in the elimination order slice
		// is the surivor and hence the winner
		if !exists {
			rankedPlayer.Win = true
			rankedPlayer.FinalPosition = 1
		}

		rankedPlayers[idx] = rankedPlayer
	}

	return rankedPlayers
}

func extractedItemToPb(item *types.ExtractedItem) *pb.Item {
	if item == nil {
		return nil
	}
	return &pb.Item{
		TemplateId:      item.TemplateID.String(),
		ItemType:        item.ItemType,
		Name:            item.Name,
		AttackPower:     int32(item.AttackPower),
		CriticalRate:    item.CriticalRate,
		WeaponType:      item.WeaponType,
		DefenseRating:   int32(item.DefenseRating),
		MagicResistance: int32(item.MagicResistance),
		ArmorSlot:       item.ArmorSlot,
		HealingAmount:   int32(item.HealingAmount),
		ManaAmount:      int32(item.ManaAmount),
		BuffDuration:    int32(item.BuffDuration),
		BuyPrice:        int32(item.BuyPrice),
		SellPrice:       int32(item.SellPrice),
		Description:     item.Description,
	}
}

func extractedInventoryToPb(inventory []*types.ExtractedItem) []*pb.Item {
	items := make([]*pb.Item, 0, len(inventory))
	for _, item := range inventory {
		if pbItem := extractedItemToPb(item); pbItem != nil {
			items = append(items, pbItem)
		}
	}
	return items
}
