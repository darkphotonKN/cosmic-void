-- Revert defaults to NOW()
ALTER TABLE subscriptions
  ALTER COLUMN created_at SET DEFAULT NOW(),
  ALTER COLUMN updated_at SET DEFAULT NOW();
