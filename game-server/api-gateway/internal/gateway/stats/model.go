package stats

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
)

type StatsClient interface {
	GetPlayerStats(ctx context.Context, req *pb.GetPlayerMatchStatsRequest) (*pb.PlayerMatchStats, error)
}

