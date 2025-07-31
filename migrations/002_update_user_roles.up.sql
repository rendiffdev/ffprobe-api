-- Update user roles to match middleware and documentation expectations
-- This migration updates the user_role enum to include pro and premium roles
-- and removes viewer and guest roles which are not used in the application

-- First, we need to handle existing data
-- Convert any 'viewer' to 'user' and 'guest' to 'user'
UPDATE users SET role = 'user' WHERE role IN ('viewer', 'guest');

-- Drop the old enum type and create new one
-- PostgreSQL doesn't allow direct modification of enums, so we need to:
-- 1. Create a new enum type
-- 2. Update the column to use the new type
-- 3. Drop the old enum type

-- Create new enum type with correct roles
CREATE TYPE user_role_new AS ENUM ('admin', 'user', 'pro', 'premium');

-- Alter the column to use the new enum
ALTER TABLE users 
    ALTER COLUMN role TYPE user_role_new 
    USING role::text::user_role_new;

-- Drop the old enum type
DROP TYPE user_role;

-- Rename the new enum to the original name
ALTER TYPE user_role_new RENAME TO user_role;

-- Add indexes for better performance on role-based queries
CREATE INDEX idx_users_role ON users(role);

-- Add comment for documentation
COMMENT ON TYPE user_role IS 'User roles: admin (full access), user (basic access), pro (enhanced features), premium (highest tier)';