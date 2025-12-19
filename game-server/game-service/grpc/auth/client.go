package grpcauth

import (
	"context"
	"fmt"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
)

const (
	serviceName = "auth"
)

type Client struct {
	registry discovery.Registry
}

func NewClient(registry discovery.Registry) AuthClient {
	return &Client{
		registry: registry,
	}
}

func (c *Client) GetMember(ctx context.Context, req *pb.GetMemberRequest) (*pb.Member, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	member, err := client.GetMember(ctx, req)
	return member, err
}

func (c *Client) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	response, err := client.ValidateToken(ctx, req)
	return response, err
}
