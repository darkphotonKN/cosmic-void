CREATE TABLE IF NOT EXISTS ascendancies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE, -- FK to classes table
    name TEXT NOT NULL UNIQUE,
    description TEXT, -- Description for the ascendancy
    image_url TEXT, -- URL for ascendancy image
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 