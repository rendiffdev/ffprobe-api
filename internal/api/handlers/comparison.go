package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rendiffdev/ffprobe-api/internal/models"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// ComparisonHandler handles video comparison API endpoints
type ComparisonHandler struct {
	comparisonService *services.ComparisonService
}

// NewComparisonHandler creates a new comparison handler
func NewComparisonHandler(comparisonService *services.ComparisonService) *ComparisonHandler {
	return &ComparisonHandler{
		comparisonService: comparisonService,
	}
}

// CreateComparison creates a new video comparison
// @Summary Create video comparison
// @Description Compare two video analyses to evaluate improvements
// @Tags comparisons
// @Accept json
// @Produce json
// @Param comparison body models.CreateComparisonRequest true "Comparison request"
// @Success 201 {object} models.ComparisonResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/comparisons [post]
func (h *ComparisonHandler) CreateComparison(c *gin.Context) {
	var req models.CreateComparisonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Validate request
	if req.OriginalAnalysisID == uuid.Nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation error",
			Message: "original_analysis_id is required",
		})
		return
	}

	if req.ModifiedAnalysisID == uuid.Nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation error",
			Message: "modified_analysis_id is required",
		})
		return
	}

	if req.OriginalAnalysisID == req.ModifiedAnalysisID {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation error",
			Message: "original_analysis_id and modified_analysis_id cannot be the same",
		})
		return
	}

	if req.ComparisonType == "" {
		req.ComparisonType = models.ComparisonTypeFullAnalysis
	}

	response, err := h.comparisonService.CreateComparison(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create comparison",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetComparison retrieves a comparison by ID
// @Summary Get comparison by ID
// @Description Retrieve detailed comparison results
// @Tags comparisons
// @Produce json
// @Param id path string true "Comparison ID"
// @Success 200 {object} models.ComparisonResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/comparisons/{id} [get]
func (h *ComparisonHandler) GetComparison(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid comparison ID",
			Message: "ID must be a valid UUID",
		})
		return
	}

	response, err := h.comparisonService.GetComparison(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "comparison not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Comparison not found",
				Message: "The requested comparison does not exist",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get comparison",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListComparisons lists comparisons with pagination
// @Summary List comparisons
// @Description Get a paginated list of comparisons
// @Tags comparisons
// @Produce json
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Param user_id query string false "Filter by user ID"
// @Success 200 {object} ComparisonListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/comparisons [get]
func (h *ComparisonHandler) ListComparisons(c *gin.Context) {
	// Parse pagination parameters
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Parse user ID filter
	var userID *uuid.UUID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := uuid.Parse(userIDStr); err == nil {
			userID = &id
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid user ID",
				Message: "user_id must be a valid UUID",
			})
			return
		}
	}

	comparisons, err := h.comparisonService.ListComparisons(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to list comparisons",
			Message: err.Error(),
		})
		return
	}

	response := ComparisonListResponse{
		Comparisons: comparisons,
		Total:       len(comparisons),
		Limit:       limit,
		Offset:      offset,
	}

	c.JSON(http.StatusOK, response)
}

// CompareVideos is a convenience endpoint for quick comparison
// @Summary Quick video comparison
// @Description Compare two videos using their analysis IDs with simplified response
// @Tags comparisons
// @Accept json
// @Produce json
// @Param request body QuickComparisonRequest true "Quick comparison request"
// @Success 200 {object} models.ComparisonSummaryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/comparisons/quick [post]
func (h *ComparisonHandler) CompareVideos(c *gin.Context) {
	var req QuickComparisonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Validate request
	if req.OriginalAnalysisID == uuid.Nil || req.ModifiedAnalysisID == uuid.Nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation error",
			Message: "Both original_analysis_id and modified_analysis_id are required",
		})
		return
	}

	// Create full comparison request
	fullReq := &models.CreateComparisonRequest{
		OriginalAnalysisID: req.OriginalAnalysisID,
		ModifiedAnalysisID: req.ModifiedAnalysisID,
		ComparisonType:     models.ComparisonTypeFullAnalysis,
		IncludeLLM:         req.IncludeLLM,
		Focus:              req.Focus,
		Threshold:          req.Threshold,
	}

	// Create comparison
	response, err := h.comparisonService.CreateComparison(c.Request.Context(), fullReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create comparison",
			Message: err.Error(),
		})
		return
	}

	// Wait for completion or return processing status
	if response.Status == models.ComparisonStatusCompleted {
		// Get full comparison details
		fullResponse, err := h.comparisonService.GetComparison(c.Request.Context(), response.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to get comparison results",
				Message: err.Error(),
			})
			return
		}

		// Create summary response
		summary := createSummaryResponse(fullResponse)
		c.JSON(http.StatusOK, summary)
	} else {
		// Return processing status
		c.JSON(http.StatusAccepted, gin.H{
			"comparison_id": response.ID,
			"status":        response.Status,
			"message":       "Comparison is being processed. Check back later for results.",
		})
	}
}

// GetComparisonReport generates a detailed comparison report
// @Summary Get comparison report
// @Description Generate a detailed report for a comparison with formatting options
// @Tags comparisons
// @Produce json
// @Param id path string true "Comparison ID"
// @Param format query string false "Report format" Enums(json, summary, detailed) default(detailed)
// @Success 200 {object} ComparisonReportResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/comparisons/{id}/report [get]
func (h *ComparisonHandler) GetComparisonReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid comparison ID",
			Message: "ID must be a valid UUID",
		})
		return
	}

	format := c.DefaultQuery("format", "detailed")
	if format != "json" && format != "summary" && format != "detailed" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid format",
			Message: "Format must be one of: json, summary, detailed",
		})
		return
	}

	comparison, err := h.comparisonService.GetComparison(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "comparison not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Comparison not found",
				Message: "The requested comparison does not exist",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get comparison",
			Message: err.Error(),
		})
		return
	}

	switch format {
	case "json":
		c.JSON(http.StatusOK, comparison)
	case "summary":
		summary := createSummaryResponse(comparison)
		c.JSON(http.StatusOK, summary)
	case "detailed":
		report := createDetailedReport(comparison)
		c.JSON(http.StatusOK, report)
	}
}

// Request/Response types

type QuickComparisonRequest struct {
	OriginalAnalysisID uuid.UUID `json:"original_analysis_id" binding:"required"`
	ModifiedAnalysisID uuid.UUID `json:"modified_analysis_id" binding:"required"`
	IncludeLLM         bool      `json:"include_llm"`
	Focus              []string  `json:"focus,omitempty"`
	Threshold          *float64  `json:"threshold,omitempty"`
}

type ComparisonListResponse struct {
	Comparisons []*models.ComparisonSummaryResponse `json:"comparisons"`
	Total       int                                 `json:"total"`
	Limit       int                                 `json:"limit"`
	Offset      int                                 `json:"offset"`
}

type ComparisonReportResponse struct {
	ComparisonID       uuid.UUID                  `json:"comparison_id"`
	GeneratedAt        string                     `json:"generated_at"`
	ReportType         string                     `json:"report_type"`
	ExecutiveSummary   *ExecutiveSummary          `json:"executive_summary,omitempty"`
	TechnicalDetails   *TechnicalDetails          `json:"technical_details,omitempty"`
	QualityAssessment  *QualityAssessment         `json:"quality_assessment,omitempty"`
	RecommendationList []string                   `json:"recommendations,omitempty"`
	RawData           *models.ComparisonResponse `json:"raw_data,omitempty"`
}

type ExecutiveSummary struct {
	OverallVerdict    string  `json:"overall_verdict"`
	QualityImprovement float64 `json:"quality_improvement"`
	FileSizeChange     string  `json:"file_size_change"`
	RecommendedAction  string  `json:"recommended_action"`
	KeyFindings        []string `json:"key_findings"`
}

type TechnicalDetails struct {
	VideoChanges  interface{} `json:"video_changes,omitempty"`
	AudioChanges  interface{} `json:"audio_changes,omitempty"`
	FormatChanges interface{} `json:"format_changes,omitempty"`
	BitrateAnalysis interface{} `json:"bitrate_analysis,omitempty"`
}

type QualityAssessment struct {
	OverallScore    float64 `json:"overall_score"`
	VideoScore      float64 `json:"video_score"`
	AudioScore      float64 `json:"audio_score"`
	CompressionScore float64 `json:"compression_score"`
	ComplianceScore float64 `json:"compliance_score"`
	LLMAssessment   *string `json:"llm_assessment,omitempty"`
}

// Helper functions

func createSummaryResponse(comparison *models.ComparisonResponse) *models.ComparisonSummaryResponse {
	summary := &models.ComparisonSummaryResponse{
		ID:             comparison.ID,
		QualityScore:   comparison.QualityScore,
		CreatedAt:      comparison.CreatedAt,
		ProcessingTime: comparison.UpdatedAt.Sub(comparison.CreatedAt),
	}

	if comparison.ComparisonData != nil {
		summary.IssuesFixed = len(comparison.ComparisonData.IssuesFixed)
		summary.NewIssues = len(comparison.ComparisonData.NewIssues)
		summary.FileSizeChange = comparison.ComparisonData.FileSize

		if comparison.ComparisonData.Summary != nil {
			summary.OverallImprovement = comparison.ComparisonData.Summary.OverallImprovement
			summary.QualityVerdict = comparison.ComparisonData.Summary.QualityVerdict
			summary.RecommendedAction = comparison.ComparisonData.Summary.RecommendedAction
		}
	}

	return summary
}

func createDetailedReport(comparison *models.ComparisonResponse) *ComparisonReportResponse {
	report := &ComparisonReportResponse{
		ComparisonID: comparison.ID,
		GeneratedAt:  comparison.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ReportType:   "detailed",
		RawData:      comparison,
	}

	if comparison.ComparisonData != nil && comparison.ComparisonData.Summary != nil {
		summary := comparison.ComparisonData.Summary
		
		// Executive Summary
		fileSizeChange := "No change"
		if comparison.ComparisonData.FileSize != nil {
			if comparison.ComparisonData.FileSize.PercentageChange > 0 {
				fileSizeChange = fmt.Sprintf("Increased by %.1f%%", comparison.ComparisonData.FileSize.PercentageChange)
			} else if comparison.ComparisonData.FileSize.PercentageChange < 0 {
				fileSizeChange = fmt.Sprintf("Reduced by %.1f%%", -comparison.ComparisonData.FileSize.PercentageChange)
			}
		}

		report.ExecutiveSummary = &ExecutiveSummary{
			OverallVerdict:     string(summary.QualityVerdict),
			QualityImprovement: summary.OverallImprovement,
			FileSizeChange:     fileSizeChange,
			RecommendedAction:  string(summary.RecommendedAction),
			KeyFindings:        append(summary.ImprovementAreas, summary.RegressionAreas...),
		}

		// Technical Details
		report.TechnicalDetails = &TechnicalDetails{
			VideoChanges:    comparison.ComparisonData.VideoQuality,
			AudioChanges:    comparison.ComparisonData.AudioQuality,
			FormatChanges:   comparison.ComparisonData.FormatChanges,
			BitrateAnalysis: comparison.ComparisonData.BitrateAnalysis,
		}

		// Recommendations
		report.RecommendationList = comparison.ComparisonData.Recommendations
	}

	// Quality Assessment
	if comparison.QualityScore != nil {
		report.QualityAssessment = &QualityAssessment{
			OverallScore:     comparison.QualityScore.OverallScore,
			VideoScore:       comparison.QualityScore.VideoScore,
			AudioScore:       comparison.QualityScore.AudioScore,
			CompressionScore: comparison.QualityScore.CompressionScore,
			ComplianceScore:  comparison.QualityScore.ComplianceScore,
			LLMAssessment:    comparison.LLMAssessment,
		}
	}

	return report
}