package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/ffmpeg"
	"github.com/rendiffdev/ffprobe-api/internal/models"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// ProbeHandler handles ffprobe-related API endpoints
type ProbeHandler struct {
	analysisService *services.AnalysisService
	logger          zerolog.Logger
}

// NewProbeHandler creates a new probe handler
func NewProbeHandler(analysisService *services.AnalysisService, logger zerolog.Logger) *ProbeHandler {
	return &ProbeHandler{
		analysisService: analysisService,
		logger:          logger,
	}
}

// ProbeFileRequest represents a request to probe a file
type ProbeFileRequest struct {
	FilePath    string                 `json:"file_path" binding:"required"`
	Options     *ffmpeg.FFprobeOptions `json:"options,omitempty"`
	Async       bool                   `json:"async,omitempty"`
	SourceType  string                 `json:"source_type,omitempty"`
}

// ProbeURLRequest represents a request to probe a URL
type ProbeURLRequest struct {
	URL         string                 `json:"url" binding:"required"`
	Options     *ffmpeg.FFprobeOptions `json:"options,omitempty"`
	Async       bool                   `json:"async,omitempty"`
	Timeout     int                    `json:"timeout,omitempty"` // seconds
}

// ProbeResponse represents the response from a probe operation
type ProbeResponse struct {
	AnalysisID uuid.UUID              `json:"analysis_id"`
	Status     string                 `json:"status"`
	Result     *ffmpeg.FFprobeResult  `json:"result,omitempty"`
	Analysis   *models.Analysis       `json:"analysis,omitempty"`
	Message    string                 `json:"message,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status        string `json:"status"`
	FFprobeVersion string `json:"ffprobe_version,omitempty"`
	Error         string `json:"error,omitempty"`
}

// ProbeFile probes a local file
// @Summary Probe a local file
// @Description Analyze a local media file using ffprobe
// @Tags probe
// @Accept json
// @Produce json
// @Param request body ProbeFileRequest true "Probe request"
// @Success 200 {object} ProbeResponse
// @Success 202 {object} ProbeResponse "Accepted for async processing"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/file [post]
func (h *ProbeHandler) ProbeFile(c *gin.Context) {
	var req ProbeFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Set default source type
	if req.SourceType == "" {
		req.SourceType = "local"
	}

	// Create analysis request
	fileName := filepath.Base(req.FilePath)
	createReq := &models.CreateAnalysisRequest{
		FileName:   fileName,
		FilePath:   req.FilePath,
		SourceType: req.SourceType,
	}

	// Create analysis record
	analysis, err := h.analysisService.CreateAnalysis(c.Request.Context(), createReq)
	if err != nil {
		h.logger.Error().Err(err).Str("file_path", req.FilePath).Msg("Failed to create analysis")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create analysis",
			Details: err.Error(),
		})
		return
	}

	if req.Async {
		// Start async processing
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()

			if err := h.analysisService.ProcessAnalysis(ctx, analysis.ID, req.Options); err != nil {
				h.logger.Error().Err(err).Str("analysis_id", analysis.ID.String()).Msg("Async analysis failed")
			}
		}()

		c.JSON(http.StatusAccepted, ProbeResponse{
			AnalysisID: analysis.ID,
			Status:     "processing",
			Message:    "Analysis started, check status endpoint for progress",
		})
		return
	}

	// Synchronous processing
	if err := h.analysisService.ProcessAnalysis(c.Request.Context(), analysis.ID, req.Options); err != nil {
		h.logger.Error().Err(err).Str("analysis_id", analysis.ID.String()).Msg("Analysis failed")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Analysis failed",
			Details: err.Error(),
		})
		return
	}

	// Get updated analysis
	updatedAnalysis, err := h.analysisService.GetAnalysis(c.Request.Context(), analysis.ID)
	if err != nil {
		h.logger.Error().Err(err).Str("analysis_id", analysis.ID.String()).Msg("Failed to get updated analysis")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get analysis result",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ProbeResponse{
		AnalysisID: analysis.ID,
		Status:     "completed",
		Analysis:   updatedAnalysis,
	})
}

// ProbeURL probes a remote URL
// @Summary Probe a remote URL
// @Description Analyze a remote media file using ffprobe
// @Tags probe
// @Accept json
// @Produce json
// @Param request body ProbeURLRequest true "Probe URL request"
// @Success 200 {object} ProbeResponse
// @Success 202 {object} ProbeResponse "Accepted for async processing"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/url [post]
func (h *ProbeHandler) ProbeURL(c *gin.Context) {
	var req ProbeURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Set timeout if provided
	if req.Options == nil {
		req.Options = ffmpeg.NewOptionsBuilder().BasicInfo().Build()
	}
	if req.Timeout > 0 {
		req.Options.Timeout = time.Duration(req.Timeout) * time.Second
	}

	// Create analysis request
	createReq := &models.CreateAnalysisRequest{
		FileName:   req.URL,
		FilePath:   req.URL,
		SourceType: "url",
	}

	// Create analysis record
	analysis, err := h.analysisService.CreateAnalysis(c.Request.Context(), createReq)
	if err != nil {
		h.logger.Error().Err(err).Str("url", req.URL).Msg("Failed to create analysis")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create analysis",
			Details: err.Error(),
		})
		return
	}

	if req.Async {
		// Start async processing
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()

			if err := h.analysisService.ProcessAnalysis(ctx, analysis.ID, req.Options); err != nil {
				h.logger.Error().Err(err).Str("analysis_id", analysis.ID.String()).Msg("Async URL analysis failed")
			}
		}()

		c.JSON(http.StatusAccepted, ProbeResponse{
			AnalysisID: analysis.ID,
			Status:     "processing",
			Message:    "URL analysis started, check status endpoint for progress",
		})
		return
	}

	// Synchronous processing
	if err := h.analysisService.ProcessAnalysis(c.Request.Context(), analysis.ID, req.Options); err != nil {
		h.logger.Error().Err(err).Str("analysis_id", analysis.ID.String()).Msg("URL analysis failed")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "URL analysis failed",
			Details: err.Error(),
		})
		return
	}

	// Get updated analysis
	updatedAnalysis, err := h.analysisService.GetAnalysis(c.Request.Context(), analysis.ID)
	if err != nil {
		h.logger.Error().Err(err).Str("analysis_id", analysis.ID.String()).Msg("Failed to get updated analysis")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get analysis result",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ProbeResponse{
		AnalysisID: analysis.ID,
		Status:     "completed",
		Analysis:   updatedAnalysis,
	})
}

// GetAnalysisStatus gets the status of an analysis
// @Summary Get analysis status
// @Description Get the current status and result of an analysis
// @Tags probe
// @Produce json
// @Param id path string true "Analysis ID"
// @Success 200 {object} ProbeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/status/{id} [get]
func (h *ProbeHandler) GetAnalysisStatus(c *gin.Context) {
	analysisIDStr := c.Param("id")
	analysisID, err := uuid.Parse(analysisIDStr)
	if err != nil {
		h.logger.Error().Err(err).Str("id", analysisIDStr).Msg("Invalid analysis ID")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid analysis ID",
			Details: err.Error(),
		})
		return
	}

	analysis, err := h.analysisService.GetAnalysis(c.Request.Context(), analysisID)
	if err != nil {
		h.logger.Error().Err(err).Str("analysis_id", analysisID.String()).Msg("Failed to get analysis")
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Analysis not found",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ProbeResponse{
		AnalysisID: analysis.ID,
		Status:     analysis.Status,
		Analysis:   analysis,
	})
}

// ListAnalyses lists analyses for a user
// @Summary List user analyses
// @Description List all analyses for the current user
// @Tags probe
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} ListAnalysesResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/analyses [get]
func (h *ProbeHandler) ListAnalyses(c *gin.Context) {
	// For now, use a default user ID since we don't have auth yet
	// This will be replaced when authentication is implemented
	defaultUserID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	analyses, err := h.analysisService.GetAnalysesByUser(c.Request.Context(), defaultUserID, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get analyses")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get analyses",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ListAnalysesResponse{
		Analyses: analyses,
		Limit:    limit,
		Offset:   offset,
		Count:    len(analyses),
	})
}

// DeleteAnalysis deletes an analysis
// @Summary Delete analysis
// @Description Delete an analysis and its results
// @Tags probe
// @Param id path string true "Analysis ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/analyses/{id} [delete]
func (h *ProbeHandler) DeleteAnalysis(c *gin.Context) {
	analysisIDStr := c.Param("id")
	analysisID, err := uuid.Parse(analysisIDStr)
	if err != nil {
		h.logger.Error().Err(err).Str("id", analysisIDStr).Msg("Invalid analysis ID")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid analysis ID",
			Details: err.Error(),
		})
		return
	}

	if err := h.analysisService.DeleteAnalysis(c.Request.Context(), analysisID); err != nil {
		h.logger.Error().Err(err).Str("analysis_id", analysisID.String()).Msg("Failed to delete analysis")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete analysis",
			Details: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// Health checks the health of the ffprobe service
// @Summary Health check
// @Description Check if ffprobe is available and working
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /api/v1/probe/health [get]
func (h *ProbeHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Check ffprobe availability
	if err := h.analysisService.CheckFFprobeAvailability(ctx); err != nil {
		h.logger.Error().Err(err).Msg("FFprobe health check failed")
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status: "unhealthy",
			Error:  err.Error(),
		})
		return
	}

	// Get ffprobe version
	version, err := h.analysisService.GetFFprobeVersion(ctx)
	if err != nil {
		h.logger.Warn().Err(err).Msg("Failed to get ffprobe version")
		version = "unknown"
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:        "healthy",
		FFprobeVersion: version,
	})
}

// QuickProbe performs a quick probe with minimal information
// @Summary Quick probe
// @Description Perform a fast analysis with basic information only
// @Tags probe
// @Accept json
// @Produce json
// @Param request body ProbeFileRequest true "Quick probe request"
// @Success 200 {object} ProbeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/quick [post]
func (h *ProbeHandler) QuickProbe(c *gin.Context) {
	var req ProbeFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Force quick analysis options
	req.Options = ffmpeg.NewOptionsBuilder().
		Input(req.FilePath).
		QuickInfo().
		Build()

	// Set default source type
	if req.SourceType == "" {
		req.SourceType = "local"
	}

	// Create and process analysis
	fileName := filepath.Base(req.FilePath)
	createReq := &models.CreateAnalysisRequest{
		FileName:   fileName,
		FilePath:   req.FilePath,
		SourceType: req.SourceType,
	}

	analysis, err := h.analysisService.CreateAnalysis(c.Request.Context(), createReq)
	if err != nil {
		h.logger.Error().Err(err).Str("file_path", req.FilePath).Msg("Failed to create quick analysis")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create analysis",
			Details: err.Error(),
		})
		return
	}

	// Process synchronously with quick options
	if err := h.analysisService.ProcessAnalysis(c.Request.Context(), analysis.ID, req.Options); err != nil {
		h.logger.Error().Err(err).Str("analysis_id", analysis.ID.String()).Msg("Quick analysis failed")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Quick analysis failed",
			Details: err.Error(),
		})
		return
	}

	// Get updated analysis
	updatedAnalysis, err := h.analysisService.GetAnalysis(c.Request.Context(), analysis.ID)
	if err != nil {
		h.logger.Error().Err(err).Str("analysis_id", analysis.ID.String()).Msg("Failed to get quick analysis result")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get analysis result",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ProbeResponse{
		AnalysisID: analysis.ID,
		Status:     "completed",
		Analysis:   updatedAnalysis,
	})
}

// Response types
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type ListAnalysesResponse struct {
	Analyses []models.Analysis `json:"analyses"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	Count    int               `json:"count"`
}