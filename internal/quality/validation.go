package quality

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ValidateQualityRequest validates quality comparison request
func ValidateQualityRequest(request *QualityComparisonRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate files
	if err := validateQualityFile(request.ReferenceFile, "reference"); err != nil {
		return err
	}

	if err := validateQualityFile(request.DistortedFile, "distorted"); err != nil {
		return err
	}

	// Validate metrics
	if len(request.Metrics) == 0 {
		return fmt.Errorf("at least one quality metric must be specified")
	}

	for _, metric := range request.Metrics {
		if err := validateMetricType(metric); err != nil {
			return fmt.Errorf("invalid metric %s: %w", metric, err)
		}
	}

	// Validate configuration
	if request.Configuration != nil {
		if err := validateQualityConfiguration(request.Configuration); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}
	}

	return nil
}

// validateQualityFile validates a quality comparison file
func validateQualityFile(filePath, fileType string) error {
	if strings.TrimSpace(filePath) == "" {
		return fmt.Errorf("%s file path cannot be empty", fileType)
	}

	// Check for dangerous characters
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">"}
	for _, char := range dangerousChars {
		if strings.Contains(filePath, char) {
			return fmt.Errorf("%s file path contains dangerous character: %s", fileType, char)
		}
	}

	// For local files, validate existence and readability
	if !strings.Contains(filePath, "://") {
		info, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%s file does not exist: %s", fileType, filePath)
			}
			return fmt.Errorf("cannot access %s file: %w", fileType, err)
		}

		if info.IsDir() {
			return fmt.Errorf("%s file is a directory: %s", fileType, filePath)
		}

		// Check file extension
		ext := strings.ToLower(filepath.Ext(filePath))
		validExts := []string{".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".mpg", ".mpeg"}
		validExt := false
		for _, validExtension := range validExts {
			if ext == validExtension {
				validExt = true
				break
			}
		}

		if !validExt {
			return fmt.Errorf("unsupported %s file format: %s", fileType, ext)
		}

		// Check file size (max 50GB)
		maxSize := int64(50 * 1024 * 1024 * 1024)
		if info.Size() > maxSize {
			return fmt.Errorf("%s file too large: %d bytes (max %d)", fileType, info.Size(), maxSize)
		}
	}

	return nil
}

// validateMetricType validates quality metric type
func validateMetricType(metric QualityMetricType) error {
	validMetrics := []QualityMetricType{
		MetricVMAF,
		MetricPSNR,
		MetricSSIM,
		MetricMSSSIM,
		MetricLPIPS,
	}

	for _, valid := range validMetrics {
		if metric == valid {
			return nil
		}
	}

	return fmt.Errorf("unsupported metric type: %s", metric)
}

// validateQualityConfiguration validates quality analysis configuration
func validateQualityConfiguration(config *QualityConfiguration) error {
	if config.VMAF != nil {
		if err := validateVMAFConfiguration(config.VMAF); err != nil {
			return fmt.Errorf("invalid VMAF configuration: %w", err)
		}
	}

	if config.PSNR != nil {
		if err := validatePSNRConfiguration(config.PSNR); err != nil {
			return fmt.Errorf("invalid PSNR configuration: %w", err)
		}
	}

	if config.SSIM != nil {
		if err := validateSSIMConfiguration(config.SSIM); err != nil {
			return fmt.Errorf("invalid SSIM configuration: %w", err)
		}
	}

	if config.Timeout > 0 && config.Timeout > 2*time.Hour {
		return fmt.Errorf("timeout cannot exceed 2 hours")
	}

	return nil
}

// validateVMAFConfiguration validates VMAF-specific configuration
func validateVMAFConfiguration(config *VMAFConfiguration) error {
	if config.Model != "" {
		// Validate VMAF model format
		validModels := []string{
			"version=vmaf_v0.6.1",
			"version=vmaf_v0.6.1neg",
			"version=vmaf_4k_v0.6.1",
			"version=vmaf_b_v0.6.3",
		}

		validModel := false
		for _, valid := range validModels {
			if strings.Contains(config.Model, valid) {
				validModel = true
				break
			}
		}

		if !validModel {
			return fmt.Errorf("unsupported VMAF model: %s", config.Model)
		}
	}

	if config.SubSampling < 1 || config.SubSampling > 10 {
		return fmt.Errorf("VMAF sub-sampling must be between 1 and 10")
	}

	if config.NThreads < 1 || config.NThreads > 32 {
		return fmt.Errorf("VMAF thread count must be between 1 and 32")
	}

	validOutputFormats := []string{"json", "xml", "csv"}
	validFormat := false
	for _, format := range validOutputFormats {
		if config.OutputFormat == format {
			validFormat = true
			break
		}
	}

	if !validFormat {
		return fmt.Errorf("unsupported VMAF output format: %s", config.OutputFormat)
	}

	return nil
}

// validatePSNRConfiguration validates PSNR-specific configuration
func validatePSNRConfiguration(config *PSNRConfiguration) error {
	if config.ComponentMask != "" {
		validMasks := []string{"Y", "U", "V", "A", "YUV", "YUVA"}
		validMask := false
		for _, mask := range validMasks {
			if config.ComponentMask == mask {
				validMask = true
				break
			}
		}

		if !validMask {
			return fmt.Errorf("invalid PSNR component mask: %s", config.ComponentMask)
		}
	}

	return nil
}

// validateSSIMConfiguration validates SSIM-specific configuration
func validateSSIMConfiguration(config *SSIMConfiguration) error {
	if config.WindowSize < 3 || config.WindowSize > 255 {
		return fmt.Errorf("SSIM window size must be between 3 and 255")
	}

	if config.K1 < 0 || config.K1 > 1 {
		return fmt.Errorf("SSIM K1 parameter must be between 0 and 1")
	}

	if config.K2 < 0 || config.K2 > 1 {
		return fmt.Errorf("SSIM K2 parameter must be between 0 and 1")
	}

	return nil
}

// ValidateQualityResult validates quality analysis result
func ValidateQualityResult(result *QualityResult) error {
	if result == nil {
		return fmt.Errorf("result cannot be nil")
	}

	// Validate status
	validStatuses := []QualityStatus{
		QualityStatusPending,
		QualityStatusProcessing,
		QualityStatusCompleted,
		QualityStatusFailed,
	}

	validStatus := false
	for _, status := range validStatuses {
		if result.Status == status {
			validStatus = true
			break
		}
	}

	if !validStatus {
		return fmt.Errorf("invalid quality status: %s", result.Status)
	}

	// Validate analysis results
	for i, analysis := range result.Analysis {
		if err := validateQualityAnalysis(analysis); err != nil {
			return fmt.Errorf("invalid analysis %d: %w", i, err)
		}
	}

	// Validate processing time
	if result.ProcessingTime < 0 {
		return fmt.Errorf("negative processing time")
	}

	return nil
}

// validateQualityAnalysis validates individual quality analysis
func validateQualityAnalysis(analysis *QualityAnalysis) error {
	if analysis == nil {
		return fmt.Errorf("analysis cannot be nil")
	}

	// Validate metric type
	if err := validateMetricType(analysis.MetricType); err != nil {
		return err
	}

	// Validate scores
	if analysis.Status == QualityStatusCompleted {
		if err := validateScoreRanges(analysis.MetricType, analysis); err != nil {
			return err
		}
	}

	return nil
}

// validateScoreRanges validates score ranges for different metrics
func validateScoreRanges(metric QualityMetricType, analysis *QualityAnalysis) error {
	switch metric {
	case MetricVMAF:
		if analysis.OverallScore < 0 || analysis.OverallScore > 100 {
			return fmt.Errorf("VMAF score out of range [0-100]: %.2f", analysis.OverallScore)
		}
	case MetricSSIM:
		if analysis.OverallScore < 0 || analysis.OverallScore > 1 {
			return fmt.Errorf("SSIM score out of range [0-1]: %.2f", analysis.OverallScore)
		}
	case MetricPSNR:
		if analysis.OverallScore < 0 || analysis.OverallScore > 100 {
			return fmt.Errorf("PSNR score out of range [0-100]: %.2f", analysis.OverallScore)
		}
	}

	// Validate min/max consistency
	if analysis.MinScore > analysis.MaxScore {
		return fmt.Errorf("minimum score (%.2f) cannot be greater than maximum score (%.2f)", 
			analysis.MinScore, analysis.MaxScore)
	}

	if analysis.OverallScore < analysis.MinScore || analysis.OverallScore > analysis.MaxScore {
		return fmt.Errorf("overall score (%.2f) must be within min/max range [%.2f-%.2f]", 
			analysis.OverallScore, analysis.MinScore, analysis.MaxScore)
	}

	return nil
}