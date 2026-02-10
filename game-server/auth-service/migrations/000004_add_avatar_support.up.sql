-- Add avatar_url to members table
ALTER TABLE members ADD COLUMN avatar_url TEXT;

-- Create avatar_uploads tracking table
CREATE TABLE avatar_uploads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    s3_key TEXT NOT NULL,
    upload_status VARCHAR(20) DEFAULT 'pending' CHECK (upload_status IN ('pending', 'uploaded', 'synced', 'failed')),
    file_size BIGINT,
    content_type TEXT,
    presigned_url_expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_avatar_uploads_member_id ON avatar_uploads(member_id);
CREATE INDEX idx_avatar_uploads_status ON avatar_uploads(upload_status);

-- Update trigger for avatar_uploads
CREATE TRIGGER avatar_uploads_updated_at
BEFORE UPDATE ON avatar_uploads
FOR EACH ROW
EXECUTE FUNCTION update_members_updated_at();
