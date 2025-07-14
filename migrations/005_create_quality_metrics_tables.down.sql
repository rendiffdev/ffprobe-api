-- Drop quality metrics tables and related objects

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_quality_metrics_updated_at ON quality_metrics;
DROP TRIGGER IF EXISTS trigger_quality_comparisons_updated_at ON quality_comparisons;
DROP TRIGGER IF EXISTS trigger_quality_thresholds_updated_at ON quality_thresholds;

-- Drop function
DROP FUNCTION IF EXISTS update_quality_updated_at();

-- Drop indexes (will be dropped automatically with tables, but explicit for clarity)
DROP INDEX IF EXISTS idx_quality_metrics_analysis_id;
DROP INDEX IF EXISTS idx_quality_metrics_metric_type;
DROP INDEX IF EXISTS idx_quality_metrics_status;
DROP INDEX IF EXISTS idx_quality_metrics_created_at;
DROP INDEX IF EXISTS idx_quality_metrics_score;
DROP INDEX IF EXISTS idx_quality_metrics_files;
DROP INDEX IF EXISTS idx_quality_metrics_composite;
DROP INDEX IF EXISTS idx_quality_metrics_configuration;

DROP INDEX IF EXISTS idx_quality_frames_quality_id;
DROP INDEX IF EXISTS idx_quality_frames_frame_number;
DROP INDEX IF EXISTS idx_quality_frames_timestamp;
DROP INDEX IF EXISTS idx_quality_frames_score;
DROP INDEX IF EXISTS idx_quality_frames_composite;
DROP INDEX IF EXISTS idx_quality_frames_additional_data;

DROP INDEX IF EXISTS idx_quality_comparisons_batch_id;
DROP INDEX IF EXISTS idx_quality_comparisons_status;
DROP INDEX IF EXISTS idx_quality_comparisons_created_at;
DROP INDEX IF EXISTS idx_quality_comparisons_rating;
DROP INDEX IF EXISTS idx_quality_comparisons_summary;
DROP INDEX IF EXISTS idx_quality_comparisons_visualization;

DROP INDEX IF EXISTS idx_quality_thresholds_metric_type;
DROP INDEX IF EXISTS idx_quality_thresholds_default;

DROP INDEX IF EXISTS idx_quality_issues_quality_id;
DROP INDEX IF EXISTS idx_quality_issues_type;
DROP INDEX IF EXISTS idx_quality_issues_severity;
DROP INDEX IF EXISTS idx_quality_issues_composite;
DROP INDEX IF EXISTS idx_quality_issues_additional_data;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS quality_issues;
DROP TABLE IF EXISTS quality_frames;
DROP TABLE IF EXISTS quality_comparisons;
DROP TABLE IF EXISTS quality_thresholds;
DROP TABLE IF EXISTS quality_metrics;

COMMIT;