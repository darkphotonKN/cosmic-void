package types

import (
	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/google/uuid"
)

type Player struct {
	ID                   uuid.UUID
	Username             string
	CurrentGameSessionId uuid.UUID
	ConnectState         *constants.ConnectState
}

type PlayerState struct {
	ID        uuid.UUID        `json:"id"`
	EntityID  uuid.UUID        `json:"entity_id"`
	Username  string           `json:"username"`
	Position  *Position        `json:"position"`
	Direction *PlayerDirection `json:"direction"`
	Inventory []*ItemState     `json:"inventory"`
	Escape    bool             `json:"escape"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type PlayerDirection struct {
	VX    float64 `json:"vx"`
	VY    float64 `json:"vy"`
	Speed float64 `json:"speed"`
}

type DoorState struct {
	EntityID uuid.UUID `json:"entity_id"`
	Position *Position `json:"position"`
	Width    float64   `json:"width"`
	Height   float64   `json:"height"`
	IsOpen   bool      `json:"is_open"`
}

type ItemState struct {
	ItemID        uuid.UUID `json:"item_id"`   // base item id
	EntityID      uuid.UUID `json:"entity_id"` // unique items entity id
	Name          string    `json:"name"`
	Quantity      int       `json:"quantity"`
	AttackPower   int32     `json:"attack_power,omitempty"`   // weapon
	CriticalRate  float32   `json:"critical_rate,omitempty"`  // weapon
	WeaponType    string    `json:"weapon_type,omitempty"`    // weapon
	DefenseRating int32     `json:"defense_rating,omitempty"` // armor
	ArmorSlot     ArmorSlot `json:"armor_slot,omitempty"`     // armor
	HealingAmount int32     `json:"healing_amount,omitempty"` // consumable
	ManaAmount    int32     `json:"mana_amount,omitempty"`    // consumable
	Description   string    `json:"description,omitempty"`    // all types
}

type ContainerState struct {
	ContainerID uuid.UUID    `json:"container_id"`
	EntityID    uuid.UUID    `json:"entity_id"`
	Position    *Position    `json:"position"`
	IsOpen      bool         `json:"is_open"`
	Items       []*ItemState `json:"items"`
}

type RawMatchState struct {
	SessionID uuid.UUID
	StartedAt time.Time
	EndedAt   time.Time
	Players   []RawPlayerState
	// [player's memberID] Position
	EliminationOrder map[uuid.UUID]int
}

type RawPlayerState struct {
	MemberID  string
	Username  string
	Kills     int32
	Deaths    int32
	Escape    bool
	Equipment ExtractedEquipment
	Inventory []*ExtractedItem
}

type RankedPlayerState struct {
	MemberID      string
	Username      string
	Kills         int32
	Deaths        int32
	FinalPosition int32
	Win           bool
	Escape        bool
	Equipment     ExtractedEquipment
	Inventory     []*ExtractedItem
}

type ExtractedItem struct {
	TemplateID uuid.UUID
	ItemType   string
	Name       string

	// Weapon stats
	AttackPower  int
	CriticalRate float64
	WeaponType   string

	// Armor stats
	DefenseRating   int
	MagicResistance int
	ArmorSlot       string

	// Consumable stats
	HealingAmount int
	ManaAmount    int
	BuffDuration  int

	// Shared
	BuyPrice    int
	SellPrice   int
	Description string
}

type ExtractedEquipment struct {
	// Weapons
	WeaponSlot *ExtractedItem

	// Armor
	HeadSlot   *ExtractedItem
	ChestSlot  *ExtractedItem
	GlovesSlot *ExtractedItem
	LegsSlot   *ExtractedItem

	// Accessories
	Ring1Slot *ExtractedItem
	Ring2Slot *ExtractedItem

	// Consumablesected events
	Consumable1 *ExtractedItem
	Consumable2 *ExtractedItem
	Consumable3 *ExtractedItem
}

type FormattedMatchData struct {
	MatchEndedEvent     []byte
	ItemsExtractedEvent []byte
}

type WallState struct {
	HouseID  uuid.UUID `json:"house_id"`
	EntityID uuid.UUID `json:"entity_id"`
	Position *Position `json:"position"`
	Width    float64   `json:"width"`
	Height   float64   `json:"height"`
}

type EscapeDoorState struct {
	EntityID uuid.UUID `json:"entity_id"`
	Position *Position `json:"position"`
	IsOpen   bool      `json:"is_open"`
	IsLocked bool      `json:"is_locked"`
}

type SwitchState struct {
	EntityID    uuid.UUID `json:"entity_id"`
	Position    *Position `json:"position"`
	SwitchID    int       `json:"switch_id"`
	IsActivated bool      `json:"is_activated"`
}

type ItemPool struct {
	Count       int
	Weapons     []*ItemConfig
	Armor       []*ItemConfig
	Consumables []*ItemConfig
}

type ItemConfig struct {
	TemplateID      uuid.UUID
	ItemType        ItemType
	Name            string
	AttackPower     int
	CriticalRate    float64
	WeaponType      string
	DefenseRating   int
	MagicResistance int
	ArmorSlot       ArmorSlot
	HealingAmount   int
	ManaAmount      int
	BuffDuration    int
	BuyPrice        int
	SellPrice       int
	Description     string
}

type ItemType string

const (
	ItemTypeWeapon     ItemType = "weapon"
	ItemTypeArmor      ItemType = "armor"
	ItemTypeConsumable ItemType = "consumable"
)

type ArmorSlot string

const (
	ArmorSlotHead   ArmorSlot = "head"
	ArmorSlotChest  ArmorSlot = "chest"
	ArmorSlotLegs   ArmorSlot = "legs"
	ArmorSlotGloves ArmorSlot = "gloves"
)
