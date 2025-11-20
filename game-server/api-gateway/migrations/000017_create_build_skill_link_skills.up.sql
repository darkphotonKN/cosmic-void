CREATE TABLE IF NOT EXISTS build_skill_link_skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    build_skill_link_id UUID NOT NULL REFERENCES build_skill_links(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE
); 