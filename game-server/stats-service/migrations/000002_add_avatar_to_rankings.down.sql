-- Remove avatar_url column from player_ranking_stats table
ALTER TABLE player_ranking_stats DROP COLUMN IF EXISTS avatar_url;