package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret    string
	APIKey       string
	TokenExpiry  time.Duration
	RefreshExpiry time.Duration
}

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	config AuthConfig
	logger zerolog.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config AuthConfig, logger zerolog.Logger) *AuthMiddleware {
	if config.TokenExpiry == 0 {
		config.TokenExpiry = 24 * time.Hour // Default 24 hours
	}
	if config.RefreshExpiry == 0 {
		config.RefreshExpiry = 7 * 24 * time.Hour // Default 7 days
	}

	return &AuthMiddleware{
		config: config,
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
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
				"code":  "MISSING_API_KEY",
			})
			c.Abort()
			return
		}

		// Validate API key using constant time comparison
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(m.config.APIKey)) != 1 {
			m.logger.Warn().
				Str("path", c.Request.URL.Path).
				Str("ip", c.ClientIP()).
				Msg("Invalid API key")
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
				"code":  "INVALID_API_KEY",
			})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// In production, validate against database
	// For now, use hardcoded credentials
	if !m.validateCredentials(req.Username, req.Password) {
		m.logger.Warn().
			Str("username", req.Username).
			Str("ip", c.ClientIP()).
			Msg("Failed login attempt")
		
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
			"code":  "INVALID_CREDENTIALS",
		})
		return
	}

	// Generate tokens
	userID := uuid.New().String()
	accessToken, refreshToken, err := m.generateTokens(userID, req.Username, []string{"user"})
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to generate tokens")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate tokens",
			"code":  "TOKEN_GENERATION_FAILED",
		})
		return
	}

	m.logger.Info().
		Str("username", req.Username).
		Str("user_id", userID).
		Str("ip", c.ClientIP()).
		Msg("Successful login")

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(m.config.TokenExpiry.Seconds()),
	})
}

// RefreshToken handles token refresh
func (m *AuthMiddleware) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	claims, err := m.validateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid refresh token",
			"code":  "INVALID_REFRESH_TOKEN",
		})
		return
	}

	// Generate new tokens
	accessToken, refreshToken, err := m.generateTokens(claims.UserID, claims.Username, claims.Roles)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to refresh tokens")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to refresh tokens",
			"code":  "TOKEN_REFRESH_FAILED",
		})
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

func (m *AuthMiddleware) validateCredentials(username, password string) bool {
	// SECURITY: This should be replaced with proper password hashing (bcrypt) 
	// and database validation in production
	// DO NOT use hardcoded credentials in production!
	
	// For demo purposes only - replace with database lookup
	if username == "" || password == "" {
		return false
	}
	
	// In production, this should:
	// 1. Query user from database by username
	// 2. Compare hashed password using bcrypt.CompareHashAndPassword
	// 3. Implement account lockout after failed attempts
	// 4. Log all authentication attempts
	
	return false // Disabled hardcoded auth - implement proper auth
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
	
	return nil, jwt.ErrTokenInvalid
}