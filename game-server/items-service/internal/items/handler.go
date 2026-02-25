package items

import (
	"context"
	"log/slog"
	"strings"

	authpb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/items"
	commontypes "github.com/darkphotonKN/cosmic-void-server/common/constants/types"
	"github.com/darkphotonKN/cosmic-void-server/items-service/grpc/auth"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	pb.UnimplementedItemsServiceServer
	service    Service
	authClient auth.AuthClient
}

func NewHandler(service Service, authClient auth.AuthClient) *Handler {
	return &Handler{
		service:    service,
		authClient: authClient,
	}
}

// checkAdminPermission checks if the user has admin permission
func (h *Handler) checkAdminPermission(ctx context.Context, userID string) error {
	member, err := h.authClient.GetMember(ctx, &authpb.GetMemberRequest{Id: userID})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get member info: %v", err)
	}

	// Convert string role to enum for comparison
	memberRole := stringToRole(member.Role)
	if memberRole != commontypes.Role_ROLE_ADMIN {
		return status.Error(codes.PermissionDenied, "admin permission required")
	}

	return nil
}

// stringToRole converts database role string to Role enum
func stringToRole(roleStr string) commontypes.Role {
	switch strings.ToLower(roleStr) {
	case "player":
		return commontypes.Role_ROLE_PLAYER
	case "admin":
		return commontypes.Role_ROLE_ADMIN
	default:
		return commontypes.Role_ROLE_UNSPECIFIED
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

// ListArmorsWithTemplate retrieves all armors with their template information (gRPC endpoint)
func (h *Handler) ListArmorsWithTemplate(ctx context.Context, _ *emptypb.Empty) (*pb.ListArmorsResponse, error) {
	armors, err := h.service.ListArmorsWithTemplate(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list armors: %v", err)
	}

	pbArmors := make([]*pb.ArmorDetail, len(armors))
	for i, armor := range armors {
		var magicResistance int32
		if armor.MagicResistance != nil {
			magicResistance = int32(*armor.MagicResistance)
		}
		var armorSlot, description, iconURL string
		if armor.ArmorSlot != nil {
			armorSlot = *armor.ArmorSlot
		}
		if armor.Description != nil {
			description = *armor.Description
		}
		if armor.IconURL != nil {
			iconURL = *armor.IconURL
		}

		pbArmors[i] = &pb.ArmorDetail{
			Id:              armor.ID.String(),
			TypeId:          armor.TypeID.String(),
			RarityId:        armor.RarityID.String(),
			DefenseRating:   int32(armor.DefenseRating),
			Durability:      int32(armor.Durability),
			MagicResistance: magicResistance,
			ArmorSlot:       armorSlot,
			Description:     description,

			ItemTemplateId: armor.ItemTemplateID.String(),
			ItemName:       armor.ItemName,
			ItemCode:       armor.ItemCode,
			IconUrl:        iconURL,
			IsTradeable:    armor.IsTradeable,
			IsDroppable:    armor.IsDroppable,
			RequiredLevel:  int32(armor.RequiredLevel),
			BaseSellPrice:  int32(armor.BaseSellPrice),
			BaseBuyPrice:   int32(armor.BaseBuyPrice),

			CreatedAt: timestamppb.New(armor.CreatedAt),
			UpdatedAt: timestamppb.New(armor.UpdatedAt),
		}
	}

	return &pb.ListArmorsResponse{
		Armors: pbArmors,
	}, nil
}

// ListConsumablesWithTemplate retrieves all consumables with their template information (gRPC endpoint)
func (h *Handler) ListConsumablesWithTemplate(ctx context.Context, _ *emptypb.Empty) (*pb.ListConsumablesResponse, error) {
	consumables, err := h.service.ListConsumablesWithTemplate(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list consumables: %v", err)
	}

	pbConsumables := make([]*pb.ConsumableDetail, len(consumables))
	for i, consumable := range consumables {
		var healingAmount, manaAmount, buffDuration int32
		if consumable.HealingAmount != nil {
			healingAmount = int32(*consumable.HealingAmount)
		}
		if consumable.ManaAmount != nil {
			manaAmount = int32(*consumable.ManaAmount)
		}
		if consumable.BuffDuration != nil {
			buffDuration = int32(*consumable.BuffDuration)
		}
		var description, iconURL string
		if consumable.Description != nil {
			description = *consumable.Description
		}
		if consumable.IconURL != nil {
			iconURL = *consumable.IconURL
		}

		pbConsumables[i] = &pb.ConsumableDetail{
			Id:            consumable.ID.String(),
			TypeId:        consumable.TypeID.String(),
			RarityId:      consumable.RarityID.String(),
			HealingAmount: healingAmount,
			ManaAmount:    manaAmount,
			BuffDuration:  buffDuration,
			MaxStackSize:  int32(consumable.MaxStackSize),
			Description:   description,

			ItemTemplateId: consumable.ItemTemplateID.String(),
			ItemName:       consumable.ItemName,
			ItemCode:       consumable.ItemCode,
			IconUrl:        iconURL,
			IsTradeable:    consumable.IsTradeable,
			IsDroppable:    consumable.IsDroppable,
			RequiredLevel:  int32(consumable.RequiredLevel),
			BaseSellPrice:  int32(consumable.BaseSellPrice),
			BaseBuyPrice:   int32(consumable.BaseBuyPrice),

			CreatedAt: timestamppb.New(consumable.CreatedAt),
			UpdatedAt: timestamppb.New(consumable.UpdatedAt),
		}
	}

	return &pb.ListConsumablesResponse{
		Consumables: pbConsumables,
	}, nil
}

// CreateItemTemplate creates a new item template (gRPC endpoint)
// This will also send an event to RabbitMQ for notification-service
func (h *Handler) CreateItemTemplate(ctx context.Context, req *pb.CreateItemTemplateRequest) (*pb.ItemTemplate, error) {

	err := h.checkAdminPermission(ctx, req.UserId)
	if err != nil {
		slog.Error("role is not Admin", "err", err)
		return nil, err
	}
	// Parse UUIDs
	typeID, err := uuid.Parse(req.TypeId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid type_id: %v", err)
	}

	rarityID, err := uuid.Parse(req.RarityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rarity_id: %v", err)
	}

	itemID, err := uuid.Parse(req.ItemId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid item_id: %v", err)
	}

	// Build service request with optional fields
	createReq := &CreateItemTemplateRequest{
		UserId:   req.UserId,
		ItemName: req.ItemName,
		ItemCode: req.ItemCode,
		TypeID:   typeID,
		RarityID: rarityID,
		ItemType: req.ItemType,
		ItemID:   itemID,
	}

	// Handle optional fields
	if req.IconUrl != nil {
		createReq.IconURL = req.IconUrl
	}
	if req.IsTradeable != nil {
		createReq.IsTradeable = req.IsTradeable
	}
	if req.IsDroppable != nil {
		createReq.IsDroppable = req.IsDroppable
	}
	if req.RequiredLevel != nil {
		reqLevel := int(*req.RequiredLevel)
		createReq.RequiredLevel = &reqLevel
	}
	if req.BaseSellPrice != nil {
		sellPrice := int(*req.BaseSellPrice)
		createReq.BaseSellPrice = &sellPrice
	}
	if req.BaseBuyPrice != nil {
		buyPrice := int(*req.BaseBuyPrice)
		createReq.BaseBuyPrice = &buyPrice
	}

	// Call service (will send RabbitMQ event)
	template, err := h.service.CreateItemTemplate(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create item template: %v", err)
	}

	// Convert to proto message
	var iconURL string
	if template.IconURL != nil {
		iconURL = *template.IconURL
	}

	return &pb.ItemTemplate{
		Id:            template.ID.String(),
		ItemName:      template.ItemName,
		ItemCode:      template.ItemCode,
		TypeId:        template.TypeID.String(),
		RarityId:      template.RarityID.String(),
		ItemType:      template.ItemType,
		ItemId:        template.ItemID.String(),
		IconUrl:       iconURL,
		IsTradeable:   template.IsTradeable,
		IsDroppable:   template.IsDroppable,
		RequiredLevel: int32(template.RequiredLevel),
		BaseSellPrice: int32(template.BaseSellPrice),
		BaseBuyPrice:  int32(template.BaseBuyPrice),
		CreatedAt:     timestamppb.New(template.CreatedAt),
		UpdatedAt:     timestamppb.New(template.UpdatedAt),
	}, nil
}

// CreateCompleteWeapon creates a complete weapon (weapon + template) in one operation
// This will automatically send an event to RabbitMQ for notification-service
func (h *Handler) CreateCompleteWeapon(ctx context.Context, req *pb.CreateCompleteWeaponRequest) (*pb.WeaponDetail, error) {
	// Check admin permission
	if err := h.checkAdminPermission(ctx, req.UserId); err != nil {
		slog.Error("role is not Admin", "err", err)
		return nil, err
	}

	// Parse UUIDs
	typeID, err := uuid.Parse(req.TypeId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid type_id: %v", err)
	}

	rarityID, err := uuid.Parse(req.RarityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rarity_id: %v", err)
	}

	// Build service request
	critRate := float64(req.CriticalRate)
	createReq := &CreateCompleteWeaponRequest{
		UserId:       req.UserId,
		ItemName:     req.ItemName,
		ItemCode:     req.ItemCode,
		TypeID:       typeID,
		RarityID:     rarityID,
		AttackPower:  int(req.AttackPower),
		Durability:   int(req.Durability),
		CriticalRate: &critRate,
		WeaponType:   &req.WeaponType,
		Description:  &req.Description,
	}

	// Handle optional template fields
	if req.IconUrl != nil {
		createReq.IconURL = req.IconUrl
	}
	if req.IsTradeable != nil {
		createReq.IsTradeable = req.IsTradeable
	}
	if req.IsDroppable != nil {
		createReq.IsDroppable = req.IsDroppable
	}
	if req.RequiredLevel != nil {
		reqLevel := int(*req.RequiredLevel)
		createReq.RequiredLevel = &reqLevel
	}
	if req.BaseSellPrice != nil {
		sellPrice := int(*req.BaseSellPrice)
		createReq.BaseSellPrice = &sellPrice
	}
	if req.BaseBuyPrice != nil {
		buyPrice := int(*req.BaseBuyPrice)
		createReq.BaseBuyPrice = &buyPrice
	}

	// Call service (will create weapon, create template, send RabbitMQ event)
	weaponWithTemplate, err := h.service.CreateCompleteWeapon(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create complete weapon: %v", err)
	}

	// Convert to proto message
	var critRatePb float32
	if weaponWithTemplate.CriticalRate != nil {
		critRatePb = float32(*weaponWithTemplate.CriticalRate)
	}
	var weaponType, description, iconURL string
	if weaponWithTemplate.WeaponType != nil {
		weaponType = *weaponWithTemplate.WeaponType
	}
	if weaponWithTemplate.Description != nil {
		description = *weaponWithTemplate.Description
	}
	if weaponWithTemplate.IconURL != nil {
		iconURL = *weaponWithTemplate.IconURL
	}

	return &pb.WeaponDetail{
		// Weapon fields
		Id:           weaponWithTemplate.ID.String(),
		TypeId:       weaponWithTemplate.TypeID.String(),
		RarityId:     weaponWithTemplate.RarityID.String(),
		AttackPower:  int32(weaponWithTemplate.AttackPower),
		Durability:   int32(weaponWithTemplate.Durability),
		CriticalRate: critRatePb,
		WeaponType:   weaponType,
		Description:  description,

		// ItemTemplate fields
		ItemTemplateId: weaponWithTemplate.ItemTemplateID.String(),
		ItemName:       weaponWithTemplate.ItemName,
		ItemCode:       weaponWithTemplate.ItemCode,
		IconUrl:        iconURL,
		IsTradeable:    weaponWithTemplate.IsTradeable,
		IsDroppable:    weaponWithTemplate.IsDroppable,
		RequiredLevel:  int32(weaponWithTemplate.RequiredLevel),
		BaseSellPrice:  int32(weaponWithTemplate.BaseSellPrice),
		BaseBuyPrice:   int32(weaponWithTemplate.BaseBuyPrice),

		CreatedAt: timestamppb.New(weaponWithTemplate.CreatedAt),
		UpdatedAt: timestamppb.New(weaponWithTemplate.UpdatedAt),
	}, nil
}

// CreateCompleteArmor creates a complete armor (armor + template) in one operation
func (h *Handler) CreateCompleteArmor(ctx context.Context, req *pb.CreateCompleteArmorRequest) (*pb.ArmorDetail, error) {
	// Check admin permission
	if err := h.checkAdminPermission(ctx, req.UserId); err != nil {
		slog.Error("role is not Admin", "err", err)
		return nil, err
	}

	// Parse UUIDs
	typeID, err := uuid.Parse(req.TypeId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid type_id: %v", err)
	}

	rarityID, err := uuid.Parse(req.RarityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rarity_id: %v", err)
	}

	// Build service request
	magicRes := int(req.MagicResistance)
	createReq := &CreateCompleteArmorRequest{
		UserId:          req.UserId,
		ItemName:        req.ItemName,
		ItemCode:        req.ItemCode,
		TypeID:          typeID,
		RarityID:        rarityID,
		DefenseRating:   int(req.DefenseRating),
		Durability:      int(req.Durability),
		MagicResistance: &magicRes,
		ArmorSlot:       &req.ArmorSlot,
		Description:     &req.Description,
	}

	// Handle optional template fields
	if req.IconUrl != nil {
		createReq.IconURL = req.IconUrl
	}
	if req.IsTradeable != nil {
		createReq.IsTradeable = req.IsTradeable
	}
	if req.IsDroppable != nil {
		createReq.IsDroppable = req.IsDroppable
	}
	if req.RequiredLevel != nil {
		reqLevel := int(*req.RequiredLevel)
		createReq.RequiredLevel = &reqLevel
	}
	if req.BaseSellPrice != nil {
		sellPrice := int(*req.BaseSellPrice)
		createReq.BaseSellPrice = &sellPrice
	}
	if req.BaseBuyPrice != nil {
		buyPrice := int(*req.BaseBuyPrice)
		createReq.BaseBuyPrice = &buyPrice
	}

	// Call service
	armorWithTemplate, err := h.service.CreateCompleteArmor(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create complete armor: %v", err)
	}

	// Convert to proto message
	var magicResPb int32
	if armorWithTemplate.MagicResistance != nil {
		magicResPb = int32(*armorWithTemplate.MagicResistance)
	}
	var armorSlot, description, iconURL string
	if armorWithTemplate.ArmorSlot != nil {
		armorSlot = *armorWithTemplate.ArmorSlot
	}
	if armorWithTemplate.Description != nil {
		description = *armorWithTemplate.Description
	}
	if armorWithTemplate.IconURL != nil {
		iconURL = *armorWithTemplate.IconURL
	}

	return &pb.ArmorDetail{
		Id:              armorWithTemplate.ID.String(),
		TypeId:          armorWithTemplate.TypeID.String(),
		RarityId:        armorWithTemplate.RarityID.String(),
		DefenseRating:   int32(armorWithTemplate.DefenseRating),
		Durability:      int32(armorWithTemplate.Durability),
		MagicResistance: magicResPb,
		ArmorSlot:       armorSlot,
		Description:     description,

		ItemTemplateId: armorWithTemplate.ItemTemplateID.String(),
		ItemName:       armorWithTemplate.ItemName,
		ItemCode:       armorWithTemplate.ItemCode,
		IconUrl:        iconURL,
		IsTradeable:    armorWithTemplate.IsTradeable,
		IsDroppable:    armorWithTemplate.IsDroppable,
		RequiredLevel:  int32(armorWithTemplate.RequiredLevel),
		BaseSellPrice:  int32(armorWithTemplate.BaseSellPrice),
		BaseBuyPrice:   int32(armorWithTemplate.BaseBuyPrice),

		CreatedAt: timestamppb.New(armorWithTemplate.CreatedAt),
		UpdatedAt: timestamppb.New(armorWithTemplate.UpdatedAt),
	}, nil
}

// CreateCompleteConsumable creates a complete consumable (consumable + template) in one operation
func (h *Handler) CreateCompleteConsumable(ctx context.Context, req *pb.CreateCompleteConsumableRequest) (*pb.ConsumableDetail, error) {
	// Check admin permission
	if err := h.checkAdminPermission(ctx, req.UserId); err != nil {
		slog.Error("role is not Admin", "err", err)
		return nil, err
	}

	// Parse UUIDs
	typeID, err := uuid.Parse(req.TypeId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid type_id: %v", err)
	}

	rarityID, err := uuid.Parse(req.RarityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid rarity_id: %v", err)
	}

	// Build service request
	healingAmt := int(req.HealingAmount)
	manaAmt := int(req.ManaAmount)
	buffDur := int(req.BuffDuration)
	createReq := &CreateCompleteConsumableRequest{
		UserId:        req.UserId,
		ItemName:      req.ItemName,
		ItemCode:      req.ItemCode,
		TypeID:        typeID,
		RarityID:      rarityID,
		HealingAmount: &healingAmt,
		ManaAmount:    &manaAmt,
		BuffDuration:  &buffDur,
		MaxStackSize:  int(req.MaxStackSize),
		Description:   &req.Description,
	}

	// Handle optional template fields
	if req.IconUrl != nil {
		createReq.IconURL = req.IconUrl
	}
	if req.IsTradeable != nil {
		createReq.IsTradeable = req.IsTradeable
	}
	if req.IsDroppable != nil {
		createReq.IsDroppable = req.IsDroppable
	}
	if req.RequiredLevel != nil {
		reqLevel := int(*req.RequiredLevel)
		createReq.RequiredLevel = &reqLevel
	}
	if req.BaseSellPrice != nil {
		sellPrice := int(*req.BaseSellPrice)
		createReq.BaseSellPrice = &sellPrice
	}
	if req.BaseBuyPrice != nil {
		buyPrice := int(*req.BaseBuyPrice)
		createReq.BaseBuyPrice = &buyPrice
	}

	// Call service
	consumableWithTemplate, err := h.service.CreateCompleteConsumable(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create complete consumable: %v", err)
	}

	// Convert to proto message
	var healingAmtPb, manaAmtPb, buffDurPb int32
	if consumableWithTemplate.HealingAmount != nil {
		healingAmtPb = int32(*consumableWithTemplate.HealingAmount)
	}
	if consumableWithTemplate.ManaAmount != nil {
		manaAmtPb = int32(*consumableWithTemplate.ManaAmount)
	}
	if consumableWithTemplate.BuffDuration != nil {
		buffDurPb = int32(*consumableWithTemplate.BuffDuration)
	}
	var description, iconURL string
	if consumableWithTemplate.Description != nil {
		description = *consumableWithTemplate.Description
	}
	if consumableWithTemplate.IconURL != nil {
		iconURL = *consumableWithTemplate.IconURL
	}

	return &pb.ConsumableDetail{
		Id:            consumableWithTemplate.ID.String(),
		TypeId:        consumableWithTemplate.TypeID.String(),
		RarityId:      consumableWithTemplate.RarityID.String(),
		HealingAmount: healingAmtPb,
		ManaAmount:    manaAmtPb,
		BuffDuration:  buffDurPb,
		MaxStackSize:  int32(consumableWithTemplate.MaxStackSize),
		Description:   description,

		ItemTemplateId: consumableWithTemplate.ItemTemplateID.String(),
		ItemName:       consumableWithTemplate.ItemName,
		ItemCode:       consumableWithTemplate.ItemCode,
		IconUrl:        iconURL,
		IsTradeable:    consumableWithTemplate.IsTradeable,
		IsDroppable:    consumableWithTemplate.IsDroppable,
		RequiredLevel:  int32(consumableWithTemplate.RequiredLevel),
		BaseSellPrice:  int32(consumableWithTemplate.BaseSellPrice),
		BaseBuyPrice:   int32(consumableWithTemplate.BaseBuyPrice),

		CreatedAt: timestamppb.New(consumableWithTemplate.CreatedAt),
		UpdatedAt: timestamppb.New(consumableWithTemplate.UpdatedAt),
	}, nil
}

// ListItemTypes returns all item types (gRPC endpoint)
func (h *Handler) ListItemTypes(ctx context.Context, req *emptypb.Empty) (*pb.ListItemTypesResponse, error) {
	// Call service to get all item types
	itemTypes, err := h.service.ListItemTypes(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list item types: %v", err)
	}

	// Convert to protobuf
	pbItemTypes := make([]*pb.ItemType, 0, len(itemTypes))
	for _, itemType := range itemTypes {
		var description string
		if itemType.Description != nil {
			description = *itemType.Description
		}

		pbItemTypes = append(pbItemTypes, &pb.ItemType{
			Id:          itemType.ID.String(),
			Name:        itemType.Name,
			Description: description,
			CreatedAt:   timestamppb.New(itemType.CreatedAt),
			UpdatedAt:   timestamppb.New(itemType.UpdatedAt),
		})
	}

	return &pb.ListItemTypesResponse{
		ItemTypes: pbItemTypes,
	}, nil
}

// ListItemRarities returns all item rarities (gRPC endpoint)
func (h *Handler) ListItemRarities(ctx context.Context, req *emptypb.Empty) (*pb.ListItemRaritiesResponse, error) {
	// Call service to get all item rarities
	rarities, err := h.service.ListItemRarities(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list item rarities: %v", err)
	}

	// Convert to protobuf
	pbRarities := make([]*pb.ItemRarity, 0, len(rarities))
	for _, rarity := range rarities {
		pbRarities = append(pbRarities, &pb.ItemRarity{
			Id:          rarity.ID.String(),
			Name:        rarity.RarityName,
			Description: "", // ItemRarity doesn't have a description field in DB
			CreatedAt:   timestamppb.New(rarity.CreatedAt),
			UpdatedAt:   timestamppb.New(rarity.UpdatedAt),
		})
	}

	return &pb.ListItemRaritiesResponse{
		ItemRarities: pbRarities,
	}, nil
}
