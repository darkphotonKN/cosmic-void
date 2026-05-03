package items

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type repository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{
		DB: db,
	}
}

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
			id, rarity_id, attack_power,
			critical_rate, weapon_type, description
		) VALUES (
			:id, :rarity_id, :attack_power,
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
			id, rarity_id, defense_rating,
			magic_resistance, armor_slot, description
		) VALUES (
			:id, :rarity_id, :defense_rating,
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
			id, rarity_id, healing_amount, mana_amount,
			buff_duration, max_stack_size, description
		) VALUES (
			:id, :rarity_id, :healing_amount, :mana_amount,
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
			id, item_name, rarity_id, item_type, item_id,
			icon_url, required_level, base_sell_price, base_buy_price
		) VALUES (
			:id, :item_name, :rarity_id, :item_type, :item_id,
			:icon_url, :required_level, :base_sell_price, :base_buy_price
		)`

	_, err := r.DB.NamedExecContext(ctx, query, template)
	if err != nil {
		return fmt.Errorf("failed to create item template: %w", err)
	}

	return nil
}

func (r *repository) GetItemTemplateByID(ctx context.Context, id uuid.UUID) (*ItemTemplate, error) {
	ctx, span := itemRepositoryTracer.Start(ctx, "Repo.GetItemTemplateByID", trace.WithAttributes(
		attribute.String("item.id", id.String()),
	))

	defer span.End()
	var template ItemTemplate

	query := `SELECT * FROM item_templates WHERE id = $1`

	err := r.DB.GetContext(ctx, &template, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get item template: %w", err)
	}
	slog.Info("Debug GetItemTemplateByID")
	return &template, nil
}

func (r *repository) ListItemTemplates(ctx context.Context) ([]*ItemTemplate, error) {
	var templates []*ItemTemplate

	query := `SELECT id, item_name, rarity_id, item_type, item_id, icon_url,
	           required_level, base_sell_price, base_buy_price,
	           created_at, created_by, updated_at, updated_by
	           FROM item_templates ORDER BY created_at DESC`

	err := r.DB.SelectContext(ctx, &templates, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list item templates: %w", err)
	}

	return templates, nil
}

func (r *repository) GetItemTemplateByCode(ctx context.Context, code string) (*ItemTemplate, error) {
	// This method is deprecated as item_code no longer exists
	// Return nil for now, to be removed in future refactor
	return nil, fmt.Errorf("GetItemTemplateByCode is deprecated: item_code column no longer exists")
}

func (r *repository) ListItemTemplateAggregates(ctx context.Context) ([]*ItemTemplateAggregate, error) {

	query := `
SELECT
    it.id,
    it.item_name,
    it.item_type,
    it.icon_url,
    it.required_level,
    it.base_sell_price,
    it.base_buy_price,
    r.rarity_name AS rarity,
    w.attack_power,
    w.critical_rate,
    w.weapon_type,
    a.defense_rating,
    a.magic_resistance,
    a.armor_slot,
    c.healing_amount,
    c.mana_amount,
    c.buff_duration,
    c.max_stack_size,
    COALESCE(w.description, a.description, c.description) AS description,
    it.created_at,
    it.created_by,
    it.updated_at,
    it.updated_by
FROM item_templates AS it
LEFT JOIN weapons AS w
    ON w.id = it.item_id AND it.item_type = 'weapon'
LEFT JOIN armors AS a
    ON a.id = it.item_id AND it.item_type = 'armor'
LEFT JOIN consumables AS c
    ON c.id = it.item_id AND it.item_type = 'consumable'
LEFT JOIN item_rarities AS r
    ON r.id = it.rarity_id
	`

	var items []*ItemTemplateAggregate

	err := r.DB.SelectContext(ctx, &items, query)

	if err != nil {
		return nil, commonhelpers.AnalyzeDBErr(err)
	}

	return items, nil
}

func (r *repository) GetWeaponWithTemplateByID(ctx context.Context, id uuid.UUID) (*WeaponWithTemplate, error) {
	var weapon WeaponWithTemplate

	query := `
		SELECT w.id, w.rarity_id, w.attack_power,
		       w.critical_rate, w.weapon_type, w.description, w.created_at, w.updated_at,
		       t.id AS item_template_id, t.item_name, t.icon_url,
		       t.required_level, t.base_sell_price, t.base_buy_price
		FROM weapons w
		JOIN item_templates t ON t.item_id = w.id AND t.item_type = 'weapon'
		WHERE w.id = $1`

	err := r.DB.GetContext(ctx, &weapon, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get weapon with template: %w", err)
	}

	return &weapon, nil
}

// ListArmorsWithTemplate retrieves all armors with their item template information
func (r *repository) ListArmorsWithTemplate(ctx context.Context) ([]*ArmorWithTemplate, error) {
	ctx, span := itemRepositoryTracer.Start(ctx, "Repo.ListArmorsWithTemplate")
	defer span.End()

	var armors []*ArmorWithTemplate

	query := `
		SELECT a.id, a.rarity_id, a.defense_rating,
		       a.magic_resistance, a.armor_slot, a.description, a.created_at, a.updated_at,
		       t.id AS item_template_id, t.item_name, t.icon_url,
		       t.required_level, t.base_sell_price, t.base_buy_price
		FROM armors a
		JOIN item_templates t ON t.item_id = a.id AND t.item_type = 'armor'
		ORDER BY a.created_at DESC`

	err := r.DB.SelectContext(ctx, &armors, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list armors with template: %w", err)
	}

	return armors, nil
}

// ListConsumablesWithTemplate retrieves all consumables with their item template information
func (r *repository) ListConsumablesWithTemplate(ctx context.Context) ([]*ConsumableWithTemplate, error) {
	ctx, span := itemRepositoryTracer.Start(ctx, "Repo.ListConsumablesWithTemplate")
	defer span.End()

	var consumables []*ConsumableWithTemplate

	query := `
		SELECT c.id, c.rarity_id, c.healing_amount, c.mana_amount,
		       c.buff_duration, c.max_stack_size, c.description, c.created_at, c.updated_at,
		       t.id AS item_template_id, t.item_name, t.icon_url,
		       t.required_level, t.base_sell_price, t.base_buy_price
		FROM consumables c
		JOIN item_templates t ON t.item_id = c.id AND t.item_type = 'consumable'
		ORDER BY c.created_at DESC`

	err := r.DB.SelectContext(ctx, &consumables, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list consumables with template: %w", err)
	}

	return consumables, nil
}

// ListWeaponsWithTemplate retrieves all weapons with their item template information
func (r *repository) ListWeaponsWithTemplate(ctx context.Context) ([]*WeaponWithTemplate, error) {
	ctx, span := itemRepositoryTracer.Start(ctx, "Repo.ListWeaponsWithTemplate")
	defer span.End()

	var weapons []*WeaponWithTemplate

	query := `
		SELECT w.id, w.rarity_id, w.attack_power,
		       w.critical_rate, w.weapon_type, w.description, w.created_at, w.updated_at,
		       t.id AS item_template_id, t.item_name, t.icon_url,
		       t.required_level, t.base_sell_price, t.base_buy_price
		FROM weapons w
		JOIN item_templates t ON t.item_id = w.id AND t.item_type = 'weapon'
		ORDER BY w.created_at DESC`

	err := r.DB.SelectContext(ctx, &weapons, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list weapons with template: %w", err)
	}

	return weapons, nil
}

// ==========================================
// Transaction-aware Create Methods
// ==========================================

func (r *repository) CreateWeaponTx(ctx context.Context, tx *sqlx.Tx, weapon *Weapon) error {
	weapon.ID = uuid.New()

	query := `
		INSERT INTO weapons (
			id, rarity_id, attack_power,
			critical_rate, weapon_type, description
		) VALUES (
			:id, :rarity_id, :attack_power,
			:critical_rate, :weapon_type, :description
		)`

	_, err := tx.NamedExecContext(ctx, query, weapon)
	if err != nil {
		return fmt.Errorf("failed to create weapon (tx): %w", err)
	}

	return nil
}

func (r *repository) CreateArmorTx(ctx context.Context, tx *sqlx.Tx, armor *Armor) error {
	armor.ID = uuid.New()

	query := `
		INSERT INTO armors (
			id, rarity_id, defense_rating,
			magic_resistance, armor_slot, description
		) VALUES (
			:id, :rarity_id, :defense_rating,
			:magic_resistance, :armor_slot, :description
		)`

	_, err := tx.NamedExecContext(ctx, query, armor)
	if err != nil {
		return fmt.Errorf("failed to create armor (tx): %w", err)
	}

	return nil
}

func (r *repository) CreateConsumableTx(ctx context.Context, tx *sqlx.Tx, consumable *Consumable) error {
	consumable.ID = uuid.New()

	query := `
		INSERT INTO consumables (
			id, rarity_id, healing_amount, mana_amount,
			buff_duration, max_stack_size, description
		) VALUES (
			:id, :rarity_id, :healing_amount, :mana_amount,
			:buff_duration, :max_stack_size, :description
		)`

	_, err := tx.NamedExecContext(ctx, query, consumable)
	if err != nil {
		return fmt.Errorf("failed to create consumable (tx): %w", err)
	}

	return nil
}

func (r *repository) CreateItemTemplateTx(ctx context.Context, tx *sqlx.Tx, template *ItemTemplate) error {
	template.ID = uuid.New()

	query := `
		INSERT INTO item_templates (
			id, item_name, rarity_id, item_type, item_id,
			icon_url, required_level, base_sell_price, base_buy_price
		) VALUES (
			:id, :item_name, :rarity_id, :item_type, :item_id,
			:icon_url, :required_level, :base_sell_price, :base_buy_price
		)`

	_, err := tx.NamedExecContext(ctx, query, template)
	if err != nil {
		return fmt.Errorf("failed to create item template (tx): %w", err)
	}

	return nil
}

func (r *repository) UpsertItemInstanceTx(ctx context.Context, tx *sqlx.Tx, instance *ItemInstance) error {
	if instance.ID == uuid.Nil {
		instance.ID = uuid.New()
	}

	query := `
		INSERT INTO item_instances (
			id, template_id, owner_member_id, source,
			item_type, name, rarity_id,
			attack_power, critical_rate, weapon_type,
			defense_rating, magic_resistance, armor_slot,
			healing_amount, mana_amount, buff_duration,
			durability, description, buy_price, sell_price,
			created_by, updated_by
		) VALUES (
			:id, :template_id, :owner_member_id, :source,
			:item_type, :name, :rarity_id,
			:attack_power, :critical_rate, :weapon_type,
			:defense_rating, :magic_resistance, :armor_slot,
			:healing_amount, :mana_amount, :buff_duration,
			:durability, :description, :buy_price, :sell_price,
			:created_by, :updated_by
		)
		ON CONFLICT (id) DO UPDATE SET
			template_id      = EXCLUDED.template_id,
			owner_member_id  = EXCLUDED.owner_member_id,
			source           = EXCLUDED.source,
			item_type        = EXCLUDED.item_type,
			name             = EXCLUDED.name,
			rarity_id        = EXCLUDED.rarity_id,
			attack_power     = EXCLUDED.attack_power,
			critical_rate    = EXCLUDED.critical_rate,
			weapon_type      = EXCLUDED.weapon_type,
			defense_rating   = EXCLUDED.defense_rating,
			magic_resistance = EXCLUDED.magic_resistance,
			armor_slot       = EXCLUDED.armor_slot,
			healing_amount   = EXCLUDED.healing_amount,
			mana_amount      = EXCLUDED.mana_amount,
			buff_duration    = EXCLUDED.buff_duration,
			durability       = EXCLUDED.durability,
			description      = EXCLUDED.description,
			buy_price        = EXCLUDED.buy_price,
			sell_price       = EXCLUDED.sell_price,
			updated_by       = EXCLUDED.updated_by`

	_, err := tx.NamedExecContext(ctx, query, instance)
	if err != nil {
		return fmt.Errorf("failed to upsert item instance (tx): %w", err)
	}

	return nil
}

func (r *repository) GetLoadout(ctx context.Context, req *GetLoadoutRequest) (*Loadout, error) {
	memberId := req.MemberId

	loadout := &Loadout{}

	// Explicit columns — table has audit columns (created_by/updated_by)
	// that Loadout struct doesn't map, so SELECT * would fail sqlx scan with
	// "missing destination name".
	query := `
	 SELECT id, member_id,
	        weapon_instance_id, head_instance_id, chest_instance_id,
	        gloves_instance_id, legs_instance_id,
	        ring_1_instance_id, ring_2_instance_id,
	        consumable_1_id, consumable_2_id, consumable_3_id,
	        created_at, updated_at
	 FROM player_loadouts
	 WHERE member_id = $1
	`

	err := r.DB.GetContext(ctx, loadout, query, memberId)
	if err != nil {
		if err == sql.ErrNoRows {
			return loadout, nil
		}
		return nil, fmt.Errorf("failed to get loadout: %w", err)
	}

	return loadout, nil
}

func (r *repository) GetItemInstanceByID(ctx context.Context, id uuid.UUID) (*ItemInstance, error) {
	item := &ItemInstance{}

	query := `
	 SELECT id, template_id, owner_member_id, source, item_type, name, rarity_id,
	        attack_power, critical_rate, weapon_type, defense_rating, magic_resistance,
	        armor_slot, healing_amount, mana_amount, buff_duration, buy_price, sell_price,
	        description, acquired_at, created_at, updated_at
	 FROM item_instances
	 WHERE id = $1
	`

	err := r.DB.GetContext(ctx, item, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get item instance: %w", err)
	}

	return item, nil
}

func (r *repository) ListItemInstances(ctx context.Context, req *ListItemInstancesRequest) ([]*ItemInstance, error) {
	items := []*ItemInstance{}

	query := `
	 SELECT id, template_id, owner_member_id, source, item_type, name, rarity_id,
	        attack_power, critical_rate, weapon_type, defense_rating, magic_resistance,
	        armor_slot, healing_amount, mana_amount, buff_duration, buy_price, sell_price,
	        description, acquired_at, created_at, updated_at
	 FROM item_instances
	 WHERE owner_member_id = $1
	 ORDER BY created_at DESC
	`

	err := r.DB.SelectContext(ctx, &items, query, req.MemberId)
	if err != nil {
		return nil, fmt.Errorf("failed to list item instances: %w", err)
	}

	return items, nil
}

func (r *repository) UpsertLoadoutSlot(ctx context.Context, req *UpdateLoadoutRequest) error {
	validSlots := map[string]string{
		"weapon":       "weapon_instance_id",
		"head":         "head_instance_id",
		"chest":        "chest_instance_id",
		"gloves":       "gloves_instance_id",
		"legs":         "legs_instance_id",
		"ring_1":       "ring_1_instance_id",
		"ring_2":       "ring_2_instance_id",
		"consumable_1": "consumable_1_id",
		"consumable_2": "consumable_2_id",
		"consumable_3": "consumable_3_id",
	}

	column, ok := validSlots[req.Slot]
	if !ok {
		return fmt.Errorf("invalid slot: %s", req.Slot)
	}

	// Upsert: first-time equippers won't have a player_loadouts row yet,
	// so a plain UPDATE would silently affect 0 rows.
	query := fmt.Sprintf(`
	 INSERT INTO player_loadouts (member_id, %s)
	 VALUES ($1, $2)
	 ON CONFLICT (member_id) DO UPDATE
	 SET %s = EXCLUDED.%s, updated_at = NOW()
	`, column, column, column)

	_, err := r.DB.ExecContext(ctx, query, req.MemberId, req.ItemInstanceId)
	if err != nil {
		return fmt.Errorf("failed to update loadout: %w", err)
	}

	return nil
}

func (r *repository) UpsertLoadoutSlotTx(ctx context.Context, tx *sqlx.Tx, req *UpdateLoadoutRequest) error {
	validSlots := map[string]string{
		"weapon":       "weapon_instance_id",
		"head":         "head_instance_id",
		"chest":        "chest_instance_id",
		"gloves":       "gloves_instance_id",
		"legs":         "legs_instance_id",
		"ring_1":       "ring_1_instance_id",
		"ring_2":       "ring_2_instance_id",
		"consumable_1": "consumable_1_id",
		"consumable_2": "consumable_2_id",
		"consumable_3": "consumable_3_id",
	}

	column, ok := validSlots[req.Slot]
	if !ok {
		return fmt.Errorf("invalid slot: %s", req.Slot)
	}

	query := fmt.Sprintf(`
	 INSERT INTO player_loadouts (member_id, %s)
	 VALUES ($1, $2)
	 ON CONFLICT (member_id) DO UPDATE
	 SET %s = EXCLUDED.%s, updated_at = NOW()
	`, column, column, column)

	_, err := tx.ExecContext(ctx, query, req.MemberId, req.ItemInstanceId)
	if err != nil {
		return fmt.Errorf("failed to update loadout (tx): %w", err)
	}

	return nil
}

// UpsertPlayerLoadoutTx writes the full loadout snapshot for a member.
// Nil slot pointers in req become NULL (slot cleared); non-nil values are set.
// updated_at is left to the BEFORE UPDATE trigger.
func (r *repository) UpsertPlayerLoadoutTx(ctx context.Context, tx *sqlx.Tx, req *UpsertPlayerLoadoutRequest) error {
	query := `
		INSERT INTO player_loadouts (
			member_id,
			weapon_instance_id,
			head_instance_id, chest_instance_id, legs_instance_id, gloves_instance_id,
			ring_1_instance_id, ring_2_instance_id,
			consumable_1_id, consumable_2_id, consumable_3_id
		) VALUES (
			:member_id,
			:weapon_instance_id,
			:head_instance_id, :chest_instance_id, :legs_instance_id, :gloves_instance_id,
			:ring_1_instance_id, :ring_2_instance_id,
			:consumable_1_id, :consumable_2_id, :consumable_3_id
		)
		ON CONFLICT (member_id) DO UPDATE SET
			weapon_instance_id = EXCLUDED.weapon_instance_id,
			head_instance_id   = EXCLUDED.head_instance_id,
			chest_instance_id  = EXCLUDED.chest_instance_id,
			legs_instance_id   = EXCLUDED.legs_instance_id,
			gloves_instance_id = EXCLUDED.gloves_instance_id,
			ring_1_instance_id = EXCLUDED.ring_1_instance_id,
			ring_2_instance_id = EXCLUDED.ring_2_instance_id,
			consumable_1_id    = EXCLUDED.consumable_1_id,
			consumable_2_id    = EXCLUDED.consumable_2_id,
			consumable_3_id    = EXCLUDED.consumable_3_id`

	_, err := tx.NamedExecContext(ctx, query, req)
	if err != nil {
		return fmt.Errorf("failed to upsert player loadout (tx): %w", err)
	}

	return nil
}

func (r *repository) BatchUpsertItemInstances(ctx context.Context, tx *sqlx.Tx, items []*ItemInstance) error {
	if len(items) == 0 {
		return nil
	}

	const colsPerRow = 20

	var b strings.Builder
	b.WriteString(`
		INSERT INTO item_instances (
			id, template_id, owner_member_id, source,
			item_type, name, rarity_id,
			attack_power, critical_rate, weapon_type,
			defense_rating, magic_resistance, armor_slot,
			healing_amount, mana_amount, buff_duration,
			durability, description, buy_price, sell_price
		) VALUES `)

	args := make([]any, 0, len(items)*colsPerRow)

	for i, item := range items {
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}

		if i > 0 {
			b.WriteString(", ")
		}

		base := i * colsPerRow
		fmt.Fprintf(&b,
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10,
			base+11, base+12, base+13, base+14, base+15, base+16, base+17, base+18, base+19, base+20,
		)

		args = append(args,
			item.ID,
			item.TemplateID,
			item.OwnerMemberID,
			item.Source,
			item.ItemType,
			item.Name,
			item.RarityID,
			item.AttackPower,
			item.CriticalRate,
			item.WeaponType,
			item.DefenseRating,
			item.MagicResistance,
			item.ArmorSlot,
			item.HealingAmount,
			item.ManaAmount,
			item.BuffDuration,
			item.Durability,
			item.Description,
			item.BuyPrice,
			item.SellPrice,
		)
	}

	b.WriteString(`
		ON CONFLICT (id) DO UPDATE SET
			template_id      = EXCLUDED.template_id,
			owner_member_id  = EXCLUDED.owner_member_id,
			source           = EXCLUDED.source,
			item_type        = EXCLUDED.item_type,
			name             = EXCLUDED.name,
			rarity_id        = EXCLUDED.rarity_id,
			attack_power     = EXCLUDED.attack_power,
			critical_rate    = EXCLUDED.critical_rate,
			weapon_type      = EXCLUDED.weapon_type,
			defense_rating   = EXCLUDED.defense_rating,
			magic_resistance = EXCLUDED.magic_resistance,
			armor_slot       = EXCLUDED.armor_slot,
			healing_amount   = EXCLUDED.healing_amount,
			mana_amount      = EXCLUDED.mana_amount,
			buff_duration    = EXCLUDED.buff_duration,
			durability       = EXCLUDED.durability,
			description      = EXCLUDED.description,
			buy_price        = EXCLUDED.buy_price,
			sell_price       = EXCLUDED.sell_price`)

	_, err := tx.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("failed to batch upsert item instances (tx): %w", err)
	}

	return nil
}
