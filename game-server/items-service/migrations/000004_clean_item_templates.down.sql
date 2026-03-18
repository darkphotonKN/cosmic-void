-- Rollback Migration 004: Restore item_templates columns

-- Re-add columns
ALTER TABLE item_templates ADD COLUMN item_code VARCHAR(100) UNIQUE NOT NULL DEFAULT 'temp_code';
ALTER TABLE item_templates ADD COLUMN item_type VARCHAR(50) NOT NULL DEFAULT 'weapon';
ALTER TABLE item_templates ADD COLUMN item_id UUID NOT NULL DEFAULT uuid_generate_v4();

-- Re-add check constraint
ALTER TABLE item_templates ADD CONSTRAINT valid_item_type
    CHECK (item_type IN ('weapon', 'armor', 'consumable'));

-- Re-add unique constraint
ALTER TABLE item_templates ADD CONSTRAINT item_templates_item_type_item_id_key UNIQUE(item_type, item_id);

-- Recreate indexes
CREATE INDEX idx_item_templates_code ON item_templates(item_code);
CREATE INDEX idx_item_templates_itemable ON item_templates(item_type, item_id);

-- Note: You'll need to update the default values with actual data after rollback