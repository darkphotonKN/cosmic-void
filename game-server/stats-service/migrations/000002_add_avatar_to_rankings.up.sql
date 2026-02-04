-- Add avatar_url to player_ranking_stats for denormalized leaderboard performance
ALTER TABLE player_ranking_stats ADD COLUMN avatar_url TEXT;