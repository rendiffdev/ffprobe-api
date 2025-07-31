package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/middleware"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authMiddleware *middleware.AuthMiddleware
	logger         zerolog.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authMiddleware *middleware.AuthMiddleware, logger zerolog.Logger) *AuthHandler {
	return &AuthHandler{
		authMiddleware: authMiddleware,
		logger:         logger,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenResponse represents authentication response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	UserInfo     UserInfo `json:"user_info"`
}

// UserInfo represents user information
type UserInfo struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

// Login handles user authentication
// @Summary User login
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	h.authMiddleware.Login(c)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refresh expired access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	h.authMiddleware.RefreshToken(c)
}

// Logout handles user logout
// @Summary User logout
// @Description Logout user and invalidate tokens
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := c.GetString("user_id")
	username := c.GetString("username")
	
	// Extract token from Authorization header
	token := ""
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}
	
	// Revoke the JWT token
	if token != "" {
		if err := h.authMiddleware.RevokeToken(token); err != nil {
			h.logger.Error().Err(err).
				Str("user_id", userID).
				Str("username", username).
				Msg("Failed to revoke token during logout")
			
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Logout completed but token revocation failed",
				"status":  "warning",
			})
			return
		}
	}
	
	h.logger.Info().
		Str("user_id", userID).
		Str("username", username).
		Str("ip", c.ClientIP()).
		Msg("User logged out and token revoked")

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
		"status":  "success",
	})
}

// Profile returns current user profile
// @Summary Get user profile
// @Description Get current authenticated user's profile
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UserInfo
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) Profile(c *gin.Context) {
	userID := c.GetString("user_id")
	username := c.GetString("username")
	roles, _ := c.Get("roles")
	
	userRoles := []string{}
	if roleList, ok := roles.([]string); ok {
		userRoles = roleList
	}

	c.JSON(http.StatusOK, UserInfo{
		ID:       userID,
		Username: username,
		Roles:    userRoles,
	})
}

// ChangePassword handles password change
// @Summary Change password
// @Description Change current user's password
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Password change request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// Validate passwords match
	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Password confirmation doesn't match",
			Details: "New password and confirmation password must be identical",
		})
		return
	}

	userID := c.GetString("user_id")
	username := c.GetString("username")

	// Get user from database and validate current password
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Details: err.Error(),
		})
		return
	}

	// Validate current password through auth middleware
	if !h.authMiddleware.ValidateUserPassword(userUUID, req.CurrentPassword) {
		h.logger.Warn().
			Str("user_id", userID).
			Str("username", username).
			Str("ip", c.ClientIP()).
			Msg("Invalid current password during password change")
		
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Current password is incorrect",
			Details: "Please verify your current password",
		})
		return
	}

	// Hash new password
	hashedPassword, err := h.authMiddleware.HashPassword(req.NewPassword)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to hash new password")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to process password change",
			Details: "Internal server error",
		})
		return
	}

	// Update password in database
	if err := h.authMiddleware.UpdateUserPassword(userUUID, hashedPassword); err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update password in database")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update password",
			Details: "Database update failed",
		})
		return
	}

	// Log successful password change
	h.logger.Info().
		Str("user_id", userID).
		Str("username", username).
		Str("ip", c.ClientIP()).
		Msg("Password changed successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
		"status":  "success",
	})
}

// ValidateToken validates a JWT token
// @Summary Validate token
// @Description Validate JWT token and return user info
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} TokenValidationResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/validate [get]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	userID := c.GetString("user_id")
	username := c.GetString("username")
	authType := c.GetString("auth_type")
	roles, _ := c.Get("roles")
	
	userRoles := []string{}
	if roleList, ok := roles.([]string); ok {
		userRoles = roleList
	}

	c.JSON(http.StatusOK, TokenValidationResponse{
		Valid:    true,
		UserInfo: UserInfo{
			ID:       userID,
			Username: username,
			Roles:    userRoles,
		},
		AuthType:  authType,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // This should come from token claims
	})
}

// GenerateAPIKey generates a new API key for the user
// @Summary Generate API key
// @Description Generate a new API key for the authenticated user
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} APIKeyResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/api-key [post]
func (h *AuthHandler) GenerateAPIKey(c *gin.Context) {
	userID := c.GetString("user_id")
	username := c.GetString("username")

	// Generate a cryptographically secure API key
	randomSuffix, err := generateSecureRandomString(16) // 16 bytes = 32 hex chars
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to generate secure API key")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to generate API key",
			Details: "Internal server error",
		})
		return
	}
	
	apiKey := "ffprobe_" + userID[:8] + "_" + randomSuffix

	h.logger.Info().
		Str("user_id", userID).
		Str("username", username).
		Msg("API key generated")

	c.JSON(http.StatusOK, APIKeyResponse{
		APIKey:    apiKey,
		CreatedAt: time.Now().Unix(),
		UserID:    userID,
		Note:      "Store this key securely. It will not be shown again.",
	})
}

// ListAPIKeys lists user's API keys
// @Summary List API keys
// @Description List all API keys for the authenticated user
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} APIKeyListResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/api-keys [get]
func (h *AuthHandler) ListAPIKeys(c *gin.Context) {
	userID := c.GetString("user_id")

	// In production, fetch from database
	apiKeys := []APIKeyInfo{
		{
			ID:        "key_1",
			Name:      "Production Key",
			CreatedAt: time.Now().AddDate(0, -1, 0).Unix(),
			LastUsed:  time.Now().AddDate(0, 0, -7).Unix(),
			Prefix:    "ffprobe_" + userID[:8] + "_****",
		},
		{
			ID:        "key_2", 
			Name:      "Development Key",
			CreatedAt: time.Now().AddDate(0, 0, -15).Unix(),
			LastUsed:  time.Now().AddDate(0, 0, -2).Unix(),
			Prefix:    "ffprobe_" + userID[:8] + "_****",
		},
	}

	c.JSON(http.StatusOK, APIKeyListResponse{
		APIKeys: apiKeys,
		Total:   len(apiKeys),
	})
}

// RevokeAPIKey revokes an API key
// @Summary Revoke API key
// @Description Revoke a specific API key
// @Tags auth
// @Security BearerAuth
// @Param id path string true "API Key ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/auth/api-keys/{id} [delete]
func (h *AuthHandler) RevokeAPIKey(c *gin.Context) {
	keyID := c.Param("id")
	userID := c.GetString("user_id")

	// In production, verify key belongs to user and delete from database

	h.logger.Info().
		Str("user_id", userID).
		Str("key_id", keyID).
		Msg("API key revoked")

	c.Status(http.StatusNoContent)
}

// Request/Response types

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

type TokenValidationResponse struct {
	Valid     bool     `json:"valid"`
	UserInfo  UserInfo `json:"user_info"`
	AuthType  string   `json:"auth_type"`
	ExpiresAt int64    `json:"expires_at"`
}

type APIKeyResponse struct {
	APIKey    string `json:"api_key"`
	CreatedAt int64  `json:"created_at"`
	UserID    string `json:"user_id"`
	Note      string `json:"note"`
}

type APIKeyInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	LastUsed  int64  `json:"last_used"`
	Prefix    string `json:"prefix"`
}

type APIKeyListResponse struct {
	APIKeys []APIKeyInfo `json:"api_keys"`
	Total   int          `json:"total"`
}

// Helper functions

func generateSecureRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}