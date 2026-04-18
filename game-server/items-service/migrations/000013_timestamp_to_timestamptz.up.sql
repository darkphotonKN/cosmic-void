-- Convert TIMESTAMP columns to TIMESTAMPTZ for correct timezone handling.
-- Existing rows are interpreted as the server's local timezone and converted to UTC.
-- item_types was dropped in 000003 and is excluded.
ALTER TABLE item_rarities
  ALTER COLUMN created_at TYPE TIMESTAMPTZ,
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ;

ALTER TABLE weapons
  ALTER COLUMN created_at TYPE TIMESTAMPTZ,
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ;

ALTER TABLE armors
  ALTER COLUMN created_at TYPE TIMESTAMPTZ,
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ;

ALTER TABLE consumables
  ALTER COLUMN created_at TYPE TIMESTAMPTZ,
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ;

ALTER TABLE item_templates
  ALTER COLUMN created_at TYPE TIMESTAMPTZ,
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ;

ALTER TABLE item_instances
  ALTER COLUMN acquired_at TYPE TIMESTAMPTZ,
  ALTER COLUMN created_at TYPE TIMESTAMPTZ,
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ;

ALTER TABLE player_loadouts
  ALTER COLUMN created_at TYPE TIMESTAMPTZ,
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ;
