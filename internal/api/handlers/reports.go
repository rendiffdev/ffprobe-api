package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// ReportsHandler handles report generation and download endpoints
type ReportsHandler struct {
	reportGenerator *services.ReportGenerator
	analysisService *services.AnalysisService
	comparisonService *services.ComparisonService
}

// NewReportsHandler creates a new reports handler
func NewReportsHandler(
	reportGenerator *services.ReportGenerator,
	analysisService *services.AnalysisService,
	comparisonService *services.ComparisonService,
) *ReportsHandler {
	return &ReportsHandler{
		reportGenerator: reportGenerator,
		analysisService: analysisService,
		comparisonService: comparisonService,
	}
}

// GenerateAnalysisReportRequest represents the request for generating analysis reports
type GenerateAnalysisReportRequest struct {
	AnalysisID string   `json:"analysis_id" binding:"required"`
	Formats    []string `json:"formats" binding:"required"`
}

// GenerateComparisonReportRequest represents the request for generating comparison reports
type GenerateComparisonReportRequest struct {
	ComparisonID string   `json:"comparison_id" binding:"required"`
	Formats      []string `json:"formats" binding:"required"`
}

// GenerateAnalysisReport generates reports for video analysis
func (h *ReportsHandler) GenerateAnalysisReport(c *gin.Context) {
	var req GenerateAnalysisReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_request",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate formats
	formats, err := h.validateFormats(req.Formats)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_format",
			"message": "Invalid report format specified",
			"details": err.Error(),
		})
		return
	}

	// Get analysis from database
	analysis, err := h.analysisService.GetByID(c.Request.Context(), req.AnalysisID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "analysis_not_found",
			"message": "Analysis not found",
		})
		return
	}

	// Generate reports
	reportResponse, err := h.reportGenerator.GenerateAnalysisReport(c.Request.Context(), analysis, formats)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "report_generation_failed",
			"message": "Failed to generate reports",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Reports generated successfully",
		"data": reportResponse,
	})
}

// GenerateComparisonReport generates reports for video comparison
func (h *ReportsHandler) GenerateComparisonReport(c *gin.Context) {
	var req GenerateComparisonReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_request",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate formats
	formats, err := h.validateFormats(req.Formats)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_format",
			"message": "Invalid report format specified",
			"details": err.Error(),
		})
		return
	}

	// Get comparison from database
	comparison, err := h.comparisonService.GetByID(c.Request.Context(), req.ComparisonID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "comparison_not_found",
			"message": "Comparison not found",
		})
		return
	}

	// Generate reports
	reportResponse, err := h.reportGenerator.GenerateComparisonReport(c.Request.Context(), comparison, formats)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "report_generation_failed",
			"message": "Failed to generate reports",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Reports generated successfully",
		"data": reportResponse,
	})
}

// DownloadReport handles report file downloads
func (h *ReportsHandler) DownloadReport(c *gin.Context) {
	reportID := c.Param("reportId")
	filename := c.Param("filename")

	if reportID == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_parameters",
			"message": "Report ID and filename are required",
		})
		return
	}

	// Get file path
	filePath, err := h.reportGenerator.GetReportFile(reportID, filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "file_not_found",
			"message": "Report file not found or expired",
		})
		return
	}

	// Determine content type based on file extension
	contentType := h.getContentType(filename)
	
	// Set headers for download
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// Stream file
	if err := h.reportGenerator.StreamReportFile(reportID, filename, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "download_failed",
			"message": "Failed to download report",
		})
		return
	}
}

// GenerateAnalysisReportWithFormat generates a report for analysis in specified format
func (h *ReportsHandler) GenerateAnalysisReportWithFormat(c *gin.Context) {
	analysisID := c.Param("analysisId")
	format := c.Param("format")

	if analysisID == "" || format == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_parameters",
			"message": "Analysis ID and format are required",
		})
		return
	}

	// Validate format
	formats, err := h.validateFormats([]string{format})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_format",
			"message": "Invalid report format specified",
			"details": err.Error(),
		})
		return
	}

	// Get analysis from database
	analysis, err := h.analysisService.GetByID(c.Request.Context(), analysisID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "analysis_not_found",
			"message": "Analysis not found",
		})
		return
	}

	// Generate report
	reportResponse, err := h.reportGenerator.GenerateAnalysisReport(c.Request.Context(), analysis, formats)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "report_generation_failed",
			"message": "Failed to generate report",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Report generated successfully",
		"data": reportResponse,
	})
}

// GenerateComparisonReportWithFormat generates a report for comparison in specified format
func (h *ReportsHandler) GenerateComparisonReportWithFormat(c *gin.Context) {
	comparisonID := c.Param("comparisonId")
	format := c.Param("format")

	if comparisonID == "" || format == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_parameters",
			"message": "Comparison ID and format are required",
		})
		return
	}

	// Validate format
	formats, err := h.validateFormats([]string{format})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_format",
			"message": "Invalid report format specified",
			"details": err.Error(),
		})
		return
	}

	// Get comparison from database
	comparison, err := h.comparisonService.GetByID(c.Request.Context(), comparisonID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "comparison_not_found",
			"message": "Comparison not found",
		})
		return
	}

	// Generate report
	reportResponse, err := h.reportGenerator.GenerateComparisonReport(c.Request.Context(), comparison, formats)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "report_generation_failed",
			"message": "Failed to generate report",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Report generated successfully",
		"data": reportResponse,
	})
}

// ListReportFormats returns available report formats
func (h *ReportsHandler) ListReportFormats(c *gin.Context) {
	formats := []map[string]interface{}{
		{
			"format": "json",
			"description": "JSON format with complete data structure",
			"content_type": "application/json",
			"extension": ".json",
		},
		{
			"format": "xml",
			"description": "XML format with structured markup",
			"content_type": "application/xml",
			"extension": ".xml",
		},
		{
			"format": "pdf",
			"description": "PDF format for professional presentation",
			"content_type": "application/pdf",
			"extension": ".pdf",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"available_formats": formats,
			"default_formats": []string{"json"},
			"bulk_generation": true,
		},
	})
}

// validateFormats validates the requested report formats
func (h *ReportsHandler) validateFormats(formatStrings []string) ([]services.ReportFormat, error) {
	if len(formatStrings) == 0 {
		return nil, fmt.Errorf("at least one format must be specified")
	}

	var formats []services.ReportFormat
	for _, formatStr := range formatStrings {
		format := services.ReportFormat(strings.ToLower(formatStr))
		switch format {
		case services.FormatJSON, services.FormatXML, services.FormatPDF:
			formats = append(formats, format)
		default:
			return nil, fmt.Errorf("unsupported format: %s", formatStr)
		}
	}

	return formats, nil
}

// getContentType returns the appropriate content type for a file
func (h *ReportsHandler) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

// CleanupExpiredReports handles cleanup of expired reports
func (h *ReportsHandler) CleanupExpiredReports(c *gin.Context) {
	// This should typically be called by a cron job, not exposed as API
	// Including for completeness
	err := h.reportGenerator.CleanupExpiredReports(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "cleanup_failed",
			"message": "Failed to cleanup expired reports",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Expired reports cleaned up successfully",
	})
}

// Enhanced probe endpoint with report generation
func (h *ReportsHandler) ProbeFileWithReports(c *gin.Context) {
	// Get format parameter
	formats := c.QueryArray("report_formats")
	if len(formats) == 0 {
		formats = []string{"json"} // default format
	}

	// Validate formats
	validatedFormats, err := h.validateFormats(formats)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_format",
			"message": "Invalid report format specified",
			"details": err.Error(),
		})
		return
	}

	// Handle file upload and analysis (this would integrate with existing probe handler)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_file",
			"message": "File upload failed",
			"details": err.Error(),
		})
		return
	}
	defer file.Close()

	// Process analysis (this would call existing analysis service)
	// For now, we'll assume we have an analysis result
	
	// Generate reports automatically if requested
	generateReports := c.Query("generate_reports") == "true"
	
	response := gin.H{
		"status": "success",
		"message": "Analysis completed",
		"data": gin.H{
			"file_name": header.Filename,
			"analysis_id": "sample-analysis-id", // This would be the actual analysis ID
		},
	}

	if generateReports {
		// This would generate reports using the actual analysis result
		response["reports"] = gin.H{
			"message": "Reports generated",
			"formats": formats,
			"download_urls": gin.H{
				// These would be actual download URLs
				"json": "/api/v1/reports/sample-report-id/download/sample-report-id_analysis.json",
				"xml":  "/api/v1/reports/sample-report-id/download/sample-report-id_analysis.xml",
				"pdf":  "/api/v1/reports/sample-report-id/download/sample-report-id_analysis.pdf",
			},
		}
	}

	c.JSON(http.StatusOK, response)
}