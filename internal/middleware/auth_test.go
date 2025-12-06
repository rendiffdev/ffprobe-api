package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewAuthMiddleware(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret:     "test-secret-key-32-characters-long",
		APIKey:        "test-api-key",
		TokenExpiry:   time.Hour,
		RefreshExpiry: 24 * time.Hour,
	}

	middleware := NewAuthMiddleware(config, nil, nil, logger)

	if middleware == nil {
		t.Fatal("Expected middleware to be created, got nil")
	}
	if middleware.config.JWTSecret != config.JWTSecret {
		t.Errorf("Expected JWTSecret %s, got %s", config.JWTSecret, middleware.config.JWTSecret)
	}
}

func TestNewAuthMiddleware_DefaultExpiry(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret: "test-secret-key-32-characters-long",
		APIKey:    "test-api-key",
	}

	middleware := NewAuthMiddleware(config, nil, nil, logger)

	if middleware.config.TokenExpiry != 24*time.Hour {
		t.Errorf("Expected default TokenExpiry 24h, got %v", middleware.config.TokenExpiry)
	}
	if middleware.config.RefreshExpiry != 7*24*time.Hour {
		t.Errorf("Expected default RefreshExpiry 7 days, got %v", middleware.config.RefreshExpiry)
	}
}

func TestIsPublicEndpoint(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret: "test-secret-key-32-characters-long",
		APIKey:    "test-api-key",
	}
	middleware := NewAuthMiddleware(config, nil, nil, logger)

	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/health/ready", true},
		{"/docs", true},
		{"/docs/swagger", true},
		{"/api/v1/auth/login", true},
		{"/api/v1/auth/refresh", true},
		{"/api/v1/system/version", true},
		{"/api/v1/analyze", false},
		{"/api/v1/users", false},
		{"/admin", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := middleware.isPublicEndpoint(tt.path)
			if result != tt.expected {
				t.Errorf("isPublicEndpoint(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestExtractAPIKey(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret: "test-secret-key-32-characters-long",
		APIKey:    "test-api-key",
	}
	middleware := NewAuthMiddleware(config, nil, nil, logger)

	tests := []struct {
		name       string
		header     string
		query      string
		expectedKey string
	}{
		{
			name:       "API key in Authorization header",
			header:     "ApiKey test-key-123",
			query:      "",
			expectedKey: "test-key-123",
		},
		{
			name:       "API key in X-API-Key header",
			header:     "",
			query:      "",
			expectedKey: "",
		},
		{
			name:       "No API key",
			header:     "",
			query:      "",
			expectedKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			if tt.header != "" {
				c.Request.Header.Set("Authorization", tt.header)
			}

			result := middleware.extractAPIKey(c)
			if result != tt.expectedKey {
				t.Errorf("extractAPIKey() = %v, want %v", result, tt.expectedKey)
			}
		})
	}
}

func TestExtractToken(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret: "test-secret-key-32-characters-long",
		APIKey:    "test-api-key",
	}
	middleware := NewAuthMiddleware(config, nil, nil, logger)

	tests := []struct {
		name          string
		authHeader    string
		expectedToken string
	}{
		{
			name:          "Bearer token",
			authHeader:    "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
			expectedToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
		},
		{
			name:          "No token",
			authHeader:    "",
			expectedToken: "",
		},
		{
			name:          "Invalid format",
			authHeader:    "InvalidFormat token123",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.authHeader)
			}

			result := middleware.extractToken(c)
			if result != tt.expectedToken {
				t.Errorf("extractToken() = %v, want %v", result, tt.expectedToken)
			}
		})
	}
}

func TestAPIKeyAuth_PublicEndpoint(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret: "test-secret-key-32-characters-long",
		APIKey:    "test-api-key",
	}
	middleware := NewAuthMiddleware(config, nil, nil, logger)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	// Add the middleware
	r.Use(middleware.APIKeyAuth())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	c.Request = httptest.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for public endpoint, got %d", http.StatusOK, w.Code)
	}
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret: "test-secret-key-32-characters-long",
		APIKey:    "test-api-key",
	}
	middleware := NewAuthMiddleware(config, nil, nil, logger)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(middleware.APIKeyAuth())
	r.GET("/api/v1/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for missing API key, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret: "test-secret-key-32-characters-long",
		APIKey:    "valid-api-key",
	}
	middleware := NewAuthMiddleware(config, nil, nil, logger)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(middleware.APIKeyAuth())
	r.GET("/api/v1/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	req.Header.Set("Authorization", "ApiKey invalid-key")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for invalid API key, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret: "test-secret-key-32-characters-long",
		APIKey:    "valid-api-key",
	}
	middleware := NewAuthMiddleware(config, nil, nil, logger)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(middleware.APIKeyAuth())
	r.GET("/api/v1/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	req.Header.Set("Authorization", "ApiKey valid-api-key")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for valid API key, got %d", http.StatusOK, w.Code)
	}
}

func TestGenerateTokens(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret:     "test-secret-key-32-characters-long",
		APIKey:        "test-api-key",
		TokenExpiry:   time.Hour,
		RefreshExpiry: 24 * time.Hour,
	}
	middleware := NewAuthMiddleware(config, nil, nil, logger)

	accessToken, refreshToken, err := middleware.generateTokens("user123", "testuser", []string{"user"})
	if err != nil {
		t.Fatalf("generateTokens() error = %v", err)
	}

	if accessToken == "" {
		t.Error("Expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}
	if accessToken == refreshToken {
		t.Error("Access and refresh tokens should be different")
	}
}

func TestValidateToken(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret:     "test-secret-key-32-characters-long",
		APIKey:        "test-api-key",
		TokenExpiry:   time.Hour,
		RefreshExpiry: 24 * time.Hour,
	}
	middleware := NewAuthMiddleware(config, nil, nil, logger)

	// Generate a valid token
	accessToken, _, err := middleware.generateTokens("user123", "testuser", []string{"user", "admin"})
	if err != nil {
		t.Fatalf("generateTokens() error = %v", err)
	}

	// Validate the token
	claims, err := middleware.validateToken(accessToken)
	if err != nil {
		t.Fatalf("validateToken() error = %v", err)
	}

	if claims.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", claims.UserID)
	}
	if claims.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", claims.Username)
	}
	if len(claims.Roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(claims.Roles))
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret:     "test-secret-key-32-characters-long",
		APIKey:        "test-api-key",
		TokenExpiry:   time.Hour,
		RefreshExpiry: 24 * time.Hour,
	}
	middleware := NewAuthMiddleware(config, nil, nil, logger)

	_, err := middleware.validateToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret:     "test-secret-key-32-characters-long",
		APIKey:        "test-api-key",
		TokenExpiry:   -time.Hour, // Already expired
		RefreshExpiry: 24 * time.Hour,
	}
	middleware := NewAuthMiddleware(config, nil, nil, logger)

	accessToken, _, err := middleware.generateTokens("user123", "testuser", []string{"user"})
	if err != nil {
		t.Fatalf("generateTokens() error = %v", err)
	}

	_, err = middleware.validateToken(accessToken)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func TestHashPassword(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret: "test-secret-key-32-characters-long",
		APIKey:    "test-api-key",
	}
	middleware := NewAuthMiddleware(config, nil, nil, logger)

	password := "securePassword123!"
	hash, err := middleware.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Error("Expected non-empty hash")
	}
	if hash == password {
		t.Error("Hash should not equal original password")
	}

	// Hash should be different each time due to salt
	hash2, _ := middleware.HashPassword(password)
	if hash == hash2 {
		t.Error("Two hashes of same password should be different (due to salt)")
	}
}

func TestRequireRole(t *testing.T) {
	logger := zerolog.Nop()
	config := AuthConfig{
		JWTSecret:     "test-secret-key-32-characters-long",
		APIKey:        "test-api-key",
		TokenExpiry:   time.Hour,
		RefreshExpiry: 24 * time.Hour,
	}
	middleware := NewAuthMiddleware(config, nil, nil, logger)

	tests := []struct {
		name           string
		userRoles      []string
		requiredRoles  []string
		expectedStatus int
	}{
		{
			name:           "User has required role",
			userRoles:      []string{"admin"},
			requiredRoles:  []string{"admin"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "User has one of required roles",
			userRoles:      []string{"user"},
			requiredRoles:  []string{"admin", "user"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "User missing required role",
			userRoles:      []string{"user"},
			requiredRoles:  []string{"admin"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "User has no roles",
			userRoles:      []string{},
			requiredRoles:  []string{"admin"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			// Generate token with user roles
			accessToken, _, _ := middleware.generateTokens("user123", "testuser", tt.userRoles)

			// Set up the route with role requirement
			r.Use(middleware.JWTAuth())
			r.Use(middleware.RequireRole(tt.requiredRoles...))
			r.GET("/admin", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/admin", nil)
			req.Header.Set("Authorization", "Bearer "+accessToken)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
