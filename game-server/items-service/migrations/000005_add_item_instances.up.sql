-- Migration 005: Add item_instances table

CREATE TABLE item_instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id UUID NOT NULL REFERENCES item_templates(id),
    owner_member_id UUID NOT NULL,
    source VARCHAR(50) NOT NULL CHECK (source IN ('extracted', 'starting_gear', 'reward')),

    -- Rolled/actual stats (copied from base, randomized in future)
    item_type VARCHAR(20) NOT NULL CHECK (item_type IN ('weapon', 'armor', 'consumable')),
    name VARCHAR(200) NOT NULL,
    rarity_id UUID REFERENCES item_rarities(id),

    -- Weapon stats (null if not weapon)
    attack_power INTEGER,
    critical_rate DECIMAL(5,2),
    weapon_type VARCHAR(50),

    -- Armor stats (null if not armor)
    defense_rating INTEGER,
    magic_resistance INTEGER,
    armor_slot VARCHAR(50),

    -- Consumable stats (null if not consumable)
    healing_amount INTEGER,
    mana_amount INTEGER,
    buff_duration INTEGER,

    -- Shared stats
    durability INTEGER,

    -- Pricing (at time of acquisition)
    buy_price INTEGER,
    sell_price INTEGER,

    acquired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_item_instances_owner ON item_instances(owner_member_id);
CREATE INDEX idx_item_instances_template ON item_instances(template_id);

-- Add trigger for updated_at
CREATE TRIGGER item_instances_updated_at
    BEFORE UPDATE ON item_instances
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
