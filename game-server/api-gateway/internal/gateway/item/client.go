package item

import (
	"context"
	"fmt"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/items"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Client struct {
	registry discovery.Registry
}

const (
	serviceName = "items"
)

func NewClient(registry discovery.Registry) ItemClient {
	return &Client{
		registry: registry,
	}
}

func (c *Client) CreateWeapon(ctx context.Context, req *pb.CreateWeaponRequest) (*pb.Weapon, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to items service: %w", err)
	}
	defer conn.Close()
	client := pb.NewItemsServiceClient(conn)
	weapon, err := client.CreateWeapon(ctx, req)
	return weapon, err
}

func (c *Client) ListWeaponsWithTemplate(ctx context.Context) (*pb.ListWeaponsResponse, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to items service: %w", err)
	}
	defer conn.Close()

	client := pb.NewItemsServiceClient(conn)
	weapons, err := client.ListWeaponsWithTemplate(ctx, &emptypb.Empty{})
	return weapons, err
}

func (c *Client) CreateItemTemplate(ctx context.Context, req *pb.CreateItemTemplateRequest) (*pb.ItemTemplate, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to items service: %w", err)
	}
	defer conn.Close()

	client := pb.NewItemsServiceClient(conn)
	template, err := client.CreateItemTemplate(ctx, req)
	return template, err
}
