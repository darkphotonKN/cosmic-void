-- Remove index
DROP INDEX IF EXISTS idx_members_role;

-- Remove role column from members table
ALTER TABLE members DROP COLUMN IF EXISTS role;
