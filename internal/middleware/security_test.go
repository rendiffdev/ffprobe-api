package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func TestNewSecurityMiddleware(t *testing.T) {
	logger := zerolog.Nop()
	config := SecurityConfig{
		EnableCSRF:         true,
		EnableXSS:          true,
		EnableFrameGuard:   true,
		EnableHSTS:         true,
		ContentTypeNoSniff: true,
	}

	middleware := NewSecurityMiddleware(config, logger)

	if middleware == nil {
		t.Fatal("Expected middleware to be created, got nil")
	}
	if middleware.config.HSTSMaxAge != 31536000 {
		t.Errorf("Expected default HSTSMaxAge 31536000, got %d", middleware.config.HSTSMaxAge)
	}
	if middleware.config.ReferrerPolicy != "strict-origin-when-cross-origin" {
		t.Errorf("Expected default ReferrerPolicy, got %s", middleware.config.ReferrerPolicy)
	}
}

func TestNewSecurityMiddleware_Defaults(t *testing.T) {
	logger := zerolog.Nop()
	config := SecurityConfig{}

	middleware := NewSecurityMiddleware(config, logger)

	// Check default allowed origins
	if len(middleware.config.AllowedOrigins) != 1 || middleware.config.AllowedOrigins[0] != "*" {
		t.Errorf("Expected default AllowedOrigins [*], got %v", middleware.config.AllowedOrigins)
	}

	// Check default allowed methods
	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"}
	if len(middleware.config.AllowedMethods) != len(expectedMethods) {
		t.Errorf("Expected %d default methods, got %d", len(expectedMethods), len(middleware.config.AllowedMethods))
	}
}

func TestSecurity_Headers(t *testing.T) {
	logger := zerolog.Nop()
	config := SecurityConfig{
		EnableXSS:          true,
		EnableFrameGuard:   true,
		ContentTypeNoSniff: true,
	}
	middleware := NewSecurityMiddleware(config, logger)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(middleware.Security())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// Check CSP header
	csp := w.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("Expected Content-Security-Policy header to be set")
	}

	// Check X-Frame-Options
	xfo := w.Header().Get("X-Frame-Options")
	if xfo != "DENY" {
		t.Errorf("Expected X-Frame-Options DENY, got %s", xfo)
	}

	// Check X-Content-Type-Options
	xcto := w.Header().Get("X-Content-Type-Options")
	if xcto != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options nosniff, got %s", xcto)
	}

	// Check X-XSS-Protection
	xxp := w.Header().Get("X-XSS-Protection")
	if xxp != "1; mode=block" {
		t.Errorf("Expected X-XSS-Protection '1; mode=block', got %s", xxp)
	}

	// Check Referrer-Policy
	rp := w.Header().Get("Referrer-Policy")
	if rp != "strict-origin-when-cross-origin" {
		t.Errorf("Expected Referrer-Policy strict-origin-when-cross-origin, got %s", rp)
	}
}

func TestCORS_Preflight(t *testing.T) {
	logger := zerolog.Nop()
	config := SecurityConfig{
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}
	middleware := NewSecurityMiddleware(config, logger)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(middleware.CORS())
	r.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d for preflight, got %d", http.StatusNoContent, w.Code)
	}

	// Check CORS headers
	acao := w.Header().Get("Access-Control-Allow-Origin")
	if acao != "https://example.com" {
		t.Errorf("Expected Access-Control-Allow-Origin https://example.com, got %s", acao)
	}
}

func TestCORS_ActualRequest(t *testing.T) {
	logger := zerolog.Nop()
	config := SecurityConfig{
		AllowedOrigins: []string{"*"},
	}
	middleware := NewSecurityMiddleware(config, logger)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(middleware.CORS())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// When origin is set and wildcard is allowed, CORS uses the actual origin for security
	acao := w.Header().Get("Access-Control-Allow-Origin")
	if acao != "https://example.com" {
		t.Errorf("Expected Access-Control-Allow-Origin https://example.com (matching origin), got %s", acao)
	}
}

func TestIPWhitelist_AllowedIP(t *testing.T) {
	logger := zerolog.Nop()
	config := SecurityConfig{}
	middleware := NewSecurityMiddleware(config, logger)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(middleware.IPWhitelist([]string{"192.168.1.1", "10.0.0.1"}))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for whitelisted IP, got %d", http.StatusOK, w.Code)
	}
}

func TestIPWhitelist_BlockedIP(t *testing.T) {
	logger := zerolog.Nop()
	config := SecurityConfig{}
	middleware := NewSecurityMiddleware(config, logger)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(middleware.IPWhitelist([]string{"192.168.1.1"}))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.5:12345"
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for blocked IP, got %d", http.StatusForbidden, w.Code)
	}
}

func TestInputSanitization(t *testing.T) {
	logger := zerolog.Nop()
	config := SecurityConfig{}
	middleware := NewSecurityMiddleware(config, logger)

	tests := []struct {
		name           string
		input          string
		expectedStatus int
	}{
		{
			name:           "Normal input",
			input:          `{"name": "test"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "SQL injection attempt",
			input:          `{"query": "'; DROP TABLE users; --"}`,
			expectedStatus: http.StatusOK, // Input is sanitized, not blocked
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			r.Use(middleware.InputSanitization())
			r.POST("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.input))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestThreatDetection_NormalRequest(t *testing.T) {
	logger := zerolog.Nop()
	config := SecurityConfig{}
	middleware := NewSecurityMiddleware(config, logger)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(middleware.ThreatDetection())
	r.GET("/api/v1/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for normal request, got %d", http.StatusOK, w.Code)
	}
}

func TestRequestLogging(t *testing.T) {
	logger := zerolog.Nop()
	config := SecurityConfig{}
	middleware := NewSecurityMiddleware(config, logger)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(middleware.RequestLogging())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// RequestLogging middleware should pass through requests
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCSRF_TokenGeneration(t *testing.T) {
	logger := zerolog.Nop()
	config := SecurityConfig{
		EnableCSRF: true,
	}
	middleware := NewSecurityMiddleware(config, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	middleware.GenerateCSRFToken(c)

	// Check that token was set in header
	xcsrf := w.Header().Get("X-CSRF-Token")
	if xcsrf == "" {
		t.Error("Expected X-CSRF-Token header to be set")
	} else if len(xcsrf) < 32 {
		t.Errorf("Expected CSRF token to be at least 32 characters, got %d", len(xcsrf))
	}

	// Check that token was set in context
	token, exists := c.Get("csrf_token")
	if !exists {
		t.Error("Expected csrf_token to be set in context")
	} else if token.(string) != xcsrf {
		t.Error("Expected context token to match header token")
	}
}
