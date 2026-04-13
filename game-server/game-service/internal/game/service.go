package game

import (
	"context"
	"log/slog"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service struct {
	publishCh commonbroker.Publisher
}

func NewService(publishCh commonbroker.Publisher) *service {
	return &service{
		publishCh: publishCh,
	}
}

func (s *service) PublishMatchComplete(ctx context.Context, data *types.RawMatchState) error {

	// calculate winner and determine final positions
	rankedPlayers := s.rankPlayers(data.Players, data.EliminationOrder)

	// proto marshal
	protoData, err := s.formatMatchData(data.SessionID, data.StartedAt, data.EndedAt, rankedPlayers)

	if err != nil {
		slog.Error("Error publishing game match end event", "error", err)
		return err
	}

	err = s.publishCh.PublishWithContext(
		ctx,
		commonconstants.GameEventsExchange,
		commonconstants.GameMatchEnded,
		commonbroker.Message{
			ContentType:  "application/protobuf",
			Body:         protoData.MatchEndedEvent,
			DeliveryMode: amqp.Persistent,
		})

	if err != nil {
		slog.Error("Error publishing game match end event", "error", err)
		return err
	}

	return nil
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

	// marshal to protobuf
	protoData, matchEndedErr := proto.Marshal(&pb.MatchEndedEvent{
		SessionId:      string(sessionID.String()),
		MatchStartedAt: timestamppb.New(startedAt),
		MatchEndedAt:   timestamppb.New(endedAt),
		Players:        playerMatchRes,
	})

	if matchEndedErr != nil {
		slog.Error("could not marshal end match data to MatchEndedEvent proto",
			"session_id", sessionID,
			"error", matchEndedErr,
		)
	}

	playerItems := make([]*pb.PlayerItems, len(players))

	for idx, player := range players {
		playerItems[idx].Equipment = &pb.Equipment{
			// auto complete this
		}

		// TODO: add inventory, pass it over and complete extraction
		items := make([]*pb.Item, 5)

		playerItems[idx].Inventory = items
	}

	itemsExtractedProtoData, itemsExtractErr := proto.Marshal(&pb.ItemsExtractedEvent{
		SessionId:   string(sessionID.String()),
		PlayerItems: nil,
	})

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
