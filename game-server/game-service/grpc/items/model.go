package grpcitems

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/items"
)

// ItemsClient defines the interface for items service gRPC client
type ItemsClient interface {
	// CreateWeapon creates a new weapon
	CreateWeapon(ctx context.Context, req *pb.CreateWeaponRequest) (*pb.Weapon, error)

	// GetWeaponWithTemplateByID gets a weapon with template information by ID
	GetWeaponWithTemplateByID(ctx context.Context, req *pb.GetWeaponRequest) (*pb.WeaponDetail, error)

	// ListWeaponsWithTemplate lists all weapons with template information
	ListWeaponsWithTemplate(ctx context.Context) (*pb.ListWeaponsResponse, error)
}
