package items

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/items"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	pb.UnimplementedItemsServiceServer
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateWeapon creates a new weapon (gRPC endpoint)
func (h *Handler) CreateWeapon(ctx context.Context, req *pb.CreateWeaponRequest) (*pb.Weapon, error) {
	// Parse UUIDs
	typeID, err := uuid.Parse(req.TypeId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid type_id: %v", err)
	}

	rarityID, err := uuid.Parse(req.RarityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rarity_id: %v", err)
	}

	// Create weapon request
	critRate := float64(req.CriticalRate)
	createReq := &CreateWeaponRequest{
		TypeID:       typeID,
		RarityID:     rarityID,
		AttackPower:  int(req.AttackPower),
		Durability:   int(req.Durability),
		CriticalRate: &critRate,
		WeaponType:   &req.WeaponType,
		Description:  &req.Description,
	}

	// Call service
	weapon, err := h.service.CreateWeapon(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create weapon: %v", err)
	}

	// Convert to proto message
	var pbCritRate float32
	if weapon.CriticalRate != nil {
		pbCritRate = float32(*weapon.CriticalRate)
	}
	var pbWeaponType, pbDescription string
	if weapon.WeaponType != nil {
		pbWeaponType = *weapon.WeaponType
	}
	if weapon.Description != nil {
		pbDescription = *weapon.Description
	}

	return &pb.Weapon{
		Id:           weapon.ID.String(),
		TypeId:       weapon.TypeID.String(),
		RarityId:     weapon.RarityID.String(),
		AttackPower:  int32(weapon.AttackPower),
		Durability:   int32(weapon.Durability),
		CriticalRate: pbCritRate,
		WeaponType:   pbWeaponType,
		Description:  pbDescription,
		CreatedAt:    timestamppb.New(weapon.CreatedAt),
		UpdatedAt:    timestamppb.New(weapon.UpdatedAt),
	}, nil
}

// GetWeaponWithTemplateByID retrieves a weapon with its template information by ID (gRPC endpoint)
func (h *Handler) GetWeaponWithTemplateByID(ctx context.Context, req *pb.GetWeaponRequest) (*pb.WeaponDetail, error) {
	// Parse weapon ID
	weaponID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid weapon id: %v", err)
	}

	// Call service
	weapon, err := h.service.GetWeaponWithTemplateByID(ctx, weaponID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get weapon: %v", err)
	}

	// Convert to proto message
	var critRate float32
	if weapon.CriticalRate != nil {
		critRate = float32(*weapon.CriticalRate)
	}
	var weaponType, description, iconURL string
	if weapon.WeaponType != nil {
		weaponType = *weapon.WeaponType
	}
	if weapon.Description != nil {
		description = *weapon.Description
	}
	if weapon.IconURL != nil {
		iconURL = *weapon.IconURL
	}

	return &pb.WeaponDetail{
		// Weapon fields
		Id:           weapon.ID.String(),
		TypeId:       weapon.TypeID.String(),
		RarityId:     weapon.RarityID.String(),
		AttackPower:  int32(weapon.AttackPower),
		Durability:   int32(weapon.Durability),
		CriticalRate: critRate,
		WeaponType:   weaponType,
		Description:  description,

		// ItemTemplate fields
		ItemTemplateId: weapon.ItemTemplateID.String(),
		ItemName:       weapon.ItemName,
		ItemCode:       weapon.ItemCode,
		IconUrl:        iconURL,
		IsTradeable:    weapon.IsTradeable,
		IsDroppable:    weapon.IsDroppable,
		RequiredLevel:  int32(weapon.RequiredLevel),
		BaseSellPrice:  int32(weapon.BaseSellPrice),
		BaseBuyPrice:   int32(weapon.BaseBuyPrice),

		CreatedAt: timestamppb.New(weapon.CreatedAt),
		UpdatedAt: timestamppb.New(weapon.UpdatedAt),
	}, nil
}

// ListWeaponsWithTemplate retrieves all weapons with their template information (gRPC endpoint)
func (h *Handler) ListWeaponsWithTemplate(ctx context.Context, _ *emptypb.Empty) (*pb.ListWeaponsResponse, error) {
	// Call service
	weapons, err := h.service.ListWeaponsWithTemplate(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list weapons: %v", err)
	}

	// Convert to proto messages
	pbWeapons := make([]*pb.WeaponDetail, len(weapons))
	for i, weapon := range weapons {
		// Convert pointer types
		var critRate float32
		if weapon.CriticalRate != nil {
			critRate = float32(*weapon.CriticalRate)
		}
		var weaponType, description, iconURL string
		if weapon.WeaponType != nil {
			weaponType = *weapon.WeaponType
		}
		if weapon.Description != nil {
			description = *weapon.Description
		}
		if weapon.IconURL != nil {
			iconURL = *weapon.IconURL
		}

		pbWeapons[i] = &pb.WeaponDetail{
			// Weapon fields
			Id:           weapon.ID.String(),
			TypeId:       weapon.TypeID.String(),
			RarityId:     weapon.RarityID.String(),
			AttackPower:  int32(weapon.AttackPower),
			Durability:   int32(weapon.Durability),
			CriticalRate: critRate,
			WeaponType:   weaponType,
			Description:  description,

			// ItemTemplate fields
			ItemTemplateId: weapon.ItemTemplateID.String(),
			ItemName:       weapon.ItemName,
			ItemCode:       weapon.ItemCode,
			IconUrl:        iconURL,
			IsTradeable:    weapon.IsTradeable,
			IsDroppable:    weapon.IsDroppable,
			RequiredLevel:  int32(weapon.RequiredLevel),
			BaseSellPrice:  int32(weapon.BaseSellPrice),
			BaseBuyPrice:   int32(weapon.BaseBuyPrice),

			CreatedAt: timestamppb.New(weapon.CreatedAt),
			UpdatedAt: timestamppb.New(weapon.UpdatedAt),
		}
	}

	return &pb.ListWeaponsResponse{
		Weapons: pbWeapons,
	}, nil
}
