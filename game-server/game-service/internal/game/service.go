package game

import (
	"context"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	commontypes "github.com/darkphotonKN/cosmic-void-server/common/types"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
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

func (s *service) PublishMatchComplete(ctx context.Context, data *commontypes.MatchEndState) error {

	// TODO: pass game state
	// formatData, err := s.formatMatchData()
	//
	// if err != nil {
	// 	return err
	// }

	// format data for marshalling as protobuf
	playerMatchRes := make([]*pb.PlayerMatchResult, len(data.PlayerMatchResults))

	for i, player := range data.PlayerMatchResults {
		playerMatchRes[i] = &pb.PlayerMatchResult{
			MemberId:      player.MemberID,
			Username:      player.Username,
			Win:           player.Win,
			FinalPosition: player.FinalPosition,
			Kills:         player.Kills,
			Deaths:        player.Deaths,
		}
	}

	// marshal to protobuf
	protoData, err := proto.Marshal(&pb.MatchEndedEvent{
		SessionId:      string(data.SessionID.String()),
		MatchStartedAt: timestamppb.New(data.MatchStartedAt),
		MatchEndedAt:   timestamppb.New(data.MatchEndedAt),
		Players:        playerMatchRes,
	})

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
func (s *service) formatMatchData(rawState []*ecs.Entity) (*commontypes.MatchEndState, error) {
	return nil, nil
}
