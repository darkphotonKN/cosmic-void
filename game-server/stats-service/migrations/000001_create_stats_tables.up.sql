-- Player match stats (aggregated stats)
CREATE TABLE player_match_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    member_id UUID NOT NULL,

    -- Match statistics
    games_played INTEGER DEFAULT 0 NOT NULL,
    wins INTEGER DEFAULT 0 NOT NULL,
    losses INTEGER DEFAULT 0 NOT NULL,
    kills INTEGER DEFAULT 0 NOT NULL,
    deaths INTEGER DEFAULT 0 NOT NULL,
    times_placed_top_three INTEGER DEFAULT 0 NOT NULL,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT player_match_stats_wins_check CHECK (wins >= 0),
    CONSTRAINT player_match_stats_losses_check CHECK (losses >= 0),
    CONSTRAINT player_match_stats_games_check CHECK (games_played >= 0),

    -- Ensure one stats record per member
    UNIQUE(member_id)
);

-- Player ranking stats (denormalized for leaderboard performance)
CREATE TABLE player_ranking_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id UUID NOT NULL,

    -- Denormalized member data for leaderboard performance
    username VARCHAR(255) NOT NULL,
    wins INTEGER DEFAULT 0 NOT NULL,
    top_threes INTEGER DEFAULT 0 NOT NULL,

    -- ELO / MMR (calculated behind the scenes)
    rating INTEGER DEFAULT 1000 NOT NULL,

    -- Ranking position (updated periodically)
    rank_position INTEGER,

    -- Timestamps
    last_calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Ensure one ranking record per member
    UNIQUE(member_id)
);

-- Match history (event sourcing / audit trail)
CREATE TABLE match_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL,
    member_id UUID NOT NULL,

    -- Match result
    win BOOLEAN NOT NULL,
    final_position INTEGER NOT NULL,

    -- Match stats
    kills INTEGER DEFAULT 0,
    deaths INTEGER DEFAULT 0,

    -- Rating change
    rating_before INTEGER,
    rating_after INTEGER,
    rating_change INTEGER,

    -- Timestamps
    match_started_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT match_history_position_check CHECK (final_position >= 1)
);

-- Indexes for performance

-- Fast member lookup (unique constraints already provide these)
-- CREATE INDEX idx_player_match_stats_member ON player_match_stats(member_id); -- Not needed due to UNIQUE constraint
-- CREATE INDEX idx_player_ranking_stats_member ON player_ranking_stats(member_id); -- Not needed due to UNIQUE constraint

-- Leaderboard queries (sorted by rating descending)
CREATE INDEX idx_player_ranking_leaderboard ON player_ranking_stats(rating DESC, wins DESC);

-- Match history by member (recent matches first)
CREATE INDEX idx_match_history_member ON match_history(member_id, created_at DESC);

-- Match history by session
CREATE INDEX idx_match_history_session ON match_history(session_id);

-- Additional performance indexes
CREATE INDEX idx_player_ranking_username ON player_ranking_stats(username);
CREATE INDEX idx_player_ranking_rank_position ON player_ranking_stats(rank_position) WHERE rank_position IS NOT NULL;

-- Auto-update triggers
CREATE OR REPLACE FUNCTION update_stats_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER player_match_stats_updated_at
    BEFORE UPDATE ON player_match_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_stats_updated_at();

CREATE TRIGGER player_ranking_stats_updated_at
    BEFORE UPDATE ON player_ranking_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_stats_updated_at();


