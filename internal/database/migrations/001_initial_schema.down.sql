-- Rollback script for initial schema migration
-- Drop all tables and objects created in 001_initial_schema.up.sql

-- Drop triggers first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_analyses_updated_at ON analyses;
DROP TRIGGER IF EXISTS update_comparisons_updated_at ON comparisons;
DROP TRIGGER IF EXISTS update_batches_updated_at ON batches;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS hls_analyses;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS batches;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS comparisons;
DROP TABLE IF EXISTS quality_metrics;
DROP TABLE IF EXISTS streams;
DROP TABLE IF EXISTS analyses;
DROP TABLE IF EXISTS users;

-- Note: UUID extension is kept as it might be used by other applications
-- DROP EXTENSION IF EXISTS "uuid-ossp";