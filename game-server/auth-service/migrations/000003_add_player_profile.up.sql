
-- Player profile (progression data)
CREATE TABLE player_profile (
    member_id UUID PRIMARY KEY REFERENCES members(id) ON DELETE CASCADE,
    
    -- Progression
    level INTEGER DEFAULT 1 NOT NULL,
    current_xp INTEGER DEFAULT 0 NOT NULL,
    xp_to_next_level INTEGER DEFAULT 100 NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast lookups
CREATE INDEX idx_player_profile_member ON player_profile(member_id);

-- Trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_player_profile_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER player_profile_updated_at
    BEFORE UPDATE ON player_profile
    FOR EACH ROW
    EXECUTE FUNCTION update_player_profile_updated_at();
