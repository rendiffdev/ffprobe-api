package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// RequestLimitMiddleware limits concurrent requests per IP/user
type RequestLimitMiddleware struct {
	maxConcurrent int
	activeReqs    map[string]int
	mutex         sync.RWMutex
	logger        zerolog.Logger
}

// NewRequestLimitMiddleware creates a new request limiting middleware
func NewRequestLimitMiddleware(maxConcurrent int, logger zerolog.Logger) *RequestLimitMiddleware {
	return &RequestLimitMiddleware{
		maxConcurrent: maxConcurrent,
		activeReqs:    make(map[string]int),
		logger:        logger,
	}
}

// Limit limits concurrent requests
func (m *RequestLimitMiddleware) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP or user ID)
		clientID := m.getClientID(c)

		m.mutex.Lock()
		current := m.activeReqs[clientID]
		
		if current >= m.maxConcurrent {
			m.mutex.Unlock()
			m.logger.Warn().
				Str("client_id", clientID).
				Int("current_requests", current).
				Int("max_concurrent", m.maxConcurrent).
				Msg("Request limit exceeded")
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many concurrent requests",
				"code":  "RATE_LIMIT_EXCEEDED",
				"retry_after": 30,
			})
			c.Abort()
			return
		}

		m.activeReqs[clientID]++
		m.mutex.Unlock()

		// Clean up on request completion
		defer func() {
			m.mutex.Lock()
			m.activeReqs[clientID]--
			if m.activeReqs[clientID] <= 0 {
				delete(m.activeReqs, clientID)
			}
			m.mutex.Unlock()
		}()

		c.Next()
	}
}

// getClientID gets a unique identifier for the client
func (m *RequestLimitMiddleware) getClientID(c *gin.Context) string {
	// Try to get user ID first (authenticated requests)
	if userID := c.GetString("user_id"); userID != "" {
		return "user:" + userID
	}

	// Fall back to IP address
	return "ip:" + c.ClientIP()
}

// GetStats returns current statistics
func (m *RequestLimitMiddleware) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]interface{})
	totalActive := 0
	
	for clientID, count := range m.activeReqs {
		stats[clientID] = count
		totalActive += count
	}

	return map[string]interface{}{
		"total_active_requests": totalActive,
		"max_concurrent":        m.maxConcurrent,
		"active_clients":        len(m.activeReqs),
		"client_details":        stats,
	}
}