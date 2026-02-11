package item

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/items"
)

type ItemClient interface {
	// Item type and rarity operations (for dropdown options)
	ListItemTypes(ctx context.Context) (*pb.ListItemTypesResponse, error)
	ListItemRarities(ctx context.Context) (*pb.ListItemRaritiesResponse, error)

	CreateWeapon(ctx context.Context, req *pb.CreateWeaponRequest) (*pb.Weapon, error)
	ListWeaponsWithTemplate(ctx context.Context) (*pb.ListWeaponsResponse, error)
	CreateItemTemplate(ctx context.Context, req *pb.CreateItemTemplateRequest) (*pb.ItemTemplate, error)

	// Complete item operations
	CreateCompleteWeapon(ctx context.Context, req *pb.CreateCompleteWeaponRequest) (*pb.WeaponDetail, error)
	CreateCompleteArmor(ctx context.Context, req *pb.CreateCompleteArmorRequest) (*pb.ArmorDetail, error)
	CreateCompleteConsumable(ctx context.Context, req *pb.CreateCompleteConsumableRequest) (*pb.ConsumableDetail, error)
}
