package stats

import (
	"context"
	"fmt"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
)

const (
	serviceName = "stats"
)

type Client struct {
	registry discovery.Registry
}

func NewClient(registry discovery.Registry) StatsClient {
	return &Client{
		registry: registry,
	}
}

func (c *Client) GetPlayerStats(ctx context.Context, req *pb.GetPlayerMatchStatsRequest) (*pb.PlayerMatchStats, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to stats service: %w", err)
	}
	defer conn.Close()

	client := pb.NewStatsServiceClient(conn)
	stats, err := client.GetPlayerMatchStats(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get player stats: %w", err)
	}

	return stats, nil
}
