package types

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
