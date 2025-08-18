-- FFprobe API SQLite Schema Migration
-- SQLite-compatible version of the core schema

-- Users table
CREATE TABLE users (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT DEFAULT 'user' CHECK (role IN ('admin', 'user', 'viewer', 'guest')),
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME
);

-- Analyses table
CREATE TABLE analyses (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    file_name TEXT NOT NULL,
    file_path TEXT,
    file_size INTEGER,
    content_hash TEXT,
    source_type TEXT DEFAULT 'local',
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    ffprobe_data TEXT, -- JSON stored as TEXT in SQLite
    llm_report TEXT,
    processed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    error_msg TEXT
);

-- Streams table
CREATE TABLE streams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    analysis_id TEXT REFERENCES analyses(id) ON DELETE CASCADE,
    stream_index INTEGER NOT NULL,
    codec_type TEXT,
    codec_name TEXT,
    codec_long_name TEXT,
    bit_rate INTEGER,
    width INTEGER,
    height INTEGER,
    sample_rate INTEGER,
    channels INTEGER,
    duration REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Quality metrics table
CREATE TABLE quality_metrics (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    analysis_id TEXT REFERENCES analyses(id) ON DELETE CASCADE,
    metric_type TEXT NOT NULL,
    metric_value REAL,
    frame_number INTEGER,
    timestamp_ms INTEGER,
    model_version TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Comparisons table
CREATE TABLE comparisons (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    source_analysis_id TEXT REFERENCES analyses(id) ON DELETE CASCADE,
    target_analysis_id TEXT REFERENCES analyses(id) ON DELETE CASCADE,
    comparison_type TEXT DEFAULT 'quality',
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    result_data TEXT, -- JSON stored as TEXT
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Reports table
CREATE TABLE reports (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    analysis_id TEXT REFERENCES analyses(id) ON DELETE CASCADE,
    report_type TEXT NOT NULL,
    file_path TEXT,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    download_count INTEGER DEFAULT 0
);

-- Batch processing table
CREATE TABLE batches (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    name TEXT,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    total_files INTEGER DEFAULT 0,
    processed_files INTEGER DEFAULT 0,
    failed_files INTEGER DEFAULT 0,
    configuration TEXT, -- JSON stored as TEXT
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- API Keys table
CREATE TABLE api_keys (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    tenant_id TEXT,
    key_hash TEXT NOT NULL,
    key_prefix TEXT NOT NULL,
    name TEXT,
    permissions TEXT, -- JSON array stored as TEXT
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'expired', 'revoked', 'suspended')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    last_used_at DATETIME,
    last_rotated DATETIME DEFAULT CURRENT_TIMESTAMP,
    rotation_due DATETIME,
    usage_count INTEGER DEFAULT 0
);

-- HLS Analysis table
CREATE TABLE hls_analyses (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    analysis_id TEXT REFERENCES analyses(id) ON DELETE CASCADE,
    manifest_url TEXT NOT NULL,
    variant_count INTEGER,
    total_segments INTEGER,
    average_segment_duration REAL,
    bandwidth_range_min INTEGER,
    bandwidth_range_max INTEGER,
    resolution_profiles TEXT, -- JSON stored as TEXT
    codec_profiles TEXT, -- JSON stored as TEXT
    validation_errors TEXT, -- JSON stored as TEXT
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_analyses_user_id ON analyses(user_id);
CREATE INDEX idx_analyses_status ON analyses(status);
CREATE INDEX idx_analyses_created_at ON analyses(created_at);
CREATE INDEX idx_analyses_source_type ON analyses(source_type);
CREATE INDEX idx_analyses_content_hash ON analyses(content_hash);

CREATE INDEX idx_streams_analysis_id ON streams(analysis_id);
CREATE INDEX idx_streams_codec_type ON streams(codec_type);

CREATE INDEX idx_quality_metrics_analysis_id ON quality_metrics(analysis_id);
CREATE INDEX idx_quality_metrics_metric_type ON quality_metrics(metric_type);

CREATE INDEX idx_comparisons_source_analysis_id ON comparisons(source_analysis_id);
CREATE INDEX idx_comparisons_target_analysis_id ON comparisons(target_analysis_id);
CREATE INDEX idx_comparisons_status ON comparisons(status);

CREATE INDEX idx_reports_analysis_id ON reports(analysis_id);
CREATE INDEX idx_reports_status ON reports(status);

CREATE INDEX idx_batches_user_id ON batches(user_id);
CREATE INDEX idx_batches_status ON batches(status);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_status ON api_keys(status);

CREATE INDEX idx_hls_analyses_analysis_id ON hls_analyses(analysis_id);

-- SQLite triggers for updated_at columns
CREATE TRIGGER update_users_updated_at 
    AFTER UPDATE ON users
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_analyses_updated_at 
    AFTER UPDATE ON analyses
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE analyses SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_comparisons_updated_at 
    AFTER UPDATE ON comparisons
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE comparisons SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_batches_updated_at 
    AFTER UPDATE ON batches
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE batches SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;