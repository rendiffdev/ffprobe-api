package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// SessionIDKey is the context key for session ID
	SessionIDKey ContextKey = "session_id"
)

// Config holds logger configuration
type Config struct {
	Level      string
	Format     string // "json" or "console"
	Output     string // "stdout", "stderr", or file path
	TimeFormat string
	Structured bool
}

// New creates a new logger with the specified level and configuration
func New(level string) zerolog.Logger {
	return NewWithConfig(Config{
		Level:      level,
		Format:     "json",
		Output:     "stderr",
		TimeFormat: time.RFC3339,
		Structured: true,
	})
}

// NewWithConfig creates a new logger with custom configuration
func NewWithConfig(cfg Config) zerolog.Logger {
	// Configure time format
	if cfg.TimeFormat != "" {
		zerolog.TimeFieldFormat = cfg.TimeFormat
	} else {
		zerolog.TimeFieldFormat = time.RFC3339
	}

	// Configure output
	var output *os.File
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		output = os.Stderr
		// Could add file output support here
	}

	// Set up logger based on environment and format
	var logger zerolog.Logger
	if cfg.Format == "console" || (strings.ToLower(os.Getenv("GO_ENV")) != "production" && cfg.Format != "json") {
		// Console output with color and human-readable format
		consoleWriter := zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: "2006-01-02 15:04:05",
			FormatLevel: func(i interface{}) string {
				return strings.ToUpper(fmt.Sprintf("| %-5s |", i))
			},
			FormatMessage: func(i interface{}) string {
				return fmt.Sprintf("%-50s", i)
			},
			FormatFieldName: func(i interface{}) string {
				return fmt.Sprintf("%s:", i)
			},
			FormatFieldValue: func(i interface{}) string {
				return strings.ToUpper(fmt.Sprintf("%s", i))
			},
		}
		logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	} else {
		// JSON structured logging for production
		logger = zerolog.New(output).With().Timestamp().Logger()
	}

	// Parse log level
	logLevel, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Set global log level
	zerolog.SetGlobalLevel(logLevel)

	// Add service information
	logger = logger.With().
		Str("service", "rendiff-probe").
		Str("version", getVersion()).
		Logger()

	return logger
}

// WithRequestID adds a request ID to the logger context
func WithRequestID(logger zerolog.Logger, requestID string) zerolog.Logger {
	return logger.With().Str("request_id", requestID).Logger()
}

// WithUserID adds a user ID to the logger context
func WithUserID(logger zerolog.Logger, userID string) zerolog.Logger {
	return logger.With().Str("user_id", userID).Logger()
}

// WithSessionID adds a session ID to the logger context
func WithSessionID(logger zerolog.Logger, sessionID string) zerolog.Logger {
	return logger.With().Str("session_id", sessionID).Logger()
}

// WithContext adds context information to the logger
func WithContext(logger zerolog.Logger, ctx context.Context) zerolog.Logger {
	contextLogger := logger

	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		contextLogger = contextLogger.With().Str("request_id", requestID.(string)).Logger()
	}

	if userID := ctx.Value(UserIDKey); userID != nil {
		contextLogger = contextLogger.With().Str("user_id", userID.(string)).Logger()
	}

	if sessionID := ctx.Value(SessionIDKey); sessionID != nil {
		contextLogger = contextLogger.With().Str("session_id", sessionID.(string)).Logger()
	}

	return contextLogger
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to gin context
		c.Set("request_id", requestID)

		// Add to Go context
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// Add to response header
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// LoggingMiddleware logs HTTP requests with structured information
func LoggingMiddleware(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		end := time.Now()
		latency := end.Sub(start)

		// Get request ID from context
		requestID := c.GetString("request_id")

		// Create structured log entry
		logEvent := logger.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Int("status", c.Writer.Status()).
			Int("body_size", c.Writer.Size()).
			Dur("latency", latency).
			Str("request_id", requestID)

		// Add query parameters if present
		if raw != "" {
			logEvent.Str("query", raw)
		}

		// Add user context if available
		if userID := c.GetString("user_id"); userID != "" {
			logEvent.Str("user_id", userID)
		}

		if username := c.GetString("username"); username != "" {
			logEvent.Str("username", username)
		}

		// Add authentication method
		if authType := c.GetString("auth_type"); authType != "" {
			logEvent.Str("auth_type", authType)
		}

		// Add forwarded headers for proxy information
		if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
			logEvent.Str("x_forwarded_for", forwardedFor)
		}

		if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
			logEvent.Str("x_real_ip", realIP)
		}

		// Log errors if present
		if len(c.Errors) > 0 {
			logEvent.Str("errors", c.Errors.String())
		}

		// Determine log level based on status code
		switch {
		case c.Writer.Status() >= 500:
			logEvent = logger.Error().
				Str("method", c.Request.Method).
				Str("path", path).
				Str("ip", c.ClientIP()).
				Int("status", c.Writer.Status()).
				Dur("latency", latency).
				Str("request_id", requestID)
		case c.Writer.Status() >= 400:
			logEvent = logger.Warn().
				Str("method", c.Request.Method).
				Str("path", path).
				Str("ip", c.ClientIP()).
				Int("status", c.Writer.Status()).
				Dur("latency", latency).
				Str("request_id", requestID)
		}

		logEvent.Msg("HTTP request processed")
	}
}

// getVersion returns the application version
func getVersion() string {
	if version := os.Getenv("APP_VERSION"); version != "" {
		return version
	}
	return "development"
}
