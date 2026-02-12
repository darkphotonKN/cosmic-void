DROP INDEX IF EXISTS idx_members_stripe_customer_id;
ALTER TABLE members DROP COLUMN IF EXISTS stripe_customer_id;
