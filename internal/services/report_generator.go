package services

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/rendiffdev/ffprobe-api/internal/models"
)

// ReportFormat represents the available report formats
type ReportFormat string

const (
	FormatJSON ReportFormat = "json"
	FormatXML  ReportFormat = "xml"
	FormatPDF  ReportFormat = "pdf"
)

// ReportGenerator handles generation of reports in multiple formats
type ReportGenerator struct {
	reportsDir   string
	baseURL      string
	expiryHours  int
	templates    map[string]*template.Template
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(reportsDir, baseURL string, expiryHours int) *ReportGenerator {
	generator := &ReportGenerator{
		reportsDir:  reportsDir,
		baseURL:     baseURL,
		expiryHours: expiryHours,
		templates:   make(map[string]*template.Template),
	}
	
	// Load templates
	generator.loadTemplates()
	
	return generator
}

// ReportResponse contains download URLs and metadata
type ReportResponse struct {
	ReportID     string            `json:"report_id"`
	GeneratedAt  time.Time         `json:"generated_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
	DownloadURLs map[string]string `json:"download_urls"`
	Formats      []string          `json:"formats"`
	FileSize     map[string]int64  `json:"file_size"`
}

// GenerateAnalysisReport generates analysis reports in all requested formats
func (rg *ReportGenerator) GenerateAnalysisReport(ctx context.Context, analysis *models.Analysis, formats []ReportFormat) (*ReportResponse, error) {
	reportID := uuid.New().String()
	reportDir := filepath.Join(rg.reportsDir, reportID)
	
	// Create report directory
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create report directory: %w", err)
	}
	
	response := &ReportResponse{
		ReportID:     reportID,
		GeneratedAt:  time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(rg.expiryHours) * time.Hour),
		DownloadURLs: make(map[string]string),
		Formats:      make([]string, 0, len(formats)),
		FileSize:     make(map[string]int64),
	}
	
	// Generate reports in each requested format
	for _, format := range formats {
		filename, size, err := rg.generateAnalysisReportFormat(analysis, reportDir, reportID, format)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s report: %w", format, err)
		}
		
		response.DownloadURLs[string(format)] = fmt.Sprintf("%s/api/v1/reports/%s/download/%s", rg.baseURL, reportID, filename)
		response.Formats = append(response.Formats, string(format))
		response.FileSize[string(format)] = size
	}
	
	return response, nil
}

// GenerateComparisonReport generates comparison reports in all requested formats
func (rg *ReportGenerator) GenerateComparisonReport(ctx context.Context, comparison *models.VideoComparison, formats []ReportFormat) (*ReportResponse, error) {
	reportID := uuid.New().String()
	reportDir := filepath.Join(rg.reportsDir, reportID)
	
	// Create report directory
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create report directory: %w", err)
	}
	
	response := &ReportResponse{
		ReportID:     reportID,
		GeneratedAt:  time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(rg.expiryHours) * time.Hour),
		DownloadURLs: make(map[string]string),
		Formats:      make([]string, 0, len(formats)),
		FileSize:     make(map[string]int64),
	}
	
	// Generate reports in each requested format
	for _, format := range formats {
		filename, size, err := rg.generateComparisonReportFormat(comparison, reportDir, reportID, format)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s report: %w", format, err)
		}
		
		response.DownloadURLs[string(format)] = fmt.Sprintf("%s/api/v1/reports/%s/download/%s", rg.baseURL, reportID, filename)
		response.Formats = append(response.Formats, string(format))
		response.FileSize[string(format)] = size
	}
	
	return response, nil
}

// generateAnalysisReportFormat generates a single format for analysis report
func (rg *ReportGenerator) generateAnalysisReportFormat(analysis *models.Analysis, reportDir, reportID string, format ReportFormat) (string, int64, error) {
	switch format {
	case FormatJSON:
		return rg.generateAnalysisJSON(analysis, reportDir, reportID)
	case FormatXML:
		return rg.generateAnalysisXML(analysis, reportDir, reportID)
	case FormatPDF:
		return rg.generateAnalysisPDF(analysis, reportDir, reportID)
	default:
		return "", 0, fmt.Errorf("unsupported format: %s", format)
	}
}

// generateComparisonReportFormat generates a single format for comparison report
func (rg *ReportGenerator) generateComparisonReportFormat(comparison *models.VideoComparison, reportDir, reportID string, format ReportFormat) (string, int64, error) {
	switch format {
	case FormatJSON:
		return rg.generateComparisonJSON(comparison, reportDir, reportID)
	case FormatXML:
		return rg.generateComparisonXML(comparison, reportDir, reportID)
	case FormatPDF:
		return rg.generateComparisonPDF(comparison, reportDir, reportID)
	default:
		return "", 0, fmt.Errorf("unsupported format: %s", format)
	}
}

// generateAnalysisJSON generates JSON format analysis report
func (rg *ReportGenerator) generateAnalysisJSON(analysis *models.Analysis, reportDir, reportID string) (string, int64, error) {
	filename := fmt.Sprintf("%s_analysis.json", reportID)
	filepath := filepath.Join(reportDir, filename)
	
	// Create enhanced JSON report with metadata
	report := map[string]interface{}{
		"report_metadata": map[string]interface{}{
			"report_id":      reportID,
			"report_type":    "video_analysis",
			"generated_at":   time.Now().Format(time.RFC3339),
			"format":         "json",
			"version":        "1.0",
		},
		"analysis": analysis,
		"summary": map[string]interface{}{
			"file_name":    analysis.FileName,
			"file_size":    analysis.FileSize,
			"source_type":  analysis.SourceType,
			"status":       analysis.Status,
			"video_codec":  rg.extractVideoCodec(analysis),
			"audio_codec":  rg.extractAudioCodec(analysis),
			"resolution":   rg.extractResolution(analysis),
			"bitrate":      rg.extractBitrate(analysis),
		},
	}
	
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", 0, err
	}
	
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return "", 0, err
	}
	
	stat, err := os.Stat(filepath)
	if err != nil {
		return "", 0, err
	}
	
	return filename, stat.Size(), nil
}

// generateAnalysisXML generates XML format analysis report
func (rg *ReportGenerator) generateAnalysisXML(analysis *models.Analysis, reportDir, reportID string) (string, int64, error) {
	filename := fmt.Sprintf("%s_analysis.xml", reportID)
	filepath := filepath.Join(reportDir, filename)
	
	// XML structure for analysis report
	type XMLAnalysisReport struct {
		XMLName     xml.Name    `xml:"video_analysis_report"`
		ReportID    string      `xml:"report_id,attr"`
		GeneratedAt string      `xml:"generated_at,attr"`
		Version     string      `xml:"version,attr"`
		Analysis    interface{} `xml:"analysis"`
		Summary     struct {
			FileName    string `xml:"file_name"`
			FileSize    int64  `xml:"file_size"`
			SourceType  string `xml:"source_type"`
			Status      string `xml:"status"`
			VideoCodec  string `xml:"video_codec"`
			AudioCodec  string `xml:"audio_codec"`
			Resolution  string `xml:"resolution"`
			Bitrate     string `xml:"bitrate"`
		} `xml:"summary"`
	}
	
	report := XMLAnalysisReport{
		ReportID:    reportID,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     "1.0",
		Analysis:    analysis,
	}
	
	report.Summary.FileName = analysis.FileName
	report.Summary.FileSize = analysis.FileSize
	report.Summary.SourceType = analysis.SourceType
	report.Summary.Status = string(analysis.Status)
	report.Summary.VideoCodec = rg.extractVideoCodec(analysis)
	report.Summary.AudioCodec = rg.extractAudioCodec(analysis)
	report.Summary.Resolution = rg.extractResolution(analysis)
	report.Summary.Bitrate = rg.extractBitrate(analysis)
	
	xmlData, err := xml.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", 0, err
	}
	
	// Add XML header
	xmlContent := []byte(xml.Header + string(xmlData))
	
	if err := os.WriteFile(filepath, xmlContent, 0644); err != nil {
		return "", 0, err
	}
	
	stat, err := os.Stat(filepath)
	if err != nil {
		return "", 0, err
	}
	
	return filename, stat.Size(), nil
}

// generateAnalysisPDF generates PDF format analysis report
func (rg *ReportGenerator) generateAnalysisPDF(analysis *models.Analysis, reportDir, reportID string) (string, int64, error) {
	filename := fmt.Sprintf("%s_analysis.pdf", reportID)
	filepath := filepath.Join(reportDir, filename)
	
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Video Analysis Report")
	pdf.Ln(15)
	
	// Report metadata
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, "Report ID:")
	pdf.Cell(150, 6, reportID)
	pdf.Ln(6)
	pdf.Cell(40, 6, "Generated:")
	pdf.Cell(150, 6, time.Now().Format("2006-01-02 15:04:05"))
	pdf.Ln(12)
	
	// File information
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "File Information")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, "File Name:")
	pdf.Cell(150, 6, analysis.FileName)
	pdf.Ln(6)
	pdf.Cell(40, 6, "File Size:")
	pdf.Cell(150, 6, fmt.Sprintf("%d bytes (%.2f MB)", analysis.FileSize, float64(analysis.FileSize)/1024/1024))
	pdf.Ln(6)
	pdf.Cell(40, 6, "Source Type:")
	pdf.Cell(150, 6, analysis.SourceType)
	pdf.Ln(6)
	pdf.Cell(40, 6, "Status:")
	pdf.Cell(150, 6, string(analysis.Status))
	pdf.Ln(12)
	
	// Technical specifications
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Technical Specifications")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, "Video Codec:")
	pdf.Cell(150, 6, rg.extractVideoCodec(analysis))
	pdf.Ln(6)
	pdf.Cell(40, 6, "Audio Codec:")
	pdf.Cell(150, 6, rg.extractAudioCodec(analysis))
	pdf.Ln(6)
	pdf.Cell(40, 6, "Resolution:")
	pdf.Cell(150, 6, rg.extractResolution(analysis))
	pdf.Ln(6)
	pdf.Cell(40, 6, "Bitrate:")
	pdf.Cell(150, 6, rg.extractBitrate(analysis))
	pdf.Ln(12)
	
	// FFprobe Analysis Data
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Analysis Data")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(190, 5, "Raw FFprobe data available in JSON/XML formats")
	pdf.Ln(12)
	
	// AI Analysis (if available)
	if analysis.LLMReport != nil && *analysis.LLMReport != "" {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(190, 8, "AI Analysis")
		pdf.Ln(10)
		
		pdf.SetFont("Arial", "", 9)
		// Split long text into multiple lines
		lines := rg.splitText(*analysis.LLMReport, 80)
		for _, line := range lines {
			pdf.Cell(190, 5, line)
			pdf.Ln(5)
		}
	}
	
	if err := pdf.OutputFileAndClose(filepath); err != nil {
		return "", 0, err
	}
	
	stat, err := os.Stat(filepath)
	if err != nil {
		return "", 0, err
	}
	
	return filename, stat.Size(), nil
}

// generateComparisonJSON generates JSON format comparison report
func (rg *ReportGenerator) generateComparisonJSON(comparison *models.VideoComparison, reportDir, reportID string) (string, int64, error) {
	filename := fmt.Sprintf("%s_comparison.json", reportID)
	filepath := filepath.Join(reportDir, filename)
	
	report := map[string]interface{}{
		"report_metadata": map[string]interface{}{
			"report_id":      reportID,
			"report_type":    "video_comparison",
			"generated_at":   time.Now().Format(time.RFC3339),
			"format":         "json",
			"version":        "1.0",
		},
		"comparison": comparison,
	}
	
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", 0, err
	}
	
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return "", 0, err
	}
	
	stat, err := os.Stat(filepath)
	if err != nil {
		return "", 0, err
	}
	
	return filename, stat.Size(), nil
}

// generateComparisonXML generates XML format comparison report
func (rg *ReportGenerator) generateComparisonXML(comparison *models.VideoComparison, reportDir, reportID string) (string, int64, error) {
	filename := fmt.Sprintf("%s_comparison.xml", reportID)
	filepath := filepath.Join(reportDir, filename)
	
	type XMLComparisonReport struct {
		XMLName     xml.Name    `xml:"video_comparison_report"`
		ReportID    string      `xml:"report_id,attr"`
		GeneratedAt string      `xml:"generated_at,attr"`
		Version     string      `xml:"version,attr"`
		Comparison  interface{} `xml:"comparison"`
	}
	
	report := XMLComparisonReport{
		ReportID:    reportID,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     "1.0",
		Comparison:  comparison,
	}
	
	xmlData, err := xml.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", 0, err
	}
	
	xmlContent := []byte(xml.Header + string(xmlData))
	
	if err := os.WriteFile(filepath, xmlContent, 0644); err != nil {
		return "", 0, err
	}
	
	stat, err := os.Stat(filepath)
	if err != nil {
		return "", 0, err
	}
	
	return filename, stat.Size(), nil
}

// generateComparisonPDF generates PDF format comparison report
func (rg *ReportGenerator) generateComparisonPDF(comparison *models.VideoComparison, reportDir, reportID string) (string, int64, error) {
	filename := fmt.Sprintf("%s_comparison.pdf", reportID)
	filepath := filepath.Join(reportDir, filename)
	
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Video Comparison Report")
	pdf.Ln(15)
	
	// Report metadata
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, "Report ID:")
	pdf.Cell(150, 6, reportID)
	pdf.Ln(6)
	pdf.Cell(40, 6, "Generated:")
	pdf.Cell(150, 6, time.Now().Format("2006-01-02 15:04:05"))
	pdf.Ln(12)
	
	// Comparison summary
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Comparison Summary")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, "Comparison Type:")
	pdf.Cell(150, 6, string(comparison.ComparisonType))
	pdf.Ln(6)
	pdf.Cell(40, 6, "Status:")
	pdf.Cell(150, 6, string(comparison.Status))
	pdf.Ln(12)
	
	// Quality verdict (if available)
	if comparison.QualityScore != nil {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(190, 8, "Quality Assessment")
		pdf.Ln(10)
		
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(40, 6, "Verdict:")
		pdf.Cell(150, 6, string(comparison.QualityScore.Verdict))
		pdf.Ln(6)
		pdf.Cell(40, 6, "Confidence:")
		pdf.Cell(150, 6, fmt.Sprintf("%.2f", comparison.QualityScore.ConfidenceScore))
		pdf.Ln(12)
	}
	
	// AI Assessment (if available)
	if comparison.AIAssessment != "" {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(190, 8, "AI Assessment")
		pdf.Ln(10)
		
		pdf.SetFont("Arial", "", 9)
		lines := rg.splitText(comparison.AIAssessment, 80)
		for _, line := range lines {
			pdf.Cell(190, 5, line)
			pdf.Ln(5)
		}
	}
	
	if err := pdf.OutputFileAndClose(filepath); err != nil {
		return "", 0, err
	}
	
	stat, err := os.Stat(filepath)
	if err != nil {
		return "", 0, err
	}
	
	return filename, stat.Size(), nil
}

// GetReportFile returns the file path for a specific report
func (rg *ReportGenerator) GetReportFile(reportID, filename string) (string, error) {
	reportDir := filepath.Join(rg.reportsDir, reportID)
	filePath := filepath.Join(reportDir, filename)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("report file not found")
	}
	
	return filePath, nil
}

// CleanupExpiredReports removes expired report files
func (rg *ReportGenerator) CleanupExpiredReports(ctx context.Context) error {
	entries, err := os.ReadDir(rg.reportsDir)
	if err != nil {
		return err
	}
	
	cutoff := time.Now().Add(-time.Duration(rg.expiryHours) * time.Hour)
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		reportPath := filepath.Join(rg.reportsDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		if info.ModTime().Before(cutoff) {
			if err := os.RemoveAll(reportPath); err != nil {
				// Log error but continue cleanup
				continue
			}
		}
	}
	
	return nil
}

// Helper functions

func (rg *ReportGenerator) extractVideoCodec(analysis *models.Analysis) string {
	// Extract video codec from FFprobe data
	if len(analysis.FFprobeData.Streams) > 0 {
		var streams []interface{}
		if err := json.Unmarshal(analysis.FFprobeData.Streams, &streams); err == nil {
			for _, stream := range streams {
				if s, ok := stream.(map[string]interface{}); ok {
					if codecType, ok := s["codec_type"].(string); ok && codecType == "video" {
						if codecName, ok := s["codec_name"].(string); ok {
							return codecName
						}
					}
				}
			}
		}
	}
	return "Unknown"
}

func (rg *ReportGenerator) extractAudioCodec(analysis *models.Analysis) string {
	if len(analysis.FFprobeData.Streams) > 0 {
		var streams []interface{}
		if err := json.Unmarshal(analysis.FFprobeData.Streams, &streams); err == nil {
			for _, stream := range streams {
				if s, ok := stream.(map[string]interface{}); ok {
					if codecType, ok := s["codec_type"].(string); ok && codecType == "audio" {
						if codecName, ok := s["codec_name"].(string); ok {
							return codecName
						}
					}
				}
			}
		}
	}
	return "Unknown"
}

func (rg *ReportGenerator) extractResolution(analysis *models.Analysis) string {
	if len(analysis.FFprobeData.Streams) > 0 {
		var streams []interface{}
		if err := json.Unmarshal(analysis.FFprobeData.Streams, &streams); err == nil {
			for _, stream := range streams {
				if s, ok := stream.(map[string]interface{}); ok {
					if codecType, ok := s["codec_type"].(string); ok && codecType == "video" {
						if width, ok := s["width"].(float64); ok {
							if height, ok := s["height"].(float64); ok {
								return fmt.Sprintf("%.0fx%.0f", width, height)
							}
						}
					}
				}
			}
		}
	}
	return "Unknown"
}

func (rg *ReportGenerator) extractBitrate(analysis *models.Analysis) string {
	if len(analysis.FFprobeData.Format) > 0 {
		var format map[string]interface{}
		if err := json.Unmarshal(analysis.FFprobeData.Format, &format); err == nil {
			if bitRate, ok := format["bit_rate"].(string); ok {
				return bitRate + " bps"
			}
		}
	}
	return "Unknown"
}

func (rg *ReportGenerator) splitText(text string, maxLen int) []string {
	var lines []string
	words := bytes.Fields([]byte(text))
	
	var currentLine bytes.Buffer
	for _, word := range words {
		if currentLine.Len()+len(word)+1 > maxLen {
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}
		}
		if currentLine.Len() > 0 {
			currentLine.WriteByte(' ')
		}
		currentLine.Write(word)
	}
	
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}
	
	return lines
}

func (rg *ReportGenerator) loadTemplates() {
	// Load HTML templates for PDF generation if needed
	// This can be expanded to load custom templates
}

// StreamReportFile streams a report file for download
func (rg *ReportGenerator) StreamReportFile(reportID, filename string, w io.Writer) error {
	filePath, err := rg.GetReportFile(reportID, filename)
	if err != nil {
		return err
	}
	
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	_, err = io.Copy(w, file)
	return err
}