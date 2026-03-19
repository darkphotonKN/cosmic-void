package types

import grpcitems "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/items"

// List of Item Types

// top level category
type ItemCategory string

const (
	Equipment      ItemCategory = "equipment"
	Mischellaneous ItemCategory = "mischellanous"
)

// Equipment
type EquipmentItemType string

const (
	Weapon    EquipmentItemType = "weapon"
	BodyArmor EquipmentItemType = "bodyArmor"
	Jewellry  EquipmentItemType = "jewellry"
	Gloves    EquipmentItemType = "gloves"
	Boots     EquipmentItemType = "boots"
	Helmet    EquipmentItemType = "helmet"
)

// Currency and Others
type MiscellaneousItemType string

const (
	Currency EquipmentItemType = "currency"
	Map      EquipmentItemType = "map"
	Gem      EquipmentItemType = "gem"
)

type ItemConfig struct {
	Name          string
	ItemTool      grpcitems.ItemsClient
	AttackPower   int32
	Durability    int32
	CriticalRate  float32
	WeaponType    string
	DefenseRating int32
	ArmorSlot     string
	HealingAmount int32
	ManaAmount    int32
	Description   string
}
