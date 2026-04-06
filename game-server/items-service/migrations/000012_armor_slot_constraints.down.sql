-- Rollback Migration 012: Remove armor_slot constraints + revert loadout column names

-- Drop CHECK constraints
ALTER TABLE item_instances DROP CONSTRAINT IF EXISTS item_instances_armor_slot_check;
ALTER TABLE armors         DROP CONSTRAINT IF EXISTS armors_armor_slot_check;

-- Rename player_loadouts columns back
ALTER TABLE player_loadouts RENAME COLUMN gloves_instance_id TO hands_instance_id;
ALTER TABLE player_loadouts RENAME COLUMN legs_instance_id   TO feet_instance_id;
ALTER TABLE player_loadouts RENAME COLUMN chest_instance_id  TO body_instance_id;
