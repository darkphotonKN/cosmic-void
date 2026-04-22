-- Revert TIMESTAMPTZ back to TIMESTAMP using UTC explicitly,
-- so the conversion is independent of the session's TimeZone setting
-- and symmetrical with the up migration.
ALTER TABLE item_rarities
  ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE weapons
  ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE armors
  ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE consumables
  ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE item_templates
  ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE item_instances
  ALTER COLUMN acquired_at TYPE TIMESTAMP USING acquired_at AT TIME ZONE 'UTC',
  ALTER COLUMN created_at  TYPE TIMESTAMP USING created_at  AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at  TYPE TIMESTAMP USING updated_at  AT TIME ZONE 'UTC';

ALTER TABLE player_loadouts
  ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';
