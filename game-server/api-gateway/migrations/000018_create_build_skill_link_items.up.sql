CREATE TABLE IF NOT EXISTS build_skill_link_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    build_skill_link_id UUID NOT NULL REFERENCES build_skill_links(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    slot TEXT
); 