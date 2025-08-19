package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Recovery middleware with custom panic recovery
func Recovery(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := debug.Stack()

				// Log the panic
				logger.Error().
					Interface("error", err).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Str("ip", c.ClientIP()).
					Bytes("stack", stack).
					Msg("Panic recovered")

				// Return error response
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"code":    "INTERNAL_ERROR",
					"message": fmt.Sprintf("An unexpected error occurred: %v", err),
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}
