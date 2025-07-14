-- Rollback initial schema

-- Drop triggers
DROP TRIGGER IF EXISTS update_vmaf_models_updated_at ON vmaf_models;
DROP TRIGGER IF EXISTS update_processing_jobs_updated_at ON processing_jobs;
DROP TRIGGER IF EXISTS update_analyses_updated_at ON analyses;
DROP TRIGGER IF EXISTS update_api_keys_updated_at ON api_keys;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes (most will be dropped automatically with tables)
DROP INDEX IF EXISTS idx_reports_user_id;
DROP INDEX IF EXISTS idx_reports_analysis_id;
DROP INDEX IF EXISTS idx_api_keys_user_id;
DROP INDEX IF EXISTS idx_api_keys_hash;
DROP INDEX IF EXISTS idx_cache_entries_expires_at;
DROP INDEX IF EXISTS idx_cache_entries_hash_type;
DROP INDEX IF EXISTS idx_processing_jobs_priority;
DROP INDEX IF EXISTS idx_processing_jobs_type;
DROP INDEX IF EXISTS idx_processing_jobs_status;
DROP INDEX IF EXISTS idx_hls_segments_hls_analysis_id;
DROP INDEX IF EXISTS idx_hls_analyses_analysis_id;
DROP INDEX IF EXISTS idx_quality_frames_frame_number;
DROP INDEX IF EXISTS idx_quality_frames_metric_id;
DROP INDEX IF EXISTS idx_quality_metrics_type;
DROP INDEX IF EXISTS idx_quality_metrics_analysis_id;
DROP INDEX IF EXISTS idx_analyses_created_at;
DROP INDEX IF EXISTS idx_analyses_content_hash;
DROP INDEX IF EXISTS idx_analyses_status;
DROP INDEX IF EXISTS idx_analyses_user_id;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS vmaf_models;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS cache_entries;
DROP TABLE IF EXISTS processing_jobs;
DROP TABLE IF EXISTS quality_comparisons;
DROP TABLE IF EXISTS hls_segments;
DROP TABLE IF EXISTS hls_analyses;
DROP TABLE IF EXISTS quality_frames;
DROP TABLE IF EXISTS quality_metrics;

-- Drop analysis partitions
DROP TABLE IF EXISTS analyses_2025;
DROP TABLE IF EXISTS analyses_2024;
DROP TABLE IF EXISTS analyses;

DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS users;

-- Drop custom types
DROP TYPE IF EXISTS report_format;
DROP TYPE IF EXISTS report_type;
DROP TYPE IF EXISTS cache_type;
DROP TYPE IF EXISTS comparison_type;
DROP TYPE IF EXISTS hls_manifest_type;
DROP TYPE IF EXISTS quality_metric_type;
DROP TYPE IF EXISTS job_status;
DROP TYPE IF EXISTS job_type;
DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS analysis_status;

-- Drop extensions (only if not used by other databases)
-- DROP EXTENSION IF EXISTS "uuid-ossp";