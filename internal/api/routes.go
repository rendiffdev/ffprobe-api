package api

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/cache"
	"github.com/rendiffdev/ffprobe-api/internal/config"
	"github.com/rendiffdev/ffprobe-api/internal/database"
	"github.com/rendiffdev/ffprobe-api/internal/handlers"
	"github.com/rendiffdev/ffprobe-api/internal/hls"
	"github.com/rendiffdev/ffprobe-api/internal/middleware"
	"github.com/rendiffdev/ffprobe-api/internal/quality"
	"github.com/rendiffdev/ffprobe-api/internal/repositories"
	"github.com/rendiffdev/ffprobe-api/internal/services"
	"github.com/rendiffdev/ffprobe-api/internal/storage"
)


// Router contains all the route handlers
type Router struct {
	probeHandler      *handlers.ProbeHandler
	uploadHandler     *handlers.UploadHandler
	batchHandler      *handlers.BatchHandler
	streamHandler     *handlers.StreamHandler
	authHandler       *handlers.AuthHandler
	adminHandler      *handlers.AdminHandler
	qualityHandler    *handlers.QualityHandler
	hlsHandler        *handlers.HLSHandler
	reportHandler     *handlers.ReportHandler
	reportsHandler    *handlers.ReportsHandler
	compareHandler    *handlers.CompareHandler
	comparisonHandler *handlers.ComparisonHandler
	rawHandler        *handlers.RawHandler
	storageHandler    *handlers.StorageHandler
	graphqlHandler    *handlers.GraphQLHandler
	apiKeyHandler     *handlers.APIKeyHandler
	authMiddleware    *middleware.AuthMiddleware
	rateLimiter       *middleware.RateLimitMiddleware
	tenantRateLimiter *middleware.TenantRateLimiter
	security          *middleware.SecurityMiddleware
	monitoring        *middleware.MonitoringMiddleware
	cacheClient       *cache.ValkeyClient
	logger            zerolog.Logger
	db               *database.DB
	config           *config.Config
}

// NewRouter creates a new router with all handlers
func NewRouter(cfg *config.Config, db *database.DB, logger zerolog.Logger) *Router {
	// Create cache client (optional for rate limiting and caching)
	var cacheClient cache.Client
	
	if cfg.EnableRateLimit || cfg.ValkeyHost != "" {
		var err error
		cacheClient, err = cache.New(cfg, logger)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to connect to Valkey, continuing without cache")
		} else {
			logger.Info().Msg("Valkey cache enabled")
		}
	}
	// Create analysis service
	analysisService := services.NewAnalysisService(db, cfg.FFprobePath, logger)
	
	// Create report generator
	reportGenerator := services.NewReportGenerator(cfg.ReportsDir, cfg.BaseURL, 24) // 24 hour expiry

	// Create quality service
	qualityAnalyzer := quality.NewQualityAnalyzer(cfg.FFmpegPath, logger)
	qualityRepo := database.NewQualityRepository(db.DB)
	qualityService := services.NewQualityService(qualityAnalyzer, qualityRepo, db, logger)
	
	// Create HLS service
	hlsAnalyzer := hls.NewAnalyzer(cfg.FFprobePath, cfg.FFmpegPath, logger)
	hlsService := services.NewHLSService(db, hlsAnalyzer, logger)
	
	// Create report service
	reportService := services.NewReportService(db, analysisService, cfg.ReportsDir, logger)
	
	// Create LLM service
	llmService := services.NewLLMService(cfg, logger)
	
	// Create comparison repositories and service
	analysisRepo := repositories.NewPostgreSQLAnalysisRepository(db.DB)
	comparisonRepo := repositories.NewPostgreSQLComparisonRepository(db.DB)
	comparisonService := services.NewComparisonService(comparisonRepo, analysisRepo, llmService)
	
	// Create worker client for distributed processing
	ffprobeWorkerURL := os.Getenv("FFPROBE_WORKER_URL")
	llmServiceURL := os.Getenv("LLM_SERVICE_URL")
	
	if ffprobeWorkerURL != "" || llmServiceURL != "" {
		workerClient := services.NewWorkerClient(ffprobeWorkerURL, llmServiceURL, logger)
		analysisService.SetWorkerClient(workerClient)
		logger.Info().
			Str("ffprobe_worker", ffprobeWorkerURL).
			Str("llm_service", llmServiceURL).
			Msg("Worker services configured")
	}
	
	// Set LLM service on analysis service for automatic report generation (fallback)
	analysisService.SetLLMService(llmService)
	
	// Create storage service
	storageConfig := storage.Config{
		Provider:   cfg.StorageProvider,
		Region:     cfg.StorageRegion,
		Bucket:     cfg.StorageBucket,
		AccessKey:  cfg.StorageAccessKey,
		SecretKey:  cfg.StorageSecretKey,
		Endpoint:   cfg.StorageEndpoint,
		UseSSL:     cfg.StorageUseSSL,
		BaseURL:    cfg.StorageBaseURL,
	}
	storageProvider, err := storage.NewProvider(storageConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create storage provider")
	}
	storageService := services.NewStorageService(storageProvider, logger)
	
	// Create middleware
	authConfig := middleware.AuthConfig{
		JWTSecret:    cfg.JWTSecret,
		APIKey:       cfg.APIKey,
		TokenExpiry:  time.Duration(cfg.TokenExpiry) * time.Hour,
		RefreshExpiry: time.Duration(cfg.RefreshExpiry) * time.Hour,
	}
	authMiddleware := middleware.NewAuthMiddleware(authConfig, db.DB, cacheClient, logger)

	rateLimitConfig := middleware.RateLimitConfig{
		RequestsPerMinute: cfg.RateLimitPerMinute,
		RequestsPerHour:   cfg.RateLimitPerHour,
		RequestsPerDay:    cfg.RateLimitPerDay,
		EnablePerIP:       true,
		EnablePerUser:     true,
	}
	rateLimiter := middleware.NewRateLimitMiddleware(rateLimitConfig, logger)

	securityConfig := middleware.SecurityConfig{
		EnableCSRF:         cfg.EnableCSRF,
		EnableXSS:          true,
		EnableFrameGuard:   true,
		EnableHSTS:         true,
		ContentTypeNoSniff: true,
		AllowedOrigins:     cfg.AllowedOrigins,
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key", "X-Request-ID"},
		ExposeHeaders:      []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
	}
	security := middleware.NewSecurityMiddleware(securityConfig, logger)

	monitoring := middleware.NewMonitoringMiddleware(logger)
	
	// Create user service
	userService := services.NewUserService(db.DB, logger)
	
	// Create secret rotation service
	rotationConfig := services.SecretRotationConfig{
		RotationInterval:   90 * 24 * time.Hour,  // 90 days
		GracePeriod:        7 * 24 * time.Hour,   // 7 days
		MinSecretLength:    32,
		MaxActiveKeys:      5,
		EnableAutoRotation: true,
	}
	rotationService := services.NewSecretRotationService(db.DB, cacheClient, logger, rotationConfig)
	
	// Create tenant rate limiter
	tenantRateLimitConfig := middleware.TenantRateLimitConfig{
		DefaultRPM:         cfg.RateLimitPerMinute,
		DefaultRPH:         cfg.RateLimitPerHour,
		DefaultRPD:         cfg.RateLimitPerDay,
		EnableTenantLimits: true,
		EnableUserLimits:   true,
		BurstMultiplier:    1.5,
		IncludeHeaders:     true,
	}
	tenantRateLimiter := middleware.NewTenantRateLimiter(cacheClient, logger, tenantRateLimitConfig)
	
	// Create GraphQL handler (DISABLED - incomplete implementation)
	// TODO: Complete GraphQL schema implementation before enabling
	// graphqlConfig := handlers.GraphQLConfig{
	// 	EnablePlayground:      cfg.LogLevel == "debug",
	// 	EnableIntrospection:   cfg.LogLevel == "debug",
	// 	EnableQueryComplexity: true,
	// 	MaxQueryComplexity:    1000,
	// 	MaxQueryDepth:         15,
	// 	QueryCacheSize:        1000,
	// 	EnableTracing:         cfg.LogLevel == "debug",
	// 	EnableAPQ:             true,
	// }
	// graphqlHandler := handlers.NewGraphQLHandler(
	// 	db.DB, redisClient, logger,
	// 	analysisService, comparisonService, reportService,
	// 	rotationService, userService, storageService,
	// 	graphqlConfig,
	// )
	
	// Create API key handler
	apiKeyHandler := handlers.NewAPIKeyHandler(rotationService, logger)
	
	return &Router{
		probeHandler:  handlers.NewProbeHandler(analysisService, reportGenerator, logger),
		uploadHandler: func() *handlers.UploadHandler {
			handler := handlers.NewUploadHandler(analysisService, cfg.UploadDir, logger)
			handler.SetMaxFileSize(cfg.MaxFileSize)
			return handler
		}(),
		batchHandler:      handlers.NewBatchHandler(analysisService, logger),
		streamHandler:     handlers.NewStreamHandler(analysisService, logger),
		authHandler:       handlers.NewAuthHandler(authMiddleware, logger),
		adminHandler:      handlers.NewAdminHandler(userService, logger),
		qualityHandler:    handlers.NewQualityHandler(qualityService),
		hlsHandler:        handlers.NewHLSHandler(analysisService, hlsService, logger),
		reportHandler:     handlers.NewReportHandler(reportService, logger),
		reportsHandler:    handlers.NewReportsHandler(reportGenerator, analysisService, comparisonService),
		compareHandler:    handlers.NewCompareHandler(qualityService, logger),
		comparisonHandler: handlers.NewComparisonHandler(comparisonService),
		rawHandler:        handlers.NewRawHandler(analysisService, logger),
		storageHandler:    handlers.NewStorageHandler(storageService, logger),
		// graphqlHandler:    graphqlHandler,  // DISABLED - incomplete implementation
		apiKeyHandler:     apiKeyHandler,
		authMiddleware:    authMiddleware,
		rateLimiter:       rateLimiter,
		tenantRateLimiter: tenantRateLimiter,
		security:          security,
		monitoring:        monitoring,
		cacheClient:       cacheClient,
		logger:            logger,
		db:               db,
		config:           cfg,
	}
}

// SetupRoutes configures all API routes
func (r *Router) SetupRoutes() *gin.Engine {
	// Create Gin router
	router := gin.New()

	// Global middleware (applied to all routes)
	router.Use(middleware.Recovery(r.logger))
	router.Use(r.security.RequestLogging())
	router.Use(r.security.Security())
	router.Use(r.security.CORS())
	router.Use(r.requestIDMiddleware())
	router.Use(r.monitoring.Metrics())

	// Optional middleware based on configuration
	if r.config.EnableRateLimit {
		router.Use(r.rateLimiter.RateLimit())
	}

	// Threat detection
	router.Use(r.security.ThreatDetection())

	// Health check endpoint (no auth required)
	router.GET("/health", r.systemHealth)

	// Metrics endpoint (for Prometheus)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication endpoints (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", r.authHandler.Login)
			auth.POST("/refresh", r.authHandler.RefreshToken)
		}

		// Protected routes - require authentication
		var authMiddleware gin.HandlerFunc
		if r.config.EnableAuth {
			if r.config.APIKey != "" {
				authMiddleware = r.authMiddleware.APIKeyAuth()
			} else {
				authMiddleware = r.authMiddleware.JWTAuth()
			}
		} else {
			// No-op middleware when auth is disabled
			authMiddleware = func(c *gin.Context) { c.Next() }
		}


		// GraphQL endpoints (DISABLED - incomplete implementation)
		// TODO: Complete GraphQL schema implementation before enabling
		// graphql := v1.Group("/graphql")
		// {
		// 	// GraphQL playground (development only)
		// 	if r.config.LogLevel == "debug" {
		// 		graphql.GET("/playground", r.graphqlHandler.GraphQLPlaygroundHandler())
		// 	}
		// 	
		// 	// GraphQL schema introspection
		// 	graphql.GET("/schema", r.graphqlHandler.GraphQLSchemaHandler())
		// 	
		// 	// Main GraphQL endpoint with optional authentication
		// 	graphql.Use(r.graphqlHandler.GraphQLMiddleware())
		// 	graphql.POST("", r.graphqlHandler.GraphQLEndpoint())
		// 	graphql.GET("", r.graphqlHandler.GraphQLEndpoint()) // For GET queries
		// }

		// Apply authentication to protected routes
		protected := v1.Group("", authMiddleware)
		{
			// API Key Management endpoints
			keys := protected.Group("/keys")
			{
				keys.POST("/create", r.apiKeyHandler.CreateAPIKey)
				keys.POST("/rotate", r.apiKeyHandler.RotateAPIKey)
			}

			// Additional auth endpoints (require auth)
			authProtected := protected.Group("/auth")
			{
				authProtected.POST("/logout", r.authHandler.Logout)
				authProtected.GET("/profile", r.authHandler.Profile)
				authProtected.POST("/change-password", r.authHandler.ChangePassword)
				authProtected.GET("/validate", r.authHandler.ValidateToken)
				authProtected.POST("/api-key", r.authHandler.GenerateAPIKey)
				authProtected.GET("/api-keys", r.authHandler.ListAPIKeys)
				authProtected.DELETE("/api-keys/:id", r.authHandler.RevokeAPIKey)
			}

			// Admin endpoints (require admin role)
			admin := protected.Group("/admin")
			{
				admin.POST("/rotate-jwt", r.apiKeyHandler.RotateJWTSecret)
				admin.POST("/rate-limits", r.apiKeyHandler.UpdateRateLimits)
				admin.GET("/rotation-status", r.apiKeyHandler.CheckRotationStatus)
				admin.POST("/cleanup", r.apiKeyHandler.CleanupExpiredKeys)
			}

			// Probe endpoints
			probe := protected.Group("/probe")
			{
				probe.Use(r.monitoring.FFprobeMetrics())
				probe.POST("/file", r.probeHandler.ProbeFile)
				probe.POST("/url", r.probeHandler.ProbeURL)
				probe.POST("/quick", r.probeHandler.QuickProbe)
				probe.POST("/hls", r.hlsHandler.AnalyzeHLS)
				probe.GET("/hls/:id", r.hlsHandler.GetHLSAnalysis)
				probe.GET("/hls", r.hlsHandler.ListHLSAnalyses)
				probe.POST("/hls/validate", r.hlsHandler.ValidateHLSPlaylist)
				probe.POST("/compare", r.compareHandler.CompareQuality)
				probe.GET("/compare/:id", r.compareHandler.GetComparisonStatus)
				probe.DELETE("/compare/:id", r.compareHandler.DeleteComparison)
				probe.GET("/comparisons", r.compareHandler.ListComparisons)
				probe.GET("/raw/:id", r.rawHandler.GetRawData)
				probe.GET("/raw/:id/streams", r.rawHandler.GetRawStreams)
				probe.GET("/raw/:id/format", r.rawHandler.GetRawFormat)
				probe.POST("/report", r.reportHandler.GenerateReport)
				probe.GET("/report/:id", r.reportHandler.GetReportStatus)
				probe.DELETE("/report/:id", r.reportHandler.DeleteReport)
				probe.GET("/reports", r.reportHandler.ListReports)
				probe.GET("/download/:id", r.reportHandler.DownloadReport)
				probe.GET("/status/:id", r.probeHandler.GetAnalysisStatus)
				probe.GET("/analyses", r.probeHandler.ListAnalyses)
				probe.DELETE("/analyses/:id", r.probeHandler.DeleteAnalysis)
				probe.GET("/health", r.probeHandler.Health)
			}

			// Upload endpoints
			upload := protected.Group("/upload")
			{
				upload.Use(r.monitoring.UploadMetrics())
				upload.POST("", r.uploadHandler.UploadFile)
				upload.POST("/chunk", r.uploadHandler.UploadChunk)
				upload.GET("/status/:id", r.uploadHandler.GetUploadStatus)
			}

			// Storage endpoints
			storage := protected.Group("/storage")
			{
				storage.POST("/upload", r.storageHandler.UploadFile)
				storage.GET("/download/:key", r.storageHandler.DownloadFile)
				storage.DELETE("/:key", r.storageHandler.DeleteFile)
				storage.GET("/info/:key", r.storageHandler.GetFileInfo)
				storage.POST("/signed-url", r.storageHandler.GetSignedURL)
			}

			// Batch processing endpoints
			batch := protected.Group("/batch")
			{
				batch.Use(r.authMiddleware.RequireRole("user", "admin"))
				batch.POST("/analyze", r.batchHandler.CreateBatch)
				batch.GET("/status/:id", r.batchHandler.GetBatchStatus)
				batch.POST("/:id/cancel", r.batchHandler.CancelBatch)
				batch.GET("", r.batchHandler.ListBatches)
			}

			// Report generation endpoints
			reports := protected.Group("/reports")
			{
				reports.POST("/analysis", r.reportsHandler.GenerateAnalysisReport)
				reports.POST("/comparison", r.reportsHandler.GenerateComparisonReport)
				reports.POST("/analysis/:analysisId/:format", r.reportsHandler.GenerateAnalysisReportWithFormat)
				reports.POST("/comparison/:comparisonId/:format", r.reportsHandler.GenerateComparisonReportWithFormat)
				reports.GET("/formats", r.reportsHandler.ListReportFormats)
				reports.GET("/:reportId/download/:filename", r.reportsHandler.DownloadReport)
				reports.DELETE("/cleanup", r.reportsHandler.CleanupExpiredReports) // Admin only
			}

			// Streaming endpoints
			stream := protected.Group("/stream")
			{
				stream.GET("/analysis", r.streamHandler.StreamAnalysis)
				stream.GET("/progress/:id", r.streamHandler.StreamProgress)
				stream.POST("/live", r.streamHandler.LiveStreamAnalysis)
			}

			// Quality analysis endpoints
			quality := protected.Group("/quality")
			{
				quality.POST("/compare", r.qualityHandler.CompareQuality)
				quality.POST("/batch", r.qualityHandler.BatchCompareQuality)
				quality.GET("/analysis/:id", r.qualityHandler.GetQualityAnalysis)
				quality.DELETE("/analysis/:id", r.qualityHandler.DeleteQualityAnalysis)
				quality.GET("/analysis/:id/frames", r.qualityHandler.GetQualityFrames)
				quality.GET("/analysis/:id/issues", r.qualityHandler.GetQualityIssues)
				quality.GET("/comparison/:id", r.qualityHandler.GetQualityComparison)
				quality.GET("/statistics", r.qualityHandler.GetQualityStatistics)
				quality.GET("/thresholds", r.qualityHandler.GetQualityThresholds)
			}

			// Video comparison endpoints
			comparisons := protected.Group("/comparisons")
			{
				comparisons.POST("", r.comparisonHandler.CreateComparison)
				comparisons.GET("/:id", r.comparisonHandler.GetComparison)
				comparisons.GET("", r.comparisonHandler.ListComparisons)
				comparisons.POST("/quick", r.comparisonHandler.CompareVideos)
				comparisons.GET("/:id/report", r.comparisonHandler.GetComparisonReport)
			}


			// Admin endpoints
			admin := protected.Group("/admin")
			{
				admin.Use(r.authMiddleware.RequireRole("admin"))
				// User management
				admin.GET("/users", r.adminHandler.ListUsers)
				admin.GET("/users/:id/role", r.adminHandler.GetUserRole)
				admin.PUT("/users/:id/role", r.adminHandler.UpdateUserRole)
				admin.GET("/stats", r.adminHandler.GetSystemStats)
			}

			// Admin-only system endpoints
			system := protected.Group("/system")
			{
				system.Use(r.authMiddleware.RequireRole("admin"))
				system.GET("/version", r.getVersion)
				system.GET("/stats", r.getStats)
			}
		}
	}

	// API documentation (when Swagger is added)
	router.GET("/docs/*any", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "API documentation will be available here",
			"swagger": "/docs/swagger.json",
		})
	})

	return router
}

// Middleware functions

func (r *Router) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		c.Next()
	}
}

// System handlers

func (r *Router) systemHealth(c *gin.Context) {
	health := gin.H{
		"status":  "healthy",
		"service": "ffprobe-api",
		"version": "v1.0.0",
	}

	// Check database health
	if err := r.db.Health(c.Request.Context()); err != nil {
		health["status"] = "unhealthy"
		health["database"] = gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		dbStats := r.db.Stats()
		health["database"] = gin.H{
			"status": "healthy",
			"type":   dbStats["database_type"],
			"stats":  dbStats,
		}
	}

	// Check cache health (Valkey/Redis) if available
	if cacheClient := r.getCacheClient(); cacheClient != nil {
		if err := cacheClient.Health(c.Request.Context()); err != nil {
			health["cache"] = gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		} else {
			cacheStats := cacheClient.Stats()
			health["cache"] = gin.H{
				"status": "healthy",
				"type":   "valkey",
				"stats":  cacheStats,
			}
		}
	} else {
		health["cache"] = gin.H{
			"status": "disabled",
			"type":   "none",
		}
	}

	// Overall health status
	if health["status"] == "unhealthy" {
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	c.JSON(http.StatusOK, health)
}

// getCacheClient returns the cache client if available
func (r *Router) getCacheClient() *cache.ValkeyClient {
	return r.cacheClient
}

func (r *Router) getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":     "ffprobe-api",
		"version":     "v1.0.0",
		"api_version": "v1",
		"build_time":  "2024-01-01T00:00:00Z", // This would be set during build
		"commit":      "unknown",               // This would be set during build
	})
}

func (r *Router) getStats(c *gin.Context) {
	// Get database stats
	dbStats := r.db.Stats()

	c.JSON(http.StatusOK, gin.H{
		"uptime":         "0s",
		"requests_total": 0,
		"active_jobs":    0,
		"memory_usage":   "0MB",
		"database":       dbStats,
	})
}

// Helper functions

func generateRequestID() string {
	// Simple request ID generation
	// In production, you might want to use a more sophisticated approach
	return "req-" + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)] // Simple pseudo-random
	}
	return string(b)
}

// Legacy function for backward compatibility
func SetupRoutes(router *gin.Engine, cfg *config.Config, db *database.DB, logger zerolog.Logger) {
	r := NewRouter(cfg, db, logger)
	*router = *r.SetupRoutes()
}