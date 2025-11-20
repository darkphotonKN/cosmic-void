CREATE TABLE IF NOT EXISTS build_item_set_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    build_item_set_id UUID NOT NULL REFERENCES build_item_sets(id) ON DELETE CASCADE,
    item_id UUID REFERENCES items(id) ON DELETE CASCADE,
    slot TEXT
); 