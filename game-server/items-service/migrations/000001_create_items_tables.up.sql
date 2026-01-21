CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- item_types table
CREATE TABLE item_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type_code VARCHAR(50) UNIQUE NOT NULL,  -- 'weapon', 'armor', 'consumable'
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0 NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,

    CONSTRAINT non_negative_sort_order CHECK (sort_order >= 0)
);

-- item_rarities table
CREATE TABLE item_rarities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rarity_code VARCHAR(50) UNIQUE NOT NULL,  -- 'common', 'rare', 'epic', 'legendary'
    rarity_name VARCHAR(100) NOT NULL,
    color_hex VARCHAR(7),
    drop_rate_multiplier DECIMAL(5,2) NOT NULL,
    sort_order INTEGER DEFAULT 0 NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,

    CONSTRAINT valid_color_hex CHECK (color_hex ~ '^#[0-9A-Fa-f]{6}$') NOT VALID,
    CONSTRAINT non_negative_sort_order CHECK (sort_order >= 0),
    CONSTRAINT positive_drop_rate_multiplier CHECK (drop_rate_multiplier > 0 AND drop_rate_multiplier <= 1)
);

-- weapons table
CREATE TABLE weapons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type_id UUID NOT NULL REFERENCES item_types(id),
    rarity_id UUID NOT NULL REFERENCES item_rarities(id),

    -- 武器專屬屬性
    attack_power INTEGER NOT NULL,
    durability INTEGER NOT NULL,
    critical_rate DECIMAL(5,2) DEFAULT 0,
    weapon_type VARCHAR(50),              -- 'sword', 'axe', 'bow'
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    
    CONSTRAINT weapon_attack_positive CHECK (attack_power >= 0),
    CONSTRAINT weapon_durability_positive CHECK (durability >= 0)
);

-- armors table
CREATE TABLE armors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type_id UUID NOT NULL REFERENCES item_types(id),
    rarity_id UUID NOT NULL REFERENCES item_rarities(id),
    
    -- 防具專屬屬性
    defense_rating INTEGER NOT NULL,
    durability INTEGER NOT NULL,
    magic_resistance INTEGER DEFAULT 0,
    armor_slot VARCHAR(50),               -- 'head', 'chest', 'legs', 'shield'
    description TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    
    CONSTRAINT armor_defense_positive CHECK (defense_rating >= 0),
    CONSTRAINT armor_durability_positive CHECK (durability >= 0)
);

-- consumables table
CREATE TABLE consumables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type_id UUID NOT NULL REFERENCES item_types(id),
    rarity_id UUID NOT NULL REFERENCES item_rarities(id),
    
    -- 消耗品專屬屬性
    healing_amount INTEGER DEFAULT 0,
    mana_amount INTEGER DEFAULT 0,
    buff_duration INTEGER DEFAULT 0,      -- 秒數
    max_stack_size INTEGER DEFAULT 99,
    description TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    
    CONSTRAINT consumable_healing_positive CHECK (healing_amount >= 0),
    CONSTRAINT consumable_stack_positive CHECK (max_stack_size > 0)
);

-- item_templates table
CREATE TABLE item_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    item_name VARCHAR(200) NOT NULL,
    item_code VARCHAR(100) UNIQUE NOT NULL,
    
    type_id UUID NOT NULL REFERENCES item_types(id),
    rarity_id UUID NOT NULL REFERENCES item_rarities(id),
    
    item_type VARCHAR(50) NOT NULL,           -- 'weapon', 'armor', 'consumable'
    item_id UUID NOT NULL,                    -- item ID
    
    icon_url VARCHAR(255),
    is_tradeable BOOLEAN DEFAULT TRUE,
    is_droppable BOOLEAN DEFAULT TRUE,
    required_level INTEGER DEFAULT 1,
    base_sell_price INTEGER DEFAULT 0,
    base_buy_price INTEGER DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    


    CONSTRAINT valid_item_type 
        CHECK (item_type IN ('weapon', 'armor', 'consumable')),
    
    -- (item_type + item_id)
    UNIQUE(item_type, item_id)
);

-- item_elements table

-- status_effects table

-- item_template_attributes table

-- item_template_effects table

-- Create Indexes
CREATE INDEX idx_item_types_active ON item_types(is_active);
CREATE INDEX idx_item_types_sort ON item_types(sort_order);

CREATE INDEX idx_item_rarities_sort ON item_rarities(sort_order);

CREATE INDEX idx_weapons_type ON weapons(type_id);
CREATE INDEX idx_weapons_rarity ON weapons(rarity_id);

CREATE INDEX idx_armors_type ON armors(type_id);
CREATE INDEX idx_armors_rarity ON armors(rarity_id);

CREATE INDEX idx_consumables_type ON consumables(type_id);
CREATE INDEX idx_consumables_rarity ON consumables(rarity_id);

CREATE INDEX idx_item_templates_type ON item_templates(type_id);
CREATE INDEX idx_item_templates_rarity ON item_templates(rarity_id);
CREATE INDEX idx_item_templates_code ON item_templates(item_code);
CREATE INDEX idx_item_templates_itemable ON item_templates(item_type, item_id);

-- Trigger function to update updated_at column
CREATE or REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;

    BEGIN
      NEW.updated_by = current_setting('app.current_user_id', true)::UUID;
    EXCEPTION WHEN OTHERS THEN
      NEW.updated_by = NULL;
    END;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- auto-update triggers
CREATE TRIGGER item_rarities_updated_at
    BEFORE UPDATE ON item_rarities
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER weapons_updated_at
    BEFORE UPDATE ON weapons
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER armors_updated_at
    BEFORE UPDATE ON armors
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER consumables_updated_at
    BEFORE UPDATE ON consumables
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER item_templates_updated_at
    BEFORE UPDATE ON item_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();