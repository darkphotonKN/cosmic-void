-- Drop indexes
DROP INDEX IF EXISTS idx_consumables_type;
DROP INDEX IF EXISTS idx_armors_type;
DROP INDEX IF EXISTS idx_weapons_type;
DROP INDEX IF EXISTS idx_item_templates_itemable;

-- Drop constraints
ALTER TABLE item_templates DROP CONSTRAINT IF EXISTS item_templates_item_type_item_id_key;
ALTER TABLE item_templates DROP CONSTRAINT IF EXISTS valid_item_type;

-- Drop columns from item_templates
ALTER TABLE item_templates DROP COLUMN IF EXISTS item_id;
ALTER TABLE item_templates DROP COLUMN IF EXISTS item_type;

-- Drop type_id columns
ALTER TABLE consumables DROP COLUMN IF EXISTS type_id;
ALTER TABLE armors DROP COLUMN IF EXISTS type_id;
ALTER TABLE weapons DROP COLUMN IF EXISTS type_id;
