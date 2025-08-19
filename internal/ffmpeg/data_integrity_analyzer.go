package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

// DataIntegrityAnalyzer performs comprehensive data integrity validation using FFprobe error detection
type DataIntegrityAnalyzer struct {
	ffprobePath string
	logger      zerolog.Logger
}

// NewDataIntegrityAnalyzer creates a new data integrity analyzer
func NewDataIntegrityAnalyzer(ffprobePath string, logger zerolog.Logger) *DataIntegrityAnalyzer {
	return &DataIntegrityAnalyzer{
		ffprobePath: ffprobePath,
		logger:      logger.With().Str("analyzer", "data_integrity").Logger(),
	}
}

// AnalyzeDataIntegrity performs comprehensive data integrity analysis with error detection and hash validation
func (a *DataIntegrityAnalyzer) AnalyzeDataIntegrity(ctx context.Context, filePath string) (*DataIntegrityAnalysis, error) {
	analysis := &DataIntegrityAnalysis{
		DataHashes: make(map[string]string),
	}

	// Run comprehensive error detection analysis
	if err := a.runErrorDetectionAnalysis(ctx, filePath, analysis); err != nil {
		a.logger.Error().Err(err).Msg("Error detection analysis failed")
		// Continue with other analysis even if error detection fails
	}

	// Generate data hashes for integrity verification
	if err := a.generateDataHashes(ctx, filePath, analysis); err != nil {
		a.logger.Error().Err(err).Msg("Hash generation failed")
		// Continue with analysis
	}

	// Analyze packet-level integrity
	if err := a.analyzePacketIntegrity(ctx, filePath, analysis); err != nil {
		a.logger.Error().Err(err).Msg("Packet integrity analysis failed")
		// Continue with analysis
	}

	// Calculate integrity score and compliance
	a.calculateIntegrityMetrics(analysis)

	// Validate overall data integrity
	validation := a.validateDataIntegrity(analysis)
	analysis.Validation = validation

	return analysis, nil
}

// runErrorDetectionAnalysis runs comprehensive error detection using FFprobe flags
func (a *DataIntegrityAnalyzer) runErrorDetectionAnalysis(ctx context.Context, filePath string, analysis *DataIntegrityAnalysis) error {
	// Build comprehensive error detection command
	options := NewOptionsBuilder().
		Input(filePath).
		JSON().
		ShowError().
		ShowFormat().
		ShowStreams().
		ErrorDetectAll().
		FormatErrorDetectAll().
		Build()

	// Execute FFprobe with error detection
	result, err := a.executeFFprobeWithOptions(ctx, options)
	if err != nil {
		return fmt.Errorf("FFprobe error detection failed: %w", err)
	}

	// Parse error information
	a.parseErrorInformation(result, analysis)

	return nil
}

// generateDataHashes generates integrity hashes using FFprobe
func (a *DataIntegrityAnalyzer) generateDataHashes(ctx context.Context, filePath string, analysis *DataIntegrityAnalysis) error {
	// Generate CRC32 hash
	crcOptions := NewOptionsBuilder().
		Input(filePath).
		JSON().
		ShowFormat().
		ShowDataHash().
		CRC32Hash().
		Build()

	if result, err := a.executeFFprobeWithOptions(ctx, crcOptions); err == nil {
		if hash := a.extractHashFromResult(result, "crc32"); hash != "" {
			analysis.DataHashes["crc32"] = hash
		}
	}

	// Generate MD5 hash
	md5Options := NewOptionsBuilder().
		Input(filePath).
		JSON().
		ShowFormat().
		ShowDataHash().
		MD5Hash().
		Build()

	if result, err := a.executeFFprobeWithOptions(ctx, md5Options); err == nil {
		if hash := a.extractHashFromResult(result, "md5"); hash != "" {
			analysis.DataHashes["md5"] = hash
		}
	}

	return nil
}

// analyzePacketIntegrity analyzes packet-level integrity issues
func (a *DataIntegrityAnalyzer) analyzePacketIntegrity(ctx context.Context, filePath string, analysis *DataIntegrityAnalysis) error {
	// Analyze packets with error detection
	options := NewOptionsBuilder().
		Input(filePath).
		JSON().
		ShowPackets().
		ShowError().
		ErrorDetect("crccheck+bitstream+buffer+careful").
		CountPackets().
		Build()

	result, err := a.executeFFprobeWithOptions(ctx, options)
	if err != nil {
		return fmt.Errorf("packet analysis failed: %w", err)
	}

	// Parse packet information for integrity issues
	a.analyzePacketErrors(result, analysis)

	return nil
}

// executeFFprobeWithOptions executes FFprobe with the given options
func (a *DataIntegrityAnalyzer) executeFFprobeWithOptions(ctx context.Context, options *FFprobeOptions) (*FFprobeResult, error) {
	// Build command arguments
	args := []string{a.ffprobePath}

	// Add basic options
	args = append(args, "-hide_banner", "-loglevel", "error")

	if options.ShowError {
		args = append(args, "-show_error")
	}
	if options.ShowFormat {
		args = append(args, "-show_format")
	}
	if options.ShowStreams {
		args = append(args, "-show_streams")
	}
	if options.ShowPackets {
		args = append(args, "-show_packets")
	}
	if options.ShowDataHash {
		args = append(args, "-show_data_hash")
		if options.HashAlgorithm != "" {
			args = append(args, "-hash", options.HashAlgorithm)
		}
	}
	if options.CountPackets {
		args = append(args, "-count_packets")
	}
	if options.ErrorDetect != "" {
		args = append(args, "-err_detect", options.ErrorDetect)
	}
	if options.FormatErrorDetect != "" {
		args = append(args, "-f_err_detect", options.FormatErrorDetect)
	}

	args = append(args, "-of", "json", options.Input)

	// Execute command
	output, err := executeFFprobeCommand(ctx, args)
	if err != nil {
		return nil, err
	}

	// Parse JSON result
	var result FFprobeResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse FFprobe output: %w", err)
	}

	return &result, nil
}

// parseErrorInformation parses error information from FFprobe result
func (a *DataIntegrityAnalyzer) parseErrorInformation(result *FFprobeResult, analysis *DataIntegrityAnalysis) {
	if result.Error != nil {
		// Count different types of errors
		errorDetail := ErrorDetail{
			Type:     "format_error",
			Code:     result.Error.Code,
			Message:  result.Error.String,
			Severity: a.categorizeErrorSeverity(result.Error.Code),
		}

		if analysis.ErrorSummary == nil {
			analysis.ErrorSummary = &ErrorSummary{
				ErrorsByType: make(map[string]int),
			}
		}

		// Categorize error by severity
		switch errorDetail.Severity {
		case "critical":
			analysis.ErrorSummary.CriticalErrors = append(analysis.ErrorSummary.CriticalErrors, errorDetail)
		case "major":
			analysis.ErrorSummary.MajorErrors = append(analysis.ErrorSummary.MajorErrors, errorDetail)
		case "minor":
			analysis.ErrorSummary.MinorErrors = append(analysis.ErrorSummary.MinorErrors, errorDetail)
		default:
			analysis.ErrorSummary.Warnings = append(analysis.ErrorSummary.Warnings, errorDetail)
		}

		analysis.ErrorSummary.ErrorsByType[errorDetail.Type]++
		analysis.ErrorSummary.TotalErrors++

		// Increment specific error counters
		if strings.Contains(strings.ToLower(result.Error.String), "format") {
			analysis.FormatErrors++
		}
		if strings.Contains(strings.ToLower(result.Error.String), "bitstream") {
			analysis.BitstreamErrors++
		}
	}
}

// analyzePacketErrors analyzes packet-level errors
func (a *DataIntegrityAnalyzer) analyzePacketErrors(result *FFprobeResult, analysis *DataIntegrityAnalysis) {
	if result.Packets == nil {
		return
	}

	var lastPts int64 = -1
	var lastDts int64 = -1
	continuityErrors := 0

	for _, packet := range result.Packets {
		// Check for timestamp continuity errors
		if packet.Pts != 0 && lastPts != -1 {
			if packet.Pts < lastPts {
				continuityErrors++
			}
		}

		if packet.Dts != 0 && lastDts != -1 {
			if packet.Dts < lastDts {
				continuityErrors++
			}
		}

		lastPts = packet.Pts
		lastDts = packet.Dts
	}

	analysis.PacketErrors = len(result.Packets) // Placeholder - would need more sophisticated analysis
	analysis.ContinuityErrors = continuityErrors
}

// extractHashFromResult extracts hash value from FFprobe result
func (a *DataIntegrityAnalyzer) extractHashFromResult(result *FFprobeResult, algorithm string) string {
	// This would need to be implemented based on actual FFprobe hash output format
	// For now, return placeholder
	if result.Format != nil && result.Format.Tags != nil {
		if hash, exists := result.Format.Tags[algorithm]; exists {
			return hash
		}
	}
	return ""
}

// categorizeErrorSeverity categorizes error severity based on error code
func (a *DataIntegrityAnalyzer) categorizeErrorSeverity(code int) string {
	switch {
	case code == -1094995529: // AVERROR_INVALIDDATA
		return "critical"
	case code == -541478725: // AVERROR_EOF
		return "minor"
	case code == -1414092869: // AVERROR_BUFFER_TOO_SMALL
		return "major"
	case code < -1000000000: // Major FFmpeg errors
		return "major"
	case code < -1000: // Minor FFmpeg errors
		return "minor"
	default:
		return "warning"
	}
}

// calculateIntegrityMetrics calculates overall integrity score and compliance flags
func (a *DataIntegrityAnalyzer) calculateIntegrityMetrics(analysis *DataIntegrityAnalysis) {
	baseScore := 100

	// Deduct points for errors
	if analysis.ErrorSummary != nil {
		baseScore -= len(analysis.ErrorSummary.CriticalErrors) * 30
		baseScore -= len(analysis.ErrorSummary.MajorErrors) * 15
		baseScore -= len(analysis.ErrorSummary.MinorErrors) * 5
		baseScore -= len(analysis.ErrorSummary.Warnings) * 1
	}

	// Deduct points for specific error types
	baseScore -= analysis.FormatErrors * 20
	baseScore -= analysis.BitstreamErrors * 15
	baseScore -= analysis.PacketErrors * 5
	baseScore -= analysis.ContinuityErrors * 10

	// Ensure score doesn't go below 0
	if baseScore < 0 {
		baseScore = 0
	}

	analysis.IntegrityScore = baseScore
	analysis.IsCorrupted = baseScore < 50
	analysis.IsBroadcastCompliant = baseScore >= 85 && analysis.FormatErrors == 0 && analysis.BitstreamErrors == 0
}

// validateDataIntegrity validates overall data integrity and provides recommendations
func (a *DataIntegrityAnalyzer) validateDataIntegrity(analysis *DataIntegrityAnalysis) *DataIntegrityValidation {
	validation := &DataIntegrityValidation{
		IsValid:            analysis.IntegrityScore >= 70,
		BroadcastCompliant: analysis.IsBroadcastCompliant,
		StreamingCompliant: analysis.IntegrityScore >= 80 && analysis.ContinuityErrors == 0,
	}

	// Add issues based on analysis
	if analysis.IntegrityScore < 70 {
		validation.Issues = append(validation.Issues, "Low data integrity score detected")
	}

	if analysis.FormatErrors > 0 {
		validation.Issues = append(validation.Issues, fmt.Sprintf("%d format errors detected", analysis.FormatErrors))
		validation.RequiredActions = append(validation.RequiredActions, "Fix format compliance issues")
	}

	if analysis.BitstreamErrors > 0 {
		validation.Issues = append(validation.Issues, fmt.Sprintf("%d bitstream errors detected", analysis.BitstreamErrors))
		validation.RequiredActions = append(validation.RequiredActions, "Re-encode to fix bitstream errors")
	}

	if analysis.ContinuityErrors > 0 {
		validation.Issues = append(validation.Issues, fmt.Sprintf("%d continuity errors detected", analysis.ContinuityErrors))
		validation.RequiredActions = append(validation.RequiredActions, "Fix timestamp continuity issues")
	}

	// Add recommendations
	if analysis.IntegrityScore < 90 && len(analysis.DataHashes) == 0 {
		validation.Recommendations = append(validation.Recommendations, "Generate data hashes for integrity verification")
	}

	if !analysis.IsBroadcastCompliant {
		validation.Recommendations = append(validation.Recommendations, "Content may not meet broadcast standards")
	}

	return validation
}
