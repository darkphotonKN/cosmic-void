-- Restore is_tradeable and is_droppable to item_templates
ALTER TABLE item_templates ADD COLUMN IF NOT EXISTS is_tradeable BOOLEAN DEFAULT TRUE;
ALTER TABLE item_templates ADD COLUMN IF NOT EXISTS is_droppable BOOLEAN DEFAULT TRUE;
