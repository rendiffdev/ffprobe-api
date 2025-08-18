-- Rollback script for SQLite schema migration

-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at;
DROP TRIGGER IF EXISTS update_analyses_updated_at;
DROP TRIGGER IF EXISTS update_comparisons_updated_at;
DROP TRIGGER IF EXISTS update_batches_updated_at;

-- Drop tables in reverse order
DROP TABLE IF EXISTS hls_analyses;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS batches;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS comparisons;
DROP TABLE IF EXISTS quality_metrics;
DROP TABLE IF EXISTS streams;
DROP TABLE IF EXISTS analyses;
DROP TABLE IF EXISTS users;