-- Migration 004: Clean up item_templates redundant columns

-- Drop unique constraint on (item_type, item_id) if it exists
ALTER TABLE item_templates DROP CONSTRAINT IF EXISTS item_templates_item_type_item_id_key;

-- Drop indexes
DROP INDEX IF EXISTS idx_item_templates_code;
DROP INDEX IF EXISTS idx_item_templates_itemable;

-- Drop columns
ALTER TABLE item_templates DROP COLUMN IF EXISTS item_code;
ALTER TABLE item_templates DROP COLUMN IF EXISTS item_id;
ALTER TABLE item_templates DROP COLUMN IF EXISTS item_type;

-- Drop check constraint on item_type if it exists (should be removed with column, but being explicit)
ALTER TABLE item_templates DROP CONSTRAINT IF EXISTS valid_item_type;