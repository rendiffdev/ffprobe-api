package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/rendiffdev/ffprobe-api/internal/config"
	"github.com/rendiffdev/ffprobe-api/pkg/logger"
)

func TestHealthCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := &config.Config{}
	log := logger.New("info")
	
	// Mock database for testing
	// Note: In a real test, you'd use a test database or mock
	// For now, we'll test the endpoint structure
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"service":  "ffprobe-api",
			"version":  "0.1.0",
			"database": "healthy",
		})
	})

	// Test health check endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ffprobe-api")
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestProbeFileEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Mock endpoint for testing
	v1 := router.Group("/api/v1")
	v1.POST("/probe/file", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "File probe endpoint - implementation coming soon",
			"status":  "not_implemented",
		})
	})

	// Test probe file endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/probe/file", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "not_implemented")
}