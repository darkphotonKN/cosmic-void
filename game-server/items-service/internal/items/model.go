package items

import (
	"time"

	"github.com/google/uuid"
)

// ItemType represents item type enum (weapon, armor, consumable, etc.)
type ItemType struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	TypeCode    string     `db:"type_code" json:"type_code"`       // 'weapon', 'armor', 'consumable'
	Name        string     `db:"name" json:"name"`                 // 'Weapon', 'Armor'
	Description *string    `db:"description" json:"description"`
	IsActive    bool       `db:"is_active" json:"is_active"`
	SortOrder   int        `db:"sort_order" json:"sort_order"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	CreatedBy   *uuid.UUID `db:"created_by" json:"created_by"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	UpdatedBy   *uuid.UUID `db:"updated_by" json:"updated_by"`
}

// ItemRarity represents rarity level (common, rare, epic, legendary)
type ItemRarity struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	RarityCode           string     `db:"rarity_code" json:"rarity_code"`                       // 'common', 'legendary'
	RarityName           string     `db:"rarity_name" json:"rarity_name"`                       // 'Common', 'Legendary'
	ColorHex             *string    `db:"color_hex" json:"color_hex"`                           // '#FFD700'
	DropRateMultiplier   float64    `db:"drop_rate_multiplier" json:"drop_rate_multiplier"`     // 0.01 ~ 1.00
	SortOrder            int        `db:"sort_order" json:"sort_order"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
	CreatedBy            *uuid.UUID `db:"created_by" json:"created_by"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updated_at"`
	UpdatedBy            *uuid.UUID `db:"updated_by" json:"updated_by"`
}

// Weapon represents weapon-specific attributes
type Weapon struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	TypeID       uuid.UUID  `db:"type_id" json:"type_id"`
	RarityID     uuid.UUID  `db:"rarity_id" json:"rarity_id"`
	AttackPower  int        `db:"attack_power" json:"attack_power"`
	Durability   int        `db:"durability" json:"durability"`
	CriticalRate *float64   `db:"critical_rate" json:"critical_rate"`   // 0.00 ~ 1.00
	WeaponType   *string    `db:"weapon_type" json:"weapon_type"`       // 'sword', 'axe', 'bow'
	Description  *string    `db:"description" json:"description"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	CreatedBy    *uuid.UUID `db:"created_by" json:"created_by"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	UpdatedBy    *uuid.UUID `db:"updated_by" json:"updated_by"`
}

// Armor represents armor-specific attributes
type Armor struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	TypeID           uuid.UUID  `db:"type_id" json:"type_id"`
	RarityID         uuid.UUID  `db:"rarity_id" json:"rarity_id"`
	DefenseRating    int        `db:"defense_rating" json:"defense_rating"`
	Durability       int        `db:"durability" json:"durability"`
	MagicResistance  *int       `db:"magic_resistance" json:"magic_resistance"`
	ArmorSlot        *string    `db:"armor_slot" json:"armor_slot"`         // 'head', 'chest', 'legs', 'shield'
	Description      *string    `db:"description" json:"description"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	CreatedBy        *uuid.UUID `db:"created_by" json:"created_by"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
	UpdatedBy        *uuid.UUID `db:"updated_by" json:"updated_by"`
}

// Consumable represents consumable item attributes
type Consumable struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	TypeID        uuid.UUID  `db:"type_id" json:"type_id"`
	RarityID      uuid.UUID  `db:"rarity_id" json:"rarity_id"`
	HealingAmount *int       `db:"healing_amount" json:"healing_amount"`
	ManaAmount    *int       `db:"mana_amount" json:"mana_amount"`
	BuffDuration  *int       `db:"buff_duration" json:"buff_duration"`       // seconds
	MaxStackSize  int        `db:"max_stack_size" json:"max_stack_size"`     // default 99
	Description   *string    `db:"description" json:"description"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	CreatedBy     *uuid.UUID `db:"created_by" json:"created_by"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
	UpdatedBy     *uuid.UUID `db:"updated_by" json:"updated_by"`
}

// ItemTemplate represents the unified item template (polymorphic pattern)
type ItemTemplate struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	ItemName       string     `db:"item_name" json:"item_name"`
	ItemCode       string     `db:"item_code" json:"item_code"`               // unique identifier
	TypeID         uuid.UUID  `db:"type_id" json:"type_id"`
	RarityID       uuid.UUID  `db:"rarity_id" json:"rarity_id"`
	ItemType       string     `db:"item_type" json:"item_type"`               // 'weapon', 'armor', 'consumable'
	ItemID         uuid.UUID  `db:"item_id" json:"item_id"`                   // references weapon/armor/consumable id
	IconURL        *string    `db:"icon_url" json:"icon_url"`
	IsTradeable    bool       `db:"is_tradeable" json:"is_tradeable"`
	IsDroppable    bool       `db:"is_droppable" json:"is_droppable"`
	RequiredLevel  int        `db:"required_level" json:"required_level"`
	BaseSellPrice  int        `db:"base_sell_price" json:"base_sell_price"`
	BaseBuyPrice   int        `db:"base_buy_price" json:"base_buy_price"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	CreatedBy      *uuid.UUID `db:"created_by" json:"created_by"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
	UpdatedBy      *uuid.UUID `db:"updated_by" json:"updated_by"`
}

// CreateItemTypeRequest represents the request to create an item type
type CreateItemTypeRequest struct {
	TypeCode    string  `json:"type_code" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	SortOrder   int     `json:"sort_order"`
}

// CreateItemRarityRequest represents the request to create an item rarity
type CreateItemRarityRequest struct {
	RarityCode         string   `json:"rarity_code" binding:"required"`
	RarityName         string   `json:"rarity_name" binding:"required"`
	ColorHex           *string  `json:"color_hex"`
	DropRateMultiplier float64  `json:"drop_rate_multiplier" binding:"required,gt=0,lte=1"`
	SortOrder          int      `json:"sort_order"`
}

// CreateWeaponRequest represents the request to create a weapon
type CreateWeaponRequest struct {
	TypeID       uuid.UUID `json:"type_id" binding:"required"`
	RarityID     uuid.UUID `json:"rarity_id" binding:"required"`
	AttackPower  int       `json:"attack_power" binding:"required,gte=0"`
	Durability   int       `json:"durability" binding:"required,gte=0"`
	CriticalRate *float64  `json:"critical_rate"`
	WeaponType   *string   `json:"weapon_type"`
	Description  *string   `json:"description"`
}

// CreateArmorRequest represents the request to create an armor
type CreateArmorRequest struct {
	TypeID          uuid.UUID `json:"type_id" binding:"required"`
	RarityID        uuid.UUID `json:"rarity_id" binding:"required"`
	DefenseRating   int       `json:"defense_rating" binding:"required,gte=0"`
	Durability      int       `json:"durability" binding:"required,gte=0"`
	MagicResistance *int      `json:"magic_resistance"`
	ArmorSlot       *string   `json:"armor_slot"`
	Description     *string   `json:"description"`
}

// CreateConsumableRequest represents the request to create a consumable
type CreateConsumableRequest struct {
	TypeID        uuid.UUID `json:"type_id" binding:"required"`
	RarityID      uuid.UUID `json:"rarity_id" binding:"required"`
	HealingAmount *int      `json:"healing_amount"`
	ManaAmount    *int      `json:"mana_amount"`
	BuffDuration  *int      `json:"buff_duration"`
	MaxStackSize  int       `json:"max_stack_size" binding:"required,gt=0"`
	Description   *string   `json:"description"`
}

// CreateItemTemplateRequest represents the request to create an item template
type CreateItemTemplateRequest struct {
	ItemName      string    `json:"item_name" binding:"required"`
	ItemCode      string    `json:"item_code" binding:"required"`
	TypeID        uuid.UUID `json:"type_id" binding:"required"`
	RarityID      uuid.UUID `json:"rarity_id" binding:"required"`
	ItemType      string    `json:"item_type" binding:"required,oneof=weapon armor consumable"`
	ItemID        uuid.UUID `json:"item_id" binding:"required"`
	IconURL       *string   `json:"icon_url"`
	IsTradeable   *bool     `json:"is_tradeable"`
	IsDroppable   *bool     `json:"is_droppable"`
	RequiredLevel *int      `json:"required_level"`
	BaseSellPrice *int      `json:"base_sell_price"`
	BaseBuyPrice  *int      `json:"base_buy_price"`
}

// ArmorWithTemplate represents an armor joined with its item template
type ArmorWithTemplate struct {
	// Armor fields
	ID              uuid.UUID `db:"id"`
	TypeID          uuid.UUID `db:"type_id"`
	RarityID        uuid.UUID `db:"rarity_id"`
	DefenseRating   int       `db:"defense_rating"`
	Durability      int       `db:"durability"`
	MagicResistance *int      `db:"magic_resistance"`
	ArmorSlot       *string   `db:"armor_slot"`
	Description     *string   `db:"description"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`

	// ItemTemplate fields
	ItemTemplateID uuid.UUID `db:"item_template_id"`
	ItemName       string    `db:"item_name"`
	ItemCode       string    `db:"item_code"`
	IconURL        *string   `db:"icon_url"`
	RequiredLevel  int       `db:"required_level"`

	IsTradeable    bool      `db:"is_tradeable"`
	IsDroppable    bool      `db:"is_droppable"`

	BaseSellPrice  int       `db:"base_sell_price"`
	BaseBuyPrice   int       `db:"base_buy_price"`
}

// ConsumableWithTemplate represents a consumable joined with its item template
type ConsumableWithTemplate struct {
	// Consumable fields
	ID            uuid.UUID `db:"id"`
	TypeID        uuid.UUID `db:"type_id"`
	RarityID      uuid.UUID `db:"rarity_id"`
	HealingAmount *int      `db:"healing_amount"`
	ManaAmount    *int      `db:"mana_amount"`
	BuffDuration  *int      `db:"buff_duration"`
	MaxStackSize  int       `db:"max_stack_size"`
	Description   *string   `db:"description"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`

	// ItemTemplate fields
	ItemTemplateID uuid.UUID `db:"item_template_id"`
	ItemName       string    `db:"item_name"`
	ItemCode       string    `db:"item_code"`
	IconURL        *string   `db:"icon_url"`
	IsTradeable    bool      `db:"is_tradeable"`
	IsDroppable    bool      `db:"is_droppable"`
	RequiredLevel  int       `db:"required_level"`
	BaseSellPrice  int       `db:"base_sell_price"`
	BaseBuyPrice   int       `db:"base_buy_price"`
}

// WeaponWithTemplate represents a weapon joined with its item template
// This is used for detailed queries that need both weapon and template information
type WeaponWithTemplate struct {
	// Weapon fields
	ID           uuid.UUID  `db:"id"`
	TypeID       uuid.UUID  `db:"type_id"`
	RarityID     uuid.UUID  `db:"rarity_id"`
	AttackPower  int        `db:"attack_power"`
	Durability   int        `db:"durability"`
	CriticalRate *float64   `db:"critical_rate"`
	WeaponType   *string    `db:"weapon_type"`
	Description  *string    `db:"description"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`

	// ItemTemplate fields
	ItemTemplateID   uuid.UUID `db:"item_template_id"`
	ItemName         string    `db:"item_name"`
	ItemCode         string    `db:"item_code"`
	IconURL          *string   `db:"icon_url"`
	IsTradeable      bool      `db:"is_tradeable"`
	IsDroppable      bool      `db:"is_droppable"`
	RequiredLevel    int       `db:"required_level"`
	BaseSellPrice    int       `db:"base_sell_price"`
	BaseBuyPrice     int       `db:"base_buy_price"`
}
