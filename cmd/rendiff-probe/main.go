// Rendiff Probe - Professional Video Analysis API
// Powered by FFprobe (FFmpeg)
//
// This API provides comprehensive video/audio file analysis with 19 professional
// quality control analysis categories. It uses FFprobe as its core media analysis engine.
//
// FFprobe is part of the FFmpeg project (https://ffmpeg.org/)
// and is licensed under the LGPL/GPL license.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/rendiffdev/rendiff-probe/internal/config"
	"github.com/rendiffdev/rendiff-probe/internal/database"
	"github.com/rendiffdev/rendiff-probe/internal/ffmpeg"
	"github.com/rendiffdev/rendiff-probe/internal/hls"
	"github.com/rendiffdev/rendiff-probe/internal/models"
	"github.com/rendiffdev/rendiff-probe/internal/services"
	"github.com/rendiffdev/rendiff-probe/internal/validator"
	"github.com/rendiffdev/rendiff-probe/pkg/logger"
	"github.com/rs/zerolog"
)

// Production constants
const (
	maxFileSize       = 5 * 1024 * 1024 * 1024 // 5GB max file size
	maxRequestBodyMB  = 10                      // 10MB max JSON request body
	maxBatchItems     = 100                     // Max items in batch processing
	defaultTimeout    = 60 * time.Second
	maxTimeout        = 30 * time.Minute
	shutdownTimeout   = 30 * time.Second
	wsReadBufferSize  = 1024
	wsWriteBufferSize = 1024
)

// Global instances for services
var (
	ffprobeInstance *ffmpeg.FFprobe
	hlsAnalyzer     *hls.HLSAnalyzer
	llmService      *services.LLMService
	appLogger       zerolog.Logger
	appConfig       *config.Config

	// Shutdown context for graceful termination
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc

	// WebSocket upgrader with secure origin checking
	wsUpgrader websocket.Upgrader

	// Active WebSocket connections for progress updates
	wsConnections = make(map[string]*websocket.Conn)
	wsLock        sync.RWMutex

	// Batch job status tracking
	batchJobs = make(map[string]*BatchJob)
	batchLock sync.RWMutex

	// File path validator
	fileValidator *validator.FilePathValidator
)

// BatchJob represents a batch processing job
type BatchJob struct {
	ID        string                   `json:"id"`
	Status    string                   `json:"status"`
	Total     int                      `json:"total"`
	Completed int                      `json:"completed"`
	Failed    int                      `json:"failed"`
	Results   []map[string]interface{} `json:"results"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
	ctx       context.Context
	cancel    context.CancelFunc
}

// ProgressUpdate represents a WebSocket progress message
type ProgressUpdate struct {
	Type      string  `json:"type"`
	JobID     string  `json:"job_id"`
	Progress  float64 `json:"progress"`
	Message   string  `json:"message"`
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	appConfig = cfg

	// Set Gin mode based on environment (CloudMode = development, !CloudMode = production)
	if !cfg.CloudMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize logger
	appLogger = logger.New(cfg.LogLevel)
	appLogger.Info().
		Bool("cloud_mode", cfg.CloudMode).
		Msg("Starting rendiff-probe with full feature set")

	// Initialize shutdown context
	shutdownCtx, shutdownCancel = context.WithCancel(context.Background())

	// Initialize file validator
	fileValidator = validator.NewFilePathValidator()

	// Initialize WebSocket upgrader with secure origin checking
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  wsReadBufferSize,
		WriteBufferSize: wsWriteBufferSize,
		CheckOrigin:     checkWebSocketOrigin,
	}

	// Initialize database
	db, err := database.New(cfg, appLogger)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Validate FFmpeg/FFprobe binary at startup
	appLogger.Info().Msg("Validating FFmpeg/FFprobe binaries...")
	ffprobeInstance = ffmpeg.NewFFprobe(cfg.FFprobePath, appLogger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ffprobeInstance.ValidateBinaryAtStartup(ctx); err != nil {
		appLogger.Fatal().
			Err(err).
			Str("ffprobe_path", cfg.FFprobePath).
			Msg("FFprobe binary validation failed")
	}

	// Initialize HLS Analyzer
	hlsAnalyzer = hls.NewHLSAnalyzer(appLogger)
	appLogger.Info().Msg("HLS Analyzer initialized")

	// Initialize LLM Service
	llmService = services.NewLLMService(cfg, appLogger)
	appLogger.Info().Msg("LLM Service initialized")

	appLogger.Info().Msg("All services initialized successfully")

	// Create Gin router with production settings
	router := gin.New()

	// Add recovery middleware with logging
	router.Use(gin.Recovery())

	// Add custom logging middleware
	router.Use(requestLoggingMiddleware())

	// Add security headers middleware
	router.Use(securityHeadersMiddleware())

	// Add request size limit middleware
	router.Use(requestSizeLimitMiddleware(maxRequestBodyMB * 1024 * 1024))

	// Setup routes
	setupRoutes(router, cfg)

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      5 * time.Minute, // Longer for file uploads
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	// Start server
	go func() {
		appLogger.Info().
			Int("port", cfg.Port).
			Bool("cloud_mode", cfg.CloudMode).
			Msg("Server starting with full feature set")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info().Msg("Shutting down server...")

	// Cancel all batch jobs
	shutdownCancel()
	cancelAllBatchJobs()

	// Close all WebSocket connections
	closeAllWebSocketConnections()

	ctx, cancel = context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Error().Err(err).Msg("Server forced to shutdown")
	}

	appLogger.Info().Msg("Server exited gracefully")
}

// checkWebSocketOrigin validates WebSocket connection origins
func checkWebSocketOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // Allow requests without origin (same-origin)
	}

	// In cloud/development mode, allow all origins (for testing)
	if appConfig.CloudMode {
		return true
	}

	// Check against allowed origins
	for _, allowed := range appConfig.AllowedOrigins {
		if allowed == "*" {
			appLogger.Warn().Msg("WebSocket allowing all origins - not recommended for production")
			return true
		}
		if allowed == origin {
			return true
		}
	}

	appLogger.Warn().
		Str("origin", origin).
		Msg("WebSocket connection rejected: origin not allowed")
	return false
}

// cancelAllBatchJobs cancels all running batch jobs during shutdown
func cancelAllBatchJobs() {
	batchLock.Lock()
	defer batchLock.Unlock()

	for id, job := range batchJobs {
		if job.cancel != nil && job.Status == "processing" {
			appLogger.Info().Str("job_id", id).Msg("Cancelling batch job")
			job.cancel()
			job.Status = "cancelled"
		}
	}
}

// closeAllWebSocketConnections closes all active WebSocket connections
func closeAllWebSocketConnections() {
	wsLock.Lock()
	defer wsLock.Unlock()

	for id, conn := range wsConnections {
		appLogger.Info().Str("job_id", id).Msg("Closing WebSocket connection")
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "Server shutting down"))
		conn.Close()
	}
	wsConnections = make(map[string]*websocket.Conn)
}

// requestLoggingMiddleware logs HTTP requests
func requestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		appLogger.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Str("client_ip", c.ClientIP()).
			Msg("HTTP request")
	}
}

// securityHeadersMiddleware adds security headers to responses
func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Remove server identification
		c.Header("Server", "")
		c.Header("X-Powered-By", "")

		c.Next()
	}
}

// requestSizeLimitMiddleware limits request body size
// Note: Multipart form requests (file uploads) are excluded - they use maxFileSize limit
func requestSizeLimitMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip limit for multipart form data (file uploads)
		contentType := c.GetHeader("Content-Type")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			// For file uploads, use the much larger maxFileSize limit
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)
		} else {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

func setupRoutes(router *gin.Engine, cfg *config.Config) {
	// Health check (no auth required)
	router.GET("/health", healthHandler)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// File probing
		v1.POST("/probe/file", probeFileHandler)

		// URL probing
		v1.POST("/probe/url", probeURLHandler)

		// HLS analysis
		v1.POST("/probe/hls", probeHLSHandler)

		// Batch processing
		v1.POST("/batch/analyze", batchAnalyzeHandler)
		v1.GET("/batch/status/:id", batchStatusHandler)

		// WebSocket for progress
		v1.GET("/ws/progress/:id", wsProgressHandler)
	}

	// GraphQL endpoint
	schema := createGraphQLSchema()
	graphqlHandler := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   appConfig.CloudMode, // Only enable pretty output in cloud/dev mode
		GraphiQL: appConfig.CloudMode, // Only enable GraphiQL in cloud/dev mode
	})
	router.POST("/api/v1/graphql", gin.WrapH(graphqlHandler))
	router.GET("/api/v1/graphql", gin.WrapH(graphqlHandler))
}

// Health check handler
func healthHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"service": "rendiff-probe",
		"version": "2.0.0",
		"features": gin.H{
			"file_probe":       true,
			"url_probe":        true,
			"hls_analysis":     true,
			"batch_processing": true,
			"websocket":        true,
			"graphql":          true,
			"llm_insights":     true,
		},
		"qc_tools": []string{
			"AFD Analysis", "Dead Pixel Detection", "PSE Flash Analysis",
			"HDR Analysis", "Audio Wrapping Analysis", "Endianness Detection",
			"Codec Analysis", "Container Validation", "Resolution Analysis",
			"Frame Rate Analysis", "Bitdepth Analysis", "Timecode Analysis",
			"MXF Analysis", "IMF Compliance", "Transport Stream Analysis",
			"Content Analysis", "Enhanced Analysis", "Stream Disposition Analysis",
			"Data Integrity Analysis",
		},
		"ffmpeg_validated": true,
		"timestamp":        time.Now(),
	})
}

// File probe handler with security validations
func probeFileHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > maxFileSize {
		c.JSON(413, gin.H{"error": "File too large", "max_size_bytes": maxFileSize})
		return
	}

	// Sanitize filename to prevent path traversal
	safeFilename := validator.SanitizeFilename(header.Filename)
	if safeFilename == "" {
		safeFilename = fmt.Sprintf("upload_%s", uuid.New().String()[:8])
	}

	// Check if LLM insights requested
	includeLLM := c.PostForm("include_llm") == "true"

	// Create temp file with sanitized name
	tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("ffprobe_%d_%s", time.Now().UnixNano(), safeFilename))
	tempFile, err := os.Create(tempPath)
	if err != nil {
		appLogger.Error().Err(err).Msg("Failed to create temporary file")
		c.JSON(500, gin.H{"error": "Failed to process file"})
		return
	}
	defer tempFile.Close()
	defer func() {
		if err := os.Remove(tempPath); err != nil {
			appLogger.Warn().Err(err).Str("path", tempPath).Msg("Failed to cleanup temp file")
		}
	}()

	// Copy file with size limit
	written, err := io.CopyN(tempFile, file, maxFileSize+1)
	if err != nil && err != io.EOF {
		appLogger.Error().Err(err).Msg("Failed to save uploaded file")
		c.JSON(500, gin.H{"error": "Failed to process file"})
		return
	}
	if written > maxFileSize {
		c.JSON(413, gin.H{"error": "File too large", "max_size_bytes": maxFileSize})
		return
	}

	// Perform analysis
	result, err := analyzeFile(c.Request.Context(), tempPath)
	if err != nil {
		appLogger.Error().Err(err).Str("filename", safeFilename).Msg("Analysis failed")
		c.JSON(500, gin.H{"error": "Analysis failed"})
		return
	}

	response := gin.H{
		"status":                 "success",
		"analysis_id":            uuid.New().String(),
		"filename":               safeFilename,
		"size":                   written,
		"analysis":               result,
		"qc_categories_analyzed": 19,
		"timestamp":              time.Now(),
	}

	// Add LLM insights if requested
	if includeLLM {
		llmReport, err := generateLLMInsights(c.Request.Context(), result, safeFilename)
		if err != nil {
			appLogger.Warn().Err(err).Msg("LLM insights generation failed")
			response["llm_error"] = "LLM analysis unavailable"
		} else {
			response["llm_report"] = llmReport
			response["llm_enabled"] = true
		}
	}

	c.JSON(200, response)
}

// URL probe handler with security validations
func probeURLHandler(c *gin.Context) {
	var request struct {
		URL        string `json:"url" binding:"required"`
		IncludeLLM bool   `json:"include_llm"`
		Timeout    int    `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Validate URL for security (SSRF prevention)
	if err := validator.ValidateURL(request.URL); err != nil {
		appLogger.Warn().Str("url", request.URL).Err(err).Msg("URL validation failed")
		c.JSON(400, gin.H{"error": "Invalid or blocked URL"})
		return
	}

	// Set timeout with bounds
	timeout := defaultTimeout
	if request.Timeout > 0 {
		timeout = time.Duration(request.Timeout) * time.Second
		if timeout > maxTimeout {
			timeout = maxTimeout
		}
	}

	// Download file from URL
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	tempPath, filename, err := downloadURL(ctx, request.URL)
	if err != nil {
		appLogger.Warn().Err(err).Str("url", request.URL).Msg("URL download failed")
		c.JSON(500, gin.H{"error": "Failed to download from URL"})
		return
	}
	defer func() {
		if err := os.Remove(tempPath); err != nil {
			appLogger.Warn().Err(err).Str("path", tempPath).Msg("Failed to cleanup temp file")
		}
	}()

	// Perform analysis
	result, err := analyzeFile(ctx, tempPath)
	if err != nil {
		appLogger.Error().Err(err).Msg("Analysis failed")
		c.JSON(500, gin.H{"error": "Analysis failed"})
		return
	}

	response := gin.H{
		"status":                 "success",
		"analysis_id":            uuid.New().String(),
		"url":                    request.URL,
		"filename":               filename,
		"analysis":               result,
		"qc_categories_analyzed": 19,
		"timestamp":              time.Now(),
	}

	// Add LLM insights if requested
	if request.IncludeLLM {
		llmReport, err := generateLLMInsights(ctx, result, filename)
		if err != nil {
			response["llm_error"] = "LLM analysis unavailable"
		} else {
			response["llm_report"] = llmReport
			response["llm_enabled"] = true
		}
	}

	c.JSON(200, response)
}

// HLS probe handler with validation
func probeHLSHandler(c *gin.Context) {
	var request struct {
		ManifestURL         string `json:"manifest_url" binding:"required"`
		AnalyzeSegments     bool   `json:"analyze_segments"`
		AnalyzeQuality      bool   `json:"analyze_quality"`
		ValidateCompliance  bool   `json:"validate_compliance"`
		PerformanceAnalysis bool   `json:"performance_analysis"`
		MaxSegments         int    `json:"max_segments"`
		IncludeLLM          bool   `json:"include_llm"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Validate URL
	if err := validator.ValidateURL(request.ManifestURL); err != nil {
		c.JSON(400, gin.H{"error": "Invalid or blocked URL"})
		return
	}

	hlsRequest := &hls.HLSAnalysisRequest{
		ManifestURL:         request.ManifestURL,
		AnalyzeSegments:     request.AnalyzeSegments,
		AnalyzeQuality:      request.AnalyzeQuality,
		ValidateCompliance:  request.ValidateCompliance,
		PerformanceAnalysis: request.PerformanceAnalysis,
		MaxSegments:         request.MaxSegments,
	}

	if hlsRequest.MaxSegments <= 0 || hlsRequest.MaxSegments > 100 {
		hlsRequest.MaxSegments = 10
	}

	result, err := hlsAnalyzer.AnalyzeHLS(c.Request.Context(), hlsRequest)
	if err != nil {
		appLogger.Error().Err(err).Msg("HLS analysis failed")
		c.JSON(500, gin.H{"error": "HLS analysis failed"})
		return
	}

	response := gin.H{
		"status":          "success",
		"analysis_id":     result.ID.String(),
		"manifest_url":    request.ManifestURL,
		"analysis":        result.Analysis,
		"processing_time": result.ProcessingTime.String(),
		"timestamp":       time.Now(),
	}

	c.JSON(200, response)
}

// Batch analyze handler with validation and limits
func batchAnalyzeHandler(c *gin.Context) {
	var request struct {
		Files      []string `json:"files"`
		URLs       []string `json:"urls"`
		IncludeLLM bool     `json:"include_llm"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	total := len(request.Files) + len(request.URLs)
	if total == 0 {
		c.JSON(400, gin.H{"error": "No files or URLs provided"})
		return
	}

	// Enforce batch size limit
	if total > maxBatchItems {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Batch size exceeds limit of %d items", maxBatchItems)})
		return
	}

	// Validate all URLs upfront
	for _, url := range request.URLs {
		if err := validator.ValidateURL(url); err != nil {
			c.JSON(400, gin.H{"error": "Invalid or blocked URL", "url": url})
			return
		}
	}

	// Validate file paths
	for _, filePath := range request.Files {
		if err := fileValidator.ValidateFilePath(filePath); err != nil {
			c.JSON(400, gin.H{"error": "Invalid file path", "path": filePath})
			return
		}
	}

	// Create batch job with cancellation context
	jobCtx, jobCancel := context.WithCancel(shutdownCtx)
	jobID := uuid.New().String()
	job := &BatchJob{
		ID:        jobID,
		Status:    "processing",
		Total:     total,
		Completed: 0,
		Failed:    0,
		Results:   make([]map[string]interface{}, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ctx:       jobCtx,
		cancel:    jobCancel,
	}

	batchLock.Lock()
	batchJobs[jobID] = job
	batchLock.Unlock()

	// Process in background with cancellation support
	go processBatchJob(job, request.Files, request.URLs, request.IncludeLLM)

	c.JSON(202, gin.H{
		"status":     "accepted",
		"job_id":     jobID,
		"total":      total,
		"message":    "Batch job started",
		"status_url": fmt.Sprintf("/api/v1/batch/status/%s", jobID),
		"ws_url":     fmt.Sprintf("/api/v1/ws/progress/%s", jobID),
	})
}

// Batch status handler
func batchStatusHandler(c *gin.Context) {
	jobID := c.Param("id")

	// Validate UUID format
	if _, err := uuid.Parse(jobID); err != nil {
		c.JSON(400, gin.H{"error": "Invalid job ID format"})
		return
	}

	batchLock.RLock()
	job, exists := batchJobs[jobID]
	batchLock.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Job not found"})
		return
	}

	// Return job status without internal fields
	c.JSON(200, gin.H{
		"id":         job.ID,
		"status":     job.Status,
		"total":      job.Total,
		"completed":  job.Completed,
		"failed":     job.Failed,
		"results":    job.Results,
		"created_at": job.CreatedAt,
		"updated_at": job.UpdatedAt,
	})
}

// WebSocket progress handler
func wsProgressHandler(c *gin.Context) {
	jobID := c.Param("id")

	// Validate UUID format
	if _, err := uuid.Parse(jobID); err != nil {
		c.JSON(400, gin.H{"error": "Invalid job ID format"})
		return
	}

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		appLogger.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}
	defer conn.Close()

	// Set connection limits
	conn.SetReadLimit(512) // Small limit for ping/pong
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	wsLock.Lock()
	wsConnections[jobID] = conn
	wsLock.Unlock()

	defer func() {
		wsLock.Lock()
		delete(wsConnections, jobID)
		wsLock.Unlock()
	}()

	// Send initial status
	batchLock.RLock()
	job, exists := batchJobs[jobID]
	batchLock.RUnlock()

	if exists {
		progress := float64(job.Completed) / float64(job.Total) * 100
		sendProgressUpdate(jobID, progress, job.Status, "Connected to progress stream")
	}

	// Keep connection alive with ping/pong
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-shutdownCtx.Done():
			return
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		default:
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}
}

// Helper functions

func analyzeFile(ctx context.Context, filePath string) (*ffmpeg.FFprobeResult, error) {
	options := ffmpeg.NewOptionsBuilder().
		Input(filePath).
		JSON().
		ShowAll().
		ShowError().
		ShowDataHash().
		ShowPrivateData().
		CountFrames().
		CountPackets().
		ErrorDetectBroadcast().
		FormatErrorDetectAll().
		CRC32Hash().
		ProbeSizeMB(100).
		AnalyzeDurationSeconds(60).
		Build()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	return ffprobeInstance.Probe(ctx, options)
}

func downloadURL(ctx context.Context, urlStr string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set reasonable headers
	req.Header.Set("User-Agent", "rendiff-probe/2.0")

	client := &http.Client{
		Timeout: 5 * time.Minute,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			// Validate redirect URL
			if err := validator.ValidateURL(req.URL.String()); err != nil {
				return fmt.Errorf("redirect blocked: %w", err)
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Check content length
	if resp.ContentLength > maxFileSize {
		return "", "", fmt.Errorf("file too large: %d bytes", resp.ContentLength)
	}

	// Extract and sanitize filename
	filename := extractFilename(urlStr, resp.Header.Get("Content-Disposition"))
	safeFilename := validator.SanitizeFilename(filename)
	if safeFilename == "" {
		safeFilename = fmt.Sprintf("download_%s", uuid.New().String()[:8])
	}

	tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("ffprobe_%d_%s", time.Now().UnixNano(), safeFilename))
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Copy with size limit
	written, err := io.CopyN(tempFile, resp.Body, maxFileSize+1)
	if err != nil && err != io.EOF {
		os.Remove(tempPath)
		return "", "", fmt.Errorf("failed to save file: %w", err)
	}
	if written > maxFileSize {
		os.Remove(tempPath)
		return "", "", fmt.Errorf("file too large: %d bytes", written)
	}

	return tempPath, safeFilename, nil
}

// extractFilename safely extracts filename from URL or Content-Disposition
func extractFilename(urlStr, contentDisposition string) string {
	// Try Content-Disposition first
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			if filename, ok := params["filename"]; ok {
				return filename
			}
		}
	}

	// Fall back to URL path
	return filepath.Base(strings.Split(urlStr, "?")[0])
}

func generateLLMInsights(ctx context.Context, result *ffmpeg.FFprobeResult, filename string) (string, error) {
	// Create analysis model from FFprobe result
	analysis := &models.Analysis{
		ID:       uuid.New(),
		FileName: filename,
		Status:   models.StatusCompleted,
	}

	// Convert FFprobe result components to JSON for FFprobeData
	if result.Format != nil {
		formatJSON, err := json.Marshal(result.Format)
		if err != nil {
			appLogger.Warn().Err(err).Msg("Failed to marshal format data")
		} else {
			analysis.FFprobeData.Format = formatJSON
		}
	}
	if result.Streams != nil {
		streamsJSON, err := json.Marshal(result.Streams)
		if err != nil {
			appLogger.Warn().Err(err).Msg("Failed to marshal streams data")
		} else {
			analysis.FFprobeData.Streams = streamsJSON
		}
	}

	return llmService.GenerateAnalysis(ctx, analysis)
}

func processBatchJob(job *BatchJob, files []string, urls []string, includeLLM bool) {
	ctx := job.ctx

	// Process files
	for _, filePath := range files {
		select {
		case <-ctx.Done():
			appLogger.Info().Str("job_id", job.ID).Msg("Batch job cancelled")
			batchLock.Lock()
			job.Status = "cancelled"
			job.UpdatedAt = time.Now()
			batchLock.Unlock()
			return
		default:
		}

		result, err := analyzeFile(ctx, filePath)

		batchLock.Lock()
		if err != nil {
			job.Failed++
			job.Results = append(job.Results, map[string]interface{}{
				"type":   "file",
				"path":   filePath,
				"status": "failed",
				"error":  "Analysis failed",
			})
		} else {
			job.Completed++
			resultMap := map[string]interface{}{
				"type":     "file",
				"path":     filePath,
				"status":   "success",
				"analysis": result,
			}
			if includeLLM {
				llmReport, err := generateLLMInsights(ctx, result, filepath.Base(filePath))
				if err == nil {
					resultMap["llm_report"] = llmReport
				}
			}
			job.Results = append(job.Results, resultMap)
		}
		job.UpdatedAt = time.Now()
		batchLock.Unlock()

		// Send progress update
		progress := float64(job.Completed+job.Failed) / float64(job.Total) * 100
		sendProgressUpdate(job.ID, progress, "processing", fmt.Sprintf("Processed: %s", filepath.Base(filePath)))
	}

	// Process URLs
	for _, url := range urls {
		select {
		case <-ctx.Done():
			appLogger.Info().Str("job_id", job.ID).Msg("Batch job cancelled")
			batchLock.Lock()
			job.Status = "cancelled"
			job.UpdatedAt = time.Now()
			batchLock.Unlock()
			return
		default:
		}

		tempPath, filename, err := downloadURL(ctx, url)
		if err != nil {
			batchLock.Lock()
			job.Failed++
			job.Results = append(job.Results, map[string]interface{}{
				"type":   "url",
				"url":    url,
				"status": "failed",
				"error":  "Download failed",
			})
			job.UpdatedAt = time.Now()
			batchLock.Unlock()

			progress := float64(job.Completed+job.Failed) / float64(job.Total) * 100
			sendProgressUpdate(job.ID, progress, "processing", fmt.Sprintf("Failed: %s", url))
			continue
		}

		result, err := analyzeFile(ctx, tempPath)
		if removeErr := os.Remove(tempPath); removeErr != nil {
			appLogger.Warn().Err(removeErr).Str("path", tempPath).Msg("Failed to cleanup temp file")
		}

		batchLock.Lock()
		if err != nil {
			job.Failed++
			job.Results = append(job.Results, map[string]interface{}{
				"type":   "url",
				"url":    url,
				"status": "failed",
				"error":  "Analysis failed",
			})
		} else {
			job.Completed++
			resultMap := map[string]interface{}{
				"type":     "url",
				"url":      url,
				"filename": filename,
				"status":   "success",
				"analysis": result,
			}
			if includeLLM {
				llmReport, err := generateLLMInsights(ctx, result, filename)
				if err == nil {
					resultMap["llm_report"] = llmReport
				}
			}
			job.Results = append(job.Results, resultMap)
		}
		job.UpdatedAt = time.Now()
		batchLock.Unlock()

		progress := float64(job.Completed+job.Failed) / float64(job.Total) * 100
		sendProgressUpdate(job.ID, progress, "processing", fmt.Sprintf("Processed: %s", filename))
	}

	// Mark job as completed
	batchLock.Lock()
	job.Status = "completed"
	job.UpdatedAt = time.Now()
	batchLock.Unlock()

	sendProgressUpdate(job.ID, 100, "completed", "Batch processing completed")
}

func sendProgressUpdate(jobID string, progress float64, status, message string) {
	wsLock.RLock()
	conn, exists := wsConnections[jobID]
	wsLock.RUnlock()

	if !exists {
		return
	}

	update := ProgressUpdate{
		Type:      "progress",
		JobID:     jobID,
		Progress:  progress,
		Message:   message,
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := conn.WriteJSON(update); err != nil {
		appLogger.Warn().Err(err).Str("job_id", jobID).Msg("Failed to send WebSocket update")
	}
}

// GraphQL Schema
func createGraphQLSchema() graphql.Schema {
	// Define stream type
	streamType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Stream",
		Fields: graphql.Fields{
			"index":       &graphql.Field{Type: graphql.Int},
			"codec_name":  &graphql.Field{Type: graphql.String},
			"codec_type":  &graphql.Field{Type: graphql.String},
			"width":       &graphql.Field{Type: graphql.Int},
			"height":      &graphql.Field{Type: graphql.Int},
			"sample_rate": &graphql.Field{Type: graphql.String},
			"channels":    &graphql.Field{Type: graphql.Int},
		},
	})

	// Define format type
	formatType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Format",
		Fields: graphql.Fields{
			"filename":         &graphql.Field{Type: graphql.String},
			"nb_streams":       &graphql.Field{Type: graphql.Int},
			"format_name":      &graphql.Field{Type: graphql.String},
			"format_long_name": &graphql.Field{Type: graphql.String},
			"duration":         &graphql.Field{Type: graphql.String},
			"size":             &graphql.Field{Type: graphql.String},
			"bit_rate":         &graphql.Field{Type: graphql.String},
		},
	})

	// Define analysis result type
	analysisType := graphql.NewObject(graphql.ObjectConfig{
		Name: "AnalysisResult",
		Fields: graphql.Fields{
			"id":          &graphql.Field{Type: graphql.String},
			"filename":    &graphql.Field{Type: graphql.String},
			"status":      &graphql.Field{Type: graphql.String},
			"streams":     &graphql.Field{Type: graphql.NewList(streamType)},
			"format":      &graphql.Field{Type: formatType},
			"llm_report":  &graphql.Field{Type: graphql.String},
			"llm_enabled": &graphql.Field{Type: graphql.Boolean},
			"timestamp":   &graphql.Field{Type: graphql.String},
		},
	})

	// Define query
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"health": &graphql.Field{
				Type: graphql.NewObject(graphql.ObjectConfig{
					Name: "Health",
					Fields: graphql.Fields{
						"status":  &graphql.Field{Type: graphql.String},
						"version": &graphql.Field{Type: graphql.String},
					},
				}),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return map[string]interface{}{
						"status":  "healthy",
						"version": "2.0.0",
					}, nil
				},
			},
		},
	})

	// Define mutation with URL validation
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"analyzeURL": &graphql.Field{
				Type: analysisType,
				Args: graphql.FieldConfigArgument{
					"url": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"include_llm": &graphql.ArgumentConfig{
						Type:         graphql.Boolean,
						DefaultValue: false,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					url := p.Args["url"].(string)

					// Validate URL
					if err := validator.ValidateURL(url); err != nil {
						return nil, fmt.Errorf("invalid or blocked URL")
					}

					includeLLM := false
					if v, ok := p.Args["include_llm"].(bool); ok {
						includeLLM = v
					}

					ctx := p.Context
					tempPath, filename, err := downloadURL(ctx, url)
					if err != nil {
						return nil, fmt.Errorf("failed to download URL")
					}
					defer func() {
						if err := os.Remove(tempPath); err != nil {
							appLogger.Warn().Err(err).Str("path", tempPath).Msg("Failed to cleanup temp file")
						}
					}()

					result, err := analyzeFile(ctx, tempPath)
					if err != nil {
						return nil, fmt.Errorf("analysis failed")
					}

					response := map[string]interface{}{
						"id":          uuid.New().String(),
						"filename":    filename,
						"status":      "completed",
						"streams":     result.Streams,
						"format":      result.Format,
						"llm_enabled": false,
						"timestamp":   time.Now().Format(time.RFC3339),
					}

					if includeLLM {
						llmReport, err := generateLLMInsights(ctx, result, filename)
						if err == nil {
							response["llm_report"] = llmReport
							response["llm_enabled"] = true
						}
					}

					return response, nil
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to create GraphQL schema")
	}

	return schema
}
