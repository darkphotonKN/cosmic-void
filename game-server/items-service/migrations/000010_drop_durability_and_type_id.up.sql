-- Migration 010: Drop durability from weapons/armors, drop type_id from weapons/armors/consumables

-- Drop durability
ALTER TABLE weapons DROP COLUMN IF EXISTS durability;
ALTER TABLE weapons DROP CONSTRAINT IF EXISTS weapon_durability_positive;
ALTER TABLE armors DROP COLUMN IF EXISTS durability;
ALTER TABLE armors DROP CONSTRAINT IF EXISTS armor_durability_positive;

-- Drop type_id (dangling UUID, item_types table was dropped in 003)
DROP INDEX IF EXISTS idx_weapons_type;
ALTER TABLE weapons DROP COLUMN IF EXISTS type_id;
DROP INDEX IF EXISTS idx_armors_type;
ALTER TABLE armors DROP COLUMN IF EXISTS type_id;
DROP INDEX IF EXISTS idx_consumables_type;
ALTER TABLE consumables DROP COLUMN IF EXISTS type_id;
