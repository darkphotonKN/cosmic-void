package grpcitems

import (
	"context"
	"fmt"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/items"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	serviceName = "items"
)

// Client implements ItemsClient interface
type Client struct {
	registry discovery.Registry
}

// NewClient creates a new items gRPC client
func NewClient(registry discovery.Registry) ItemsClient {
	return &Client{
		registry: registry,
	}
}

// CreateWeapon creates a new weapon
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

// GetWeaponWithTemplateByID gets a weapon with template information by ID
func (c *Client) GetWeaponWithTemplateByID(ctx context.Context, req *pb.GetWeaponRequest) (*pb.WeaponDetail, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to items service: %w", err)
	}
	defer conn.Close()

	client := pb.NewItemsServiceClient(conn)
	weapon, err := client.GetWeaponWithTemplateByID(ctx, req)
	return weapon, err
}

// ListWeaponsWithTemplate lists all weapons with template information
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
