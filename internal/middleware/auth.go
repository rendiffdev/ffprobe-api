package middleware

import (
	"crypto/subtle"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"context"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rendiffdev/ffprobe-api/internal/cache"
	"github.com/rendiffdev/ffprobe-api/internal/errors"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string
	APIKey        string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
}

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	config AuthConfig
	db     *sqlx.DB
	cache  cache.Client
	logger zerolog.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config AuthConfig, db *sqlx.DB, cacheClient cache.Client, logger zerolog.Logger) *AuthMiddleware {
	if cacheClient == nil {
		cacheClient = &cache.NoOpClient{}
	}
	if config.TokenExpiry == 0 {
		config.TokenExpiry = 24 * time.Hour // Default 24 hours
	}
	if config.RefreshExpiry == 0 {
		config.RefreshExpiry = 7 * 24 * time.Hour // Default 7 days
	}

	return &AuthMiddleware{
		config: config,
		db:     db,
		cache:  cacheClient,
		logger: logger,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// TokenResponse represents authentication response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// AuthRequest represents authentication request
type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest represents token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// APIKeyAuth middleware for API key authentication
func (m *AuthMiddleware) APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health checks and docs
		if m.isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		apiKey := m.extractAPIKey(c)
		if apiKey == "" {
			m.logger.Warn().
				Str("path", c.Request.URL.Path).
				Str("ip", c.ClientIP()).
				Msg("Missing API key")

			errors.Unauthorized(c, "API key required", "No API key provided in request")
			c.Abort()
			return
		}

		// Validate API key using constant time comparison
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(m.config.APIKey)) != 1 {
			m.logger.Warn().
				Str("path", c.Request.URL.Path).
				Str("ip", c.ClientIP()).
				Msg("Invalid API key")

			errors.Unauthorized(c, "Invalid API key", "The provided API key is not valid")
			c.Abort()
			return
		}

		// Set user context for API key auth
		c.Set("user_id", "api_key_user")
		c.Set("auth_type", "api_key")
		c.Next()
	}
}

// JWTAuth middleware for JWT token authentication
func (m *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for public endpoints
		if m.isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		token := m.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization token required",
				"code":  "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		claims, err := m.validateToken(token)
		if err != nil {
			m.logger.Warn().
				Err(err).
				Str("path", c.Request.URL.Path).
				Str("ip", c.ClientIP()).
				Msg("Invalid token")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Set("auth_type", "jwt")
		c.Next()
	}
}

// RequireRole middleware to check user roles
func (m *AuthMiddleware) RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Role information not found",
				"code":  "MISSING_ROLES",
			})
			c.Abort()
			return
		}

		roles, ok := userRoles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid role format",
				"code":  "INVALID_ROLES",
			})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, userRole := range roles {
			for _, requiredRole := range requiredRoles {
				if userRole == requiredRole || userRole == "admin" {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			m.logger.Warn().
				Strs("user_roles", roles).
				Strs("required_roles", requiredRoles).
				Str("user_id", c.GetString("user_id")).
				Msg("Insufficient permissions")

			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Login handles user authentication
func (m *AuthMiddleware) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.ValidationError(c, "Invalid request format", "")
		return
	}

	// Validate credentials and get user info
	user, valid := m.validateCredentialsAndGetUser(req.Username, req.Password)
	if !valid || user == nil {
		m.logger.Warn().
			Str("username", req.Username).
			Str("ip", c.ClientIP()).
			Msg("Failed login attempt")

		errors.Unauthorized(c, "Invalid credentials", "")
		return
	}

	// Generate tokens with actual user data
	userID := user.ID.String()
	roles := []string{user.Role}
	accessToken, refreshToken, err := m.generateTokens(userID, user.Email, roles)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to generate tokens")
		errors.InternalError(c, "Failed to generate tokens", "")
		return
	}

	m.logger.Info().
		Str("username", user.Email).
		Str("user_id", userID).
		Str("role", user.Role).
		Str("ip", c.ClientIP()).
		Msg("Successful login")

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    int64(m.config.TokenExpiry.Seconds()),
		"user": gin.H{
			"id":    userID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// RefreshToken handles token refresh
func (m *AuthMiddleware) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.ValidationError(c, "Invalid request format", "")
		return
	}

	claims, err := m.validateToken(req.RefreshToken)
	if err != nil {
		errors.Unauthorized(c, "Invalid refresh token", "")
		return
	}

	// Generate new tokens
	accessToken, refreshToken, err := m.generateTokens(claims.UserID, claims.Username, claims.Roles)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to refresh tokens")
		errors.InternalError(c, "Failed to refresh tokens", "")
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(m.config.TokenExpiry.Seconds()),
	})
}

// Helper methods

func (m *AuthMiddleware) extractAPIKey(c *gin.Context) string {
	// Check header first
	if key := c.GetHeader("X-API-Key"); key != "" {
		return key
	}

	// Check Authorization header
	if auth := c.GetHeader("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "ApiKey ") {
			return strings.TrimPrefix(auth, "ApiKey ")
		}
	}

	// Check query parameter
	return c.Query("api_key")
}

func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	// Check Authorization header
	auth := c.GetHeader("Authorization")
	if auth != "" && strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Check query parameter (less secure, for WebSocket connections)
	return c.Query("token")
}

func (m *AuthMiddleware) isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/docs",
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
		"/api/v1/system/version",
	}

	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

// UserCredentials represents user credentials from database
type UserCredentials struct {
	ID           uuid.UUID  `db:"id"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	Role         string     `db:"role"`
	Status       string     `db:"status"`
	FailedLogins int        `db:"failed_logins"`
	LockedUntil  *time.Time `db:"locked_until"`
	LastLoginAt  *time.Time `db:"last_login_at"`
}

func (m *AuthMiddleware) validateCredentials(username, password string) bool {
	user, valid := m.validateCredentialsAndGetUser(username, password)
	return valid && user != nil
}

func (m *AuthMiddleware) validateCredentialsAndGetUser(username, password string) (*UserCredentials, bool) {
	if username == "" || password == "" {
		return nil, false
	}

	// Get user from database
	var user UserCredentials
	query := `
		SELECT id, email, password_hash, role, status, 
			   COALESCE(failed_logins, 0) as failed_logins,
			   locked_until, last_login_at
		FROM users 
		WHERE email = $1 AND deleted_at IS NULL`

	err := m.db.Get(&user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			m.logger.Warn().Str("email", username).Msg("Login attempt with non-existent user")
		} else {
			m.logger.Error().Err(err).Str("email", username).Msg("Database error during login")
		}
		return nil, false
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		m.logger.Warn().Str("email", username).Msg("Login attempt on locked account")
		return nil, false
	}

	// Check if account is active
	if user.Status != "active" {
		m.logger.Warn().Str("email", username).Str("status", user.Status).Msg("Login attempt on inactive account")
		return nil, false
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Password is incorrect - increment failed login count
		m.handleFailedLogin(user.ID, user.Email, user.FailedLogins)
		return nil, false
	}

	// Password is correct - reset failed logins and update last login
	m.handleSuccessfulLogin(user.ID, user.Email)

	return &user, true
}

// handleFailedLogin increments failed login count and locks account if necessary
func (m *AuthMiddleware) handleFailedLogin(userID uuid.UUID, email string, currentFailedLogins int) {
	newFailedLogins := currentFailedLogins + 1
	var lockedUntil *time.Time

	// Lock account after 5 failed attempts for 30 minutes
	if newFailedLogins >= 5 {
		lockTime := time.Now().Add(30 * time.Minute)
		lockedUntil = &lockTime
		m.logger.Warn().
			Str("email", email).
			Int("failed_logins", newFailedLogins).
			Time("locked_until", lockTime).
			Msg("Account locked due to too many failed login attempts")
	}

	query := `
		UPDATE users 
		SET failed_logins = $2, locked_until = $3, updated_at = NOW()
		WHERE id = $1`

	_, err := m.db.Exec(query, userID, newFailedLogins, lockedUntil)
	if err != nil {
		m.logger.Error().Err(err).Str("email", email).Msg("Failed to update failed login count")
	}
}

// handleSuccessfulLogin resets failed login count and updates last login time
func (m *AuthMiddleware) handleSuccessfulLogin(userID uuid.UUID, email string) {
	query := `
		UPDATE users 
		SET failed_logins = 0, locked_until = NULL, last_login_at = NOW(), updated_at = NOW()
		WHERE id = $1`

	_, err := m.db.Exec(query, userID)
	if err != nil {
		m.logger.Error().Err(err).Str("email", email).Msg("Failed to update successful login")
	} else {
		m.logger.Info().Str("email", email).Msg("Successful login")
	}
}

func (m *AuthMiddleware) generateTokens(userID, username string, roles []string) (string, string, error) {
	now := time.Now()

	// Access token
	accessClaims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "ffprobe-api",
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(m.config.JWTSecret))
	if err != nil {
		return "", "", err
	}

	// Refresh token
	refreshClaims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "ffprobe-api",
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(m.config.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (m *AuthMiddleware) validateToken(tokenString string) (*Claims, error) {
	// First check if token is blacklisted
	if m.isTokenBlacklisted(tokenString) {
		m.logger.Warn().Str("token_prefix", tokenString[:min(20, len(tokenString))]).Msg("Attempting to use blacklisted token")
		return nil, fmt.Errorf("token is blacklisted")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateUserPassword validates a user's password
func (m *AuthMiddleware) ValidateUserPassword(userID uuid.UUID, password string) bool {
	var user UserCredentials
	query := `
		SELECT id, email, password_hash, role, status, 
			   COALESCE(failed_logins, 0) as failed_logins,
			   locked_until, last_login_at
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL`

	err := m.db.Get(&user, query, userID)
	if err != nil {
		m.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get user for password validation")
		return false
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		m.logger.Warn().Str("user_id", userID.String()).Msg("Password validation on locked account")
		return false
	}

	// Check if account is active
	if user.Status != "active" {
		m.logger.Warn().Str("user_id", userID.String()).Str("status", user.Status).Msg("Password validation on inactive account")
		return false
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// HashPassword hashes a password using bcrypt
func (m *AuthMiddleware) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// UpdateUserPassword updates a user's password in the database
func (m *AuthMiddleware) UpdateUserPassword(userID uuid.UUID, hashedPassword string) error {
	query := `
		UPDATE users 
		SET password_hash = $2, updated_at = NOW()
		WHERE id = $1`

	_, err := m.db.Exec(query, userID, hashedPassword)
	if err != nil {
		m.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to update user password")
		return err
	}

	m.logger.Info().Str("user_id", userID.String()).Msg("User password updated successfully")
	return nil
}

// RevokeToken adds a token to the blacklist
func (m *AuthMiddleware) RevokeToken(tokenString string) error {
	if m.cache == nil {
		m.logger.Error().Msg("Cache client not available for token revocation")
		return fmt.Errorf("token revocation service unavailable")
	}

	// Parse token to get expiration time
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.config.JWTSecret), nil
	})

	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to parse token for revocation")
		return err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return fmt.Errorf("invalid token claims")
	}

	// Calculate TTL until token expires
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		// Token already expired, no need to blacklist
		m.logger.Debug().Msg("Token already expired, skipping blacklist")
		return nil
	}

	// Add token to blacklist with TTL
	ctx := context.Background()
	blacklistKey := "blacklist:token:" + claims.ID
	err = m.cache.Set(ctx, blacklistKey, "revoked", ttl)
	if err != nil {
		m.logger.Error().Err(err).Str("token_id", claims.ID).Msg("Failed to blacklist token")
		return err
	}

	m.logger.Info().
		Str("token_id", claims.ID).
		Str("user_id", claims.UserID).
		Dur("ttl", ttl).
		Msg("Token successfully blacklisted")

	return nil
}

// isTokenBlacklisted checks if a token is in the blacklist
func (m *AuthMiddleware) isTokenBlacklisted(tokenString string) bool {
	if m.cache == nil {
		return false // If cache is not available, don't block tokens
	}

	// Parse token to get the JTI (token ID)
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.config.JWTSecret), nil
	})

	if err != nil {
		return false // If we can't parse, let normal validation handle it
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return false
	}

	// Check if token is blacklisted
	ctx := context.Background()
	blacklistKey := "blacklist:token:" + claims.ID
	exists := m.cache.Exists(ctx, blacklistKey)

	return exists > 0
}

// Helper function for minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
