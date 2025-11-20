CREATE TABLE IF NOT EXISTS base_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category TEXT  NOT NULL,
    class TEXT  NOT NULL,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    equip_type TEXT NOT NULL,
    is_two_hands BOOLEAN NOT NULL,
    slot TEXT NOT NULL,

    image_url TEXT, -- Path or URL for the item's image
    
    required_level TEXT,
    required_strength TEXT,
    required_dexterity TEXT,
    required_intelligence TEXT,
    armour TEXT,
    energy_shield TEXT,
    evasion TEXT,
    ward TEXT,

    damage TEXT,
    aps TEXT,
    crit TEXT,
    dps TEXT,

    implicit TEXT[],

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 