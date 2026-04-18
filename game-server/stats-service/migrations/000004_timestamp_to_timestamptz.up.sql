-- Convert TIMESTAMP columns to TIMESTAMPTZ for correct timezone handling.
-- Existing rows are interpreted as the server's local timezone and converted to UTC.
-- The outbox table (created in 000003) is already TIMESTAMPTZ and is excluded.
ALTER TABLE player_match_stats
  ALTER COLUMN created_at TYPE TIMESTAMPTZ,
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ;

ALTER TABLE player_ranking_stats
  ALTER COLUMN last_calculated_at TYPE TIMESTAMPTZ,
  ALTER COLUMN created_at TYPE TIMESTAMPTZ,
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ;

ALTER TABLE match_history
  ALTER COLUMN match_started_at TYPE TIMESTAMPTZ,
  ALTER COLUMN created_at TYPE TIMESTAMPTZ;
