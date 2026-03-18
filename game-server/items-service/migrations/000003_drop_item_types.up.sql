-- Migration 003: Drop item_types table and references

-- Drop foreign key constraints and type_id columns from dependent tables

-- 1. Drop type_id from weapons table
ALTER TABLE weapons DROP CONSTRAINT IF EXISTS weapons_type_id_fkey;
DROP INDEX IF EXISTS idx_weapons_type;
ALTER TABLE weapons DROP COLUMN IF EXISTS type_id;

-- 2. Drop type_id from armors table
ALTER TABLE armors DROP CONSTRAINT IF EXISTS armors_type_id_fkey;
DROP INDEX IF EXISTS idx_armors_type;
ALTER TABLE armors DROP COLUMN IF EXISTS type_id;

-- 3. Drop type_id from consumables table
ALTER TABLE consumables DROP CONSTRAINT IF EXISTS consumables_type_id_fkey;
DROP INDEX IF EXISTS idx_consumables_type;
ALTER TABLE consumables DROP COLUMN IF EXISTS type_id;

-- 4. Drop type_id from item_templates table
ALTER TABLE item_templates DROP CONSTRAINT IF EXISTS item_templates_type_id_fkey;
DROP INDEX IF EXISTS idx_item_templates_type;
ALTER TABLE item_templates DROP COLUMN IF EXISTS type_id;

-- 5. Finally drop the item_types table
DROP TABLE IF EXISTS item_types;