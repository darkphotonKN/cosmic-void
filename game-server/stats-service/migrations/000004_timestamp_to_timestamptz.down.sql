-- Revert TIMESTAMPTZ back to TIMESTAMP
ALTER TABLE player_match_stats
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

ALTER TABLE player_ranking_stats
  ALTER COLUMN last_calculated_at TYPE TIMESTAMP,
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

ALTER TABLE match_history
  ALTER COLUMN match_started_at TYPE TIMESTAMP,
  ALTER COLUMN created_at TYPE TIMESTAMP;
