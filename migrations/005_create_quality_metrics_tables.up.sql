-- Create quality metrics tables for VMAF, PSNR, SSIM analysis
-- This migration adds comprehensive quality analysis storage

BEGIN;

-- Create quality_metrics table for overall quality analysis results
CREATE TABLE quality_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    analysis_id UUID NOT NULL REFERENCES analyses(id) ON DELETE CASCADE,
    reference_file TEXT NOT NULL,
    distorted_file TEXT NOT NULL,
    metric_type VARCHAR(20) NOT NULL CHECK (metric_type IN ('vmaf', 'psnr', 'ssim', 'mse')),
    overall_score DECIMAL(10,6) NOT NULL,
    min_score DECIMAL(10,6) NOT NULL,
    max_score DECIMAL(10,6) NOT NULL,
    mean_score DECIMAL(10,6) NOT NULL,
    median_score DECIMAL(10,6),
    std_dev_score DECIMAL(10,6),
    percentile_1 DECIMAL(10,6),
    percentile_5 DECIMAL(10,6),
    percentile_10 DECIMAL(10,6),
    percentile_25 DECIMAL(10,6),
    percentile_75 DECIMAL(10,6),
    percentile_90 DECIMAL(10,6),
    percentile_95 DECIMAL(10,6),
    percentile_99 DECIMAL(10,6),
    frame_count INTEGER NOT NULL DEFAULT 0,
    duration DECIMAL(10,3),
    width INTEGER,
    height INTEGER,
    frame_rate DECIMAL(10,3),
    bit_rate BIGINT,
    configuration JSONB,
    processing_time INTERVAL NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Create quality_frames table for per-frame quality metrics
CREATE TABLE quality_frames (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quality_id UUID NOT NULL REFERENCES quality_metrics(id) ON DELETE CASCADE,
    frame_number INTEGER NOT NULL,
    timestamp DECIMAL(10,6) NOT NULL,
    score DECIMAL(10,6) NOT NULL,
    component_y DECIMAL(10,6),
    component_u DECIMAL(10,6),
    component_v DECIMAL(10,6),
    additional_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create quality_comparisons table for batch quality comparisons
CREATE TABLE quality_comparisons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id UUID NOT NULL,
    reference_file TEXT NOT NULL,
    distorted_file TEXT NOT NULL,
    metrics VARCHAR(20)[] NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    overall_rating VARCHAR(20) CHECK (overall_rating IN ('excellent', 'good', 'fair', 'poor', 'bad')),
    summary JSONB,
    visualization JSONB,
    processing_time INTERVAL,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Create quality_thresholds table for configurable quality thresholds
CREATE TABLE quality_thresholds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_type VARCHAR(20) NOT NULL CHECK (metric_type IN ('vmaf', 'psnr', 'ssim', 'mse')),
    excellent_threshold DECIMAL(10,6) NOT NULL,
    good_threshold DECIMAL(10,6) NOT NULL,
    fair_threshold DECIMAL(10,6) NOT NULL,
    poor_threshold DECIMAL(10,6) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create quality_issues table for detected quality issues
CREATE TABLE quality_issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quality_id UUID NOT NULL REFERENCES quality_metrics(id) ON DELETE CASCADE,
    issue_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('high', 'medium', 'low')),
    description TEXT NOT NULL,
    frame_range_start INTEGER,
    frame_range_end INTEGER,
    timestamp_start DECIMAL(10,6),
    timestamp_end DECIMAL(10,6),
    score DECIMAL(10,6),
    additional_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance optimization

-- Primary indexes for quality_metrics
CREATE INDEX idx_quality_metrics_analysis_id ON quality_metrics(analysis_id);
CREATE INDEX idx_quality_metrics_metric_type ON quality_metrics(metric_type);
CREATE INDEX idx_quality_metrics_status ON quality_metrics(status);
CREATE INDEX idx_quality_metrics_created_at ON quality_metrics(created_at);
CREATE INDEX idx_quality_metrics_score ON quality_metrics(overall_score);
CREATE INDEX idx_quality_metrics_files ON quality_metrics(reference_file, distorted_file);
CREATE INDEX idx_quality_metrics_composite ON quality_metrics(metric_type, status, created_at);

-- Indexes for quality_frames
CREATE INDEX idx_quality_frames_quality_id ON quality_frames(quality_id);
CREATE INDEX idx_quality_frames_frame_number ON quality_frames(frame_number);
CREATE INDEX idx_quality_frames_timestamp ON quality_frames(timestamp);
CREATE INDEX idx_quality_frames_score ON quality_frames(score);
CREATE INDEX idx_quality_frames_composite ON quality_frames(quality_id, frame_number);

-- Indexes for quality_comparisons
CREATE INDEX idx_quality_comparisons_batch_id ON quality_comparisons(batch_id);
CREATE INDEX idx_quality_comparisons_status ON quality_comparisons(status);
CREATE INDEX idx_quality_comparisons_created_at ON quality_comparisons(created_at);
CREATE INDEX idx_quality_comparisons_rating ON quality_comparisons(overall_rating);

-- Indexes for quality_thresholds
CREATE INDEX idx_quality_thresholds_metric_type ON quality_thresholds(metric_type);
CREATE INDEX idx_quality_thresholds_default ON quality_thresholds(is_default);

-- Indexes for quality_issues
CREATE INDEX idx_quality_issues_quality_id ON quality_issues(quality_id);
CREATE INDEX idx_quality_issues_type ON quality_issues(issue_type);
CREATE INDEX idx_quality_issues_severity ON quality_issues(severity);
CREATE INDEX idx_quality_issues_composite ON quality_issues(quality_id, severity);

-- JSONB indexes for configuration and additional data
CREATE INDEX idx_quality_metrics_configuration ON quality_metrics USING GIN (configuration);
CREATE INDEX idx_quality_frames_additional_data ON quality_frames USING GIN (additional_data);
CREATE INDEX idx_quality_comparisons_summary ON quality_comparisons USING GIN (summary);
CREATE INDEX idx_quality_comparisons_visualization ON quality_comparisons USING GIN (visualization);
CREATE INDEX idx_quality_issues_additional_data ON quality_issues USING GIN (additional_data);

-- Create constraints for data integrity
ALTER TABLE quality_metrics ADD CONSTRAINT check_quality_scores 
    CHECK (min_score <= max_score AND min_score <= mean_score AND mean_score <= max_score);

ALTER TABLE quality_frames ADD CONSTRAINT check_frame_timestamp 
    CHECK (timestamp >= 0 AND frame_number >= 0);

ALTER TABLE quality_thresholds ADD CONSTRAINT check_threshold_order 
    CHECK (poor_threshold <= fair_threshold AND fair_threshold <= good_threshold AND good_threshold <= excellent_threshold);

-- Create unique constraints to prevent duplicate entries
ALTER TABLE quality_metrics ADD CONSTRAINT unique_quality_analysis 
    UNIQUE (analysis_id, metric_type, reference_file, distorted_file);

ALTER TABLE quality_frames ADD CONSTRAINT unique_quality_frame 
    UNIQUE (quality_id, frame_number);

ALTER TABLE quality_thresholds ADD CONSTRAINT unique_metric_default 
    UNIQUE (metric_type, is_default) DEFERRABLE INITIALLY DEFERRED;

-- Create trigger for updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_quality_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_quality_metrics_updated_at
    BEFORE UPDATE ON quality_metrics
    FOR EACH ROW
    EXECUTE FUNCTION update_quality_updated_at();

CREATE TRIGGER trigger_quality_comparisons_updated_at
    BEFORE UPDATE ON quality_comparisons
    FOR EACH ROW
    EXECUTE FUNCTION update_quality_updated_at();

CREATE TRIGGER trigger_quality_thresholds_updated_at
    BEFORE UPDATE ON quality_thresholds
    FOR EACH ROW
    EXECUTE FUNCTION update_quality_updated_at();

-- Insert default quality thresholds
INSERT INTO quality_thresholds (metric_type, excellent_threshold, good_threshold, fair_threshold, poor_threshold, is_default) VALUES
    ('vmaf', 95.0, 85.0, 75.0, 60.0, true),
    ('psnr', 40.0, 35.0, 30.0, 25.0, true),
    ('ssim', 0.95, 0.90, 0.85, 0.80, true),
    ('mse', 100.0, 300.0, 500.0, 1000.0, true);

-- Add table comments for documentation
COMMENT ON TABLE quality_metrics IS 'Stores overall quality analysis results for video quality metrics (VMAF, PSNR, SSIM)';
COMMENT ON TABLE quality_frames IS 'Stores per-frame quality metrics for detailed analysis';
COMMENT ON TABLE quality_comparisons IS 'Stores batch quality comparison results';
COMMENT ON TABLE quality_thresholds IS 'Configurable quality thresholds for different metrics';
COMMENT ON TABLE quality_issues IS 'Detected quality issues and problems';

-- Add column comments for key fields
COMMENT ON COLUMN quality_metrics.metric_type IS 'Type of quality metric: vmaf, psnr, ssim, mse';
COMMENT ON COLUMN quality_metrics.overall_score IS 'Overall quality score for the entire video';
COMMENT ON COLUMN quality_metrics.configuration IS 'JSONB configuration used for quality analysis';
COMMENT ON COLUMN quality_frames.score IS 'Quality score for this specific frame';
COMMENT ON COLUMN quality_frames.additional_data IS 'Additional per-frame data in JSONB format';
COMMENT ON COLUMN quality_comparisons.metrics IS 'Array of metrics used in comparison';
COMMENT ON COLUMN quality_comparisons.summary IS 'JSONB summary of comparison results';
COMMENT ON COLUMN quality_issues.issue_type IS 'Type of quality issue detected';
COMMENT ON COLUMN quality_issues.severity IS 'Severity level: high, medium, low';

COMMIT;