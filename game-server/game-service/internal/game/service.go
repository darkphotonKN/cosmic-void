package game

import (
	"context"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	commontypes "github.com/darkphotonKN/cosmic-void-server/common/types"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
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
	// determine win and other business logic
	protoData, err := s.formatMatchData(data)

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
			Body:         protoData,
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
func (s *service) formatMatchData(data *types.RawMatchState) ([]byte, error) {
	// format data for marshalling as protobuf
	playerMatchRes := make([]*pb.PlayerMatchResult, len(data.Players))

	for i, player := range data.Players {
		playerMatchRes[i] = &pb.PlayerMatchResult{
			MemberId: player.MemberID,
			Username: player.Username,
			Kills:    player.Kills,
			Deaths:   player.Deaths,
		}
	}

	// marshal to protobuf
	protoData, err := proto.Marshal(&pb.MatchEndedEvent{
		SessionId:      string(data.SessionID.String()),
		MatchStartedAt: timestamppb.New(data.StartedAt),
		MatchEndedAt:   timestamppb.New(data.EndedAt),
		Players:        playerMatchRes,
	})

	if err != nil {
		slog.Error("could not marshal end match data to MatchEndedEvent proto",
			"session_id", data.SessionID,
			"error", err,
		)

		return nil, err
	}

	return protoData, nil
}

/**
* Derives results from raw game state.
**/
// TODO: need to add enough raw states to deterine win
func (s *service) calculateResult(rawMatchState *types.RawMatchState) (*commontypes.MatchEndState, error) {

	for _, player := range rawMatchState.Players {
		slog.Info(
			"yeee", "player", player)
	}
	return nil, nil
}
