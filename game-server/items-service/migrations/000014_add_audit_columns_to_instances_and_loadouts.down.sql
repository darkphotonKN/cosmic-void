-- Rollback: remove audit columns from player_loadouts and item_instances.

ALTER TABLE item_instances
    DROP COLUMN IF EXISTS updated_by,
    DROP COLUMN IF EXISTS created_by;

ALTER TABLE player_loadouts
    DROP COLUMN IF EXISTS updated_by,
    DROP COLUMN IF EXISTS created_by;
