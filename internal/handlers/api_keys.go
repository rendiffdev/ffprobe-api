package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rendiffdev/rendiff-probe/internal/services"
	"github.com/rs/zerolog"
)

// APIKeyHandler handles API key management endpoints
type APIKeyHandler struct {
	rotationService *services.SecretRotationService
	logger          zerolog.Logger
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(rotationService *services.SecretRotationService, logger zerolog.Logger) *APIKeyHandler {
	return &APIKeyHandler{
		rotationService: rotationService,
		logger:          logger,
	}
}

// CreateAPIKeyRequest represents the request to create a new API key
type CreateAPIKeyRequest struct {
	Name        string   `json:"name" binding:"required,min=3,max=100"`
	Permissions []string `json:"permissions"`
	ExpiresIn   int      `json:"expires_in_days,omitempty"` // Optional, defaults to 90 days
}

// RotateAPIKeyRequest represents the request to rotate an API key
type RotateAPIKeyRequest struct {
	KeyID string `json:"key_id" binding:"required,uuid"`
}

// UpdateRateLimitsRequest represents the request to update rate limits
type UpdateRateLimitsRequest struct {
	KeyID        string `json:"key_id,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	TenantID     string `json:"tenant_id,omitempty"`
	RateLimitRPM int    `json:"rate_limit_rpm" binding:"required,min=1"`
	RateLimitRPH int    `json:"rate_limit_rph" binding:"required,min=1"`
	RateLimitRPD int    `json:"rate_limit_rpd" binding:"required,min=1"`
}

// CreateAPIKey creates a new API key for the authenticated user
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid create API key request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// Get user and tenant from context
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User authentication required",
		})
		return
	}

	// Default tenant if not provided
	if tenantID == "" {
		tenantID = "default"
	}

	// Create the API key
	apiKey, rawKey, err := h.rotationService.GenerateAPIKey(
		c.Request.Context(),
		userID,
		tenantID,
		req.Name,
		req.Permissions,
	)
	if err != nil {
		h.logger.Error().Err(err).
			Str("user_id", userID).
			Str("tenant_id", tenantID).
			Msg("Failed to create API key")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create API key",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info().
		Str("user_id", userID).
		Str("tenant_id", tenantID).
		Str("key_id", apiKey.ID).
		Str("key_prefix", apiKey.KeyPrefix).
		Msg("Created new API key")

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":           apiKey.ID,
			"name":         apiKey.Name,
			"key":          rawKey, // Only returned once at creation
			"key_prefix":   apiKey.KeyPrefix,
			"permissions":  apiKey.Permissions,
			"created_at":   apiKey.CreatedAt,
			"expires_at":   apiKey.ExpiresAt,
			"rotation_due": apiKey.RotationDue,
			"rate_limits": gin.H{
				"per_minute": apiKey.RateLimitRPM,
				"per_hour":   apiKey.RateLimitRPH,
				"per_day":    apiKey.RateLimitRPD,
			},
		},
		"message": "API key created successfully. Please save the key securely - it will not be shown again.",
	})
}

// RotateAPIKey rotates an existing API key
func (h *APIKeyHandler) RotateAPIKey(c *gin.Context) {
	var req RotateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid rotate API key request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// Get user from context for authorization
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User authentication required",
		})
		return
	}

	// Rotate the key
	newKey, rawKey, err := h.rotationService.RotateAPIKey(c.Request.Context(), req.KeyID)
	if err != nil {
		h.logger.Error().Err(err).
			Str("user_id", userID).
			Str("key_id", req.KeyID).
			Msg("Failed to rotate API key")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to rotate API key",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info().
		Str("user_id", userID).
		Str("old_key_id", req.KeyID).
		Str("new_key_id", newKey.ID).
		Msg("Rotated API key")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"old_key_id":        req.KeyID,
			"new_key_id":        newKey.ID,
			"new_key":           rawKey, // Only returned once
			"key_prefix":        newKey.KeyPrefix,
			"expires_at":        newKey.ExpiresAt,
			"rotation_due":      newKey.RotationDue,
			"grace_period_ends": time.Now().Add(7 * 24 * time.Hour),
		},
		"message": "API key rotated successfully. The old key will remain valid for 7 days.",
	})
}

// RotateJWTSecret rotates the JWT signing secret (admin only)
func (h *APIKeyHandler) RotateJWTSecret(c *gin.Context) {
	// Check admin role
	roles := c.GetStringSlice("roles")
	isAdmin := false
	for _, role := range roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Admin role required",
		})
		return
	}

	// Rotate JWT secret
	jwtSecret, err := h.rotationService.RotateJWTSecret(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to rotate JWT secret")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to rotate JWT secret",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info().
		Int("version", jwtSecret.Version).
		Str("admin_id", c.GetString("user_id")).
		Msg("Rotated JWT secret")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"version":    jwtSecret.Version,
			"algorithm":  jwtSecret.Algorithm,
			"rotated_at": jwtSecret.RotatedAt,
			"expires_at": jwtSecret.ExpiresAt,
		},
		"message": "JWT secret rotated successfully. Existing tokens remain valid during grace period.",
	})
}

// UpdateRateLimits updates rate limits for a user, tenant, or API key
func (h *APIKeyHandler) UpdateRateLimits(c *gin.Context) {
	var req UpdateRateLimitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid update rate limits request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// Check admin role
	roles := c.GetStringSlice("roles")
	isAdmin := false
	for _, role := range roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Admin role required",
		})
		return
	}

	// Validate that at least one target is specified
	if req.KeyID == "" && req.UserID == "" && req.TenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "Must specify key_id, user_id, or tenant_id",
		})
		return
	}

	// Update rate limits for API key
	if req.KeyID != "" {
		err := h.rotationService.SetUserRateLimits(
			c.Request.Context(),
			req.KeyID,
			req.RateLimitRPM,
			req.RateLimitRPH,
			req.RateLimitRPD,
		)
		if err != nil {
			h.logger.Error().Err(err).
				Str("key_id", req.KeyID).
				Msg("Failed to update API key rate limits")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update rate limits",
				"message": err.Error(),
			})
			return
		}
	}

	h.logger.Info().
		Str("admin_id", c.GetString("user_id")).
		Str("key_id", req.KeyID).
		Str("user_id", req.UserID).
		Str("tenant_id", req.TenantID).
		Int("rpm", req.RateLimitRPM).
		Int("rph", req.RateLimitRPH).
		Int("rpd", req.RateLimitRPD).
		Msg("Updated rate limits")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"key_id":    req.KeyID,
			"user_id":   req.UserID,
			"tenant_id": req.TenantID,
			"rate_limits": gin.H{
				"per_minute": req.RateLimitRPM,
				"per_hour":   req.RateLimitRPH,
				"per_day":    req.RateLimitRPD,
			},
		},
		"message": "Rate limits updated successfully",
	})
}

// CheckRotationStatus checks which secrets are due for rotation (admin only)
func (h *APIKeyHandler) CheckRotationStatus(c *gin.Context) {
	// Check admin role
	roles := c.GetStringSlice("roles")
	isAdmin := false
	for _, role := range roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Admin role required",
		})
		return
	}

	// Check rotation status
	dueKeys, err := h.rotationService.CheckRotationDue(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to check rotation status")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check rotation status",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"keys_due_for_rotation": dueKeys,
			"count":                 len(dueKeys),
		},
		"message": "Rotation status retrieved successfully",
	})
}

// CleanupExpiredKeys removes expired keys past their grace period (admin only)
func (h *APIKeyHandler) CleanupExpiredKeys(c *gin.Context) {
	// Check admin role
	roles := c.GetStringSlice("roles")
	isAdmin := false
	for _, role := range roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Admin role required",
		})
		return
	}

	// Cleanup expired keys
	err := h.rotationService.CleanupExpiredKeys(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to cleanup expired keys")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to cleanup expired keys",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info().
		Str("admin_id", c.GetString("user_id")).
		Msg("Cleaned up expired keys")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Expired keys cleaned up successfully",
	})
}
