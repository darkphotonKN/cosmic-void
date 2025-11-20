-- Drop trigger and function
DROP TRIGGER IF EXISTS set_members_updated_at ON members;
DROP FUNCTION IF EXISTS update_members_updated_at;

-- Drop members table
DROP TABLE IF EXISTS members;