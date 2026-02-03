-- Drop trigger
DROP TRIGGER IF EXISTS update_notifications_updated_at ON notifications;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop table
DROP TABLE IF EXISTS notifications;