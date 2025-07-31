-- Add authentication fields to users table
-- This migration adds fields required for proper authentication and security

-- Add password_hash column for storing bcrypt hashed passwords
ALTER TABLE users ADD COLUMN password_hash VARCHAR(512);

-- Add status column for account status management
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended', 'pending');
ALTER TABLE users ADD COLUMN status user_status NOT NULL DEFAULT 'active';

-- Add failed_logins counter for account lockout functionality
ALTER TABLE users ADD COLUMN failed_logins INTEGER NOT NULL DEFAULT 0;

-- Add locked_until timestamp for temporary account locks
ALTER TABLE users ADD COLUMN locked_until TIMESTAMP WITH TIME ZONE;

-- Add deleted_at for soft delete functionality
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- Add last_login_at to track successful logins (rename existing last_login)
ALTER TABLE users RENAME COLUMN last_login TO last_login_at;

-- Create indexes for performance
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_locked_until ON users(locked_until) WHERE locked_until IS NOT NULL;

-- Update existing users to have active status and no password initially
-- In production, you would need to handle password migration separately
UPDATE users SET status = 'active' WHERE status IS NULL;

-- Add constraint to ensure password_hash is not null for active users
-- This constraint can be enabled after passwords are set
-- ALTER TABLE users ADD CONSTRAINT users_password_required 
--   CHECK (status != 'active' OR password_hash IS NOT NULL);