-- Rollback Migration 005: Drop item_instances table

-- Drop trigger first
DROP TRIGGER IF EXISTS item_instances_updated_at ON item_instances;

-- Drop indexes
DROP INDEX IF EXISTS idx_item_instances_owner;
DROP INDEX IF EXISTS idx_item_instances_template;

-- Drop the table
DROP TABLE IF EXISTS item_instances;