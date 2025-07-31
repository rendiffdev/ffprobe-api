-- Migration: Create video_comparisons table for storing video comparison results
-- Version: 20240131000003_create_video_comparisons.sql

-- Create comparison types enum
CREATE TYPE comparison_type AS ENUM (
    'quality',
    'encoding', 
    'compliance',
    'optimization',
    'full_analysis'
);

-- Create comparison status enum
CREATE TYPE comparison_status AS ENUM (
    'pending',
    'processing', 
    'completed',
    'failed'
);

-- Create quality verdict enum
CREATE TYPE quality_verdict AS ENUM (
    'significant_improvement',
    'improvement',
    'minimal_change',
    'regression',
    'significant_regression'
);

-- Create recommended action enum
CREATE TYPE recommended_action AS ENUM (
    'accept',
    'reject',
    'further_optimize',
    'review_manually',
    'fix_issues'
);

-- Create compliance status enum
CREATE TYPE compliance_status AS ENUM (
    'pass',
    'warning',
    'fail',
    'unknown'
);

-- Create video_comparisons table
CREATE TABLE IF NOT EXISTS video_comparisons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    original_analysis_id UUID NOT NULL REFERENCES analyses(id) ON DELETE CASCADE,
    modified_analysis_id UUID NOT NULL REFERENCES analyses(id) ON DELETE CASCADE,
    comparison_type comparison_type NOT NULL DEFAULT 'full_analysis',
    status comparison_status NOT NULL DEFAULT 'pending',
    comparison_data JSONB NOT NULL DEFAULT '{}',
    llm_assessment TEXT,
    quality_score JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    error_msg TEXT,
    
    -- Constraints
    CONSTRAINT different_analyses CHECK (original_analysis_id != modified_analysis_id),
    CONSTRAINT valid_quality_score CHECK (
        quality_score IS NULL OR (
            (quality_score->>'overall_score')::numeric BETWEEN 0 AND 100 AND
            (quality_score->>'video_score')::numeric BETWEEN 0 AND 100 AND
            (quality_score->>'audio_score')::numeric BETWEEN 0 AND 100 AND
            (quality_score->>'compression_score')::numeric BETWEEN 0 AND 100 AND
            (quality_score->>'compliance_score')::numeric BETWEEN 0 AND 100
        )
    )
);

-- Create indexes for better query performance
CREATE INDEX idx_video_comparisons_user_id ON video_comparisons(user_id);
CREATE INDEX idx_video_comparisons_original_analysis ON video_comparisons(original_analysis_id);
CREATE INDEX idx_video_comparisons_modified_analysis ON video_comparisons(modified_analysis_id);
CREATE INDEX idx_video_comparisons_status ON video_comparisons(status);
CREATE INDEX idx_video_comparisons_type ON video_comparisons(comparison_type);
CREATE INDEX idx_video_comparisons_created_at ON video_comparisons(created_at DESC);

-- Composite index for analysis pair lookups
CREATE INDEX idx_video_comparisons_analysis_pair ON video_comparisons(original_analysis_id, modified_analysis_id);

-- GIN index for JSONB comparison_data for advanced queries
CREATE INDEX idx_video_comparisons_data ON video_comparisons USING GIN(comparison_data);
CREATE INDEX idx_video_comparisons_quality_score ON video_comparisons USING GIN(quality_score);

-- Partial indexes for specific use cases
CREATE INDEX idx_video_comparisons_completed ON video_comparisons(created_at DESC) 
WHERE status = 'completed';

CREATE INDEX idx_video_comparisons_failed ON video_comparisons(created_at DESC) 
WHERE status = 'failed';

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_video_comparisons_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_video_comparisons_updated_at
    BEFORE UPDATE ON video_comparisons
    FOR EACH ROW
    EXECUTE FUNCTION update_video_comparisons_updated_at();

-- Add comments for documentation
COMMENT ON TABLE video_comparisons IS 'Stores video comparison results between two analyses';
COMMENT ON COLUMN video_comparisons.id IS 'Unique identifier for the comparison';
COMMENT ON COLUMN video_comparisons.user_id IS 'User who initiated the comparison (nullable for system comparisons)';
COMMENT ON COLUMN video_comparisons.original_analysis_id IS 'ID of the original analysis being compared from';
COMMENT ON COLUMN video_comparisons.modified_analysis_id IS 'ID of the modified analysis being compared to';
COMMENT ON COLUMN video_comparisons.comparison_type IS 'Type of comparison performed';
COMMENT ON COLUMN video_comparisons.status IS 'Current processing status of the comparison';
COMMENT ON COLUMN video_comparisons.comparison_data IS 'Detailed comparison results in JSON format';
COMMENT ON COLUMN video_comparisons.llm_assessment IS 'AI-generated assessment of the comparison results';
COMMENT ON COLUMN video_comparisons.quality_score IS 'Overall quality scores in JSON format';
COMMENT ON COLUMN video_comparisons.created_at IS 'Timestamp when comparison was created';
COMMENT ON COLUMN video_comparisons.updated_at IS 'Timestamp when comparison was last updated';
COMMENT ON COLUMN video_comparisons.error_msg IS 'Error message if comparison failed';

-- Create a view for comparison summaries (commonly used data)
CREATE VIEW comparison_summaries AS
SELECT 
    c.id,
    c.user_id,
    c.original_analysis_id,
    c.modified_analysis_id,
    c.comparison_type,
    c.status,
    c.created_at,
    c.updated_at,
    -- Extract summary data from JSON
    (c.comparison_data->>'overall_improvement')::numeric as overall_improvement,
    c.comparison_data->'summary'->>'quality_verdict' as quality_verdict,
    c.comparison_data->'summary'->>'recommended_action' as recommended_action,
    (c.comparison_data->'file_size'->>'percentage_change')::numeric as file_size_change_percent,
    (c.quality_score->>'overall_score')::numeric as overall_quality_score,
    -- Analysis metadata
    oa.file_name as original_file_name,
    oa.file_size as original_file_size,
    ma.file_name as modified_file_name,
    ma.file_size as modified_file_size
FROM video_comparisons c
LEFT JOIN analyses oa ON c.original_analysis_id = oa.id
LEFT JOIN analyses ma ON c.modified_analysis_id = ma.id;

COMMENT ON VIEW comparison_summaries IS 'Simplified view of comparison data for listing and summary purposes';

-- Create indexes on the view (PostgreSQL automatically creates these as indexes on the base table)
-- These are documented here for reference:
-- - id (primary key)
-- - user_id 
-- - status
-- - created_at
-- - comparison_type

-- Grant permissions (adjust as needed for your application)
-- GRANT SELECT, INSERT, UPDATE ON video_comparisons TO ffprobe_api_app;
-- GRANT SELECT ON comparison_summaries TO ffprobe_api_app;
-- GRANT USAGE ON TYPE comparison_type TO ffprobe_api_app;
-- GRANT USAGE ON TYPE comparison_status TO ffprobe_api_app;
-- GRANT USAGE ON TYPE quality_verdict TO ffprobe_api_app;
-- GRANT USAGE ON TYPE recommended_action TO ffprobe_api_app;
-- GRANT USAGE ON TYPE compliance_status TO ffprobe_api_app;