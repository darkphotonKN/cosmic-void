package items

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	commonutils "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

type service struct {
	repo      Repository
	db        *sqlx.DB
	publishCh commonbroker.Publisher
}

func NewService(repo Repository, db *sqlx.DB, publishCh commonbroker.Publisher) *service {
	return &service{
		repo:      repo,
		db:        db,
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

	// Transaction aware create methods (for CreateComplete* flows)
	CreateWeaponTx(ctx context.Context, tx *sqlx.Tx, weapon *Weapon) error
	CreateArmorTx(ctx context.Context, tx *sqlx.Tx, armor *Armor) error
	CreateConsumableTx(ctx context.Context, tx *sqlx.Tx, consumable *Consumable) error
	CreateItemTemplateTx(ctx context.Context, tx *sqlx.Tx, template *ItemTemplate) error

	GetLoadout(ctx context.Context, req *GetLoadoutRequest) (*Loadout, error)
	GetItemInstanceByID(ctx context.Context, id uuid.UUID) (*ItemInstance, error)
	ListItemInstances(ctx context.Context, req *ListItemInstancesRequest) ([]*ItemInstance, error)
	UpsertLoadoutSlot(ctx context.Context, req *UpdateLoadoutRequest) error
	UpsertPlayerLoadoutTx(ctx context.Context, tx *sqlx.Tx, req *UpsertPlayerLoadoutRequest) error
	UpsertItemInstanceTx(ctx context.Context, tx *sqlx.Tx, instance *ItemInstance) error
	BatchUpsertItemInstances(ctx context.Context, tx *sqlx.Tx, instances []*ItemInstance) error
}

func (s *service) CreateItemInstance(createItemInstanceReq *ItemInstance) (*ItemInstance, error) {
	return nil, nil
}

func (s *service) CreatePlayerLoadout(createPlayerLoadoutReq *PlayerLoadout) error {
	return nil
}

func (s *service) ProcessItemsExtracted(ctx context.Context, req *pb.ItemsExtractedEvent) error {
	// transaction to wrap inventory upserts and player_loadout upserts
	for _, playerItems := range req.PlayerItems {
		// loop through each player
		slog.Debug("single player iterated from req.PlayerItems",
			"member_id", playerItems.MemberId,
			"equipment", playerItems.Equipment,
			"inventory", playerItems.Inventory,
		)

		// only holds one connection a time, released when committed or rolled back
		commonutils.ExecTx(ctx, s.db, func(tx *sqlx.Tx) error {

			// convert inventory and equipment into item instances
			invItemInstances, err := s.MapProtoItemToItemInstances(playerItems.Inventory)
			if err != nil {
				slog.Error("Unexpected error converting inventory from pb.Item to InventoryInstance",
					"err", err,
				)
			}

			memberId, err := uuid.Parse(playerItems.MemberId)
			if err != nil {
				slog.Error("error parsing member id when processing items extracted",
					"member_id", playerItems.MemberId,
					"err", err,
				)
				return err
			}

			equipItemInstances, upsertParams, err := s.MapProtoEquipmentToItemInstances(memberId, playerItems.Equipment)
			if err != nil {
				slog.Error("Error when attempting to map pb equipped items to item instances and create upsert player loadout params",
					"member_id", memberId,
					"err", err,
					"player_items_equipment", playerItems.Equipment,
				)
			}

			allItemIntances := append(equipItemInstances, invItemInstances...)

			// batch update items
			err = s.repo.BatchUpsertItemInstances(ctx, tx, allItemIntances)
			if err != nil {
				slog.Error("Error when attempting to batch upsert item instances",
					"item_instances", allItemIntances,
				)
				return err
			}

			// upsert player loadout with equipment ids
			err = s.repo.UpsertPlayerLoadoutTx(ctx, tx, upsertParams)
			if err != nil {
				slog.Error("Error when attempting to upsert equipment into player_loadouts",
					"member_id", memberId,
					"err", err,
					"upsert_params", upsertParams,
				)
				return err
			}
			return nil
		})
	}
	return nil
}

/**
* Converts the equipped items, equipment, extracted from items.extracted event into ItemInstance entities for
* updating the item instance table and formatted into PlayerLoadout for updating player_loadouts table.
**/
func (s *service) MapProtoEquipmentToItemInstances(memberID uuid.UUID, equipmentProto *pb.Equipment) ([]*ItemInstance, *UpsertPlayerLoadoutRequest, error) {
	// holds both existing ids and new ids
	playerLoadoutParam := &UpsertPlayerLoadoutRequest{
		MemberID: memberID,
	}

	itemInstances := make([]*ItemInstance, 0)

	// convert each item invidually to maintain mapping
	chestInstanceItem, err := s.ConvertSingleProtoItemtoItemInstance(equipmentProto.Chest)
	if err == nil {
		// update
		playerLoadoutParam.ChestInstanceID = &chestInstanceItem.ID
		itemInstances = append(itemInstances, chestInstanceItem)
	} else {
		if errors.Is(err, commonconstants.ErrUUIDCouldNotBeParsed) {
			return nil, nil, err
		}
		slog.Warn("Couldn't convert chest item to instanceItem.",
			"error", err,
		)
	}

	weaponInstanceItem, err := s.ConvertSingleProtoItemtoItemInstance(equipmentProto.Weapon)

	if err == nil {
		playerLoadoutParam.WeaponInstanceID = &weaponInstanceItem.ID
		itemInstances = append(itemInstances, weaponInstanceItem)
	} else {
		if errors.Is(err, commonconstants.ErrUUIDCouldNotBeParsed) {
			return nil, nil, err
		}
		slog.Warn("Couldn't convert weapon item to instanceItem.",
			"error", err,
		)
	}

	headInstanceItem, err := s.ConvertSingleProtoItemtoItemInstance(equipmentProto.Head)
	if err == nil {
		playerLoadoutParam.HeadInstanceID = &headInstanceItem.ID
		itemInstances = append(itemInstances, headInstanceItem)
	} else {
		if errors.Is(err, commonconstants.ErrUUIDCouldNotBeParsed) {
			return nil, nil, err
		}
		slog.Warn("Couldn't convert head item to instanceItem.",
			"error", err,
		)
	}

	glovesInstanceItem, err := s.ConvertSingleProtoItemtoItemInstance(equipmentProto.Gloves)
	if err == nil {
		playerLoadoutParam.GlovesInstanceID = &glovesInstanceItem.ID
		itemInstances = append(itemInstances, glovesInstanceItem)
	} else {
		if errors.Is(err, commonconstants.ErrUUIDCouldNotBeParsed) {
			return nil, nil, err
		}
		slog.Warn("Couldn't convert gloves item to instanceItem.",
			"error", err,
		)
	}

	legsInstanceItem, err := s.ConvertSingleProtoItemtoItemInstance(equipmentProto.Legs)
	if err == nil {
		playerLoadoutParam.LegsInstanceID = &legsInstanceItem.ID
		itemInstances = append(itemInstances, legsInstanceItem)
	} else {
		if errors.Is(err, commonconstants.ErrUUIDCouldNotBeParsed) {
			return nil, nil, err
		}
		slog.Warn("Couldn't convert legs item to instanceItem.",
			"error", err,
		)
	}

	ring1InstanceItem, err := s.ConvertSingleProtoItemtoItemInstance(equipmentProto.Ring_1)
	if err == nil {
		playerLoadoutParam.Ring1InstanceID = &ring1InstanceItem.ID
		itemInstances = append(itemInstances, ring1InstanceItem)
	} else {
		if errors.Is(err, commonconstants.ErrUUIDCouldNotBeParsed) {
			return nil, nil, err
		}
		slog.Warn("Couldn't convert ring_1 item to instanceItem.",
			"error", err,
		)
	}

	ring2InstanceItem, err := s.ConvertSingleProtoItemtoItemInstance(equipmentProto.Ring_2)
	if err == nil {
		playerLoadoutParam.Ring2InstanceID = &ring2InstanceItem.ID
		itemInstances = append(itemInstances, ring2InstanceItem)
	} else {
		if errors.Is(err, commonconstants.ErrUUIDCouldNotBeParsed) {
			return nil, nil, err
		}
		slog.Warn("Couldn't convert ring_2 item to instanceItem.",
			"error", err,
		)
	}

	consumable1InstanceItem, err := s.ConvertSingleProtoItemtoItemInstance(equipmentProto.Consumable_1)
	if err == nil {
		playerLoadoutParam.Consumable1ID = &consumable1InstanceItem.ID
		itemInstances = append(itemInstances, consumable1InstanceItem)
	} else {
		if errors.Is(err, commonconstants.ErrUUIDCouldNotBeParsed) {
			return nil, nil, err
		}
		slog.Warn("Couldn't convert consumable_1 item to instanceItem.",
			"error", err,
		)
	}

	consumable2InstanceItem, err := s.ConvertSingleProtoItemtoItemInstance(equipmentProto.Consumable_2)
	if err == nil {
		playerLoadoutParam.Consumable2ID = &consumable2InstanceItem.ID
		itemInstances = append(itemInstances, consumable2InstanceItem)
	} else {
		if errors.Is(err, commonconstants.ErrUUIDCouldNotBeParsed) {
			return nil, nil, err
		}
		slog.Warn("Couldn't convert consumable_2 item to instanceItem.",
			"error", err,
		)
	}

	consumable3InstanceItem, err := s.ConvertSingleProtoItemtoItemInstance(equipmentProto.Consumable_3)
	if err == nil {
		playerLoadoutParam.Consumable3ID = &consumable3InstanceItem.ID
		itemInstances = append(itemInstances, consumable3InstanceItem)
	} else {
		if errors.Is(err, commonconstants.ErrUUIDCouldNotBeParsed) {
			return nil, nil, err
		}
		slog.Warn("Couldn't convert consumable_3 item to instanceItem.",
			"error", err,
		)
	}

	slog.Debug("Completed building playerloadoutParam and itemInstances",
		"item_instances", itemInstances,
		"player_loadout_param", playerLoadoutParam,
	)

	return itemInstances, playerLoadoutParam, nil
}

func (s *service) ConvertSingleProtoItemtoItemInstance(protoItem *pb.Item) (*ItemInstance, error) {
	if protoItem == nil {
		slog.Debug("Nothing to convert, protoItem was nil")
		return nil, fmt.Errorf("nil pb.Item cant be converted into ItemInstance.")
	}

	var itemId uuid.UUID
	var err error
	if protoItem.InstanceId == "" {
		itemId = uuid.New()
	} else {
		itemId, err = uuid.Parse(protoItem.InstanceId)
		if err != nil {
			slog.Error("error parsing protoItem's instanceID",
				"instance_id", protoItem.InstanceId,
			)
			return nil, commonconstants.ErrUUIDCouldNotBeParsed
		}
	}

	attackPower := int(protoItem.AttackPower)
	criticalRate := protoItem.CriticalRate
	weaponType := protoItem.WeaponType
	defenseRating := int(protoItem.DefenseRating)
	magicResistance := int(protoItem.MagicResistance)
	armorSlot := protoItem.ArmorSlot
	healingAmount := int(protoItem.HealingAmount)
	manaAmount := int(protoItem.ManaAmount)
	buffDuration := int(protoItem.BuffDuration)
	buyPrice := int(protoItem.BuyPrice)
	sellPrice := int(protoItem.SellPrice)
	description := protoItem.Description

	item := &ItemInstance{
		ID:              itemId,
		ItemType:        protoItem.ItemType,
		Name:            protoItem.Name,
		AttackPower:     &attackPower,
		CriticalRate:    &criticalRate,
		WeaponType:      &weaponType,
		DefenseRating:   &defenseRating,
		MagicResistance: &magicResistance,
		ArmorSlot:       &armorSlot,
		HealingAmount:   &healingAmount,
		ManaAmount:      &manaAmount,
		BuffDuration:    &buffDuration,
		BuyPrice:        &buyPrice,
		SellPrice:       &sellPrice,
		Description:     &description,
	}

	return item, nil
}

/**
* Converts the non equipped slice of items extracted from items.extracted event into ItemInstance entities for
* updating the item instance table.
**/
func (s *service) MapProtoItemToItemInstances(itemsProto []*pb.Item) ([]*ItemInstance, error) {
	// no items from user
	if len(itemsProto) == 0 {
		slog.Error("No itemProtos to map to ItemInstances.")
		return []*ItemInstance{}, nil
	}

	// items exist, update

	itemInstances := make([]*ItemInstance, 0)

	for _, protoItem := range itemsProto {
		item, err := s.ConvertSingleProtoItemtoItemInstance(protoItem)
		if err != nil {
			slog.Warn("item couldnt be mapped into ItemInstance",
				"proto_item_instance_id", protoItem.InstanceId,
			)
			continue
		}

		itemInstances = append(itemInstances, item)
	}

	return itemInstances, nil
}

// sem := make(chan struct{}, 3) // max 3 concurrent
//
// for _, player := range players {
//     sem <- struct{}{} // acquire slot
//     go func(p Player) {
//         defer func() { <-sem }() // release slot
//         tx, _ := db.BeginTx(ctx, nil)
//         // batch update
//         tx.Commit()
//     }(player)
// }

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
	// Validate rarity exists
	if _, err := s.repo.GetItemRarityByID(ctx, req.RarityID); err != nil {
		return nil, fmt.Errorf("invalid rarity_id: %w", err)
	}

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
	if _, err := s.repo.GetItemRarityByID(ctx, req.RarityID); err != nil {
		return nil, fmt.Errorf("invalid rarity_id: %w", err)
	}

	var weapon Weapon
	var template ItemTemplate

	requiredLevel, baseSellPrice, baseBuyPrice := resolveTemplateDefaults(req.RequiredLevel, req.BaseSellPrice, req.BaseBuyPrice)

	err := commonutils.ExecTx(ctx, s.db, func(tx *sqlx.Tx) error {
		w := &Weapon{
			RarityID:     req.RarityID,
			AttackPower:  req.AttackPower,
			CriticalRate: req.CriticalRate,
			WeaponType:   req.WeaponType,
			Description:  req.Description,
		}
		if err := s.repo.CreateWeaponTx(ctx, tx, w); err != nil {
			return err
		}
		weapon = *w

		t := &ItemTemplate{
			ItemName:      req.ItemName,
			RarityID:      req.RarityID,
			ItemType:      "weapon",
			ItemID:        weapon.ID,
			IconURL:       req.IconURL,
			RequiredLevel: requiredLevel,
			BaseSellPrice: baseSellPrice,
			BaseBuyPrice:  baseBuyPrice,
		}
		if err := s.repo.CreateItemTemplateTx(ctx, tx, t); err != nil {
			return err
		}
		template = *t

		return nil
	})

	if err != nil {
		slog.Error("Failed to create complete weapon", "error", err)
		return nil, err
	}

	s.publishItemCreatedEvent(ctx, req.UserId, req.ItemName, "weapon")

	return &WeaponWithTemplate{
		ID:             weapon.ID,
		RarityID:       weapon.RarityID,
		AttackPower:    weapon.AttackPower,
		CriticalRate:   weapon.CriticalRate,
		WeaponType:     weapon.WeaponType,
		Description:    weapon.Description,
		CreatedAt:      weapon.CreatedAt,
		UpdatedAt:      weapon.UpdatedAt,
		ItemTemplateID: template.ID,
		ItemName:       template.ItemName,
		IconURL:        template.IconURL,
		RequiredLevel:  template.RequiredLevel,
		BaseSellPrice:  template.BaseSellPrice,
		BaseBuyPrice:   template.BaseBuyPrice,
	}, nil
}

func (s *service) CreateCompleteArmor(ctx context.Context, req *CreateCompleteArmorRequest) (*ArmorWithTemplate, error) {
	if _, err := s.repo.GetItemRarityByID(ctx, req.RarityID); err != nil {
		return nil, fmt.Errorf("invalid rarity_id: %w", err)
	}

	var armor Armor
	var template ItemTemplate

	requiredLevel, baseSellPrice, baseBuyPrice := resolveTemplateDefaults(req.RequiredLevel, req.BaseSellPrice, req.BaseBuyPrice)

	err := commonutils.ExecTx(ctx, s.db, func(tx *sqlx.Tx) error {
		a := &Armor{
			RarityID:        req.RarityID,
			DefenseRating:   req.DefenseRating,
			MagicResistance: req.MagicResistance,
			ArmorSlot:       req.ArmorSlot,
			Description:     req.Description,
		}
		if err := s.repo.CreateArmorTx(ctx, tx, a); err != nil {
			return err
		}
		armor = *a

		t := &ItemTemplate{
			ItemName:      req.ItemName,
			RarityID:      req.RarityID,
			ItemType:      "armor",
			ItemID:        armor.ID,
			IconURL:       req.IconURL,
			RequiredLevel: requiredLevel,
			BaseSellPrice: baseSellPrice,
			BaseBuyPrice:  baseBuyPrice,
		}
		if err := s.repo.CreateItemTemplateTx(ctx, tx, t); err != nil {
			return err
		}
		template = *t

		return nil
	})

	if err != nil {
		slog.Error("Failed to create complete armor", "error", err)
		return nil, err
	}

	s.publishItemCreatedEvent(ctx, req.UserId, req.ItemName, "armor")

	return &ArmorWithTemplate{
		ID:              armor.ID,
		RarityID:        armor.RarityID,
		DefenseRating:   armor.DefenseRating,
		MagicResistance: armor.MagicResistance,
		ArmorSlot:       armor.ArmorSlot,
		Description:     armor.Description,
		CreatedAt:       armor.CreatedAt,
		UpdatedAt:       armor.UpdatedAt,
		ItemTemplateID:  template.ID,
		ItemName:        template.ItemName,
		IconURL:         template.IconURL,
		RequiredLevel:   template.RequiredLevel,
		BaseSellPrice:   template.BaseSellPrice,
		BaseBuyPrice:    template.BaseBuyPrice,
	}, nil
}

func (s *service) CreateCompleteConsumable(ctx context.Context, req *CreateCompleteConsumableRequest) (*ConsumableWithTemplate, error) {
	if _, err := s.repo.GetItemRarityByID(ctx, req.RarityID); err != nil {
		return nil, fmt.Errorf("invalid rarity_id: %w", err)
	}

	var consumable Consumable
	var template ItemTemplate

	requiredLevel, baseSellPrice, baseBuyPrice := resolveTemplateDefaults(req.RequiredLevel, req.BaseSellPrice, req.BaseBuyPrice)

	err := commonutils.ExecTx(ctx, s.db, func(tx *sqlx.Tx) error {
		c := &Consumable{
			RarityID:      req.RarityID,
			HealingAmount: req.HealingAmount,
			ManaAmount:    req.ManaAmount,
			BuffDuration:  req.BuffDuration,
			MaxStackSize:  req.MaxStackSize,
			Description:   req.Description,
		}
		if err := s.repo.CreateConsumableTx(ctx, tx, c); err != nil {
			return err
		}
		consumable = *c

		t := &ItemTemplate{
			ItemName:      req.ItemName,
			RarityID:      req.RarityID,
			ItemType:      "consumable",
			ItemID:        consumable.ID,
			IconURL:       req.IconURL,
			RequiredLevel: requiredLevel,
			BaseSellPrice: baseSellPrice,
			BaseBuyPrice:  baseBuyPrice,
		}
		if err := s.repo.CreateItemTemplateTx(ctx, tx, t); err != nil {
			return err
		}
		template = *t

		return nil
	})

	if err != nil {
		slog.Error("Failed to create complete consumable", "error", err)
		return nil, err
	}

	s.publishItemCreatedEvent(ctx, req.UserId, req.ItemName, "consumable")

	return &ConsumableWithTemplate{
		ID:             consumable.ID,
		RarityID:       consumable.RarityID,
		HealingAmount:  consumable.HealingAmount,
		ManaAmount:     consumable.ManaAmount,
		BuffDuration:   consumable.BuffDuration,
		MaxStackSize:   consumable.MaxStackSize,
		Description:    consumable.Description,
		CreatedAt:      consumable.CreatedAt,
		UpdatedAt:      consumable.UpdatedAt,
		ItemTemplateID: template.ID,
		ItemName:       template.ItemName,
		IconURL:        template.IconURL,
		RequiredLevel:  template.RequiredLevel,
		BaseSellPrice:  template.BaseSellPrice,
		BaseBuyPrice:   template.BaseBuyPrice,
	}, nil
}

// resolveTemplateDefaults applies defaults for optional template fields
func resolveTemplateDefaults(reqLevel, sellPrice, buyPrice *int) (int, int, int) {
	requiredLevel := 1
	if reqLevel != nil {
		requiredLevel = *reqLevel
	}
	baseSellPrice := 0
	if sellPrice != nil {
		baseSellPrice = *sellPrice
	}
	baseBuyPrice := 0
	if buyPrice != nil {
		baseBuyPrice = *buyPrice
	}
	return requiredLevel, baseSellPrice, baseBuyPrice
}

// publishItemCreatedEvent sends an item creation event to RabbitMQ (fire-and-forget, outside tx)
func (s *service) publishItemCreatedEvent(ctx context.Context, userId, itemName, itemType string) {
	protoData, err := proto.Marshal(&pb.ItemCreatedEvent{
		UserId:   userId,
		Name:     itemName,
		ItemType: itemType,
	})
	if err != nil {
		slog.Error("Failed to marshal item created event", "error", err)
		return
	}

	if err := s.publishCh.PublishWithContext(ctx, commonconstants.ItemEventsExchange, commonconstants.ItemCreated, commonbroker.Message{
		ContentType:  "application/protobuf",
		Body:         protoData,
		DeliveryMode: commonbroker.Persistent,
	}); err != nil {
		slog.Error("Failed to publish item created event", "error", err)
	}
}

func (h *service) GetLoadout(ctx context.Context, req *GetLoadoutRequest) (*Loadout, error) {
	return h.repo.GetLoadout(ctx, req)
}

func (h *service) GetLoadoutWithItems(ctx context.Context, req *GetLoadoutRequest) (*LoadoutWithItems, error) {
	loadout, err := h.repo.GetLoadout(ctx, req)
	if err != nil {
		return nil, err
	}

	result := &LoadoutWithItems{}

	getItem := func(id *uuid.UUID) *ItemInstance {
		if id == nil {
			return nil
		}
		item, err := h.repo.GetItemInstanceByID(ctx, *id)
		if err != nil {
			return nil
		}
		return item
	}

	result.Weapon = getItem(loadout.WeaponId)
	result.Head = getItem(loadout.HeadId)
	result.Chest = getItem(loadout.ChestId)
	result.Gloves = getItem(loadout.GlovesId)
	result.Legs = getItem(loadout.LegsId)
	result.Ring1 = getItem(loadout.Ring1Id)
	result.Ring2 = getItem(loadout.Ring2Id)
	result.Consumable1 = getItem(loadout.Consumable1Id)
	result.Consumable2 = getItem(loadout.Consumable2Id)
	result.Consumable3 = getItem(loadout.Consumable3Id)

	return result, nil
}

func (h *service) ListItemInstances(ctx context.Context, req *ListItemInstancesRequest) ([]*ItemInstance, error) {
	return h.repo.ListItemInstances(ctx, req)
}

func (h *service) UpdateLoadout(ctx context.Context, req *UpdateLoadoutRequest) error {
	return h.repo.UpsertLoadoutSlot(ctx, req)
}
