-- Create members table
CREATE TABLE IF NOT EXISTS members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- User Specific --
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1 CHECK (status IN (1, 2)), -- account status 1: Member, 2: Author
    
    -- extra --
    average_rating DECIMAL(2, 1) DEFAULT 0 CHECK (average_rating >= 0 AND average_rating <= 5)
);

-- Create function to auto-update the updated_at column
CREATE OR REPLACE FUNCTION update_members_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for the members table
CREATE TRIGGER set_members_updated_at
BEFORE UPDATE ON members
FOR EACH ROW
EXECUTE FUNCTION update_members_updated_at(); 