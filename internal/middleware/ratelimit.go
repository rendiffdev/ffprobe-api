package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
	BurstSize         int
	EnablePerIP       bool
	EnablePerUser     bool
}

// RateCounter tracks request counts for different time windows
type RateCounter struct {
	Minute map[string]*WindowCounter
	Hour   map[string]*WindowCounter
	Day    map[string]*WindowCounter
	mutex  sync.RWMutex
}

// WindowCounter tracks requests within a time window
type WindowCounter struct {
	Count     int
	ResetTime time.Time
	mutex     sync.RWMutex
}

// RateLimitMiddleware handles rate limiting
type RateLimitMiddleware struct {
	config   RateLimitConfig
	counters *RateCounter
	logger   zerolog.Logger
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(config RateLimitConfig, logger zerolog.Logger) *RateLimitMiddleware {
	// Set defaults
	if config.RequestsPerMinute == 0 {
		config.RequestsPerMinute = 60
	}
	if config.RequestsPerHour == 0 {
		config.RequestsPerHour = 1000
	}
	if config.RequestsPerDay == 0 {
		config.RequestsPerDay = 10000
	}
	if config.BurstSize == 0 {
		config.BurstSize = 10
	}

	counters := &RateCounter{
		Minute: make(map[string]*WindowCounter),
		Hour:   make(map[string]*WindowCounter),
		Day:    make(map[string]*WindowCounter),
	}

	middleware := &RateLimitMiddleware{
		config:   config,
		counters: counters,
		logger:   logger,
	}

	// Start cleanup goroutine
	go middleware.cleanup()

	return middleware
}

// RateLimit middleware function
func (rl *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for health checks
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		var identifier string
		
		// Determine identifier based on configuration
		if rl.config.EnablePerUser {
			if userID := c.GetString("user_id"); userID != "" {
				identifier = "user:" + userID
			}
		}
		
		if identifier == "" && rl.config.EnablePerIP {
			identifier = "ip:" + c.ClientIP()
		}
		
		if identifier == "" {
			identifier = "global"
		}

		// Check rate limits
		if !rl.checkRateLimit(identifier) {
			rl.logger.Warn().
				Str("identifier", identifier).
				Str("path", c.Request.URL.Path).
				Msg("Rate limit exceeded")

			// Get retry after time
			retryAfter := rl.getRetryAfter(identifier)
			
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.RequestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(retryAfter.Unix(), 10))
			c.Header("Retry-After", strconv.FormatInt(int64(retryAfter.Sub(time.Now()).Seconds()), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"code":        "RATE_LIMIT_EXCEEDED",
				"retry_after": retryAfter.Unix(),
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := rl.getRemainingRequests(identifier)
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(rl.getResetTime(identifier).Unix(), 10))

		c.Next()
	}
}

// checkRateLimit checks if request is within rate limits
func (rl *RateLimitMiddleware) checkRateLimit(identifier string) bool {
	now := time.Now()

	// Check minute limit
	if !rl.checkWindow(identifier, "minute", now, time.Minute, rl.config.RequestsPerMinute) {
		return false
	}

	// Check hour limit
	if !rl.checkWindow(identifier, "hour", now, time.Hour, rl.config.RequestsPerHour) {
		return false
	}

	// Check day limit
	if !rl.checkWindow(identifier, "day", now, 24*time.Hour, rl.config.RequestsPerDay) {
		return false
	}

	// Increment counters
	rl.incrementCounters(identifier, now)
	return true
}

// checkWindow checks if request is within window limit
func (rl *RateLimitMiddleware) checkWindow(identifier, window string, now time.Time, duration time.Duration, limit int) bool {
	var windowMap map[string]*WindowCounter
	
	switch window {
	case "minute":
		windowMap = rl.counters.Minute
	case "hour":
		windowMap = rl.counters.Hour
	case "day":
		windowMap = rl.counters.Day
	default:
		return false
	}

	rl.counters.mutex.RLock()
	counter, exists := windowMap[identifier]
	rl.counters.mutex.RUnlock()

	if !exists {
		return true // First request
	}

	counter.mutex.RLock()
	defer counter.mutex.RUnlock()

	// Reset counter if window has expired
	if now.After(counter.ResetTime) {
		return true
	}

	return counter.Count < limit
}

// incrementCounters increments all relevant counters
func (rl *RateLimitMiddleware) incrementCounters(identifier string, now time.Time) {
	rl.counters.mutex.Lock()
	defer rl.counters.mutex.Unlock()

	// Increment minute counter
	rl.incrementCounter(rl.counters.Minute, identifier, now, time.Minute)
	
	// Increment hour counter
	rl.incrementCounter(rl.counters.Hour, identifier, now, time.Hour)
	
	// Increment day counter
	rl.incrementCounter(rl.counters.Day, identifier, now, 24*time.Hour)
}

// incrementCounter increments a specific window counter
func (rl *RateLimitMiddleware) incrementCounter(windowMap map[string]*WindowCounter, identifier string, now time.Time, duration time.Duration) {
	counter, exists := windowMap[identifier]
	
	if !exists {
		windowMap[identifier] = &WindowCounter{
			Count:     1,
			ResetTime: now.Add(duration),
		}
		return
	}

	counter.mutex.Lock()
	defer counter.mutex.Unlock()

	// Reset if window expired
	if now.After(counter.ResetTime) {
		counter.Count = 1
		counter.ResetTime = now.Add(duration)
	} else {
		counter.Count++
	}
}

// getRemainingRequests gets remaining requests for identifier
func (rl *RateLimitMiddleware) getRemainingRequests(identifier string) int {
	rl.counters.mutex.RLock()
	defer rl.counters.mutex.RUnlock()

	counter, exists := rl.counters.Minute[identifier]
	if !exists {
		return rl.config.RequestsPerMinute
	}

	counter.mutex.RLock()
	defer counter.mutex.RUnlock()

	remaining := rl.config.RequestsPerMinute - counter.Count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// getRetryAfter gets retry after time for identifier
func (rl *RateLimitMiddleware) getRetryAfter(identifier string) time.Time {
	rl.counters.mutex.RLock()
	defer rl.counters.mutex.RUnlock()

	// Check which window is blocking
	now := time.Now()
	
	// Check minute window
	if counter, exists := rl.counters.Minute[identifier]; exists {
		counter.mutex.RLock()
		if counter.Count >= rl.config.RequestsPerMinute && now.Before(counter.ResetTime) {
			resetTime := counter.ResetTime
			counter.mutex.RUnlock()
			return resetTime
		}
		counter.mutex.RUnlock()
	}

	// Check hour window
	if counter, exists := rl.counters.Hour[identifier]; exists {
		counter.mutex.RLock()
		if counter.Count >= rl.config.RequestsPerHour && now.Before(counter.ResetTime) {
			resetTime := counter.ResetTime
			counter.mutex.RUnlock()
			return resetTime
		}
		counter.mutex.RUnlock()
	}

	// Check day window
	if counter, exists := rl.counters.Day[identifier]; exists {
		counter.mutex.RLock()
		if counter.Count >= rl.config.RequestsPerDay && now.Before(counter.ResetTime) {
			resetTime := counter.ResetTime
			counter.mutex.RUnlock()
			return resetTime
		}
		counter.mutex.RUnlock()
	}

	return now.Add(time.Minute)
}

// getResetTime gets reset time for identifier
func (rl *RateLimitMiddleware) getResetTime(identifier string) time.Time {
	rl.counters.mutex.RLock()
	defer rl.counters.mutex.RUnlock()

	counter, exists := rl.counters.Minute[identifier]
	if !exists {
		return time.Now().Add(time.Minute)
	}

	counter.mutex.RLock()
	defer counter.mutex.RUnlock()
	return counter.ResetTime
}

// cleanup removes expired counters
func (rl *RateLimitMiddleware) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		
		rl.counters.mutex.Lock()
		
		// Clean minute counters
		for id, counter := range rl.counters.Minute {
			counter.mutex.RLock()
			if now.After(counter.ResetTime.Add(time.Minute)) {
				delete(rl.counters.Minute, id)
			}
			counter.mutex.RUnlock()
		}
		
		// Clean hour counters
		for id, counter := range rl.counters.Hour {
			counter.mutex.RLock()
			if now.After(counter.ResetTime.Add(time.Hour)) {
				delete(rl.counters.Hour, id)
			}
			counter.mutex.RUnlock()
		}
		
		// Clean day counters
		for id, counter := range rl.counters.Day {
			counter.mutex.RLock()
			if now.After(counter.ResetTime.Add(24 * time.Hour)) {
				delete(rl.counters.Day, id)
			}
			counter.mutex.RUnlock()
		}
		
		rl.counters.mutex.Unlock()
	}
}

// GetStats returns rate limiting statistics
func (rl *RateLimitMiddleware) GetStats() map[string]interface{} {
	rl.counters.mutex.RLock()
	defer rl.counters.mutex.RUnlock()

	stats := make(map[string]interface{})
	
	stats["active_minute_windows"] = len(rl.counters.Minute)
	stats["active_hour_windows"] = len(rl.counters.Hour)
	stats["active_day_windows"] = len(rl.counters.Day)
	
	// Get top consumers
	topMinute := make(map[string]int)
	for id, counter := range rl.counters.Minute {
		counter.mutex.RLock()
		topMinute[id] = counter.Count
		counter.mutex.RUnlock()
	}
	stats["top_minute_consumers"] = topMinute
	
	return stats
}

// IPWhitelist middleware to allow certain IPs to bypass rate limiting
func IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		for _, allowedIP := range allowedIPs {
			if clientIP == allowedIP {
				c.Set("skip_rate_limit", true)
				break
			}
		}
		
		c.Next()
	}
}

// DynamicRateLimit allows different limits based on user tier
func (rl *RateLimitMiddleware) DynamicRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if whitelisted
		if c.GetBool("skip_rate_limit") {
			c.Next()
			return
		}

		// Adjust limits based on user roles
		config := rl.config
		if roles, exists := c.Get("roles"); exists {
			if roleList, ok := roles.([]string); ok {
				for _, role := range roleList {
					switch role {
					case "admin":
						config.RequestsPerMinute *= 10
						config.RequestsPerHour *= 10
						config.RequestsPerDay *= 10
					case "premium":
						config.RequestsPerMinute *= 5
						config.RequestsPerHour *= 5
						config.RequestsPerDay *= 5
					case "pro":
						config.RequestsPerMinute *= 3
						config.RequestsPerHour *= 3
						config.RequestsPerDay *= 3
					}
				}
			}
		}

		// Apply rate limiting with adjusted config
		originalConfig := rl.config
		rl.config = config
		
		// Continue with normal rate limiting
		rl.RateLimit()(c)
		
		// Restore original config
		rl.config = originalConfig
	}
}