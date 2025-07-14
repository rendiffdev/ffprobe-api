package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/xuri/excelize/v2"
	"github.com/rendiffdev/ffprobe-api/internal/database"
	"github.com/rendiffdev/ffprobe-api/internal/models"
	"github.com/rendiffdev/ffprobe-api/internal/reports"
)

var (
	// ErrReportNotFound is returned when a report is not found
	ErrReportNotFound = errors.New("report not found")
)

// ReportService handles report generation operations
type ReportService struct {
	db             *database.DB
	analysisService *AnalysisService
	storageDir     string
	logger         zerolog.Logger
}

// NewReportService creates a new report service
func NewReportService(db *database.DB, analysisService *AnalysisService, storageDir string, logger zerolog.Logger) *ReportService {
	// Ensure storage directory exists
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		logger.Error().Err(err).Msg("Failed to create storage directory")
	}

	return &ReportService{
		db:              db,
		analysisService: analysisService,
		storageDir:      storageDir,
		logger:          logger,
	}
}

// ReportGenerationOptions contains options for report generation
type ReportGenerationOptions struct {
	AnalysisID uuid.UUID
	UserID     string
	Type       models.ReportType
	Format     models.ReportFormat
	Title      string
	Options    map[string]interface{}
}

// ReportFilters contains filters for listing reports
type ReportFilters struct {
	AnalysisID string
	Type       string
	Format     string
}

// GenerateReport generates a report for an analysis
func (s *ReportService) GenerateReport(ctx context.Context, reportID string, opts ReportGenerationOptions) error {
	// Parse report ID
	reportUUID, err := uuid.Parse(reportID)
	if err != nil {
		return fmt.Errorf("invalid report ID: %w", err)
	}

	// Get analysis data
	analysis, err := s.analysisService.GetAnalysis(ctx, opts.AnalysisID.String())
	if err != nil {
		return fmt.Errorf("failed to get analysis: %w", err)
	}

	// Create report file path
	filename := fmt.Sprintf("%s_%s.%s", reportID, time.Now().Format("20060102_150405"), getFileExtension(opts.Format))
	filePath := filepath.Join(s.storageDir, filename)

	// Generate report content based on format
	var content []byte
	switch opts.Format {
	case models.ReportFormatJSON:
		content, err = s.generateJSONReport(ctx, analysis, opts)
	case models.ReportFormatPDF:
		content, err = s.generatePDFReport(ctx, analysis, opts)
	case models.ReportFormatHTML:
		content, err = s.generateHTMLReport(ctx, analysis, opts)
	case models.ReportFormatCSV:
		content, err = s.generateCSVReport(ctx, analysis, opts)
	case models.ReportFormatXML:
		content, err = s.generateXMLReport(ctx, analysis, opts)
	case models.ReportFormatExcel:
		content, err = s.generateExcelReport(ctx, analysis, opts)
	case models.ReportFormatMarkdown:
		content, err = s.generateMarkdownReport(ctx, analysis, opts)
	case models.ReportFormatText:
		content, err = s.generateTextReport(ctx, analysis, opts)
	default:
		return fmt.Errorf("unsupported report format: %s", opts.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to generate report content: %w", err)
	}

	// Write content to file
	if err := ioutil.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	// Parse user ID
	var userUUID *uuid.UUID
	if opts.UserID != "" {
		id, err := uuid.Parse(opts.UserID)
		if err == nil {
			userUUID = &id
		}
	}

	// Create report record
	report := &models.Report{
		ID:           reportUUID,
		AnalysisID:   opts.AnalysisID,
		UserID:       userUUID,
		ReportType:   opts.Type,
		Format:       opts.Format,
		Title:        opts.Title,
		Description:  fmt.Sprintf("Generated %s report in %s format", opts.Type, opts.Format),
		FilePath:     filePath,
		FileSize:     int64(len(content)),
		DownloadCount: 0,
		IsPublic:     false,
		CreatedAt:    time.Now(),
	}

	// Save report to database
	if err := s.db.CreateReport(ctx, report); err != nil {
		// Clean up file on error
		os.Remove(filePath)
		return fmt.Errorf("failed to save report: %w", err)
	}

	return nil
}

// GetReport retrieves a report by ID
func (s *ReportService) GetReport(ctx context.Context, reportID string) (*models.Report, error) {
	// Parse UUID
	id, err := uuid.Parse(reportID)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}

	// Get report from database
	report, err := s.db.GetReport(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrReportNotFound
		}
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	return report, nil
}

// GetReportContent reads report file content
func (s *ReportService) GetReportContent(ctx context.Context, filePath string) ([]byte, error) {
	// Validate file path is within storage directory
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	absStorageDir, _ := filepath.Abs(s.storageDir)
	if !filepath.HasPrefix(absPath, absStorageDir) {
		return nil, fmt.Errorf("invalid file path: outside storage directory")
	}

	// Read file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read report file: %w", err)
	}

	return content, nil
}

// IncrementDownloadCount increments the download count for a report
func (s *ReportService) IncrementDownloadCount(ctx context.Context, reportID string) error {
	// Parse UUID
	id, err := uuid.Parse(reportID)
	if err != nil {
		return fmt.Errorf("invalid report ID: %w", err)
	}

	return s.db.IncrementReportDownloadCount(ctx, id)
}

// ListReports lists reports for a user
func (s *ReportService) ListReports(ctx context.Context, userID string, filters ReportFilters, limit, offset int) ([]*models.Report, int, error) {
	// Parse user ID if provided
	var userUUID *uuid.UUID
	if userID != "" {
		id, err := uuid.Parse(userID)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid user ID: %w", err)
		}
		userUUID = &id
	}

	// Get reports from database
	reports, total, err := s.db.ListReports(ctx, userUUID, filters.AnalysisID, filters.Type, filters.Format, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list reports: %w", err)
	}

	return reports, total, nil
}

// DeleteReport deletes a report
func (s *ReportService) DeleteReport(ctx context.Context, reportID uuid.UUID) error {
	// Get report to get file path
	report, err := s.db.GetReport(ctx, reportID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return ErrReportNotFound
		}
		return fmt.Errorf("failed to get report: %w", err)
	}

	// Delete file
	if err := os.Remove(report.FilePath); err != nil && !os.IsNotExist(err) {
		s.logger.Error().Err(err).Msg("Failed to delete report file")
	}

	// Delete from database
	if err := s.db.DeleteReport(ctx, reportID); err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	return nil
}

// Report generation methods

func (s *ReportService) generateJSONReport(ctx context.Context, analysis *models.Analysis, opts ReportGenerationOptions) ([]byte, error) {
	report := reports.JSONReport{
		Title:     opts.Title,
		Generated: time.Now(),
		Analysis:  analysis,
	}

	// Add quality metrics if available
	if opts.Type == models.ReportTypeQualityMetrics || opts.Type == models.ReportTypeComparison {
		metrics, err := s.db.GetQualityMetrics(ctx, analysis.ID)
		if err == nil {
			report.QualityMetrics = metrics
		}
	}

	// Add HLS data if available
	if opts.Type == models.ReportTypeHLS {
		hlsAnalysis, err := s.db.GetHLSAnalysisByAnalysisID(ctx, analysis.ID)
		if err == nil {
			report.HLSAnalysis = hlsAnalysis
		}
	}

	return json.MarshalIndent(report, "", "  ")
}

func (s *ReportService) generatePDFReport(ctx context.Context, analysis *models.Analysis, opts ReportGenerationOptions) ([]byte, error) {
	// Create PDF generator
	generator := reports.NewPDFGenerator()
	
	// Generate PDF
	return generator.GeneratePDF(analysis, opts.Title, opts.Options)
}

func (s *ReportService) generateHTMLReport(ctx context.Context, analysis *models.Analysis, opts ReportGenerationOptions) ([]byte, error) {
	// Create HTML generator
	generator := reports.NewHTMLGenerator()
	
	// Generate HTML
	return generator.GenerateHTML(analysis, opts.Title, opts.Options)
}

func (s *ReportService) generateCSVReport(ctx context.Context, analysis *models.Analysis, opts ReportGenerationOptions) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write headers
	headers := []string{"Property", "Value"}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	// Write basic info
	rows := [][]string{
		{"Title", opts.Title},
		{"Generated", time.Now().Format(time.RFC3339)},
		{"Analysis ID", analysis.ID.String()},
		{"File Name", analysis.FileName},
		{"File Size", fmt.Sprintf("%d", analysis.FileSize)},
		{"Status", string(analysis.Status)},
	}

	// Add ffprobe data if available
	if analysis.FFprobeData != nil {
		if format, ok := analysis.FFprobeData["format"].(map[string]interface{}); ok {
			if duration, ok := format["duration"].(string); ok {
				rows = append(rows, []string{"Duration", duration})
			}
			if bitRate, ok := format["bit_rate"].(string); ok {
				rows = append(rows, []string{"Bit Rate", bitRate})
			}
		}
	}

	// Write all rows
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), nil
}

func (s *ReportService) generateXMLReport(ctx context.Context, analysis *models.Analysis, opts ReportGenerationOptions) ([]byte, error) {
	report := reports.XMLReport{
		XMLName:   xml.Name{Local: "report"},
		Title:     opts.Title,
		Generated: time.Now(),
		Analysis:  analysis,
	}

	return xml.MarshalIndent(report, "", "  ")
}

func (s *ReportService) generateExcelReport(ctx context.Context, analysis *models.Analysis, opts ReportGenerationOptions) ([]byte, error) {
	// Create new Excel file
	f := excelize.NewFile()
	
	// Create summary sheet
	sheet := "Summary"
	f.SetCellValue(sheet, "A1", "Report Title")
	f.SetCellValue(sheet, "B1", opts.Title)
	f.SetCellValue(sheet, "A2", "Generated")
	f.SetCellValue(sheet, "B2", time.Now().Format("2006-01-02 15:04:05"))
	f.SetCellValue(sheet, "A3", "Analysis ID")
	f.SetCellValue(sheet, "B3", analysis.ID.String())
	f.SetCellValue(sheet, "A4", "File Name")
	f.SetCellValue(sheet, "B4", analysis.FileName)
	f.SetCellValue(sheet, "A5", "File Size")
	f.SetCellValue(sheet, "B5", analysis.FileSize)

	// Style the headers
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellStyle(sheet, "A1", "A5", style)

	// Add quality metrics sheet if applicable
	if opts.Type == models.ReportTypeQualityMetrics {
		metricsSheet := "Quality Metrics"
		f.NewSheet(metricsSheet)
		
		// Add headers
		f.SetCellValue(metricsSheet, "A1", "Metric Type")
		f.SetCellValue(metricsSheet, "B1", "Overall Score")
		f.SetCellValue(metricsSheet, "C1", "Min Score")
		f.SetCellValue(metricsSheet, "D1", "Max Score")
		f.SetCellValue(metricsSheet, "E1", "Mean Score")
		
		// Get quality metrics
		metrics, err := s.db.GetQualityMetrics(ctx, analysis.ID)
		if err == nil {
			for i, metric := range metrics {
				row := i + 2
				f.SetCellValue(metricsSheet, fmt.Sprintf("A%d", row), string(metric.MetricType))
				f.SetCellValue(metricsSheet, fmt.Sprintf("B%d", row), metric.OverallScore)
				f.SetCellValue(metricsSheet, fmt.Sprintf("C%d", row), metric.MinScore)
				f.SetCellValue(metricsSheet, fmt.Sprintf("D%d", row), metric.MaxScore)
				f.SetCellValue(metricsSheet, fmt.Sprintf("E%d", row), metric.MeanScore)
			}
		}
	}

	// Generate buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *ReportService) generateMarkdownReport(ctx context.Context, analysis *models.Analysis, opts ReportGenerationOptions) ([]byte, error) {
	var buf bytes.Buffer

	// Write header
	buf.WriteString(fmt.Sprintf("# %s\n\n", opts.Title))
	buf.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Write analysis info
	buf.WriteString("## Analysis Information\n\n")
	buf.WriteString(fmt.Sprintf("- **ID:** %s\n", analysis.ID))
	buf.WriteString(fmt.Sprintf("- **File:** %s\n", analysis.FileName))
	buf.WriteString(fmt.Sprintf("- **Size:** %d bytes\n", analysis.FileSize))
	buf.WriteString(fmt.Sprintf("- **Status:** %s\n", analysis.Status))

	// Add ffprobe data if available
	if analysis.FFprobeData != nil {
		buf.WriteString("\n## Media Information\n\n")
		
		if format, ok := analysis.FFprobeData["format"].(map[string]interface{}); ok {
			buf.WriteString("### Format\n\n")
			for key, value := range format {
				buf.WriteString(fmt.Sprintf("- **%s:** %v\n", key, value))
			}
		}

		if streams, ok := analysis.FFprobeData["streams"].([]interface{}); ok {
			buf.WriteString("\n### Streams\n\n")
			for i, stream := range streams {
				if streamMap, ok := stream.(map[string]interface{}); ok {
					buf.WriteString(fmt.Sprintf("#### Stream %d\n\n", i))
					for key, value := range streamMap {
						buf.WriteString(fmt.Sprintf("- **%s:** %v\n", key, value))
					}
					buf.WriteString("\n")
				}
			}
		}
	}

	return buf.Bytes(), nil
}

func (s *ReportService) generateTextReport(ctx context.Context, analysis *models.Analysis, opts ReportGenerationOptions) ([]byte, error) {
	var buf bytes.Buffer

	// Write header
	buf.WriteString(fmt.Sprintf("%s\n", opts.Title))
	buf.WriteString(fmt.Sprintf("%s\n\n", strings.Repeat("=", len(opts.Title))))
	buf.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Write analysis info
	buf.WriteString("Analysis Information:\n")
	buf.WriteString(fmt.Sprintf("  ID: %s\n", analysis.ID))
	buf.WriteString(fmt.Sprintf("  File: %s\n", analysis.FileName))
	buf.WriteString(fmt.Sprintf("  Size: %d bytes\n", analysis.FileSize))
	buf.WriteString(fmt.Sprintf("  Status: %s\n", analysis.Status))

	// Add ffprobe data if available
	if analysis.FFprobeData != nil {
		buf.WriteString("\nMedia Information:\n")
		
		// Format as indented text
		data, _ := json.MarshalIndent(analysis.FFprobeData, "  ", "  ")
		buf.Write(data)
	}

	return buf.Bytes(), nil
}

// Helper function to get file extension based on format
func getFileExtension(format models.ReportFormat) string {
	switch format {
	case models.ReportFormatJSON:
		return "json"
	case models.ReportFormatPDF:
		return "pdf"
	case models.ReportFormatHTML:
		return "html"
	case models.ReportFormatCSV:
		return "csv"
	case models.ReportFormatXML:
		return "xml"
	case models.ReportFormatExcel:
		return "xlsx"
	case models.ReportFormatMarkdown:
		return "md"
	case models.ReportFormatText:
		return "txt"
	default:
		return "bin"
	}
}