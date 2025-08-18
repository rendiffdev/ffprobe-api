-- Rollback migration: Remove secret rotation tables
-- Version: 007

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_update_rotation_due ON api_keys;
DROP TRIGGER IF EXISTS trigger_log_api_key_action ON api_keys;

-- Drop functions
DROP FUNCTION IF EXISTS update_rotation_due();
DROP FUNCTION IF EXISTS log_api_key_action();
DROP FUNCTION IF EXISTS cleanup_expired_keys();

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS rate_limit_usage;
DROP TABLE IF EXISTS user_rate_limits;
DROP TABLE IF EXISTS tenant_rate_limits;
DROP TABLE IF EXISTS api_key_rotation_log;
DROP TABLE IF EXISTS jwt_secrets;
DROP TABLE IF EXISTS api_keys;