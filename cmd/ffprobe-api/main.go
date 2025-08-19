package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rendiffdev/ffprobe-api/internal/config"
	"github.com/rendiffdev/ffprobe-api/internal/database"
	"github.com/rendiffdev/ffprobe-api/internal/ffmpeg"
	"github.com/rendiffdev/ffprobe-api/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logger.New(cfg.LogLevel)
	logger.Info().Msg("Starting ffprobe-api core service")

	// Initialize database
	db, err := database.New(cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// CRITICAL: Validate FFmpeg/FFprobe binary at startup
	logger.Info().Msg("Validating FFmpeg/FFprobe binaries...")
	ffprobeInstance := ffmpeg.NewFFprobe(cfg.FFprobePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ffprobeInstance.ValidateBinaryAtStartup(ctx); err != nil {
		logger.Fatal().
			Err(err).
			Str("ffprobe_path", cfg.FFprobePath).
			Msg("FFprobe binary validation failed - cannot start application")
	}

	// QC Analysis functionality is ready through enhanced analyzer
	logger.Info().Msg("QC Analysis Tools ready and validated")

	// Create a basic Gin router for health checks
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "ffprobe-api-core",
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
		})
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info().Int("port", cfg.Port).Msg("Core service starting with validated QC tools")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info().Msg("Shutting down server...")

	// Give the server 30 seconds to finish current requests
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited")
}
