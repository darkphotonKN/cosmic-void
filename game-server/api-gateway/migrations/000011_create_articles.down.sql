-- Drop trigger and function
DROP TRIGGER IF EXISTS set_articles_updated_at ON articles;
DROP FUNCTION IF EXISTS update_articles_updated_at;

-- Drop table
DROP TABLE IF EXISTS articles; 