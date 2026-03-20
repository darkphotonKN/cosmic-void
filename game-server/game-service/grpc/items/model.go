package grpcitems

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/items"
)

// ItemsClient defines the interface for items service gRPC client
type ItemsClient interface {
	ListItemTemplates(ctx context.Context) (*pb.ListItemTemplatesResponse, error)

	// CreateWeapon creates a new weapon
	CreateWeapon(ctx context.Context, req *pb.CreateWeaponRequest) (*pb.Weapon, error)

	// GetWeaponWithTemplateByID gets a weapon with template information by ID
	GetWeaponWithTemplateByID(ctx context.Context, req *pb.GetWeaponRequest) (*pb.WeaponDetail, error)

	// ListWeaponsWithTemplate lists all weapons with template information
	ListWeaponsWithTemplate(ctx context.Context) (*pb.ListWeaponsResponse, error)

	// ListArmorsWithTemplate lists all armors with template information
	ListArmorsWithTemplate(ctx context.Context) (*pb.ListArmorsResponse, error)

	// ListConsumablesWithTemplate lists all consumables with template information
	ListConsumablesWithTemplate(ctx context.Context) (*pb.ListConsumablesResponse, error)
}
