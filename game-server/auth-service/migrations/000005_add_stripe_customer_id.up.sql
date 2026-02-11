-- Add stripe_customer_id to members table
ALTER TABLE members ADD COLUMN stripe_customer_id TEXT;

-- Unique index to ensure one Stripe customer per member
CREATE UNIQUE INDEX idx_members_stripe_customer_id ON members(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL;
