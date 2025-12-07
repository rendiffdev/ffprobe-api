package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rendiffdev/rendiff-probe/internal/cache"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

// SecretRotationService handles API key and JWT secret rotation
type SecretRotationService struct {
	db     *sqlx.DB
	cache  cache.Client
	logger zerolog.Logger
	config SecretRotationConfig
}

// SecretRotationConfig holds configuration for secret rotation
type SecretRotationConfig struct {
	RotationInterval   time.Duration
	GracePeriod        time.Duration
	MinSecretLength    int
	MaxActiveKeys      int
	EnableAutoRotation bool
}

// APIKey represents an API key with metadata
type APIKey struct {
	ID           string    `db:"id" json:"id"`
	UserID       string    `db:"user_id" json:"user_id"`
	TenantID     string    `db:"tenant_id" json:"tenant_id"`
	KeyHash      string    `db:"key_hash" json:"-"`
	KeyPrefix    string    `db:"key_prefix" json:"key_prefix"`
	Name         string    `db:"name" json:"name"`
	Permissions  []string  `db:"permissions" json:"permissions"`
	Status       string    `db:"status" json:"status"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	ExpiresAt    time.Time `db:"expires_at" json:"expires_at"`
	LastUsedAt   time.Time `db:"last_used_at" json:"last_used_at"`
	LastRotated  time.Time `db:"last_rotated" json:"last_rotated"`
	RotationDue  time.Time `db:"rotation_due" json:"rotation_due"`
	UsageCount   int64     `db:"usage_count" json:"usage_count"`
	RateLimitRPM int       `db:"rate_limit_rpm" json:"rate_limit_rpm"`
	RateLimitRPH int       `db:"rate_limit_rph" json:"rate_limit_rph"`
	RateLimitRPD int       `db:"rate_limit_rpd" json:"rate_limit_rpd"`
}

// JWTSecret represents a JWT signing secret with versioning
type JWTSecret struct {
	ID        string    `db:"id" json:"id"`
	Version   int       `db:"version" json:"version"`
	Secret    string    `db:"secret" json:"-"`
	Algorithm string    `db:"algorithm" json:"algorithm"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	RotatedAt time.Time `db:"rotated_at" json:"rotated_at"`
	IsActive  bool      `db:"is_active" json:"is_active"`
}

// NewSecretRotationService creates a new secret rotation service
func NewSecretRotationService(db *sqlx.DB, cacheClient cache.Client, logger zerolog.Logger, config SecretRotationConfig) *SecretRotationService {
	if cacheClient == nil {
		cacheClient = &cache.NoOpClient{}
	}
	if config.RotationInterval == 0 {
		config.RotationInterval = 90 * 24 * time.Hour // 90 days default
	}
	if config.GracePeriod == 0 {
		config.GracePeriod = 7 * 24 * time.Hour // 7 days grace period
	}
	if config.MinSecretLength == 0 {
		config.MinSecretLength = 32
	}
	if config.MaxActiveKeys == 0 {
		config.MaxActiveKeys = 5
	}

	return &SecretRotationService{
		db:     db,
		cache:  cacheClient,
		logger: logger,
		config: config,
	}
}

// GenerateAPIKey creates a new API key for a user/tenant
func (s *SecretRotationService) GenerateAPIKey(ctx context.Context, userID, tenantID, name string, permissions []string) (*APIKey, string, error) {
	// Generate secure random key
	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		return nil, "", fmt.Errorf("failed to generate random key: %w", err)
	}

	keyString := hex.EncodeToString(rawKey)
	keyPrefix := keyString[:8]
	fullKey := fmt.Sprintf("ffprobe_%s_sk_%s", getEnvironment(), keyString)

	// Hash the key for storage
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash API key: %w", err)
	}

	// Check active key count for user
	var activeCount int
	err = s.db.GetContext(ctx, &activeCount,
		"SELECT COUNT(*) FROM api_keys WHERE user_id = $1 AND status = 'active'", userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check active keys: %w", err)
	}

	if activeCount >= s.config.MaxActiveKeys {
		return nil, "", fmt.Errorf("maximum number of active keys (%d) reached", s.config.MaxActiveKeys)
	}

	// Create key record
	apiKey := &APIKey{
		ID:           uuid.New().String(),
		UserID:       userID,
		TenantID:     tenantID,
		KeyHash:      string(hashedKey),
		KeyPrefix:    keyPrefix,
		Name:         name,
		Permissions:  permissions,
		Status:       "active",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(s.config.RotationInterval),
		LastRotated:  time.Now(),
		RotationDue:  time.Now().Add(s.config.RotationInterval),
		UsageCount:   0,
		RateLimitRPM: 60, // Default rate limits
		RateLimitRPH: 1000,
		RateLimitRPD: 10000,
	}

	// Store in database
	query := `
		INSERT INTO api_keys (
			id, user_id, tenant_id, key_hash, key_prefix, name, 
			permissions, status, created_at, expires_at, last_rotated, 
			rotation_due, usage_count, rate_limit_rpm, rate_limit_rph, rate_limit_rpd
		) VALUES (
			:id, :user_id, :tenant_id, :key_hash, :key_prefix, :name,
			:permissions, :status, :created_at, :expires_at, :last_rotated,
			:rotation_due, :usage_count, :rate_limit_rpm, :rate_limit_rph, :rate_limit_rpd
		)`

	_, err = s.db.NamedExecContext(ctx, query, apiKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to store API key: %w", err)
	}

	// Cache key metadata in Redis for fast lookup (best effort)
	cacheKey := fmt.Sprintf("apikey:%s:meta", keyPrefix)
	_ = s.cache.HSet(ctx, cacheKey, map[string]interface{}{
		"user_id":   userID,
		"tenant_id": tenantID,
		"key_id":    apiKey.ID,
		"status":    "active",
	})
	_ = s.cache.Expire(ctx, cacheKey, 24*time.Hour)

	s.logger.Info().
		Str("user_id", userID).
		Str("tenant_id", tenantID).
		Str("key_prefix", keyPrefix).
		Msg("Generated new API key")

	return apiKey, fullKey, nil
}

// RotateAPIKey rotates an existing API key
func (s *SecretRotationService) RotateAPIKey(ctx context.Context, keyID string) (*APIKey, string, error) {
	// Get existing key
	var oldKey APIKey
	err := s.db.GetContext(ctx, &oldKey, "SELECT * FROM api_keys WHERE id = $1", keyID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get API key: %w", err)
	}

	// Generate new key
	newKey, rawKey, err := s.GenerateAPIKey(ctx, oldKey.UserID, oldKey.TenantID,
		oldKey.Name+" (rotated)", oldKey.Permissions)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate new key: %w", err)
	}

	// Mark old key for expiration (with grace period)
	gracePeriodEnd := time.Now().Add(s.config.GracePeriod)
	_, err = s.db.ExecContext(ctx,
		"UPDATE api_keys SET status = 'rotating', expires_at = $1 WHERE id = $2",
		gracePeriodEnd, keyID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to update old key: %w", err)
	}

	// Invalidate old key cache (best effort)
	oldCacheKey := fmt.Sprintf("apikey:%s:meta", oldKey.KeyPrefix)
	_ = s.cache.Del(ctx, oldCacheKey)

	s.logger.Info().
		Str("old_key_id", keyID).
		Str("new_key_id", newKey.ID).
		Str("user_id", oldKey.UserID).
		Msg("Rotated API key")

	return newKey, rawKey, nil
}

// RotateJWTSecret rotates the JWT signing secret
func (s *SecretRotationService) RotateJWTSecret(ctx context.Context) (*JWTSecret, error) {
	// Generate new secret
	secretBytes := make([]byte, 64)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	newSecret := hex.EncodeToString(secretBytes)

	// Get current version
	var currentVersion int
	err := s.db.GetContext(ctx, &currentVersion,
		"SELECT COALESCE(MAX(version), 0) FROM jwt_secrets")
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	// Create new secret record
	jwtSecret := &JWTSecret{
		ID:        uuid.New().String(),
		Version:   currentVersion + 1,
		Secret:    newSecret,
		Algorithm: "HS256",
		Status:    "active",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.config.RotationInterval),
		RotatedAt: time.Now(),
		IsActive:  true,
	}

	// Start transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Deactivate old secrets (keep for verification during grace period)
	_, err = tx.ExecContext(ctx,
		"UPDATE jwt_secrets SET is_active = false, status = 'rotating' WHERE is_active = true")
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate old secrets: %w", err)
	}

	// Insert new secret
	query := `
		INSERT INTO jwt_secrets (
			id, version, secret, algorithm, status, 
			created_at, expires_at, rotated_at, is_active
		) VALUES (
			:id, :version, :secret, :algorithm, :status,
			:created_at, :expires_at, :rotated_at, :is_active
		)`

	_, err = tx.NamedExecContext(ctx, query, jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to insert new JWT secret: %w", err)
	}

	// Cache new secret for fast access (best effort)
	_ = s.cache.Set(ctx, "jwt:secret:active", newSecret, s.config.RotationInterval)
	_ = s.cache.Set(ctx, "jwt:version:active", jwtSecret.Version, s.config.RotationInterval)

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info().
		Int("version", jwtSecret.Version).
		Msg("Rotated JWT secret")

	return jwtSecret, nil
}

// ValidateAPIKey validates an API key and returns its metadata
func (s *SecretRotationService) ValidateAPIKey(ctx context.Context, apiKey string) (*APIKey, error) {
	// Extract key prefix for fast lookup
	if len(apiKey) < 8 {
		return nil, fmt.Errorf("invalid API key format")
	}

	prefix := apiKey[8:16] // Extract prefix from standard format

	// Check cache first
	cacheKey := fmt.Sprintf("apikey:%s:meta", prefix)
	cached, _ := s.cache.HGetAll(ctx, cacheKey)

	var keyRecord APIKey
	if len(cached) > 0 && cached["status"] == "active" {
		// Use cached metadata for initial validation
		keyID := cached["key_id"]
		err := s.db.GetContext(ctx, &keyRecord,
			"SELECT * FROM api_keys WHERE id = $1 AND status IN ('active', 'rotating')", keyID)
		if err != nil {
			return nil, fmt.Errorf("invalid API key")
		}
	} else {
		// Fallback to database lookup
		err := s.db.GetContext(ctx, &keyRecord,
			"SELECT * FROM api_keys WHERE key_prefix = $1 AND status IN ('active', 'rotating')", prefix)
		if err != nil {
			return nil, fmt.Errorf("invalid API key")
		}
	}

	// Verify key hash
	err := bcrypt.CompareHashAndPassword([]byte(keyRecord.KeyHash), []byte(apiKey))
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	// Check expiration
	if time.Now().After(keyRecord.ExpiresAt) {
		return nil, fmt.Errorf("API key expired")
	}

	// Update usage stats
	go s.updateKeyUsage(context.Background(), keyRecord.ID)

	return &keyRecord, nil
}

// updateKeyUsage updates API key usage statistics
func (s *SecretRotationService) updateKeyUsage(ctx context.Context, keyID string) {
	_, err := s.db.ExecContext(ctx,
		"UPDATE api_keys SET last_used_at = NOW(), usage_count = usage_count + 1 WHERE id = $1", keyID)
	if err != nil {
		s.logger.Error().Err(err).Str("key_id", keyID).Msg("Failed to update key usage")
	}
}

// CheckRotationDue checks if any secrets are due for rotation
func (s *SecretRotationService) CheckRotationDue(ctx context.Context) ([]string, error) {
	var dueKeys []string

	// Check API keys
	rows, err := s.db.QueryContext(ctx,
		"SELECT id FROM api_keys WHERE status = 'active' AND rotation_due < NOW()")
	if err != nil {
		return nil, fmt.Errorf("failed to check API keys: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var keyID string
		if err := rows.Scan(&keyID); err != nil {
			continue
		}
		dueKeys = append(dueKeys, keyID)
	}

	// Check JWT secrets
	var jwtDue bool
	err = s.db.GetContext(ctx, &jwtDue,
		"SELECT EXISTS(SELECT 1 FROM jwt_secrets WHERE is_active = true AND expires_at < NOW())")
	if err == nil && jwtDue {
		dueKeys = append(dueKeys, "JWT_SECRET")
	}

	return dueKeys, nil
}

// CleanupExpiredKeys removes expired keys past their grace period
func (s *SecretRotationService) CleanupExpiredKeys(ctx context.Context) error {
	// Delete expired API keys
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM api_keys WHERE status IN ('expired', 'rotating') AND expires_at < NOW() - INTERVAL '7 days'")
	if err != nil {
		return fmt.Errorf("failed to cleanup API keys: %w", err)
	}

	deletedKeys, _ := result.RowsAffected()

	// Delete old JWT secrets
	result, err = s.db.ExecContext(ctx,
		"DELETE FROM jwt_secrets WHERE is_active = false AND expires_at < NOW() - INTERVAL '7 days'")
	if err != nil {
		return fmt.Errorf("failed to cleanup JWT secrets: %w", err)
	}

	deletedSecrets, _ := result.RowsAffected()

	s.logger.Info().
		Int64("deleted_keys", deletedKeys).
		Int64("deleted_secrets", deletedSecrets).
		Msg("Cleaned up expired credentials")

	return nil
}

// getEnvironment returns the current environment
func getEnvironment() string {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}
	return env
}

// SetUserRateLimits sets custom rate limits for a user's API key
func (s *SecretRotationService) SetUserRateLimits(ctx context.Context, keyID string, rpm, rph, rpd int) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE api_keys SET rate_limit_rpm = $1, rate_limit_rph = $2, rate_limit_rpd = $3 WHERE id = $4",
		rpm, rph, rpd, keyID)
	if err != nil {
		return fmt.Errorf("failed to update rate limits: %w", err)
	}

	// Update cache (best effort)
	var key APIKey
	if err = s.db.GetContext(ctx, &key, "SELECT * FROM api_keys WHERE id = $1", keyID); err == nil {
		cacheKey := fmt.Sprintf("apikey:%s:limits", key.KeyPrefix)
		_ = s.cache.HSet(ctx, cacheKey, map[string]interface{}{
			"rpm": rpm,
			"rph": rph,
			"rpd": rpd,
		})
		_ = s.cache.Expire(ctx, cacheKey, 24*time.Hour)
	}

	return nil
}
