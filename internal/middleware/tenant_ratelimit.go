package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rendiffdev/ffprobe-api/internal/cache"
	"github.com/rs/zerolog"
)

// TenantRateLimiter provides per-user and per-tenant rate limiting
type TenantRateLimiter struct {
	cache  cache.Client
	logger zerolog.Logger
	config TenantRateLimitConfig
}

// TenantRateLimitConfig holds tenant-specific rate limiting configuration
type TenantRateLimitConfig struct {
	// Default limits
	DefaultRPM int // Requests per minute
	DefaultRPH int // Requests per hour
	DefaultRPD int // Requests per day

	// Tenant-specific overrides
	EnableTenantLimits bool
	EnableUserLimits   bool

	// Burst allowance
	BurstMultiplier float64

	// Response headers
	IncludeHeaders bool
}

// RateLimitInfo contains rate limit information for a request
type RateLimitInfo struct {
	UserID      string
	TenantID    string
	APIKeyID    string
	LimitRPM    int
	LimitRPH    int
	LimitRPD    int
	CurrentRPM  int
	CurrentRPH  int
	CurrentRPD  int
	ResetMinute time.Time
	ResetHour   time.Time
	ResetDay    time.Time
}

// NewTenantRateLimiter creates a new tenant-aware rate limiter
func NewTenantRateLimiter(cacheClient cache.Client, logger zerolog.Logger, config TenantRateLimitConfig) *TenantRateLimiter {
	if cacheClient == nil {
		cacheClient = &cache.NoOpClient{}
	}
	// Set defaults if not configured
	if config.DefaultRPM == 0 {
		config.DefaultRPM = 60
	}
	if config.DefaultRPH == 0 {
		config.DefaultRPH = 1000
	}
	if config.DefaultRPD == 0 {
		config.DefaultRPD = 10000
	}
	if config.BurstMultiplier == 0 {
		config.BurstMultiplier = 1.5
	}

	return &TenantRateLimiter{
		cache:  cacheClient,
		logger: logger,
		config: config,
	}
}

// RateLimitMiddleware enforces rate limits per user/tenant
func (rl *TenantRateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for health checks
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		// Get user and tenant information from context
		userID := c.GetString("user_id")
		tenantID := c.GetString("tenant_id")
		apiKeyID := c.GetString("api_key_id")

		// If no authentication, use IP-based limiting
		if userID == "" && tenantID == "" {
			userID = "ip:" + c.ClientIP()
			tenantID = "public"
		}

		// Get rate limits for this user/tenant
		limits, err := rl.getRateLimits(c.Request.Context(), userID, tenantID, apiKeyID)
		if err != nil {
			rl.logger.Error().Err(err).Msg("Failed to get rate limits")
			// Fall back to defaults on error
			limits = rl.getDefaultLimits()
		}

		// Check rate limits
		allowed, info, err := rl.checkRateLimit(c.Request.Context(), userID, tenantID, limits)
		if err != nil {
			rl.logger.Error().Err(err).Msg("Failed to check rate limit")
			// Allow on error but log
			c.Next()
			return
		}

		// Add rate limit headers if configured
		if rl.config.IncludeHeaders {
			rl.addRateLimitHeaders(c, info)
		}

		if !allowed {
			// Rate limit exceeded
			rl.logger.Warn().
				Str("user_id", userID).
				Str("tenant_id", tenantID).
				Str("path", c.Request.URL.Path).
				Int("limit_rpm", info.LimitRPM).
				Int("current_rpm", info.CurrentRPM).
				Msg("Rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     fmt.Sprintf("Too many requests. Limit: %d req/min", info.LimitRPM),
				"retry_after": info.ResetMinute.Unix(),
				"limits": gin.H{
					"per_minute": info.LimitRPM,
					"per_hour":   info.LimitRPH,
					"per_day":    info.LimitRPD,
				},
				"current": gin.H{
					"per_minute": info.CurrentRPM,
					"per_hour":   info.CurrentRPH,
					"per_day":    info.CurrentRPD,
				},
			})
			c.Abort()
			return
		}

		// Store rate limit info in context for logging
		c.Set("rate_limit_info", info)

		c.Next()
	}
}

// getRateLimits retrieves rate limits for a user/tenant
func (rl *TenantRateLimiter) getRateLimits(ctx context.Context, userID, tenantID, apiKeyID string) (*RateLimitInfo, error) {
	limits := &RateLimitInfo{
		UserID:   userID,
		TenantID: tenantID,
		APIKeyID: apiKeyID,
		LimitRPM: rl.config.DefaultRPM,
		LimitRPH: rl.config.DefaultRPH,
		LimitRPD: rl.config.DefaultRPD,
	}

	// Check for API key specific limits first (highest priority)
	if apiKeyID != "" {
		keyLimits := rl.getAPIKeyLimits(ctx, apiKeyID)
		if keyLimits != nil {
			limits.LimitRPM = keyLimits.RPM
			limits.LimitRPH = keyLimits.RPH
			limits.LimitRPD = keyLimits.RPD
			return limits, nil
		}
	}

	// Check for user-specific limits
	if rl.config.EnableUserLimits && userID != "" {
		userLimits := rl.getUserLimits(ctx, userID)
		if userLimits != nil {
			limits.LimitRPM = userLimits.RPM
			limits.LimitRPH = userLimits.RPH
			limits.LimitRPD = userLimits.RPD
			return limits, nil
		}
	}

	// Check for tenant-specific limits
	if rl.config.EnableTenantLimits && tenantID != "" {
		tenantLimits := rl.getTenantLimits(ctx, tenantID)
		if tenantLimits != nil {
			limits.LimitRPM = tenantLimits.RPM
			limits.LimitRPH = tenantLimits.RPH
			limits.LimitRPD = tenantLimits.RPD
			return limits, nil
		}
	}

	return limits, nil
}

// APIKeyLimits holds rate limits from API key
type APIKeyLimits struct {
	RPM int
	RPH int
	RPD int
}

// getAPIKeyLimits retrieves limits from API key cache
func (rl *TenantRateLimiter) getAPIKeyLimits(ctx context.Context, apiKeyID string) *APIKeyLimits {
	cacheKey := fmt.Sprintf("apikey:%s:limits", apiKeyID)
	result, _ := rl.cache.HGetAll(ctx, cacheKey)

	if len(result) == 0 {
		return nil
	}

	limits := &APIKeyLimits{
		RPM: rl.config.DefaultRPM,
		RPH: rl.config.DefaultRPH,
		RPD: rl.config.DefaultRPD,
	}

	if rpm, err := strconv.Atoi(result["rpm"]); err == nil {
		limits.RPM = rpm
	}
	if rph, err := strconv.Atoi(result["rph"]); err == nil {
		limits.RPH = rph
	}
	if rpd, err := strconv.Atoi(result["rpd"]); err == nil {
		limits.RPD = rpd
	}

	return limits
}

// getUserLimits retrieves user-specific rate limits
func (rl *TenantRateLimiter) getUserLimits(ctx context.Context, userID string) *APIKeyLimits {
	cacheKey := fmt.Sprintf("user:%s:limits", userID)
	result, _ := rl.cache.HGetAll(ctx, cacheKey)

	if len(result) == 0 {
		return nil
	}

	limits := &APIKeyLimits{
		RPM: rl.config.DefaultRPM,
		RPH: rl.config.DefaultRPH,
		RPD: rl.config.DefaultRPD,
	}

	if rpm, err := strconv.Atoi(result["rpm"]); err == nil {
		limits.RPM = rpm
	}
	if rph, err := strconv.Atoi(result["rph"]); err == nil {
		limits.RPH = rph
	}
	if rpd, err := strconv.Atoi(result["rpd"]); err == nil {
		limits.RPD = rpd
	}

	return limits
}

// getTenantLimits retrieves tenant-specific rate limits
func (rl *TenantRateLimiter) getTenantLimits(ctx context.Context, tenantID string) *APIKeyLimits {
	cacheKey := fmt.Sprintf("tenant:%s:limits", tenantID)
	result, _ := rl.cache.HGetAll(ctx, cacheKey)

	if len(result) == 0 {
		return nil
	}

	limits := &APIKeyLimits{
		RPM: rl.config.DefaultRPM,
		RPH: rl.config.DefaultRPH,
		RPD: rl.config.DefaultRPD,
	}

	if rpm, err := strconv.Atoi(result["rpm"]); err == nil {
		limits.RPM = rpm
	}
	if rph, err := strconv.Atoi(result["rph"]); err == nil {
		limits.RPH = rph
	}
	if rpd, err := strconv.Atoi(result["rpd"]); err == nil {
		limits.RPD = rpd
	}

	return limits
}

// checkRateLimit checks if a request is within rate limits
func (rl *TenantRateLimiter) checkRateLimit(ctx context.Context, userID, tenantID string, limits *RateLimitInfo) (bool, *RateLimitInfo, error) {
	now := time.Now()

	// Create keys for different time windows
	minuteKey := fmt.Sprintf("ratelimit:%s:%s:minute:%d", tenantID, userID, now.Unix()/60)
	hourKey := fmt.Sprintf("ratelimit:%s:%s:hour:%d", tenantID, userID, now.Unix()/3600)
	dayKey := fmt.Sprintf("ratelimit:%s:%s:day:%s", tenantID, userID, now.Format("20060102"))

	// Increment counters individually
	minuteIncr, _ := rl.cache.Incr(ctx, minuteKey)
	hourIncr, _ := rl.cache.Incr(ctx, hourKey)
	dayIncr, _ := rl.cache.Incr(ctx, dayKey)

	// Set expiration
	rl.cache.Expire(ctx, minuteKey, time.Minute)
	rl.cache.Expire(ctx, hourKey, time.Hour)
	rl.cache.Expire(ctx, dayKey, 24*time.Hour)

	// Get current counts
	limits.CurrentRPM = int(minuteIncr)
	limits.CurrentRPH = int(hourIncr)
	limits.CurrentRPD = int(dayIncr)

	// Calculate reset times
	limits.ResetMinute = now.Truncate(time.Minute).Add(time.Minute)
	limits.ResetHour = now.Truncate(time.Hour).Add(time.Hour)
	limits.ResetDay = now.Truncate(24 * time.Hour).Add(24 * time.Hour)

	// Check limits with burst allowance
	burstRPM := int(float64(limits.LimitRPM) * rl.config.BurstMultiplier)

	// Check if any limit is exceeded
	if limits.CurrentRPM > burstRPM {
		return false, limits, nil
	}
	if limits.CurrentRPH > limits.LimitRPH {
		return false, limits, nil
	}
	if limits.CurrentRPD > limits.LimitRPD {
		return false, limits, nil
	}

	return true, limits, nil
}

// addRateLimitHeaders adds rate limit information to response headers
func (rl *TenantRateLimiter) addRateLimitHeaders(c *gin.Context, info *RateLimitInfo) {
	c.Header("X-RateLimit-Limit-Minute", strconv.Itoa(info.LimitRPM))
	c.Header("X-RateLimit-Limit-Hour", strconv.Itoa(info.LimitRPH))
	c.Header("X-RateLimit-Limit-Day", strconv.Itoa(info.LimitRPD))

	c.Header("X-RateLimit-Remaining-Minute", strconv.Itoa(max(0, info.LimitRPM-info.CurrentRPM)))
	c.Header("X-RateLimit-Remaining-Hour", strconv.Itoa(max(0, info.LimitRPH-info.CurrentRPH)))
	c.Header("X-RateLimit-Remaining-Day", strconv.Itoa(max(0, info.LimitRPD-info.CurrentRPD)))

	c.Header("X-RateLimit-Reset-Minute", strconv.FormatInt(info.ResetMinute.Unix(), 10))
	c.Header("X-RateLimit-Reset-Hour", strconv.FormatInt(info.ResetHour.Unix(), 10))
	c.Header("X-RateLimit-Reset-Day", strconv.FormatInt(info.ResetDay.Unix(), 10))
}

// getDefaultLimits returns default rate limits
func (rl *TenantRateLimiter) getDefaultLimits() *RateLimitInfo {
	return &RateLimitInfo{
		LimitRPM: rl.config.DefaultRPM,
		LimitRPH: rl.config.DefaultRPH,
		LimitRPD: rl.config.DefaultRPD,
	}
}

// SetUserLimits sets custom rate limits for a user
func (rl *TenantRateLimiter) SetUserLimits(ctx context.Context, userID string, rpm, rph, rpd int) error {
	cacheKey := fmt.Sprintf("user:%s:limits", userID)

	err := rl.cache.HSet(ctx, cacheKey, map[string]interface{}{
		"rpm": rpm,
		"rph": rph,
		"rpd": rpd,
	})

	if err != nil {
		return fmt.Errorf("failed to set user limits: %w", err)
	}

	// Set expiration to 30 days
	rl.cache.Expire(ctx, cacheKey, 30*24*time.Hour)

	rl.logger.Info().
		Str("user_id", userID).
		Int("rpm", rpm).
		Int("rph", rph).
		Int("rpd", rpd).
		Msg("Updated user rate limits")

	return nil
}

// SetTenantLimits sets custom rate limits for a tenant
func (rl *TenantRateLimiter) SetTenantLimits(ctx context.Context, tenantID string, rpm, rph, rpd int) error {
	cacheKey := fmt.Sprintf("tenant:%s:limits", tenantID)

	err := rl.cache.HSet(ctx, cacheKey, map[string]interface{}{
		"rpm": rpm,
		"rph": rph,
		"rpd": rpd,
	})

	if err != nil {
		return fmt.Errorf("failed to set tenant limits: %w", err)
	}

	// Set expiration to 30 days
	rl.cache.Expire(ctx, cacheKey, 30*24*time.Hour)

	rl.logger.Info().
		Str("tenant_id", tenantID).
		Int("rpm", rpm).
		Int("rph", rph).
		Int("rpd", rpd).
		Msg("Updated tenant rate limits")

	return nil
}

// GetCurrentUsage returns current usage for a user/tenant
func (rl *TenantRateLimiter) GetCurrentUsage(ctx context.Context, userID, tenantID string) (*RateLimitInfo, error) {
	now := time.Now()

	minuteKey := fmt.Sprintf("ratelimit:%s:%s:minute:%d", tenantID, userID, now.Unix()/60)
	hourKey := fmt.Sprintf("ratelimit:%s:%s:hour:%d", tenantID, userID, now.Unix()/3600)
	dayKey := fmt.Sprintf("ratelimit:%s:%s:day:%s", tenantID, userID, now.Format("20060102"))

	// Get current counts
	minuteStr, _ := rl.cache.Get(ctx, minuteKey)
	hourStr, _ := rl.cache.Get(ctx, hourKey)
	dayStr, _ := rl.cache.Get(ctx, dayKey)

	minuteCount, _ := strconv.Atoi(minuteStr)
	hourCount, _ := strconv.Atoi(hourStr)
	dayCount, _ := strconv.Atoi(dayStr)

	// Get limits
	limits, _ := rl.getRateLimits(ctx, userID, tenantID, "")

	limits.CurrentRPM = minuteCount
	limits.CurrentRPH = hourCount
	limits.CurrentRPD = dayCount

	limits.ResetMinute = now.Truncate(time.Minute).Add(time.Minute)
	limits.ResetHour = now.Truncate(time.Hour).Add(time.Hour)
	limits.ResetDay = now.Truncate(24 * time.Hour).Add(24 * time.Hour)

	return limits, nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
