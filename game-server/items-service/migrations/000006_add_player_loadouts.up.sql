-- Migration 006: Add player_loadouts table

CREATE TABLE player_loadouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id UUID NOT NULL UNIQUE,
    weapon_instance_id UUID REFERENCES item_instances(id) ON DELETE SET NULL,
    armor_instance_id UUID REFERENCES item_instances(id) ON DELETE SET NULL,
    consumable_1_id UUID REFERENCES item_instances(id) ON DELETE SET NULL,
    consumable_2_id UUID REFERENCES item_instances(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on member_id for fast lookups
CREATE INDEX idx_player_loadouts_member ON player_loadouts(member_id);

-- Add trigger for updated_at
CREATE TRIGGER player_loadouts_updated_at
    BEFORE UPDATE ON player_loadouts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();