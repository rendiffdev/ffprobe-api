package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rendiffdev/ffprobe-api/internal/database"
	"github.com/rendiffdev/ffprobe-api/internal/quality"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// QualityHandler handles quality analysis endpoints
type QualityHandler struct {
	qualityService *services.QualityService
}

// NewQualityHandler creates a new quality handler
func NewQualityHandler(qualityService *services.QualityService) *QualityHandler {
	return &QualityHandler{
		qualityService: qualityService,
	}
}

// CompareQuality handles quality comparison requests
// @Summary Compare video quality between reference and distorted files
// @Description Performs quality analysis using VMAF, PSNR, and SSIM metrics
// @Tags quality
// @Accept json
// @Produce json
// @Param request body quality.QualityComparisonRequest true "Quality comparison request"
// @Success 200 {object} quality.QualityResult
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/quality/compare [post]
func (h *QualityHandler) CompareQuality(c *gin.Context) {
	var request quality.QualityComparisonRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request format",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Validate request
	if request.ReferenceFile == "" || request.DistortedFile == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Reference file and distorted file are required",
			Code:  "MISSING_REQUIRED_FIELDS",
		})
		return
	}

	if len(request.Metrics) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "At least one quality metric is required",
			Code:  "MISSING_METRICS",
		})
		return
	}

	// Perform analysis
	var result *quality.QualityResult
	var err error

	if request.Async {
		result, err = h.qualityService.AnalyzeQualityAsync(c.Request.Context(), &request)
	} else {
		result, err = h.qualityService.AnalyzeQuality(c.Request.Context(), &request)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "QUALITY_ANALYSIS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// BatchCompareQuality handles batch quality comparison requests
// @Summary Batch compare video quality for multiple file pairs
// @Description Performs batch quality analysis with controlled concurrency
// @Tags quality
// @Accept json
// @Produce json
// @Param request body quality.BatchQualityRequest true "Batch quality comparison request"
// @Success 200 {object} quality.BatchQualityResult
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/quality/batch [post]
func (h *QualityHandler) BatchCompareQuality(c *gin.Context) {
	var request quality.BatchQualityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request format",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Validate request
	if len(request.Comparisons) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "At least one comparison is required",
			Code:  "MISSING_COMPARISONS",
		})
		return
	}

	// Perform batch analysis
	result, err := h.qualityService.BatchAnalyzeQuality(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "BATCH_ANALYSIS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetQualityAnalysis retrieves a quality analysis by ID
// @Summary Get quality analysis by ID
// @Description Retrieves a specific quality analysis result
// @Tags quality
// @Produce json
// @Param id path string true "Quality analysis ID"
// @Success 200 {object} quality.QualityAnalysis
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/quality/analysis/{id} [get]
func (h *QualityHandler) GetQualityAnalysis(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid analysis ID format",
			Code:  "INVALID_ID",
		})
		return
	}

	analysis, err := h.qualityService.GetQualityAnalysis(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "analysis not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Quality analysis not found",
				Code:  "ANALYSIS_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetQualityComparison retrieves a quality comparison by ID
// @Summary Get quality comparison by ID
// @Description Retrieves a specific quality comparison result
// @Tags quality
// @Produce json
// @Param id path string true "Quality comparison ID"
// @Success 200 {object} quality.QualityResult
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/quality/comparison/{id} [get]
func (h *QualityHandler) GetQualityComparison(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid comparison ID format",
			Code:  "INVALID_ID",
		})
		return
	}

	comparison, err := h.qualityService.GetQualityComparison(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "analysis not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Quality comparison not found",
				Code:  "COMPARISON_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, comparison)
}

// GetQualityFrames retrieves frame-level quality metrics
// @Summary Get frame-level quality metrics
// @Description Retrieves per-frame quality metrics for detailed analysis
// @Tags quality
// @Produce json
// @Param id path string true "Quality analysis ID"
// @Param limit query int false "Number of frames to return" default(100)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} []quality.QualityFrameMetric
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/quality/analysis/{id}/frames [get]
func (h *QualityHandler) GetQualityFrames(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid analysis ID format",
			Code:  "INVALID_ID",
		})
		return
	}

	// Parse query parameters
	limit := 100
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Limit maximum frames returned
	if limit > 10000 {
		limit = 10000
	}

	frames, err := h.qualityService.GetQualityFrames(c.Request.Context(), id, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, frames)
}

// GetQualityIssues retrieves quality issues for an analysis
// @Summary Get quality issues
// @Description Retrieves detected quality issues for a specific analysis
// @Tags quality
// @Produce json
// @Param id path string true "Quality analysis ID"
// @Param severity query string false "Filter by severity (high, medium, low)"
// @Success 200 {object} []quality.QualityIssue
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/quality/analysis/{id}/issues [get]
func (h *QualityHandler) GetQualityIssues(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid analysis ID format",
			Code:  "INVALID_ID",
		})
		return
	}

	issues, err := h.qualityService.GetQualityIssues(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	// Filter by severity if specified
	if severity := c.Query("severity"); severity != "" {
		var filteredIssues []*quality.QualityIssue
		for _, issue := range issues {
			if issue.Severity == severity {
				filteredIssues = append(filteredIssues, issue)
			}
		}
		issues = filteredIssues
	}

	c.JSON(http.StatusOK, issues)
}

// GetQualityStatistics retrieves quality statistics
// @Summary Get quality statistics
// @Description Retrieves aggregated quality statistics with optional filters
// @Tags quality
// @Produce json
// @Param metric_type query string false "Filter by metric type (vmaf, psnr, ssim)"
// @Param start_date query string false "Start date for filtering (RFC3339 format)"
// @Param end_date query string false "End date for filtering (RFC3339 format)"
// @Param min_score query number false "Minimum score filter"
// @Param max_score query number false "Maximum score filter"
// @Success 200 {object} database.QualityStatistics
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/quality/statistics [get]
func (h *QualityHandler) GetQualityStatistics(c *gin.Context) {
	// Parse query parameters for filtering
	var filters struct {
		MetricType quality.QualityMetricType `form:"metric_type"`
		StartDate  string                    `form:"start_date"`
		EndDate    string                    `form:"end_date"`
		MinScore   *float64                  `form:"min_score"`
		MaxScore   *float64                  `form:"max_score"`
	}

	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid query parameters",
			Code:  "INVALID_PARAMETERS",
		})
		return
	}

	// Convert to database filters
	dbFilters := database.QualityStatisticsFilters{
		MetricType: filters.MetricType,
		MinScore:   filters.MinScore,
		MaxScore:   filters.MaxScore,
	}

	// Parse date filters
	if filters.StartDate != "" {
		startDate, err := time.Parse(time.RFC3339, filters.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid start_date format. Use RFC3339 format",
				Code:  "INVALID_DATE_FORMAT",
			})
			return
		}
		dbFilters.StartDate = &startDate
	}

	if filters.EndDate != "" {
		endDate, err := time.Parse(time.RFC3339, filters.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid end_date format. Use RFC3339 format",
				Code:  "INVALID_DATE_FORMAT",
			})
			return
		}
		dbFilters.EndDate = &endDate
	}

	statistics, err := h.qualityService.GetQualityStatistics(c.Request.Context(), dbFilters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "STATISTICS_RETRIEVAL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, statistics)
}

// GetQualityThresholds retrieves quality thresholds
// @Summary Get quality thresholds
// @Description Retrieves quality thresholds for a specific metric type
// @Tags quality
// @Produce json
// @Param metric_type query string false "Metric type (vmaf, psnr, ssim)"
// @Success 200 {object} quality.QualityThresholds
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/quality/thresholds [get]
func (h *QualityHandler) GetQualityThresholds(c *gin.Context) {
	metricType := quality.QualityMetricType(c.Query("metric_type"))
	
	if metricType == "" {
		// Return default thresholds for all metrics
		thresholds := quality.DefaultQualityThresholds()
		c.JSON(http.StatusOK, thresholds)
		return
	}

	// Validate metric type
	if metricType != quality.MetricVMAF && metricType != quality.MetricPSNR && metricType != quality.MetricSSIM {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid metric type. Must be one of: vmaf, psnr, ssim",
			Code:  "INVALID_METRIC_TYPE",
		})
		return
	}

	thresholds, err := h.qualityService.GetQualityThresholds(c.Request.Context(), metricType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "THRESHOLDS_RETRIEVAL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, thresholds)
}

// DeleteQualityAnalysis deletes a quality analysis
// @Summary Delete quality analysis
// @Description Deletes a quality analysis and all related data
// @Tags quality
// @Param id path string true "Quality analysis ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/quality/analysis/{id} [delete]
func (h *QualityHandler) DeleteQualityAnalysis(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid analysis ID format",
			Code:  "INVALID_ID",
		})
		return
	}

	err = h.qualityService.DeleteQualityAnalysis(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "analysis not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Quality analysis not found",
				Code:  "ANALYSIS_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "DELETION_FAILED",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// RegisterQualityRoutes registers quality-related routes
func (h *QualityHandler) RegisterQualityRoutes(router *gin.RouterGroup) {
	quality := router.Group("/quality")
	{
		quality.POST("/compare", h.CompareQuality)
		quality.POST("/batch", h.BatchCompareQuality)
		quality.GET("/analysis/:id", h.GetQualityAnalysis)
		quality.DELETE("/analysis/:id", h.DeleteQualityAnalysis)
		quality.GET("/analysis/:id/frames", h.GetQualityFrames)
		quality.GET("/analysis/:id/issues", h.GetQualityIssues)
		quality.GET("/comparison/:id", h.GetQualityComparison)
		quality.GET("/statistics", h.GetQualityStatistics)
		quality.GET("/thresholds", h.GetQualityThresholds)
	}
}