CREATE TABLE IF NOT EXISTS build_tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    build_id UUID NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
    -- not repeatable
    UNIQUE(build_id, tag_id)
);
-- optimize index 
CREATE INDEX idx_build_tags_build_id ON build_tags(build_id);
CREATE INDEX idx_build_tags_tag_id ON build_tags(tag_id); 