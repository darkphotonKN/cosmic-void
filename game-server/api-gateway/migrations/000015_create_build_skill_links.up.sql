CREATE TABLE build_skill_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    build_id UUID NOT NULL REFERENCES builds(id) ON DELETE CASCADE,  
    name TEXT NOT NULL, -- Name of  (e.g., "Main DPS", "Mobility", "defensive")
    is_main BOOLEAN DEFAULT FALSE, -- Indicates the primary skill group for the build
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 