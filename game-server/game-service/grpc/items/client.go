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

// get all items under an aggregated type
func (c *Client) ListItemTemplates(ctx context.Context) (*pb.ListItemTemplatesResponse, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to items service: %w", err)
	}
	defer conn.Close()

	client := pb.NewItemsServiceClient(conn)

	items, err := client.ListItemTemplates(ctx, &emptypb.Empty{})
	return items, err
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

// ListArmorsWithTemplate lists all armors with template information
func (c *Client) ListArmorsWithTemplate(ctx context.Context) (*pb.ListArmorsResponse, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to items service: %w", err)
	}
	defer conn.Close()

	client := pb.NewItemsServiceClient(conn)
	armors, err := client.ListArmorsWithTemplate(ctx, &emptypb.Empty{})
	return armors, err
}

// ListConsumablesWithTemplate lists all consumables with template information
func (c *Client) ListConsumablesWithTemplate(ctx context.Context) (*pb.ListConsumablesResponse, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to items service: %w", err)
	}
	defer conn.Close()

	client := pb.NewItemsServiceClient(conn)
	consumables, err := client.ListConsumablesWithTemplate(ctx, &emptypb.Empty{})
	return consumables, err
}

func (c *Client) GetLoadout(ctx context.Context, req *pb.GetLoadoutRequest) (*pb.GetLoadoutResponse, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to items service: %w", err)
	}
	defer conn.Close()

	client := pb.NewItemsServiceClient(conn)
	return client.GetLoadout(ctx, req)
}

func (c *Client) GetLoadoutWithItems(ctx context.Context, req *pb.GetLoadoutWithItemsRequest) (*pb.GetLoadoutWithItemsResponse, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to items service: %w", err)
	}
	defer conn.Close()

	client := pb.NewItemsServiceClient(conn)
	return client.GetLoadoutWithItems(ctx, req)
}

func (c *Client) ListItemInstances(ctx context.Context, req *pb.ListItemInstancesRequest) (*pb.ListItemInstancesResponse, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to items service: %w", err)
	}
	defer conn.Close()

	client := pb.NewItemsServiceClient(conn)
	return client.ListItemInstances(ctx, req)
}
