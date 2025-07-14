package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/models"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// ReportHandler handles report generation and download requests
type ReportHandler struct {
	reportService *services.ReportService
	logger        zerolog.Logger
}

// NewReportHandler creates a new report handler
func NewReportHandler(reportService *services.ReportService, logger zerolog.Logger) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		logger:        logger,
	}
}

// ReportGenerationRequest represents a request to generate a report
type ReportGenerationRequest struct {
	AnalysisID string                `json:"analysis_id" binding:"required"`
	Format     models.ReportFormat   `json:"format" binding:"required"`
	Type       models.ReportType     `json:"type" binding:"required"`
	Title      string                `json:"title"`
	Options    map[string]interface{} `json:"options"`
}

// ReportResponse represents a report response
type ReportResponse struct {
	ID           string `json:"id"`
	AnalysisID   string `json:"analysis_id"`
	Type         string `json:"type"`
	Format       string `json:"format"`
	Title        string `json:"title"`
	DownloadURL  string `json:"download_url"`
	FileSize     int64  `json:"file_size"`
	CreatedAt    string `json:"created_at"`
	ExpiresAt    string `json:"expires_at,omitempty"`
}

// GenerateReport generates a report for an analysis
// @Summary Generate report
// @Description Generate a report in various formats (JSON, PDF, HTML, CSV, XML, Excel, Markdown, Text)
// @Tags probe
// @Accept json
// @Produce json
// @Param request body ReportGenerationRequest true "Report generation request"
// @Success 202 {object} ReportResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/report [post]
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	var req ReportGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate analysis ID
	analysisID, err := uuid.Parse(req.AnalysisID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid analysis ID format",
		})
		return
	}

	// Get user context
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	// Set default title if not provided
	if req.Title == "" {
		req.Title = fmt.Sprintf("%s Report - %s", req.Type, req.Format)
	}

	// Generate report ID
	reportID := uuid.New().String()

	// Start async report generation
	go func() {
		ctx := c.Request.Context()
		
		// Create report options
		opts := services.ReportGenerationOptions{
			AnalysisID: analysisID,
			UserID:     userIDStr,
			Type:       req.Type,
			Format:     req.Format,
			Title:      req.Title,
			Options:    req.Options,
		}

		// Generate report
		if err := h.reportService.GenerateReport(ctx, reportID, opts); err != nil {
			h.logger.Error().Err(err).Str("report_id", reportID).Msg("Report generation failed")
		}
	}()

	// Generate download URL
	downloadURL := fmt.Sprintf("/api/v1/probe/download/%s", reportID)

	// Return immediate response
	c.JSON(http.StatusAccepted, ReportResponse{
		ID:          reportID,
		AnalysisID:  req.AnalysisID,
		Type:        string(req.Type),
		Format:      string(req.Format),
		Title:       req.Title,
		DownloadURL: downloadURL,
		CreatedAt:   "now",
	})
}

// DownloadReport downloads a generated report
// @Summary Download report
// @Description Download a generated report by ID
// @Tags probe
// @Accept json
// @Produce octet-stream
// @Param id path string true "Report ID"
// @Success 200 {file} binary
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/download/{id} [get]
func (h *ReportHandler) DownloadReport(c *gin.Context) {
	reportID := c.Param("id")
	
	// Validate UUID
	if _, err := uuid.Parse(reportID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid report ID format",
		})
		return
	}

	// Get report from service
	report, err := h.reportService.GetReport(c.Request.Context(), reportID)
	if err != nil {
		if err == services.ErrReportNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Report not found",
			})
			return
		}
		
		h.logger.Error().Err(err).Str("report_id", reportID).Msg("Failed to get report")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve report",
		})
		return
	}

	// Get file content
	content, err := h.reportService.GetReportContent(c.Request.Context(), report.FilePath)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to read report file")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read report file",
		})
		return
	}

	// Increment download count
	go func() {
		if err := h.reportService.IncrementDownloadCount(c.Request.Context(), reportID); err != nil {
			h.logger.Error().Err(err).Msg("Failed to increment download count")
		}
	}()

	// Set appropriate headers based on format
	contentType := getContentType(report.Format)
	filename := fmt.Sprintf("%s.%s", strings.ReplaceAll(report.Title, " ", "_"), strings.ToLower(string(report.Format)))
	
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Length", strconv.FormatInt(report.FileSize, 10))
	
	// Send file content
	c.Data(http.StatusOK, contentType, content)
}

// GetReportStatus gets the status of a report generation
// @Summary Get report status
// @Description Get the status of a report generation by ID
// @Tags probe
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} ReportResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/report/{id} [get]
func (h *ReportHandler) GetReportStatus(c *gin.Context) {
	reportID := c.Param("id")
	
	// Validate UUID
	if _, err := uuid.Parse(reportID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid report ID format",
		})
		return
	}

	// Get report from service
	report, err := h.reportService.GetReport(c.Request.Context(), reportID)
	if err != nil {
		if err == services.ErrReportNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Report not found",
			})
			return
		}
		
		h.logger.Error().Err(err).Str("report_id", reportID).Msg("Failed to get report")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve report",
		})
		return
	}

	// Convert to response
	response := ReportResponse{
		ID:          report.ID.String(),
		AnalysisID:  report.AnalysisID.String(),
		Type:        string(report.ReportType),
		Format:      string(report.Format),
		Title:       report.Title,
		DownloadURL: fmt.Sprintf("/api/v1/probe/download/%s", report.ID),
		FileSize:    report.FileSize,
		CreatedAt:   report.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if report.ExpiresAt != nil {
		response.ExpiresAt = report.ExpiresAt.Format("2006-01-02T15:04:05Z")
	}

	c.JSON(http.StatusOK, response)
}

// ListReports lists reports for the current user
// @Summary List reports
// @Description List all reports for the authenticated user
// @Tags probe
// @Accept json
// @Produce json
// @Param analysis_id query string false "Filter by analysis ID"
// @Param type query string false "Filter by report type"
// @Param format query string false "Filter by report format"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} ReportResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/reports [get]
func (h *ReportHandler) ListReports(c *gin.Context) {
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
	filters := services.ReportFilters{
		AnalysisID: c.Query("analysis_id"),
		Type:       c.Query("type"),
		Format:     c.Query("format"),
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	// List reports
	reports, total, err := h.reportService.ListReports(c.Request.Context(), userIDStr, filters, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list reports")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list reports",
		})
		return
	}

	// Convert to response format
	responses := make([]ReportResponse, len(reports))
	for i, report := range reports {
		responses[i] = ReportResponse{
			ID:          report.ID.String(),
			AnalysisID:  report.AnalysisID.String(),
			Type:        string(report.ReportType),
			Format:      string(report.Format),
			Title:       report.Title,
			DownloadURL: fmt.Sprintf("/api/v1/probe/download/%s", report.ID),
			FileSize:    report.FileSize,
			CreatedAt:   report.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		
		if report.ExpiresAt != nil {
			responses[i].ExpiresAt = report.ExpiresAt.Format("2006-01-02T15:04:05Z")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"reports": responses,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// DeleteReport deletes a report
// @Summary Delete report
// @Description Delete a report by ID
// @Tags probe
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/report/{id} [delete]
func (h *ReportHandler) DeleteReport(c *gin.Context) {
	reportID := c.Param("id")
	
	// Validate UUID
	id, err := uuid.Parse(reportID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid report ID format",
		})
		return
	}

	// Delete report
	if err := h.reportService.DeleteReport(c.Request.Context(), id); err != nil {
		if err == services.ErrReportNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Report not found",
			})
			return
		}
		
		h.logger.Error().Err(err).Str("report_id", reportID).Msg("Failed to delete report")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete report",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// Helper function to get content type based on format
func getContentType(format models.ReportFormat) string {
	switch format {
	case models.ReportFormatJSON:
		return "application/json"
	case models.ReportFormatPDF:
		return "application/pdf"
	case models.ReportFormatHTML:
		return "text/html"
	case models.ReportFormatCSV:
		return "text/csv"
	case models.ReportFormatXML:
		return "application/xml"
	case models.ReportFormatExcel:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case models.ReportFormatMarkdown:
		return "text/markdown"
	case models.ReportFormatText:
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}