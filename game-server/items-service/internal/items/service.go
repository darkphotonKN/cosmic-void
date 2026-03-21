package items

import (
	"context"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type service struct {
	repo      Repository
	publishCh commonbroker.Publisher
}

func NewService(repo Repository, publishCh commonbroker.Publisher) Service {
	return &service{
		repo:      repo,
		publishCh: publishCh,
	}
}

type Repository interface {
	// ItemType operations
	CreateItemType(ctx context.Context, itemType *ItemType) error
	GetItemTypeByID(ctx context.Context, id uuid.UUID) (*ItemType, error)
	GetItemTypeByCode(ctx context.Context, code string) (*ItemType, error)
	ListItemTypes(ctx context.Context) ([]*ItemType, error)

	// ItemRarity operations
	CreateItemRarity(ctx context.Context, rarity *ItemRarity) error
	GetItemRarityByID(ctx context.Context, id uuid.UUID) (*ItemRarity, error)
	GetItemRarityByCode(ctx context.Context, code string) (*ItemRarity, error)
	ListItemRarities(ctx context.Context) ([]*ItemRarity, error)

	// Weapon operations
	CreateWeapon(ctx context.Context, weapon *Weapon) error
	GetWeaponByID(ctx context.Context, id uuid.UUID) (*Weapon, error)
	ListWeapons(ctx context.Context) ([]*Weapon, error)

	// Weapon operations with item template (JOIN queries)
	GetWeaponWithTemplateByID(ctx context.Context, id uuid.UUID) (*WeaponWithTemplate, error)
	ListWeaponsWithTemplate(ctx context.Context) ([]*WeaponWithTemplate, error)

	// Armor operations with item template (JOIN queries)
	ListArmorsWithTemplate(ctx context.Context) ([]*ArmorWithTemplate, error)

	// Consumable operations with item template (JOIN queries)
	ListConsumablesWithTemplate(ctx context.Context) ([]*ConsumableWithTemplate, error)

	// Armor operations
	CreateArmor(ctx context.Context, armor *Armor) error
	GetArmorByID(ctx context.Context, id uuid.UUID) (*Armor, error)
	ListArmors(ctx context.Context) ([]*Armor, error)

	// Consumable operations
	CreateConsumable(ctx context.Context, consumable *Consumable) error
	GetConsumableByID(ctx context.Context, id uuid.UUID) (*Consumable, error)
	ListConsumables(ctx context.Context) ([]*Consumable, error)

	// ItemTemplate operations
	CreateItemTemplate(ctx context.Context, template *ItemTemplate) error
	GetItemTemplateByID(ctx context.Context, id uuid.UUID) (*ItemTemplate, error)
	GetItemTemplateByCode(ctx context.Context, code string) (*ItemTemplate, error)
	ListItemTemplates(ctx context.Context) ([]*ItemTemplate, error)
	ListItemTemplateAggregates(ctx context.Context) ([]*ItemTemplateAggregate, error)
}

// ==========================================
// ItemType Service Methods
// ==========================================

func (s *service) CreateItemType(ctx context.Context, req *CreateItemTypeRequest) (*ItemType, error) {
	itemType := &ItemType{
		TypeCode:    req.TypeCode,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
		SortOrder:   req.SortOrder,
	}

	if err := s.repo.CreateItemType(ctx, itemType); err != nil {
		return nil, err
	}

	return itemType, nil
}

func (s *service) GetItemType(ctx context.Context, id uuid.UUID) (*ItemType, error) {
	return s.repo.GetItemTypeByID(ctx, id)
}

func (s *service) GetItemTypeByCode(ctx context.Context, code string) (*ItemType, error) {
	return s.repo.GetItemTypeByCode(ctx, code)
}

func (s *service) ListItemTypes(ctx context.Context) ([]*ItemType, error) {
	return s.repo.ListItemTypes(ctx)
}

// ==========================================
// ItemRarity Service Methods
// ==========================================

func (s *service) CreateItemRarity(ctx context.Context, req *CreateItemRarityRequest) (*ItemRarity, error) {
	rarity := &ItemRarity{
		RarityCode:         req.RarityCode,
		RarityName:         req.RarityName,
		ColorHex:           req.ColorHex,
		DropRateMultiplier: req.DropRateMultiplier,
		SortOrder:          req.SortOrder,
	}

	if err := s.repo.CreateItemRarity(ctx, rarity); err != nil {
		return nil, err
	}

	return rarity, nil
}

func (s *service) GetItemRarity(ctx context.Context, id uuid.UUID) (*ItemRarity, error) {
	return s.repo.GetItemRarityByID(ctx, id)
}

func (s *service) GetItemRarityByCode(ctx context.Context, code string) (*ItemRarity, error) {
	return s.repo.GetItemRarityByCode(ctx, code)
}

func (s *service) ListItemRarities(ctx context.Context) ([]*ItemRarity, error) {
	return s.repo.ListItemRarities(ctx)
}

// ==========================================
// Weapon Service Methods
// ==========================================

func (s *service) CreateWeapon(ctx context.Context, req *CreateWeaponRequest) (*Weapon, error) {
	weapon := &Weapon{
		RarityID:     req.RarityID,
		AttackPower:  req.AttackPower,
		CriticalRate: req.CriticalRate,
		WeaponType:   req.WeaponType,
		Description:  req.Description,
	}

	if err := s.repo.CreateWeapon(ctx, weapon); err != nil {
		return nil, err
	}

	// Note: No notification sent here.
	// Notifications are sent when CreateItemTemplate is called (either directly or via CreateCompleteWeapon)

	return weapon, nil
}

func (s *service) GetWeapon(ctx context.Context, id uuid.UUID) (*Weapon, error) {
	return s.repo.GetWeaponByID(ctx, id)
}

func (s *service) ListWeapons(ctx context.Context) ([]*Weapon, error) {
	return s.repo.ListWeapons(ctx)
}

// ==========================================
// Armor Service Methods
// ==========================================

func (s *service) CreateArmor(ctx context.Context, req *CreateArmorRequest) (*Armor, error) {
	armor := &Armor{
		RarityID:        req.RarityID,
		DefenseRating:   req.DefenseRating,
		MagicResistance: req.MagicResistance,
		ArmorSlot:       req.ArmorSlot,
		Description:     req.Description,
	}

	if err := s.repo.CreateArmor(ctx, armor); err != nil {
		return nil, err
	}

	return armor, nil
}

func (s *service) GetArmor(ctx context.Context, id uuid.UUID) (*Armor, error) {
	return s.repo.GetArmorByID(ctx, id)
}

func (s *service) ListArmors(ctx context.Context) ([]*Armor, error) {
	return s.repo.ListArmors(ctx)
}

// ==========================================
// Consumable Service Methods
// ==========================================

func (s *service) CreateConsumable(ctx context.Context, req *CreateConsumableRequest) (*Consumable, error) {
	consumable := &Consumable{
		RarityID:      req.RarityID,
		HealingAmount: req.HealingAmount,
		ManaAmount:    req.ManaAmount,
		BuffDuration:  req.BuffDuration,
		MaxStackSize:  req.MaxStackSize,
		Description:   req.Description,
	}

	if err := s.repo.CreateConsumable(ctx, consumable); err != nil {
		return nil, err
	}

	return consumable, nil
}

func (s *service) GetConsumable(ctx context.Context, id uuid.UUID) (*Consumable, error) {
	return s.repo.GetConsumableByID(ctx, id)
}

func (s *service) ListConsumables(ctx context.Context) ([]*Consumable, error) {
	return s.repo.ListConsumables(ctx)
}

// ==========================================
// ItemTemplate Service Methods
// ==========================================

func (s *service) CreateItemTemplate(ctx context.Context, req *CreateItemTemplateRequest) (*ItemTemplate, error) {
	// Set defaults
	requiredLevel := 1
	if req.RequiredLevel != nil {
		requiredLevel = *req.RequiredLevel
	}

	baseSellPrice := 0
	if req.BaseSellPrice != nil {
		baseSellPrice = *req.BaseSellPrice
	}

	baseBuyPrice := 0
	if req.BaseBuyPrice != nil {
		baseBuyPrice = *req.BaseBuyPrice
	}

	template := &ItemTemplate{
		ItemName:      req.ItemName,
		RarityID:      req.RarityID,
		ItemType:      req.ItemType,
		ItemID:        req.ItemID,
		IconURL:       req.IconURL,
		RequiredLevel: requiredLevel,
		BaseSellPrice: baseSellPrice,
		BaseBuyPrice:  baseBuyPrice,
	}

	if err := s.repo.CreateItemTemplate(ctx, template); err != nil {
		return nil, err
	}

	// Send message to RabbitMQ
	protoData, err := proto.Marshal(&pb.ItemCreatedEvent{
		UserId:   req.UserId,
		Name:     req.ItemName,
		ItemType: req.ItemType,
	})

	if err != nil {
		slog.Error("Error publishing game match end event", "error", err)
		return nil, err
	}
	slog.Info("CreateItemTemplate PublishWithContext")
	if err := s.publishCh.PublishWithContext(ctx, commonconstants.ItemEventsExchange, commonconstants.ItemCreated, commonbroker.Message{
		ContentType:  "application/protobuf",
		Body:         protoData,
		DeliveryMode: commonbroker.Persistent,
	}); err != nil {
		slog.Info("CreateItemTemplate error")
		return nil, err
	}

	return template, nil
}

func (s *service) GetItemTemplate(ctx context.Context, id uuid.UUID) (*ItemTemplate, error) {
	return s.repo.GetItemTemplateByID(ctx, id)
}

func (s *service) GetItemTemplateByCode(ctx context.Context, code string) (*ItemTemplate, error) {
	return s.repo.GetItemTemplateByCode(ctx, code)
}

func (s *service) ListItemTemplates(ctx context.Context) ([]*ItemTemplate, error) {
	return s.repo.ListItemTemplates(ctx)
}

func (s *service) ListItemTemplateAggregates(ctx context.Context) ([]*ItemTemplateAggregate, error) {
	return s.repo.ListItemTemplateAggregates(ctx)
}

// ==========================================
// Weapon with Template Service Methods
// ==========================================

func (s *service) GetWeaponWithTemplateByID(ctx context.Context, id uuid.UUID) (*WeaponWithTemplate, error) {
	return s.repo.GetWeaponWithTemplateByID(ctx, id)
}

func (s *service) ListWeaponsWithTemplate(ctx context.Context) ([]*WeaponWithTemplate, error) {
	return s.repo.ListWeaponsWithTemplate(ctx)
}

func (s *service) ListArmorsWithTemplate(ctx context.Context) ([]*ArmorWithTemplate, error) {
	return s.repo.ListArmorsWithTemplate(ctx)
}

func (s *service) ListConsumablesWithTemplate(ctx context.Context) ([]*ConsumableWithTemplate, error) {
	return s.repo.ListConsumablesWithTemplate(ctx)
}

// ==========================================
// Complete Item Creation Methods
// (Creates both specific item + template, sends notification)
// ==========================================

func (s *service) CreateCompleteWeapon(ctx context.Context, req *CreateCompleteWeaponRequest) (*WeaponWithTemplate, error) {
	// Step 1: Create the weapon with specific attributes
	weaponReq := &CreateWeaponRequest{
		RarityID:     req.RarityID,
		AttackPower:  req.AttackPower,
		CriticalRate: req.CriticalRate,
		WeaponType:   req.WeaponType,
		Description:  req.Description,
	}

	weapon, err := s.CreateWeapon(ctx, weaponReq)
	if err != nil {
		slog.Error("Failed to create weapon", "error", err)
		return nil, err
	}

	// Step 2: Create the item template with common attributes
	templateReq := &CreateItemTemplateRequest{
		UserId:        req.UserId,
		ItemName:      req.ItemName,
		RarityID:      req.RarityID,
		ItemType:      "weapon",
		ItemID:        weapon.ID,
		IconURL:       req.IconURL,
		RequiredLevel: req.RequiredLevel,
		BaseSellPrice: req.BaseSellPrice,
		BaseBuyPrice:  req.BaseBuyPrice,
	}

	template, err := s.CreateItemTemplate(ctx, templateReq)
	if err != nil {
		slog.Error("Failed to create item template for weapon", "weapon_id", weapon.ID, "error", err)
		// TODO: Consider rollback weapon creation here
		return nil, err
	}

	slog.Info("Complete weapon created successfully",
		"weapon_id", weapon.ID,
		"template_id", template.ID,
		"item_name", template.ItemName,
	)

	// Step 3: Return the combined result
	return &WeaponWithTemplate{
		ID:           weapon.ID,
		RarityID:     weapon.RarityID,
		AttackPower:  weapon.AttackPower,
		CriticalRate: weapon.CriticalRate,
		WeaponType:   weapon.WeaponType,
		Description:  weapon.Description,
		CreatedAt:    weapon.CreatedAt,
		UpdatedAt:    weapon.UpdatedAt,
		// Template fields
		ItemTemplateID: template.ID,
		ItemName:       template.ItemName,
		IconURL:        template.IconURL,
		RequiredLevel:  template.RequiredLevel,
		BaseSellPrice:  template.BaseSellPrice,
		BaseBuyPrice:   template.BaseBuyPrice,
	}, nil
}

func (s *service) CreateCompleteArmor(ctx context.Context, req *CreateCompleteArmorRequest) (*ArmorWithTemplate, error) {
	// Step 1: Create the armor with specific attributes
	armorReq := &CreateArmorRequest{
		RarityID:        req.RarityID,
		DefenseRating:   req.DefenseRating,
		MagicResistance: req.MagicResistance,
		ArmorSlot:       req.ArmorSlot,
		Description:     req.Description,
	}

	armor, err := s.CreateArmor(ctx, armorReq)
	if err != nil {
		slog.Error("Failed to create armor", "error", err)
		return nil, err
	}

	// Step 2: Create the item template with common attributes
	templateReq := &CreateItemTemplateRequest{
		UserId:        req.UserId,
		ItemName:      req.ItemName,
		RarityID:      req.RarityID,
		ItemType:      "armor",
		ItemID:        armor.ID,
		IconURL:       req.IconURL,
		RequiredLevel: req.RequiredLevel,
		BaseSellPrice: req.BaseSellPrice,
		BaseBuyPrice:  req.BaseBuyPrice,
	}

	template, err := s.CreateItemTemplate(ctx, templateReq)
	if err != nil {
		slog.Error("Failed to create item template for armor", "armor_id", armor.ID, "error", err)
		return nil, err
	}

	slog.Info("Complete armor created successfully",
		"armor_id", armor.ID,
		"template_id", template.ID,
		"item_name", template.ItemName,
	)

	// Step 3: Return the combined result
	return &ArmorWithTemplate{
		ID:              armor.ID,
		RarityID:        armor.RarityID,
		DefenseRating:   armor.DefenseRating,
		MagicResistance: armor.MagicResistance,
		ArmorSlot:       armor.ArmorSlot,
		Description:     armor.Description,
		CreatedAt:       armor.CreatedAt,
		UpdatedAt:       armor.UpdatedAt,
		// Template fields
		ItemTemplateID: template.ID,
		ItemName:       template.ItemName,
		IconURL:        template.IconURL,
		RequiredLevel:  template.RequiredLevel,
		BaseSellPrice:  template.BaseSellPrice,
		BaseBuyPrice:   template.BaseBuyPrice,
	}, nil
}

func (s *service) CreateCompleteConsumable(ctx context.Context, req *CreateCompleteConsumableRequest) (*ConsumableWithTemplate, error) {
	// Step 1: Create the consumable with specific attributes
	consumableReq := &CreateConsumableRequest{
		RarityID:      req.RarityID,
		HealingAmount: req.HealingAmount,
		ManaAmount:    req.ManaAmount,
		BuffDuration:  req.BuffDuration,
		MaxStackSize:  req.MaxStackSize,
		Description:   req.Description,
	}

	consumable, err := s.CreateConsumable(ctx, consumableReq)
	if err != nil {
		slog.Error("Failed to create consumable", "error", err)
		return nil, err
	}

	// Step 2: Create the item template with common attributes
	templateReq := &CreateItemTemplateRequest{
		UserId:        req.UserId,
		ItemName:      req.ItemName,
		RarityID:      req.RarityID,
		ItemType:      "consumable",
		ItemID:        consumable.ID,
		IconURL:       req.IconURL,
		RequiredLevel: req.RequiredLevel,
		BaseSellPrice: req.BaseSellPrice,
		BaseBuyPrice:  req.BaseBuyPrice,
	}

	template, err := s.CreateItemTemplate(ctx, templateReq)
	if err != nil {
		slog.Error("Failed to create item template for consumable", "consumable_id", consumable.ID, "error", err)
		return nil, err
	}

	slog.Info("Complete consumable created successfully",
		"consumable_id", consumable.ID,
		"template_id", template.ID,
		"item_name", template.ItemName,
	)

	// Step 3: Return the combined result
	return &ConsumableWithTemplate{
		ID:            consumable.ID,
		RarityID:      consumable.RarityID,
		HealingAmount: consumable.HealingAmount,
		ManaAmount:    consumable.ManaAmount,
		BuffDuration:  consumable.BuffDuration,
		MaxStackSize:  consumable.MaxStackSize,
		Description:   consumable.Description,
		CreatedAt:     consumable.CreatedAt,
		UpdatedAt:     consumable.UpdatedAt,
		// Template fields
		ItemTemplateID: template.ID,
		ItemName:       template.ItemName,
		IconURL:        template.IconURL,
		RequiredLevel:  template.RequiredLevel,
		BaseSellPrice:  template.BaseSellPrice,
		BaseBuyPrice:   template.BaseBuyPrice,
	}, nil
}

//
