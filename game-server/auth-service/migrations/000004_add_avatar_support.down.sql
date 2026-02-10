-- Drop avatar_uploads table
DROP TABLE IF EXISTS avatar_uploads;

-- Remove avatar_url column from members table
ALTER TABLE members DROP COLUMN IF EXISTS avatar_url;