-- Initial schema for ffprobe-api
-- This creates the foundational tables for the application

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create custom types
CREATE TYPE analysis_status AS ENUM ('pending', 'processing', 'completed', 'failed');
CREATE TYPE user_role AS ENUM ('admin', 'user', 'viewer', 'guest');
CREATE TYPE job_type AS ENUM ('ffprobe_analysis', 'quality_analysis', 'hls_analysis', 'report_generation', 'batch_processing');
CREATE TYPE job_status AS ENUM ('queued', 'running', 'completed', 'failed', 'cancelled');
CREATE TYPE quality_metric_type AS ENUM ('vmaf', 'psnr', 'ssim', 'ms_ssim', 'lpips');
CREATE TYPE hls_manifest_type AS ENUM ('master', 'media', 'variant');
CREATE TYPE comparison_type AS ENUM ('full', 'vmaf_only', 'psnr_only', 'custom');
CREATE TYPE cache_type AS ENUM ('ffprobe_result', 'quality_metrics', 'genai_response', 'report_data');
CREATE TYPE report_type AS ENUM ('analysis', 'quality_metrics', 'comparison', 'hls', 'batch');
CREATE TYPE report_format AS ENUM ('json', 'pdf', 'html', 'csv', 'xml', 'excel', 'markdown', 'text');

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    role user_role NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login TIMESTAMP WITH TIME ZONE
);

-- API Keys table
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(512) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]',
    rate_limit INTEGER NOT NULL DEFAULT 1000,
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Analyses table (partitioned by date)
CREATE TABLE analyses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    file_name VARCHAR(1000) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    content_hash VARCHAR(128) NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    status analysis_status NOT NULL DEFAULT 'pending',
    ffprobe_data JSONB,
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    error_msg TEXT
) PARTITION BY RANGE (created_at);

-- Create partitions for analyses table (current year and next year)
CREATE TABLE analyses_2024 PARTITION OF analyses
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

CREATE TABLE analyses_2025 PARTITION OF analyses
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

-- Quality metrics tables will be created in migration 005
-- with more comprehensive schema

-- HLS analyses table
CREATE TABLE hls_analyses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    analysis_id UUID NOT NULL REFERENCES analyses(id) ON DELETE CASCADE,
    manifest_path TEXT NOT NULL,
    manifest_type hls_manifest_type NOT NULL,
    manifest_data JSONB NOT NULL DEFAULT '{}',
    segment_count INTEGER NOT NULL DEFAULT 0,
    total_duration DOUBLE PRECISION NOT NULL DEFAULT 0,
    bitrate_variants INTEGER[] NOT NULL DEFAULT '{}',
    segment_duration DOUBLE PRECISION NOT NULL DEFAULT 0,
    playlist_version INTEGER NOT NULL DEFAULT 1,
    status analysis_status NOT NULL DEFAULT 'pending',
    processing_time DOUBLE PRECISION,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    error_msg TEXT
);

-- HLS segments table
CREATE TABLE hls_segments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hls_analysis_id UUID NOT NULL REFERENCES hls_analyses(id) ON DELETE CASCADE,
    segment_uri TEXT NOT NULL,
    sequence_number INTEGER NOT NULL,
    duration DOUBLE PRECISION NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    bitrate INTEGER NOT NULL DEFAULT 0,
    resolution VARCHAR(20),
    frame_rate DOUBLE PRECISION,
    segment_data JSONB NOT NULL DEFAULT '{}',
    quality_score DOUBLE PRECISION,
    status analysis_status NOT NULL DEFAULT 'pending',
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    error_msg TEXT
);

-- Quality comparisons table will be created in migration 005
-- with more comprehensive schema

-- Processing jobs table
CREATE TABLE processing_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    job_type job_type NOT NULL,
    status job_status NOT NULL DEFAULT 'queued',
    priority INTEGER NOT NULL DEFAULT 0,
    progress DOUBLE PRECISION NOT NULL DEFAULT 0,
    parameters JSONB NOT NULL DEFAULT '{}',
    result JSONB,
    error_msg TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Cache entries table
CREATE TABLE cache_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_hash VARCHAR(128) NOT NULL,
    cache_type cache_type NOT NULL,
    data JSONB NOT NULL DEFAULT '{}',
    hit_count INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_hit_at TIMESTAMP WITH TIME ZONE
);

-- Reports table
CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    analysis_id UUID NOT NULL REFERENCES analyses(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    report_type report_type NOT NULL,
    format report_format NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    download_count INTEGER NOT NULL DEFAULT 0,
    is_public BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_download TIMESTAMP WITH TIME ZONE
);

-- VMAF models table (for custom model management)
CREATE TABLE vmaf_models (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    version VARCHAR(100) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    model_type VARCHAR(50) NOT NULL DEFAULT 'standard',
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_analyses_user_id ON analyses(user_id);
CREATE INDEX idx_analyses_status ON analyses(status);
CREATE INDEX idx_analyses_content_hash ON analyses(content_hash);
CREATE INDEX idx_analyses_created_at ON analyses(created_at);

-- Quality metrics indexes will be created in migration 005

CREATE INDEX idx_hls_analyses_analysis_id ON hls_analyses(analysis_id);
CREATE INDEX idx_hls_segments_hls_analysis_id ON hls_segments(hls_analysis_id);

CREATE INDEX idx_processing_jobs_status ON processing_jobs(status);
CREATE INDEX idx_processing_jobs_type ON processing_jobs(job_type);
CREATE INDEX idx_processing_jobs_priority ON processing_jobs(priority DESC);

CREATE INDEX idx_cache_entries_hash_type ON cache_entries(content_hash, cache_type);
CREATE INDEX idx_cache_entries_expires_at ON cache_entries(expires_at);

CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);

CREATE INDEX idx_reports_analysis_id ON reports(analysis_id);
CREATE INDEX idx_reports_user_id ON reports(user_id);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_api_keys_updated_at BEFORE UPDATE ON api_keys
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_analyses_updated_at BEFORE UPDATE ON analyses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_processing_jobs_updated_at BEFORE UPDATE ON processing_jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vmaf_models_updated_at BEFORE UPDATE ON vmaf_models
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();