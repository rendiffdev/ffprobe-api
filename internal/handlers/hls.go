package handlers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/hls"
	"github.com/rendiffdev/ffprobe-api/internal/models"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// HLSHandler handles HLS-specific analysis requests
type HLSHandler struct {
	analysisService *services.AnalysisService
	hlsService      *services.HLSService
	logger          zerolog.Logger
}

// NewHLSHandler creates a new HLS handler
func NewHLSHandler(analysisService *services.AnalysisService, hlsService *services.HLSService, logger zerolog.Logger) *HLSHandler {
	return &HLSHandler{
		analysisService: analysisService,
		hlsService:      hlsService,
		logger:          logger,
	}
}

// HLSAnalysisRequest represents a request to analyze HLS content
type HLSAnalysisRequest struct {
	// URL or local path to HLS manifest or folder
	Source string `json:"source" binding:"required"`
	// Whether to analyze all segments (can be time-consuming)
	AnalyzeSegments bool `json:"analyze_segments"`
	// Maximum number of segments to analyze (0 = all)
	MaxSegments int `json:"max_segments"`
	// Include quality metrics for segments
	IncludeQuality bool `json:"include_quality"`
	// Custom ffprobe arguments
	FFprobeArgs []string `json:"ffprobe_args"`
	// Enable AI-powered GenAI analysis (USP feature)
	IncludeLLM bool `json:"include_llm,omitempty"`
}

// HLSAnalysisResponse represents the response from HLS analysis
type HLSAnalysisResponse struct {
	ID              string                   `json:"id"`
	Status          string                   `json:"status"`
	ManifestType    string                   `json:"manifest_type"`
	Variants        []hls.Variant           `json:"variants,omitempty"`
	Segments        []hls.Segment           `json:"segments,omitempty"`
	TotalDuration   float64                 `json:"total_duration"`
	SegmentCount    int                     `json:"segment_count"`
	ValidationIssues []string               `json:"validation_issues,omitempty"`
	ProcessingTime  float64                 `json:"processing_time"`
	Metadata        map[string]interface{}  `json:"metadata,omitempty"`
}

// AnalyzeHLS handles HLS manifest/folder analysis
// @Summary Analyze HLS content
// @Description Analyze HLS manifest file or folder containing HLS segments
// @Tags probe
// @Accept json
// @Produce json
// @Param request body HLSAnalysisRequest true "HLS analysis request"
// @Success 202 {object} HLSAnalysisResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/hls [post]
func (h *HLSHandler) AnalyzeHLS(c *gin.Context) {
	var req HLSAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get user context
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	// Validate source
	if req.Source == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Source is required",
		})
		return
	}

	// Create analysis ID
	analysisID := uuid.New().String()

	// Start async HLS analysis
	go func() {
		// Use background context to avoid cancellation when HTTP request ends
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()
		
		// Create HLS analysis options
		opts := services.HLSAnalysisOptions{
			Source:          req.Source,
			AnalyzeSegments: req.AnalyzeSegments,
			MaxSegments:     req.MaxSegments,
			IncludeQuality:  req.IncludeQuality,
			FFprobeArgs:     req.FFprobeArgs,
			UserID:          userIDStr,
		}

		// Perform HLS analysis
		if err := h.hlsService.AnalyzeHLS(ctx, analysisID, opts); err != nil {
			h.logger.Error().Err(err).Str("analysis_id", analysisID).Msg("HLS analysis failed")
		}
	}()

	// Return immediate response with analysis ID
	c.JSON(http.StatusAccepted, gin.H{
		"id": analysisID,
		"status": "processing",
		"message": "HLS analysis started",
	})
}

// GetHLSAnalysis retrieves HLS analysis results
// @Summary Get HLS analysis results
// @Description Retrieve the results of an HLS analysis by ID
// @Tags probe
// @Accept json
// @Produce json
// @Param id path string true "Analysis ID"
// @Success 200 {object} HLSAnalysisResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/hls/{id} [get]
func (h *HLSHandler) GetHLSAnalysis(c *gin.Context) {
	analysisID := c.Param("id")
	
	// Validate UUID
	if _, err := uuid.Parse(analysisID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid analysis ID format",
		})
		return
	}

	// Get HLS analysis from service
	analysis, err := h.hlsService.GetHLSAnalysis(c.Request.Context(), analysisID)
	if err != nil {
		if err == services.ErrAnalysisNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "HLS analysis not found",
			})
			return
		}
		
		h.logger.Error().Err(err).Str("analysis_id", analysisID).Msg("Failed to get HLS analysis")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve HLS analysis",
		})
		return
	}

	// Convert to response format
	response := HLSAnalysisResponse{
		ID:              analysis.ID,
		Status:          string(analysis.Status),
		ManifestType:    string(analysis.ManifestType),
		TotalDuration:   analysis.TotalDuration,
		SegmentCount:    analysis.SegmentCount,
		ProcessingTime:  analysis.ProcessingTime,
	}

	// Add variants if present
	if manifestData, ok := analysis.ManifestData.(map[string]interface{}); ok {
		if variants, ok := manifestData["variants"].([]interface{}); ok {
			for _, v := range variants {
				if variant, ok := v.(map[string]interface{}); ok {
					response.Variants = append(response.Variants, hls.Variant{
						Bandwidth:  int(variant["bandwidth"].(float64)),
						Resolution: variant["resolution"].(string),
						Codecs:     variant["codecs"].(string),
						FrameRate:  variant["frame_rate"].(float64),
					})
				}
			}
		}
	}

	// Add segments if available
	segments, err := h.hlsService.GetHLSSegments(c.Request.Context(), analysis.ID, 100)
	if err == nil && len(segments) > 0 {
		for _, seg := range segments {
			response.Segments = append(response.Segments, hls.Segment{
				URI:            seg.SegmentURI,
				Duration:       seg.Duration,
				SequenceNumber: seg.SequenceNumber,
				Bitrate:        seg.Bitrate,
				Resolution:     seg.Resolution,
				FrameRate:      seg.FrameRate,
				QualityScore:   seg.QualityScore,
			})
		}
	}

	c.JSON(http.StatusOK, response)
}

// ValidateHLSPlaylist validates an HLS playlist for spec compliance
// @Summary Validate HLS playlist
// @Description Validate an HLS playlist against HLS specifications
// @Tags probe
// @Accept json
// @Produce json
// @Param request body HLSValidationRequest true "HLS validation request"
// @Success 200 {object} HLSValidationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/hls/validate [post]
func (h *HLSHandler) ValidateHLSPlaylist(c *gin.Context) {
	var req struct {
		Source string `json:"source" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Perform validation
	issues, err := h.hlsService.ValidatePlaylist(c.Request.Context(), req.Source)
	if err != nil {
		h.logger.Error().Err(err).Msg("HLS validation failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Validation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"source": req.Source,
		"valid": len(issues) == 0,
		"issues": issues,
	})
}

// ListHLSAnalyses lists HLS analyses for the current user
// @Summary List HLS analyses
// @Description List all HLS analyses for the authenticated user
// @Tags probe
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} HLSAnalysisResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/hls [get]
func (h *HLSHandler) ListHLSAnalyses(c *gin.Context) {
	// Get pagination params
	limit := 20
	offset := 0
	
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	// List analyses
	analyses, total, err := h.hlsService.ListHLSAnalyses(c.Request.Context(), userIDStr, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list HLS analyses")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list analyses",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analyses": analyses,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}