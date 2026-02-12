ALTER TABLE members ADD COLUMN stripe_subscription_product_id TEXT;
ALTER TABLE members ADD COLUMN stripe_subscription_status TEXT NOT NULL DEFAULT '';
