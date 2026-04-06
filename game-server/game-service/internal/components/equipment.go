package components

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
)

type EquipmentComponent struct {
	// Weapons
	WeaponSlot *uuid.UUID

	// Armor
	HeadSlot   *uuid.UUID
	ChestSlot  *uuid.UUID
	GlovesSlot *uuid.UUID
	LegsSlot   *uuid.UUID

	// Accessories
	Ring1Slot *uuid.UUID
	Ring2Slot *uuid.UUID

	// Consumables
	Consumable1 *uuid.UUID
	Consumable2 *uuid.UUID
	Consumable3 *uuid.UUID
}

type EquipmentConfig struct {
	// Weapons
	WeaponSlot *uuid.UUID

	// Armor
	HeadSlot   *uuid.UUID
	ChestSlot  *uuid.UUID
	GlovesSlot *uuid.UUID
	LegsSlot   *uuid.UUID

	// Accessories
	Ring1Slot *uuid.UUID
	Ring2Slot *uuid.UUID

	// Consumables
	Consumable1 *uuid.UUID
	Consumable2 *uuid.UUID
	Consumable3 *uuid.UUID
}

func (h *EquipmentComponent) Type() ecs.ComponentType {
	return ecs.ComponentTypeEquipment
}

func NewEquipmentComponent(config *EquipmentConfig) *EquipmentComponent {
	if config == nil {
		return &EquipmentComponent{}
	}

	return &EquipmentComponent{
		WeaponSlot:  config.WeaponSlot,
		HeadSlot:    config.HeadSlot,
		ChestSlot:   config.ChestSlot,
		GlovesSlot:  config.GlovesSlot,
		LegsSlot:    config.LegsSlot,
		Ring1Slot:   config.Ring1Slot,
		Ring2Slot:   config.Ring2Slot,
		Consumable1: config.Consumable1,
		Consumable2: config.Consumable2,
		Consumable3: config.Consumable3,
	}
}
