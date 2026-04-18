-- Convert TIMESTAMP columns to TIMESTAMPTZ for correct timezone handling.
-- Existing rows are interpreted as the server's local timezone and converted to UTC.
ALTER TABLE members
  ALTER COLUMN created_at TYPE TIMESTAMPTZ,
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ;

ALTER TABLE player_profile
  ALTER COLUMN created_at TYPE TIMESTAMPTZ,
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ;

ALTER TABLE avatar_uploads
  ALTER COLUMN presigned_url_expires_at TYPE TIMESTAMPTZ,
  ALTER COLUMN created_at TYPE TIMESTAMPTZ,
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ;
