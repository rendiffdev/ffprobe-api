-- Rollback migration: Revert user roles to original enum
-- This will convert pro and premium users back to regular users

-- Convert pro and premium users to regular users
UPDATE users SET role = 'user' WHERE role IN ('pro', 'premium');

-- Create old enum type
CREATE TYPE user_role_old AS ENUM ('admin', 'user', 'viewer', 'guest');

-- Alter the column to use the old enum
ALTER TABLE users 
    ALTER COLUMN role TYPE user_role_old 
    USING role::text::user_role_old;

-- Drop the current enum type
DROP TYPE user_role;

-- Rename the old enum back to the original name
ALTER TYPE user_role_old RENAME TO user_role;

-- Drop the index
DROP INDEX IF EXISTS idx_users_role;