package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
)

// MonitoringMiddleware handles metrics collection and monitoring
type MonitoringMiddleware struct {
	logger zerolog.Logger
}

// Prometheus metrics
var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000},
		},
		[]string{"method", "endpoint"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000},
		},
		[]string{"method", "endpoint"},
	)

	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_connections",
			Help: "Number of active HTTP connections",
		},
	)

	ffprobeAnalysesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ffprobe_analyses_total",
			Help: "Total number of ffprobe analyses",
		},
		[]string{"status", "source_type"},
	)

	ffprobeAnalysisDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ffprobe_analysis_duration_seconds",
			Help:    "FFprobe analysis duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 300, 600},
		},
		[]string{"source_type"},
	)

	uploadedFilesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uploaded_files_total",
			Help: "Total number of uploaded files",
		},
		[]string{"status"},
	)

	uploadedFileSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "uploaded_file_size_bytes",
			Help:    "Size of uploaded files in bytes",
			Buckets: []float64{1e6, 10e6, 100e6, 1e9, 10e9, 50e9}, // 1MB to 50GB
		},
	)

	batchProcessingJobs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "batch_processing_jobs",
			Help: "Number of batch processing jobs by status",
		},
		[]string{"status"},
	)

	websocketConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_connections_active",
			Help: "Number of active WebSocket connections",
		},
	)

	rateLimitExceeded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Total number of rate limit exceeded events",
		},
		[]string{"identifier_type"},
	)

	authFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_failures_total",
			Help: "Total number of authentication failures",
		},
		[]string{"reason"},
	)
)

// NewMonitoringMiddleware creates a new monitoring middleware
func NewMonitoringMiddleware(logger zerolog.Logger) *MonitoringMiddleware {
	return &MonitoringMiddleware{
		logger: logger,
	}
}

// Metrics middleware collects HTTP metrics
func (mm *MonitoringMiddleware) Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Increment active connections
		activeConnections.Inc()
		defer activeConnections.Dec()

		// Get request size
		requestSize := float64(c.Request.ContentLength)
		if requestSize < 0 {
			requestSize = 0
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()
		
		// Get normalized endpoint for metrics (remove IDs)
		endpoint := normalizeEndpoint(c.FullPath())
		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		// Record metrics
		httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
		
		if requestSize > 0 {
			httpRequestSize.WithLabelValues(method, endpoint).Observe(requestSize)
		}
		
		responseSize := float64(c.Writer.Size())
		if responseSize > 0 {
			httpResponseSize.WithLabelValues(method, endpoint).Observe(responseSize)
		}
	}
}

// FFprobeMetrics records ffprobe-specific metrics
func (mm *MonitoringMiddleware) FFprobeMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to ffprobe endpoints
		if !isFFprobeEndpoint(c.FullPath()) {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		// Record analysis metrics
		status := "unknown"
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			status = "success"
		} else if c.Writer.Status() >= 400 && c.Writer.Status() < 500 {
			status = "client_error"
		} else if c.Writer.Status() >= 500 {
			status = "server_error"
		}

		sourceType := "unknown"
		if c.FullPath() == "/api/v1/probe/file" {
			sourceType = "file"
		} else if c.FullPath() == "/api/v1/probe/url" {
			sourceType = "url"
		} else if c.FullPath() == "/api/v1/stream/live" {
			sourceType = "stream"
		}

		ffprobeAnalysesTotal.WithLabelValues(status, sourceType).Inc()
		ffprobeAnalysisDuration.WithLabelValues(sourceType).Observe(duration)
	}
}

// UploadMetrics records upload-specific metrics
func (mm *MonitoringMiddleware) UploadMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to upload endpoints
		if !isUploadEndpoint(c.FullPath()) {
			c.Next()
			return
		}

		c.Next()

		// Record upload metrics
		status := "unknown"
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			status = "success"
			
			// Try to get file size from context (set by upload handler)
			if fileSize, exists := c.Get("upload_file_size"); exists {
				if size, ok := fileSize.(int64); ok {
					uploadedFileSize.Observe(float64(size))
				}
			}
		} else if c.Writer.Status() >= 400 && c.Writer.Status() < 500 {
			status = "client_error"
		} else if c.Writer.Status() >= 500 {
			status = "server_error"
		}

		uploadedFilesTotal.WithLabelValues(status).Inc()
	}
}

// WebSocketMetrics tracks WebSocket connections
func WebSocketConnected() {
	websocketConnections.Inc()
}

func WebSocketDisconnected() {
	websocketConnections.Dec()
}

// BatchMetrics updates batch processing metrics
func BatchJobStarted(status string) {
	batchProcessingJobs.WithLabelValues(status).Inc()
}

func BatchJobCompleted(oldStatus, newStatus string) {
	batchProcessingJobs.WithLabelValues(oldStatus).Dec()
	if newStatus != "" {
		batchProcessingJobs.WithLabelValues(newStatus).Inc()
	}
}

// RateLimitMetrics records rate limiting events
func RateLimitExceeded(identifierType string) {
	rateLimitExceeded.WithLabelValues(identifierType).Inc()
}

// AuthMetrics records authentication events
func AuthFailure(reason string) {
	authFailures.WithLabelValues(reason).Inc()
}

// Helper functions

func normalizeEndpoint(path string) string {
	// Replace UUID patterns with placeholder
	// This is a simple implementation - in production you might want
	// to use regex for more sophisticated ID detection
	normalized := path
	
	// Common patterns to normalize
	patterns := map[string]string{
		"/api/v1/probe/status/":   "/api/v1/probe/status/:id",
		"/api/v1/batch/status/":   "/api/v1/batch/status/:id",
		"/api/v1/upload/status/":  "/api/v1/upload/status/:id",
		"/api/v1/stream/progress/": "/api/v1/stream/progress/:id",
		"/api/v1/probe/analyses/": "/api/v1/probe/analyses/:id",
		"/api/v1/batch/":          "/api/v1/batch/:id",
	}
	
	for pattern, replacement := range patterns {
		if len(normalized) > len(pattern) && normalized[:len(pattern)] == pattern {
			normalized = replacement
			break
		}
	}
	
	return normalized
}

func isFFprobeEndpoint(path string) bool {
	ffprobeEndpoints := []string{
		"/api/v1/probe/file",
		"/api/v1/probe/url",
		"/api/v1/probe/quick",
		"/api/v1/stream/live",
		"/api/v1/batch/analyze",
	}
	
	for _, endpoint := range ffprobeEndpoints {
		if path == endpoint {
			return true
		}
	}
	return false
}

func isUploadEndpoint(path string) bool {
	uploadEndpoints := []string{
		"/api/v1/upload",
		"/api/v1/upload/chunk",
	}
	
	for _, endpoint := range uploadEndpoints {
		if path == endpoint {
			return true
		}
	}
	return false
}

// HealthMetrics provides custom health metrics
func HealthMetrics() map[string]interface{} {
	return map[string]interface{}{
		"active_connections":    getMetricValue(activeConnections),
		"websocket_connections": getMetricValue(websocketConnections),
		"total_requests":        getCounterValue(httpRequestsTotal),
		"total_analyses":        getCounterValue(ffprobeAnalysesTotal),
		"total_uploads":         getCounterValue(uploadedFilesTotal),
		"rate_limit_exceeded":   getCounterValue(rateLimitExceeded),
		"auth_failures":         getCounterValue(authFailures),
	}
}

func getMetricValue(gauge prometheus.Gauge) float64 {
	// Simplified implementation - in practice you'd need prometheus dto package
	return 0.0
}

func getCounterValue(counter prometheus.CounterVec) float64 {
	// This is a simplified version - in practice you'd need to sum all label combinations
	return 0.0 // Implementation would require iterating through metric families
}

// CustomMetrics allows handlers to record custom metrics
func RecordCustomMetric(name string, value float64, labels map[string]string) {
	// This would be implemented based on specific requirements
	// For now, log the custom metric
	logger := zerolog.New(nil)
	logEvent := logger.Info().
		Str("metric_name", name).
		Float64("value", value)
	
	for key, val := range labels {
		logEvent.Str(key, val)
	}
	
	logEvent.Msg("Custom metric recorded")
}