-- Convert TIMESTAMP columns to TIMESTAMPTZ for correct timezone handling.
-- Existing rows are interpreted as the server's local timezone and converted to UTC.
ALTER TABLE notifications
  ALTER COLUMN sent_at TYPE TIMESTAMPTZ,
  ALTER COLUMN created_at TYPE TIMESTAMPTZ,
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ;
