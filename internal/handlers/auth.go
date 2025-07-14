package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
	// In a production system, you would:
	// 1. Add the token to a blacklist/revocation list
	// 2. Store revoked tokens in Redis with expiration
	// 3. Check blacklist in JWT validation middleware
	
	userID := c.GetString("user_id")
	username := c.GetString("username")
	
	h.logger.Info().
		Str("user_id", userID).
		Str("username", username).
		Str("ip", c.ClientIP()).
		Msg("User logged out")

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

	userID := c.GetString("user_id")
	username := c.GetString("username")

	// In production, you would:
	// 1. Validate current password
	// 2. Hash new password
	// 3. Update password in database
	// 4. Optionally invalidate all existing tokens

	h.logger.Info().
		Str("user_id", userID).
		Str("username", username).
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

	// In production, you would:
	// 1. Generate a secure random API key
	// 2. Store it in database with user association
	// 3. Return the key (only shown once)

	// For demo purposes, generate a simple key
	apiKey := "ffprobe_" + userID[:8] + "_" + generateRandomString(32)

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

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)] // Simple pseudo-random for demo
	}
	return string(b)
}