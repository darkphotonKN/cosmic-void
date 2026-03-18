-- Restore type_id to weapons, armors, consumables (plain UUID, no FK -- item_types table was dropped in 003)
ALTER TABLE weapons ADD COLUMN IF NOT EXISTS type_id UUID;
ALTER TABLE armors ADD COLUMN IF NOT EXISTS type_id UUID;
ALTER TABLE consumables ADD COLUMN IF NOT EXISTS type_id UUID;

-- Restore item_type and item_id to item_templates
ALTER TABLE item_templates ADD COLUMN IF NOT EXISTS item_type VARCHAR(50);
ALTER TABLE item_templates ADD COLUMN IF NOT EXISTS item_id UUID;

-- Constraints and indexes
ALTER TABLE item_templates ADD CONSTRAINT valid_item_type CHECK (item_type IN ('weapon', 'armor', 'consumable'));
ALTER TABLE item_templates ADD CONSTRAINT item_templates_item_type_item_id_key UNIQUE(item_type, item_id);
CREATE INDEX IF NOT EXISTS idx_item_templates_itemable ON item_templates(item_type, item_id);
CREATE INDEX IF NOT EXISTS idx_weapons_type ON weapons(type_id);
CREATE INDEX IF NOT EXISTS idx_armors_type ON armors(type_id);
CREATE INDEX IF NOT EXISTS idx_consumables_type ON consumables(type_id);
