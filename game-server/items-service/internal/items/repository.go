package items

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

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
}

type repository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{
		DB: db,
	}
}

// ==========================================
// ItemType Repository Methods
// ==========================================

func (r *repository) CreateItemType(ctx context.Context, itemType *ItemType) error {
	itemType.ID = uuid.New()

	query := `
		INSERT INTO item_types (
			id, type_code, name, description, is_active, sort_order
		) VALUES (
			:id, :type_code, :name, :description, :is_active, :sort_order
		)`

	_, err := r.DB.NamedExecContext(ctx, query, itemType)
	if err != nil {
		return fmt.Errorf("failed to create item type: %w", err)
	}

	return nil
}

func (r *repository) GetItemTypeByID(ctx context.Context, id uuid.UUID) (*ItemType, error) {
	var itemType ItemType

	query := `SELECT * FROM item_types WHERE id = $1`

	err := r.DB.GetContext(ctx, &itemType, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get item type: %w", err)
	}

	return &itemType, nil
}

func (r *repository) GetItemTypeByCode(ctx context.Context, code string) (*ItemType, error) {
	var itemType ItemType

	query := `SELECT * FROM item_types WHERE type_code = $1`

	err := r.DB.GetContext(ctx, &itemType, query, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get item type by code: %w", err)
	}

	return &itemType, nil
}

func (r *repository) ListItemTypes(ctx context.Context) ([]*ItemType, error) {
	var itemTypes []*ItemType

	query := `SELECT * FROM item_types WHERE is_active = true ORDER BY sort_order ASC`

	err := r.DB.SelectContext(ctx, &itemTypes, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list item types: %w", err)
	}

	return itemTypes, nil
}

// ==========================================
// ItemRarity Repository Methods
// ==========================================

func (r *repository) CreateItemRarity(ctx context.Context, rarity *ItemRarity) error {
	rarity.ID = uuid.New()

	query := `
		INSERT INTO item_rarities (
			id, rarity_code, rarity_name, color_hex,
			drop_rate_multiplier, sort_order
		) VALUES (
			:id, :rarity_code, :rarity_name, :color_hex,
			:drop_rate_multiplier, :sort_order
		)`

	_, err := r.DB.NamedExecContext(ctx, query, rarity)
	if err != nil {
		return fmt.Errorf("failed to create item rarity: %w", err)
	}

	return nil
}

func (r *repository) GetItemRarityByID(ctx context.Context, id uuid.UUID) (*ItemRarity, error) {
	var rarity ItemRarity

	query := `SELECT * FROM item_rarities WHERE id = $1`

	err := r.DB.GetContext(ctx, &rarity, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get item rarity: %w", err)
	}

	return &rarity, nil
}

func (r *repository) GetItemRarityByCode(ctx context.Context, code string) (*ItemRarity, error) {
	var rarity ItemRarity

	query := `SELECT * FROM item_rarities WHERE rarity_code = $1`

	err := r.DB.GetContext(ctx, &rarity, query, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get item rarity by code: %w", err)
	}

	return &rarity, nil
}

func (r *repository) ListItemRarities(ctx context.Context) ([]*ItemRarity, error) {
	var rarities []*ItemRarity

	query := `SELECT * FROM item_rarities ORDER BY sort_order ASC`

	err := r.DB.SelectContext(ctx, &rarities, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list item rarities: %w", err)
	}

	return rarities, nil
}

// ==========================================
// Weapon Repository Methods
// ==========================================

func (r *repository) CreateWeapon(ctx context.Context, weapon *Weapon) error {
	weapon.ID = uuid.New()

	query := `
		INSERT INTO weapons (
			id, type_id, rarity_id, attack_power, durability,
			critical_rate, weapon_type, description
		) VALUES (
			:id, :type_id, :rarity_id, :attack_power, :durability,
			:critical_rate, :weapon_type, :description
		)`

	_, err := r.DB.NamedExecContext(ctx, query, weapon)
	if err != nil {
		return fmt.Errorf("failed to create weapon: %w", err)
	}

	return nil
}

func (r *repository) GetWeaponByID(ctx context.Context, id uuid.UUID) (*Weapon, error) {
	var weapon Weapon

	query := `SELECT * FROM weapons WHERE id = $1`

	err := r.DB.GetContext(ctx, &weapon, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get weapon: %w", err)
	}

	return &weapon, nil
}

func (r *repository) ListWeapons(ctx context.Context) ([]*Weapon, error) {
	var weapons []*Weapon

	query := `SELECT * FROM weapons ORDER BY created_at DESC`

	err := r.DB.SelectContext(ctx, &weapons, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list weapons: %w", err)
	}

	return weapons, nil
}

// ==========================================
// Armor Repository Methods
// ==========================================

func (r *repository) CreateArmor(ctx context.Context, armor *Armor) error {
	armor.ID = uuid.New()

	query := `
		INSERT INTO armors (
			id, type_id, rarity_id, defense_rating, durability,
			magic_resistance, armor_slot, description
		) VALUES (
			:id, :type_id, :rarity_id, :defense_rating, :durability,
			:magic_resistance, :armor_slot, :description
		)`

	_, err := r.DB.NamedExecContext(ctx, query, armor)
	if err != nil {
		return fmt.Errorf("failed to create armor: %w", err)
	}

	return nil
}

func (r *repository) GetArmorByID(ctx context.Context, id uuid.UUID) (*Armor, error) {
	var armor Armor

	query := `SELECT * FROM armors WHERE id = $1`

	err := r.DB.GetContext(ctx, &armor, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get armor: %w", err)
	}

	return &armor, nil
}

func (r *repository) ListArmors(ctx context.Context) ([]*Armor, error) {
	var armors []*Armor

	query := `SELECT * FROM armors ORDER BY created_at DESC`

	err := r.DB.SelectContext(ctx, &armors, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list armors: %w", err)
	}

	return armors, nil
}

// ==========================================
// Consumable Repository Methods
// ==========================================

func (r *repository) CreateConsumable(ctx context.Context, consumable *Consumable) error {
	consumable.ID = uuid.New()

	query := `
		INSERT INTO consumables (
			id, type_id, rarity_id, healing_amount, mana_amount,
			buff_duration, max_stack_size, description
		) VALUES (
			:id, :type_id, :rarity_id, :healing_amount, :mana_amount,
			:buff_duration, :max_stack_size, :description
		)`

	_, err := r.DB.NamedExecContext(ctx, query, consumable)
	if err != nil {
		return fmt.Errorf("failed to create consumable: %w", err)
	}

	return nil
}

func (r *repository) GetConsumableByID(ctx context.Context, id uuid.UUID) (*Consumable, error) {
	var consumable Consumable

	query := `SELECT * FROM consumables WHERE id = $1`

	err := r.DB.GetContext(ctx, &consumable, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumable: %w", err)
	}

	return &consumable, nil
}

func (r *repository) ListConsumables(ctx context.Context) ([]*Consumable, error) {
	var consumables []*Consumable

	query := `SELECT * FROM consumables ORDER BY created_at DESC`

	err := r.DB.SelectContext(ctx, &consumables, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list consumables: %w", err)
	}

	return consumables, nil
}

// ==========================================
// ItemTemplate Repository Methods
// ==========================================

func (r *repository) CreateItemTemplate(ctx context.Context, template *ItemTemplate) error {
	template.ID = uuid.New()

	query := `
		INSERT INTO item_templates (
			id, item_name, item_code, type_id, rarity_id,
			item_type, item_id, icon_url, is_tradeable, is_droppable,
			required_level, base_sell_price, base_buy_price
		) VALUES (
			:id, :item_name, :item_code, :type_id, :rarity_id,
			:item_type, :item_id, :icon_url, :is_tradeable, :is_droppable,
			:required_level, :base_sell_price, :base_buy_price
		)`

	_, err := r.DB.NamedExecContext(ctx, query, template)
	if err != nil {
		return fmt.Errorf("failed to create item template: %w", err)
	}

	return nil
}

func (r *repository) GetItemTemplateByID(ctx context.Context, id uuid.UUID) (*ItemTemplate, error) {
	var template ItemTemplate

	query := `SELECT * FROM item_templates WHERE id = $1`

	err := r.DB.GetContext(ctx, &template, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get item template: %w", err)
	}

	return &template, nil
}

func (r *repository) GetItemTemplateByCode(ctx context.Context, code string) (*ItemTemplate, error) {
	var template ItemTemplate

	query := `SELECT * FROM item_templates WHERE item_code = $1`

	err := r.DB.GetContext(ctx, &template, query, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get item template by code: %w", err)
	}

	return &template, nil
}

func (r *repository) ListItemTemplates(ctx context.Context) ([]*ItemTemplate, error) {
	var templates []*ItemTemplate

	query := `SELECT * FROM item_templates ORDER BY created_at DESC`

	err := r.DB.SelectContext(ctx, &templates, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list item templates: %w", err)
	}

	return templates, nil
}

func (r *repository) GetWeaponWithTemplateByID(ctx context.Context, id uuid.UUID) (*WeaponWithTemplate, error) {
	var template WeaponWithTemplate

	query := `
		SELECT
			w.id,
			w.type_id,
			w.rarity_id,
			w.attack_power,
			w.durability,
			w.critical_rate,
			w.weapon_type,
			w.description,
			w.created_at,
			w.updated_at,
			t.id AS item_template_id,
			t.item_name,
			t.item_code,
			t.icon_url,
			t.is_tradeable,
			t.is_droppable,
			t.required_level,
			t.base_sell_price,
			t.base_buy_price
		FROM weapons AS w
		INNER JOIN item_templates AS t ON t.item_id = w.id AND t.item_type = 'weapon'
		WHERE w.id = $1
	`
	err := r.DB.GetContext(ctx, &template, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get weapon with template: %w", err)
	}

	return &template, nil
}

// ListArmorsWithTemplate retrieves all armors with their item template information
func (r *repository) ListArmorsWithTemplate(ctx context.Context) ([]*ArmorWithTemplate, error) {
	var armorDetails []*ArmorWithTemplate

	query := `
		SELECT
			a.id,
			a.type_id,
			a.rarity_id,
			a.defense_rating,
			a.durability,
			a.magic_resistance,
			a.armor_slot,
			a.description,
			a.created_at,
			a.updated_at,
			t.id AS item_template_id,
			t.item_name,
			t.item_code,
			t.icon_url,
			t.is_tradeable,
			t.is_droppable,
			t.required_level,
			t.base_sell_price,
			t.base_buy_price
		FROM armors AS a
		INNER JOIN item_templates AS t ON t.item_id = a.id AND t.item_type = 'armor'
		ORDER BY a.created_at DESC
	`

	err := r.DB.SelectContext(ctx, &armorDetails, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list armors with template: %w", err)
	}

	return armorDetails, nil
}

// ListConsumablesWithTemplate retrieves all consumables with their item template information
func (r *repository) ListConsumablesWithTemplate(ctx context.Context) ([]*ConsumableWithTemplate, error) {
	var consumableDetails []*ConsumableWithTemplate

	query := `
		SELECT
			c.id,
			c.type_id,
			c.rarity_id,
			c.healing_amount,
			c.mana_amount,
			c.buff_duration,
			c.max_stack_size,
			c.description,
			c.created_at,
			c.updated_at,
			t.id AS item_template_id,
			t.item_name,
			t.item_code,
			t.icon_url,
			t.is_tradeable,
			t.is_droppable,
			t.required_level,
			t.base_sell_price,
			t.base_buy_price
		FROM consumables AS c
		INNER JOIN item_templates AS t ON t.item_id = c.id AND t.item_type = 'consumable'
		ORDER BY c.created_at DESC
	`

	err := r.DB.SelectContext(ctx, &consumableDetails, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list consumables with template: %w", err)
	}

	return consumableDetails, nil
}

// ListWeaponsWithTemplate retrieves all weapons with their item template information
func (r *repository) ListWeaponsWithTemplate(ctx context.Context) ([]*WeaponWithTemplate, error) {
	var weaponDetails []*WeaponWithTemplate

	query := `
		SELECT
			w.id,
			w.type_id,
			w.rarity_id,
			w.attack_power,
			w.durability,
			w.critical_rate,
			w.weapon_type,
			w.description,
			w.created_at,
			w.updated_at,
			t.id AS item_template_id,
			t.item_name,
			t.item_code,
			t.icon_url,
			t.is_tradeable,
			t.is_droppable,
			t.required_level,
			t.base_sell_price,
			t.base_buy_price
		FROM weapons AS w
		INNER JOIN item_templates AS t ON t.item_id = w.id AND t.item_type = 'weapon'
		ORDER BY w.created_at DESC
	`

	err := r.DB.SelectContext(ctx, &weaponDetails, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list weapons with template: %w", err)
	}

	return weaponDetails, nil
}
