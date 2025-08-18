-- Migration: Add secret rotation tables for API keys and JWT secrets
-- Version: 007
-- Description: Implements secure credential rotation with per-user/tenant rate limiting

-- API Keys table with rotation support
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL,
    key_hash TEXT NOT NULL, -- Bcrypt hash of the actual key
    key_prefix VARCHAR(16) NOT NULL, -- First 8 chars for quick lookup
    name VARCHAR(255) NOT NULL,
    permissions TEXT[] DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, rotating, expired, revoked
    
    -- Rate limiting per key
    rate_limit_rpm INTEGER DEFAULT 60,
    rate_limit_rph INTEGER DEFAULT 1000,
    rate_limit_rpd INTEGER DEFAULT 10000,
    
    -- Tracking
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE,
    last_rotated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    rotation_due TIMESTAMP WITH TIME ZONE NOT NULL,
    usage_count BIGINT DEFAULT 0,
    
    -- Metadata
    created_by UUID REFERENCES users(id),
    ip_whitelist INET[] DEFAULT '{}',
    allowed_origins TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    
    CONSTRAINT unique_key_prefix UNIQUE(key_prefix),
    CONSTRAINT valid_status CHECK (status IN ('active', 'rotating', 'expired', 'revoked'))
);

-- Indexes for API keys
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_status ON api_keys(status);
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at);
CREATE INDEX idx_api_keys_rotation_due ON api_keys(rotation_due);
CREATE INDEX idx_api_keys_user_tenant ON api_keys(user_id, tenant_id);

-- JWT Secrets table with versioning
CREATE TABLE IF NOT EXISTS jwt_secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version INTEGER NOT NULL,
    secret TEXT NOT NULL, -- Encrypted secret
    algorithm VARCHAR(50) DEFAULT 'HS256',
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, rotating, expired
    
    -- Timing
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    rotated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Flags
    is_active BOOLEAN DEFAULT false,
    
    CONSTRAINT unique_version UNIQUE(version),
    CONSTRAINT valid_jwt_status CHECK (status IN ('active', 'rotating', 'expired')),
    CONSTRAINT valid_algorithm CHECK (algorithm IN ('HS256', 'HS384', 'HS512', 'RS256', 'RS384', 'RS512'))
);

-- Indexes for JWT secrets
CREATE INDEX idx_jwt_secrets_is_active ON jwt_secrets(is_active);
CREATE INDEX idx_jwt_secrets_version ON jwt_secrets(version DESC);
CREATE INDEX idx_jwt_secrets_expires_at ON jwt_secrets(expires_at);

-- API Key rotation audit log
CREATE TABLE IF NOT EXISTS api_key_rotation_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id UUID NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL, -- created, rotated, expired, revoked
    performed_by UUID REFERENCES users(id),
    performed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    old_key_prefix VARCHAR(16),
    new_key_prefix VARCHAR(16),
    reason TEXT,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB DEFAULT '{}'
);

-- Indexes for rotation log
CREATE INDEX idx_rotation_log_key_id ON api_key_rotation_log(key_id);
CREATE INDEX idx_rotation_log_performed_at ON api_key_rotation_log(performed_at DESC);
CREATE INDEX idx_rotation_log_performed_by ON api_key_rotation_log(performed_by);

-- Tenant rate limits configuration
CREATE TABLE IF NOT EXISTS tenant_rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL UNIQUE,
    
    -- Rate limits
    rate_limit_rpm INTEGER DEFAULT 60,
    rate_limit_rph INTEGER DEFAULT 1000,
    rate_limit_rpd INTEGER DEFAULT 10000,
    
    -- Burst configuration
    burst_multiplier DECIMAL(3,2) DEFAULT 1.5,
    
    -- Quotas
    monthly_quota BIGINT,
    current_month_usage BIGINT DEFAULT 0,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    tier VARCHAR(50) DEFAULT 'standard', -- free, standard, premium, enterprise
    
    -- Tracking
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT valid_tier CHECK (tier IN ('free', 'standard', 'premium', 'enterprise'))
);

-- Indexes for tenant rate limits
CREATE INDEX idx_tenant_limits_tenant_id ON tenant_rate_limits(tenant_id);
CREATE INDEX idx_tenant_limits_tier ON tenant_rate_limits(tier);

-- User rate limits configuration (overrides tenant limits)
CREATE TABLE IF NOT EXISTS user_rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Rate limits
    rate_limit_rpm INTEGER DEFAULT 60,
    rate_limit_rph INTEGER DEFAULT 1000,
    rate_limit_rpd INTEGER DEFAULT 10000,
    
    -- Burst configuration
    burst_multiplier DECIMAL(3,2) DEFAULT 1.5,
    
    -- Quotas
    monthly_quota BIGINT,
    current_month_usage BIGINT DEFAULT 0,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    custom_limits BOOLEAN DEFAULT false,
    
    -- Tracking
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT unique_user_limits UNIQUE(user_id)
);

-- Indexes for user rate limits
CREATE INDEX idx_user_limits_user_id ON user_rate_limits(user_id);

-- Rate limit usage tracking (for analytics)
CREATE TABLE IF NOT EXISTS rate_limit_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255),
    api_key_id UUID REFERENCES api_keys(id) ON DELETE CASCADE,
    
    -- Time buckets
    timestamp_minute TIMESTAMP WITH TIME ZONE NOT NULL,
    timestamp_hour TIMESTAMP WITH TIME ZONE NOT NULL,
    timestamp_day DATE NOT NULL,
    
    -- Counters
    request_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    rate_limited_count INTEGER DEFAULT 0,
    
    -- Performance metrics
    avg_latency_ms INTEGER,
    p95_latency_ms INTEGER,
    p99_latency_ms INTEGER,
    
    CONSTRAINT unique_usage_minute UNIQUE(user_id, tenant_id, timestamp_minute)
);

-- Indexes for usage tracking
CREATE INDEX idx_usage_user_id ON rate_limit_usage(user_id);
CREATE INDEX idx_usage_tenant_id ON rate_limit_usage(tenant_id);
CREATE INDEX idx_usage_timestamp_day ON rate_limit_usage(timestamp_day DESC);
CREATE INDEX idx_usage_timestamp_hour ON rate_limit_usage(timestamp_hour DESC);

-- Function to automatically update rotation_due when key is rotated
CREATE OR REPLACE FUNCTION update_rotation_due()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.last_rotated != OLD.last_rotated THEN
        NEW.rotation_due := NEW.last_rotated + INTERVAL '90 days';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for automatic rotation_due update
CREATE TRIGGER trigger_update_rotation_due
    BEFORE UPDATE ON api_keys
    FOR EACH ROW
    EXECUTE FUNCTION update_rotation_due();

-- Function to log API key actions
CREATE OR REPLACE FUNCTION log_api_key_action()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO api_key_rotation_log (key_id, action, new_key_prefix)
        VALUES (NEW.id, 'created', NEW.key_prefix);
    ELSIF TG_OP = 'UPDATE' THEN
        IF NEW.status != OLD.status THEN
            INSERT INTO api_key_rotation_log (key_id, action, old_key_prefix, new_key_prefix)
            VALUES (NEW.id, NEW.status, OLD.key_prefix, NEW.key_prefix);
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for API key action logging
CREATE TRIGGER trigger_log_api_key_action
    AFTER INSERT OR UPDATE ON api_keys
    FOR EACH ROW
    EXECUTE FUNCTION log_api_key_action();

-- Function to clean up expired keys
CREATE OR REPLACE FUNCTION cleanup_expired_keys()
RETURNS void AS $$
BEGIN
    -- Mark keys as expired if past expiration
    UPDATE api_keys 
    SET status = 'expired' 
    WHERE status = 'active' 
    AND expires_at < NOW();
    
    -- Delete old expired keys (past grace period)
    DELETE FROM api_keys 
    WHERE status = 'expired' 
    AND expires_at < NOW() - INTERVAL '30 days';
    
    -- Delete old JWT secrets
    DELETE FROM jwt_secrets 
    WHERE is_active = false 
    AND expires_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;

-- Create initial JWT secret
INSERT INTO jwt_secrets (version, secret, algorithm, expires_at, is_active)
VALUES (
    1,
    encode(gen_random_bytes(64), 'hex'),
    'HS256',
    NOW() + INTERVAL '90 days',
    true
) ON CONFLICT DO NOTHING;

-- Add comments for documentation
COMMENT ON TABLE api_keys IS 'Stores API keys with rotation support and per-key rate limiting';
COMMENT ON TABLE jwt_secrets IS 'Stores JWT signing secrets with versioning for rotation';
COMMENT ON TABLE api_key_rotation_log IS 'Audit log for API key lifecycle events';
COMMENT ON TABLE tenant_rate_limits IS 'Per-tenant rate limiting configuration';
COMMENT ON TABLE user_rate_limits IS 'Per-user rate limiting configuration (overrides tenant limits)';
COMMENT ON TABLE rate_limit_usage IS 'Tracks rate limit usage for analytics and monitoring';

-- Grant permissions (adjust based on your user setup)
GRANT SELECT, INSERT, UPDATE, DELETE ON api_keys TO ffprobe;
GRANT SELECT, INSERT, UPDATE, DELETE ON jwt_secrets TO ffprobe;
GRANT SELECT, INSERT ON api_key_rotation_log TO ffprobe;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_rate_limits TO ffprobe;
GRANT SELECT, INSERT, UPDATE, DELETE ON user_rate_limits TO ffprobe;
GRANT SELECT, INSERT, UPDATE ON rate_limit_usage TO ffprobe;