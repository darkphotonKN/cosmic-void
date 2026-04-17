-- Migration 014: Add created_by / updated_by audit columns to
-- player_loadouts and item_instances.
--
-- Why:
--   The shared update_updated_at_column() trigger is attached to both tables
--   and sets NEW.updated_by. Without the column, every UPDATE (including
--   INSERT ... ON CONFLICT DO UPDATE) fails with
--   "record new has no field updated_by".
--   Matches the audit-column convention used by armors / weapons /
--   consumables / item_templates / item_rarities.

ALTER TABLE player_loadouts
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;

ALTER TABLE item_instances
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;
