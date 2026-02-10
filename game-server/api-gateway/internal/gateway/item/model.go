package item

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/items"
)

type ItemClient interface {
	CreateWeapon(ctx context.Context, req *pb.CreateWeaponRequest) (*pb.Weapon, error)
	ListWeaponsWithTemplate(ctx context.Context) (*pb.ListWeaponsResponse, error)
	CreateItemTemplate(ctx context.Context, req *pb.CreateItemTemplateRequest) (*pb.ItemTemplate, error)
}
