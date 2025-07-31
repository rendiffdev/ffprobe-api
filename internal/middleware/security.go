package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// SecurityConfig holds security middleware configuration
type SecurityConfig struct {
	EnableCSRF         bool
	EnableXSS          bool
	EnableFrameGuard   bool
	EnableHSTS         bool
	HSTSMaxAge         int
	ContentTypeNoSniff bool
	ReferrerPolicy     string
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	ExposeHeaders      []string
	MaxAge             int
}

// SecurityMiddleware handles various security measures
type SecurityMiddleware struct {
	config SecurityConfig
	logger zerolog.Logger
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(config SecurityConfig, logger zerolog.Logger) *SecurityMiddleware {
	// Set defaults
	if config.HSTSMaxAge == 0 {
		config.HSTSMaxAge = 31536000 // 1 year
	}
	if config.ReferrerPolicy == "" {
		config.ReferrerPolicy = "strict-origin-when-cross-origin"
	}
	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = []string{"*"}
	}
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"}
	}
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{
			"Origin", "Content-Type", "Content-Length", "Accept-Encoding",
			"X-CSRF-Token", "Authorization", "X-Request-ID", "X-API-Key",
		}
	}

	return &SecurityMiddleware{
		config: config,
		logger: logger,
	}
}

// Security applies various security headers and protections
func (sm *SecurityMiddleware) Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self' ws: wss:; " +
			"frame-ancestors 'none'"
		c.Header("Content-Security-Policy", csp)

		// X-Frame-Options
		if sm.config.EnableFrameGuard {
			c.Header("X-Frame-Options", "DENY")
		}

		// X-Content-Type-Options
		if sm.config.ContentTypeNoSniff {
			c.Header("X-Content-Type-Options", "nosniff")
		}

		// X-XSS-Protection
		if sm.config.EnableXSS {
			c.Header("X-XSS-Protection", "1; mode=block")
		}

		// Strict-Transport-Security
		if sm.config.EnableHSTS && c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", 
				fmt.Sprintf("max-age=%d; includeSubDomains; preload", sm.config.HSTSMaxAge))
		}

		// Referrer-Policy
		c.Header("Referrer-Policy", sm.config.ReferrerPolicy)

		// X-Permitted-Cross-Domain-Policies
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		// Remove server information
		c.Header("Server", "")
		c.Header("X-Powered-By", "")

		c.Next()
	}
}

// CORS handles Cross-Origin Resource Sharing
func (sm *SecurityMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range sm.config.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
				// Only set credentials to true for specific origins, not wildcard
				c.Header("Access-Control-Allow-Credentials", "true")
			} else {
				c.Header("Access-Control-Allow-Origin", "*")
				// Never set credentials to true with wildcard origin (security vulnerability)
			}
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(sm.config.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(sm.config.AllowedHeaders, ", "))
		
		if len(sm.config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(sm.config.ExposeHeaders, ", "))
		}
		
		if sm.config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", sm.config.MaxAge))
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Max-Age", "86400") // 24 hours
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CSRF provides CSRF protection
func (sm *SecurityMiddleware) CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !sm.config.EnableCSRF {
			c.Next()
			return
		}

		// Skip CSRF for safe methods
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Skip CSRF for API key authentication
		if c.GetString("auth_type") == "api_key" {
			c.Next()
			return
		}

		// Get CSRF token from header
		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			token = c.PostForm("_csrf_token")
		}

		// Get expected token from session/context
		expectedToken := c.GetString("csrf_token")
		
		if token == "" || expectedToken == "" || token != expectedToken {
			sm.logger.Warn().
				Str("path", c.Request.URL.Path).
				Str("ip", c.ClientIP()).
				Msg("CSRF token validation failed")
			
			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token validation failed",
				"code":  "CSRF_TOKEN_INVALID",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateCSRFToken generates a new CSRF token
func (sm *SecurityMiddleware) GenerateCSRFToken(c *gin.Context) {
	token, err := generateRandomToken(32)
	if err != nil {
		sm.logger.Error().Err(err).Msg("Failed to generate CSRF token")
		return
	}
	
	c.Set("csrf_token", token)
	c.Header("X-CSRF-Token", token)
}

// RequestLogging logs all requests for security monitoring
func (sm *SecurityMiddleware) RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request details
		end := time.Now()
		latency := end.Sub(start)

		logEvent := sm.logger.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Int("status", c.Writer.Status()).
			Dur("latency", latency).
			Int("body_size", c.Writer.Size())

		if raw != "" {
			logEvent.Str("query", raw)
		}

		if userID := c.GetString("user_id"); userID != "" {
			logEvent.Str("user_id", userID)
		}

		if requestID := c.GetString("request_id"); requestID != "" {
			logEvent.Str("request_id", requestID)
		}

		// Log errors
		if len(c.Errors) > 0 {
			logEvent.Str("errors", c.Errors.String())
		}

		logEvent.Msg("Request processed")
	}
}

// IPWhitelist middleware for trusted IPs
func (sm *SecurityMiddleware) IPWhitelist(trustedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		// Check if IP is in trusted list
		trusted := false
		for _, trustedIP := range trustedIPs {
			if clientIP == trustedIP {
				trusted = true
				break
			}
		}

		if !trusted {
			sm.logger.Warn().
				Str("ip", clientIP).
				Str("path", c.Request.URL.Path).
				Msg("Untrusted IP attempt")
			
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
				"code":  "IP_NOT_WHITELISTED",
			})
			c.Abort()
			return
		}

		c.Set("trusted_ip", true)
		c.Next()
	}
}

// GeoRestriction middleware to restrict access by country
func (sm *SecurityMiddleware) GeoRestriction(allowedCountries []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get country from header (set by load balancer/CDN)
		country := c.GetHeader("CF-IPCountry") // Cloudflare
		if country == "" {
			country = c.GetHeader("X-Country-Code") // Generic
		}

		if country != "" && len(allowedCountries) > 0 {
			allowed := false
			for _, allowedCountry := range allowedCountries {
				if country == allowedCountry {
					allowed = true
					break
				}
			}

			if !allowed {
				sm.logger.Warn().
					Str("country", country).
					Str("ip", c.ClientIP()).
					Str("path", c.Request.URL.Path).
					Msg("Geo-restricted access attempt")
				
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Access restricted from your location",
					"code":  "GEO_RESTRICTED",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// InputSanitization middleware to sanitize user input
func (sm *SecurityMiddleware) InputSanitization() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				c.Request.URL.Query()[key][i] = sanitizeInput(value)
			}
		}

		// Note: For JSON body sanitization, it would need to be done
		// in the handlers after binding, as we can't modify the body here
		// without affecting the binding process

		c.Next()
	}
}

// Helper functions

func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func sanitizeInput(input string) string {
	// Basic input sanitization
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Remove potentially dangerous characters
	dangerous := []string{
		"<script", "</script>", "javascript:", "vbscript:",
		"onload=", "onerror=", "onclick=", "onmouseover=",
	}
	
	for _, danger := range dangerous {
		input = strings.ReplaceAll(strings.ToLower(input), danger, "")
	}
	
	return strings.TrimSpace(input)
}

// ThreatDetection middleware for basic threat detection
func (sm *SecurityMiddleware) ThreatDetection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for common attack patterns
		userAgent := strings.ToLower(c.Request.UserAgent())
		path := strings.ToLower(c.Request.URL.Path)
		
		// SQL injection patterns
		sqlPatterns := []string{
			"union select", "drop table", "insert into", 
			"delete from", "' or 1=1", "' or '1'='1",
		}
		
		// XSS patterns  
		xssPatterns := []string{
			"<script", "javascript:", "vbscript:",
			"onload=", "onerror=", "eval(",
		}
		
		// Bot patterns
		botPatterns := []string{
			"sqlmap", "nikto", "nmap", "masscan",
			"nessus", "openvas", "burpsuite",
		}

		// Check patterns
		threatDetected := false
		threatType := ""
		
		for _, pattern := range sqlPatterns {
			if strings.Contains(path, pattern) {
				threatDetected = true
				threatType = "SQL_INJECTION"
				break
			}
		}
		
		if !threatDetected {
			for _, pattern := range xssPatterns {
				if strings.Contains(path, pattern) {
					threatDetected = true
					threatType = "XSS_ATTEMPT"
					break
				}
			}
		}
		
		if !threatDetected {
			for _, pattern := range botPatterns {
				if strings.Contains(userAgent, pattern) {
					threatDetected = true
					threatType = "MALICIOUS_BOT"
					break
				}
			}
		}

		if threatDetected {
			sm.logger.Warn().
				Str("threat_type", threatType).
				Str("ip", c.ClientIP()).
				Str("user_agent", c.Request.UserAgent()).
				Str("path", c.Request.URL.Path).
				Msg("Threat detected")
			
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Security threat detected",
				"code":  threatType,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}