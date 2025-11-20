-- Drop trigger and function
DROP TRIGGER IF EXISTS set_tags_updated_at ON tags;
DROP FUNCTION IF EXISTS update_tags_updated_at;

-- Drop table
DROP TABLE IF EXISTS tags; 