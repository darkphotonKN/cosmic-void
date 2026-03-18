-- Drop is_tradeable and is_droppable from item_templates
ALTER TABLE item_templates DROP COLUMN IF EXISTS is_tradeable;
ALTER TABLE item_templates DROP COLUMN IF EXISTS is_droppable;
