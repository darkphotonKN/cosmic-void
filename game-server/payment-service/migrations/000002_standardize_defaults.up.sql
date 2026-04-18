-- Standardize column defaults on CURRENT_TIMESTAMP (SQL-standard) to match
-- every other service in the monorepo. Functionally identical to NOW() in Postgres.
ALTER TABLE subscriptions
  ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP,
  ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;
