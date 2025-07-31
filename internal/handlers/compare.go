package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/models"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// CompareHandler handles video quality comparison requests
type CompareHandler struct {
	qualityService *services.QualityService
	logger         zerolog.Logger
}

// NewCompareHandler creates a new compare handler
func NewCompareHandler(qualityService *services.QualityService, logger zerolog.Logger) *CompareHandler {
	return &CompareHandler{
		qualityService: qualityService,
		logger:         logger,
	}
}

// CompareQualityRequest represents a request to compare video quality
type CompareQualityRequest struct {
	ReferenceID   string                 `json:"reference_id" binding:"required"`
	DistortedID   string                 `json:"distorted_id" binding:"required"`
	ComparisonType models.ComparisonType `json:"comparison_type"`
	Metrics       []string               `json:"metrics"` // vmaf, psnr, ssim, ms_ssim
	Options       map[string]interface{} `json:"options"`
}

// CompareQualityResponse represents the response from quality comparison
type CompareQualityResponse struct {
	ID             string                 `json:"id"`
	ReferenceID    string                 `json:"reference_id"`
	DistortedID    string                 `json:"distorted_id"`
	ComparisonType string                 `json:"comparison_type"`
	Status         string                 `json:"status"`
	Results        map[string]interface{} `json:"results,omitempty"`
	ProcessingTime float64                `json:"processing_time,omitempty"`
	CreatedAt      string                 `json:"created_at"`
}

// CompareQuality handles video quality comparison requests
// @Summary Compare video quality
// @Description Compare quality between reference and distorted videos using VMAF, PSNR, SSIM
// @Tags probe
// @Accept json
// @Produce json
// @Param request body CompareQualityRequest true "Quality comparison request"
// @Success 202 {object} CompareQualityResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/compare [post]
func (h *CompareHandler) CompareQuality(c *gin.Context) {
	var req CompareQualityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate analysis IDs
	referenceUUID, err := uuid.Parse(req.ReferenceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid reference analysis ID format",
		})
		return
	}

	distortedUUID, err := uuid.Parse(req.DistortedID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid distorted analysis ID format",
		})
		return
	}

	// Get user context
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	// Set default comparison type if not provided
	if req.ComparisonType == "" {
		req.ComparisonType = models.ComparisonTypeFull
	}

	// Set default metrics if not provided
	if len(req.Metrics) == 0 {
		req.Metrics = []string{"vmaf", "psnr", "ssim"}
	}

	// Generate comparison ID
	comparisonID := uuid.New().String()

	// Start async quality comparison
	go func() {
		// Use background context to avoid cancellation when HTTP request ends
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		// Create comparison options
		opts := services.QualityComparisonOptions{
			ReferenceID:    referenceUUID,
			DistortedID:    distortedUUID,
			ComparisonType: req.ComparisonType,
			Metrics:        req.Metrics,
			UserID:         userIDStr,
			Options:        req.Options,
		}

		// Perform quality comparison
		if err := h.qualityService.CompareQuality(ctx, comparisonID, opts); err != nil {
			h.logger.Error().Err(err).Str("comparison_id", comparisonID).Msg("Quality comparison failed")
		}
	}()

	// Return immediate response
	c.JSON(http.StatusAccepted, CompareQualityResponse{
		ID:             comparisonID,
		ReferenceID:    req.ReferenceID,
		DistortedID:    req.DistortedID,
		ComparisonType: string(req.ComparisonType),
		Status:         "processing",
		CreatedAt:      "now",
	})
}

// GetComparisonStatus gets the status of a quality comparison
// @Summary Get comparison status
// @Description Get the status and results of a quality comparison
// @Tags probe
// @Accept json
// @Produce json
// @Param id path string true "Comparison ID"
// @Success 200 {object} CompareQualityResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/compare/{id} [get]
func (h *CompareHandler) GetComparisonStatus(c *gin.Context) {
	comparisonID := c.Param("id")

	// Validate UUID
	id, err := uuid.Parse(comparisonID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid comparison ID format",
		})
		return
	}

	// Get comparison from service
	comparison, err := h.qualityService.GetQualityComparison(c.Request.Context(), id)
	if err != nil {
		if err == services.ErrComparisonNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Quality comparison not found",
			})
			return
		}

		h.logger.Error().Err(err).Str("comparison_id", comparisonID).Msg("Failed to get quality comparison")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve quality comparison",
		})
		return
	}

	// Convert to response
	response := CompareQualityResponse{
		ID:             comparison.ID.String(),
		ReferenceID:    comparison.ReferenceID.String(),
		DistortedID:    comparison.DistortedID.String(),
		ComparisonType: string(comparison.ComparisonType),
		Status:         string(comparison.Status),
		CreatedAt:      comparison.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Add results if comparison is completed
	if comparison.Status == models.StatusCompleted && comparison.ResultSummary != nil {
		response.Results = comparison.ResultSummary
		response.ProcessingTime = comparison.ProcessingTime
	}

	c.JSON(http.StatusOK, response)
}

// ListComparisons lists quality comparisons for the current user
// @Summary List quality comparisons
// @Description List all quality comparisons for the authenticated user
// @Tags probe
// @Accept json
// @Produce json
// @Param reference_id query string false "Filter by reference analysis ID"
// @Param distorted_id query string false "Filter by distorted analysis ID"
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} CompareQualityResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/comparisons [get]
func (h *CompareHandler) ListComparisons(c *gin.Context) {
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

	// Get filter params
	filters := services.ComparisonFilters{
		ReferenceID: c.Query("reference_id"),
		DistortedID: c.Query("distorted_id"),
		Status:      c.Query("status"),
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	// List comparisons
	comparisons, total, err := h.qualityService.ListQualityComparisons(c.Request.Context(), userIDStr, filters, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list quality comparisons")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list quality comparisons",
		})
		return
	}

	// Convert to response format
	responses := make([]CompareQualityResponse, len(comparisons))
	for i, comparison := range comparisons {
		responses[i] = CompareQualityResponse{
			ID:             comparison.ID.String(),
			ReferenceID:    comparison.ReferenceID.String(),
			DistortedID:    comparison.DistortedID.String(),
			ComparisonType: string(comparison.ComparisonType),
			Status:         string(comparison.Status),
			ProcessingTime: comparison.ProcessingTime,
			CreatedAt:      comparison.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		if comparison.Status == models.StatusCompleted && comparison.ResultSummary != nil {
			responses[i].Results = comparison.ResultSummary
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"comparisons": responses,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// DeleteComparison deletes a quality comparison
// @Summary Delete quality comparison
// @Description Delete a quality comparison by ID
// @Tags probe
// @Accept json
// @Produce json
// @Param id path string true "Comparison ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/compare/{id} [delete]
func (h *CompareHandler) DeleteComparison(c *gin.Context) {
	comparisonID := c.Param("id")

	// Validate UUID
	id, err := uuid.Parse(comparisonID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid comparison ID format",
		})
		return
	}

	// Delete comparison
	if err := h.qualityService.DeleteQualityComparison(c.Request.Context(), id); err != nil {
		if err == services.ErrComparisonNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Quality comparison not found",
			})
			return
		}

		h.logger.Error().Err(err).Str("comparison_id", comparisonID).Msg("Failed to delete quality comparison")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete quality comparison",
		})
		return
	}

	c.Status(http.StatusNoContent)
}