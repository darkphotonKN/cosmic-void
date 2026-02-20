-- Add role column to members table
ALTER TABLE members ADD COLUMN role VARCHAR(50) NOT NULL DEFAULT 'player';

-- Create index on role for faster queries
CREATE INDEX idx_members_role ON members(role);

-- Add comment to explain role values
COMMENT ON COLUMN members.role IS 'User role: player, admin';

-- Update the first member to be admin (if exists)
-- Or create a default admin account
DO $$
BEGIN
    -- Check if there are any members
    IF EXISTS (SELECT 1 FROM members LIMIT 1) THEN
        -- Update the first member to admin
        UPDATE members
        SET role = 'admin'
        WHERE id = (SELECT id FROM members ORDER BY created_at ASC LIMIT 1);
    ELSE
        -- Create default admin account
        -- Password is '123456' hashed with bcrypt (cost 10)
        INSERT INTO members (id, email, name, password, status, role, created_at, updated_at)
        VALUES (
            gen_random_uuid(),
            'admin@cosmicvoid.com',
            'Admin',
            '$2a$10$.2AEer5Qwhhxq0XkYndjkO6NhsPuG4KvYDsx8EHspoUWXPuo/LD2e',
            '1',
            'admin',
            NOW(),
            NOW()
        );
    END IF;
END $$;
