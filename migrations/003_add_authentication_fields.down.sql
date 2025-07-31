-- Remove authentication fields from users table
-- This reverses the authentication fields migration

-- Drop indexes
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_locked_until;

-- Remove constraint if it exists
-- ALTER TABLE users DROP CONSTRAINT IF EXISTS users_password_required;

-- Rename last_login_at back to last_login
ALTER TABLE users RENAME COLUMN last_login_at TO last_login;

-- Remove authentication columns
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS locked_until;
ALTER TABLE users DROP COLUMN IF EXISTS failed_logins;
ALTER TABLE users DROP COLUMN IF EXISTS status;
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;

-- Drop the user_status type
DROP TYPE IF EXISTS user_status;