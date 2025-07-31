package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// FFprobeWorker handles media analysis requests
type FFprobeWorker struct {
	logger zerolog.Logger
}

// AnalysisRequest represents a request for media analysis
type AnalysisRequest struct {
	FilePath string            `json:"file_path" binding:"required"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// AnalysisResponse represents the response from media analysis
type AnalysisResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	ProcessingTime time.Duration `json:"processing_time"`
}

func main() {
	// Setup logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	
	worker := &FFprobeWorker{
		logger: logger,
	}

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "ffprobe-worker",
			"timestamp": time.Now().UTC(),
		})
	})

	// Analysis endpoint
	router.POST("/analyze", worker.analyzeMedia)

	// Start server
	port := getEnv("PORT", "8081")
	logger.Info().Str("port", port).Msg("Starting FFprobe Worker service")
	
	if err := router.Run(":" + port); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}
}

func (w *FFprobeWorker) analyzeMedia(c *gin.Context) {
	start := time.Now()
	
	var req AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		w.logger.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, AnalysisResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
			ProcessingTime: time.Since(start),
		})
		return
	}

	// Execute FFprobe analysis
	result, err := w.executeFFprobe(c.Request.Context(), req.FilePath, req.Options)
	if err != nil {
		w.logger.Error().Err(err).Str("file_path", req.FilePath).Msg("FFprobe analysis failed")
		c.JSON(http.StatusInternalServerError, AnalysisResponse{
			Success: false,
			Error:   "Analysis failed: " + err.Error(),
			ProcessingTime: time.Since(start),
		})
		return
	}

	w.logger.Info().
		Str("file_path", req.FilePath).
		Dur("processing_time", time.Since(start)).
		Msg("Analysis completed successfully")

	c.JSON(http.StatusOK, AnalysisResponse{
		Success: true,
		Data:    result,
		ProcessingTime: time.Since(start),
	})
}

func (w *FFprobeWorker) executeFFprobe(ctx context.Context, filePath string, options map[string]interface{}) (map[string]interface{}, error) {
	// Get ffprobe path from environment or use default
	ffprobePath := getEnv("FFPROBE_PATH", "ffprobe")
	
	// Check if ffprobe is available
	if _, err := exec.LookPath(ffprobePath); err != nil {
		return nil, fmt.Errorf("ffprobe not found in PATH: %w", err)
	}
	
	// Validate file path
	if strings.Contains(filePath, "..") || strings.Contains(filePath, ";") || strings.Contains(filePath, "&") {
		return nil, fmt.Errorf("invalid file path: potentially malicious characters detected")
	}
	
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	
	// Build ffprobe command
	args := []string{
		"-v", "quiet",           // Suppress verbose output
		"-print_format", "json", // Output in JSON format
		"-show_format",          // Show format information
		"-show_streams",         // Show stream information
		"-show_chapters",        // Show chapter information
		"-show_programs",        // Show program information
		filePath,
	}
	
	// Apply options if provided
	if options != nil {
		// Add additional options based on the options map
		if showFrames, ok := options["show_frames"].(bool); ok && showFrames {
			args = append(args, "-show_frames")
		}
		if showPackets, ok := options["show_packets"].(bool); ok && showPackets {
			args = append(args, "-show_packets")
		}
		if selectStreams, ok := options["select_streams"].(string); ok && selectStreams != "" {
			args = append(args, "-select_streams", selectStreams)
		}
	}
	
	// Create command with timeout
	cmd := exec.CommandContext(ctx, ffprobePath, args...)
	
	// Set timeout if not provided by context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		cmd = exec.CommandContext(ctx, ffprobePath, args...)
	}
	
	w.logger.Debug().
		Str("command", ffprobePath).
		Strs("args", args).
		Str("file_path", filePath).
		Msg("Executing ffprobe command")
	
	// Execute command
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			w.logger.Error().
				Err(err).
				Str("stderr", string(exitError.Stderr)).
				Str("file_path", filePath).
				Msg("FFprobe command failed")
			return nil, fmt.Errorf("ffprobe failed: %s", string(exitError.Stderr))
		}
		return nil, fmt.Errorf("failed to execute ffprobe: %w", err)
	}
	
	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		w.logger.Error().Err(err).Str("output", string(output)).Msg("Failed to parse ffprobe output")
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}
	
	// Add worker metadata
	result["worker_info"] = map[string]interface{}{
		"service":    "ffprobe-worker",
		"timestamp":  time.Now().UTC(),
		"file_path":  filePath,
		"ffprobe_version": w.getFFprobeVersion(),
	}
	
	return result, nil
}

// getFFprobeVersion gets the ffprobe version
func (w *FFprobeWorker) getFFprobeVersion() string {
	ffprobePath := getEnv("FFPROBE_PATH", "ffprobe")
	cmd := exec.Command(ffprobePath, "-version")
	
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	
	// Extract version from first line
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	
	return "unknown"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}