-- Rollback Migration 006: Drop player_loadouts table

-- Drop trigger first
DROP TRIGGER IF EXISTS player_loadouts_updated_at ON player_loadouts;

-- Drop index
DROP INDEX IF EXISTS idx_player_loadouts_member;

-- Drop the table
DROP TABLE IF EXISTS player_loadouts;