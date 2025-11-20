CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id UUID REFERENCES members(id) ON DELETE RESTRICT,
    base_item_id UUID REFERENCES base_items(id) ON DELETE RESTRICT,
    category TEXT  NOT NULL,
    class TEXT  NOT NULL,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    unique_item BOOLEAN NOT NULL,
    slot TEXT NOT NULL,
    description TEXT, -- equip's story or description
    image_url TEXT, -- Path or URL for the item's image
    
    required_level TEXT,
    required_strength TEXT,
    required_dexterity TEXT,
    required_intelligence TEXT,
    armour TEXT,
    block TEXT,
    energy_shield TEXT,
    evasion TEXT,
    ward TEXT,

    damage TEXT,
    aps TEXT,
    crit TEXT,
    pdps TEXT,
    edps TEXT,
    dps TEXT,

    life TEXT,
    mana TEXT,
    duration TEXT,
    usage TEXT,
    capacity TEXT,

    additional TEXT,
    stats TEXT[],
    implicit TEXT[],

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 