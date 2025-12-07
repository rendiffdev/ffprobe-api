package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// AFDAnalyzer handles Active Format Description (AFD) analysis
type AFDAnalyzer struct {
	ffprobePath string
	logger      zerolog.Logger
}

// NewAFDAnalyzer creates a new AFD analyzer
func NewAFDAnalyzer(ffprobePath string, logger zerolog.Logger) *AFDAnalyzer {
	return &AFDAnalyzer{
		ffprobePath: ffprobePath,
		logger:      logger,
	}
}

// AFDAnalysis contains comprehensive AFD analysis
type AFDAnalysis struct {
	HasAFD              bool                 `json:"has_afd"`
	AFDStreams          map[int]*AFDInfo     `json:"afd_streams,omitempty"`
	PrimaryAFD          *AFDInfo             `json:"primary_afd,omitempty"`
	AFDChanges          []AFDChange          `json:"afd_changes,omitempty"`
	AspectRatioInfo     *AspectRatioInfo     `json:"aspect_ratio_info,omitempty"`
	ValidationResults   *AFDValidation       `json:"validation_results,omitempty"`
	BroadcastCompliance *BroadcastCompliance `json:"broadcast_compliance,omitempty"`
}

// AFDInfo contains detailed AFD information
type AFDInfo struct {
	StreamIndex        int      `json:"stream_index"`
	AFDValue           int      `json:"afd_value"`         // 4-bit AFD value (0-15)
	AFDDescription     string   `json:"afd_description"`   // Human-readable description
	AspectRatio        string   `json:"aspect_ratio"`      // "4:3", "16:9", etc.
	PresentationMode   string   `json:"presentation_mode"` // "letterbox", "center_cut", "full_frame", etc.
	ProtectedArea      string   `json:"protected_area"`    // "14:9", "4:3", "16:9"
	FirstDetectedFrame int      `json:"first_detected_frame"`
	LastDetectedFrame  int      `json:"last_detected_frame"`
	Confidence         float64  `json:"confidence"`
	IsValid            bool     `json:"is_valid"`
	Issues             []string `json:"issues,omitempty"`
}

// AFDChange represents changes in AFD signaling throughout the content
type AFDChange struct {
	FrameNumber    int     `json:"frame_number"`
	Timestamp      float64 `json:"timestamp"`
	OldAFDValue    int     `json:"old_afd_value"`
	NewAFDValue    int     `json:"new_afd_value"`
	OldDescription string  `json:"old_description"`
	NewDescription string  `json:"new_description"`
	Reason         string  `json:"reason"`
}

// AspectRatioInfo contains detailed aspect ratio analysis
type AspectRatioInfo struct {
	DisplayAspectRatio   string  `json:"display_aspect_ratio"`
	SampleAspectRatio    string  `json:"sample_aspect_ratio"`
	PixelAspectRatio     float64 `json:"pixel_aspect_ratio"`
	EffectiveAspectRatio string  `json:"effective_aspect_ratio"`
	IsAnamorphic         bool    `json:"is_anamorphic"`
	AspectRatioCategory  string  `json:"aspect_ratio_category"`
}

// AFDValidation contains AFD validation results
type AFDValidation struct {
	IsValid              bool     `json:"is_valid"`
	IsConsistent         bool     `json:"is_consistent"`
	HasValidTransitions  bool     `json:"has_valid_transitions"`
	IsBroadcastCompliant bool     `json:"is_broadcast_compliant"`
	Issues               []string `json:"issues,omitempty"`
	Warnings             []string `json:"warnings,omitempty"`
	Recommendations      []string `json:"recommendations,omitempty"`
}

// BroadcastCompliance contains broadcast standard compliance information
type BroadcastCompliance struct {
	ATSC             bool     `json:"atsc_compliant"`            // ATSC A/53 compliance
	DVB              bool     `json:"dvb_compliant"`             // DVB compliance
	ARIB             bool     `json:"arib_compliant"`            // ARIB compliance (Japan)
	SMPTE            bool     `json:"smpte_compliant"`           // SMPTE standards
	RecommendedAFD   int      `json:"recommended_afd,omitempty"` // Recommended AFD value
	ComplianceIssues []string `json:"compliance_issues,omitempty"`
}

// AFD value definitions according to ATSC A/53 and SMPTE 2016-1
var afdDefinitions = map[int]string{
	0:  "Undefined/Reserved",
	1:  "Reserved",
	2:  "Box 16:9 (top)",
	3:  "Box 14:9 (top)",
	4:  "Box > 16:9 (center)",
	5:  "Reserved",
	6:  "Reserved",
	7:  "Reserved",
	8:  "Full frame 4:3 (center)",
	9:  "Full frame 4:3 (center, shoot & protect 14:9)",
	10: "Full frame 16:9 (center)",
	11: "Full frame 14:9 (center)",
	12: "Reserved",
	13: "Full frame 4:3 (center, shoot & protect 4:3)",
	14: "Full frame 16:9 (center, shoot & protect 14:9)",
	15: "Full frame 16:9 (center, shoot & protect 4:3)",
}

var afdPresentationModes = map[int]string{
	0:  "undefined",
	1:  "reserved",
	2:  "letterbox",
	3:  "letterbox_with_protection",
	4:  "full_frame",
	8:  "center_cut",
	9:  "center_cut_with_protection",
	10: "full_frame",
	11: "center_cut",
	13: "center_cut_with_protection",
	14: "letterbox_with_protection",
	15: "center_cut_with_protection",
}

// AnalyzeAFD performs comprehensive AFD analysis
func (aa *AFDAnalyzer) AnalyzeAFD(ctx context.Context, filePath string, streams []StreamInfo) (*AFDAnalysis, error) {
	analysis := &AFDAnalysis{
		AFDStreams: make(map[int]*AFDInfo),
		AFDChanges: []AFDChange{},
	}

	// Step 1: Analyze aspect ratio information from stream metadata
	aa.analyzeAspectRatioInfo(streams, analysis)

	// Step 2: Extract AFD from user data (H.264/H.265 SEI messages)
	if err := aa.extractAFDFromUserData(ctx, filePath, analysis); err != nil {
		aa.logger.Warn().Err(err).Msg("Failed to extract AFD from user data")
	}

	// Step 3: Detect AFD from video characteristics and metadata
	if err := aa.detectAFDFromVideoCharacteristics(ctx, filePath, streams, analysis); err != nil {
		aa.logger.Warn().Err(err).Msg("Failed to detect AFD from video characteristics")
	}

	// Step 4: Analyze AFD changes throughout the content
	if err := aa.analyzeAFDChanges(ctx, filePath, analysis); err != nil {
		aa.logger.Warn().Err(err).Msg("Failed to analyze AFD changes")
	}

	// Step 5: Determine primary AFD
	aa.determinePrimaryAFD(analysis)

	// Step 6: Validate AFD compliance and consistency
	analysis.ValidationResults = aa.validateAFD(analysis)

	// Step 7: Check broadcast compliance
	analysis.BroadcastCompliance = aa.checkBroadcastCompliance(analysis)

	return analysis, nil
}

// analyzeAspectRatioInfo extracts aspect ratio information from stream metadata
func (aa *AFDAnalyzer) analyzeAspectRatioInfo(streams []StreamInfo, analysis *AFDAnalysis) {
	for _, stream := range streams {
		if strings.ToLower(stream.CodecType) == "video" {
			aspectInfo := &AspectRatioInfo{}

			// Extract display aspect ratio
			if stream.DisplayAspectRatio != "" {
				aspectInfo.DisplayAspectRatio = stream.DisplayAspectRatio
				aspectInfo.EffectiveAspectRatio = stream.DisplayAspectRatio
			}

			// Extract sample aspect ratio
			if stream.SampleAspectRatio != "" {
				aspectInfo.SampleAspectRatio = stream.SampleAspectRatio
				aspectInfo.PixelAspectRatio = aa.parseSampleAspectRatio(stream.SampleAspectRatio)
				aspectInfo.IsAnamorphic = aspectInfo.PixelAspectRatio != 1.0
			}

			// Calculate effective aspect ratio if not provided
			if aspectInfo.EffectiveAspectRatio == "" && stream.Width > 0 && stream.Height > 0 {
				effectiveAR := float64(stream.Width) / float64(stream.Height)
				if aspectInfo.IsAnamorphic {
					effectiveAR *= aspectInfo.PixelAspectRatio
				}
				aspectInfo.EffectiveAspectRatio = aa.formatAspectRatio(effectiveAR)
			}

			// Categorize aspect ratio
			aspectInfo.AspectRatioCategory = aa.categorizeAspectRatio(aspectInfo.EffectiveAspectRatio)

			analysis.AspectRatioInfo = aspectInfo
			break // Use first video stream
		}
	}
}

// extractAFDFromUserData extracts AFD information from H.264/H.265 user data
func (aa *AFDAnalyzer) extractAFDFromUserData(ctx context.Context, filePath string, analysis *AFDAnalysis) error {
	// Use ffprobe to extract user data that might contain AFD information
	cmd := []string{
		aa.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_entries", "frame=side_data,pkt_pts_time",
		"-select_streams", "v:0",
		"-read_intervals", "%+#20", // Sample first 20 frames
		filePath,
	}

	output, err := aa.executeCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to extract user data: %w", err)
	}

	var result struct {
		Frames []struct {
			PktPtsTime string `json:"pkt_pts_time"`
			SideData   []struct {
				Type string                 `json:"type"`
				Data map[string]interface{} `json:"data,omitempty"`
			} `json:"side_data,omitempty"`
		} `json:"frames"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return fmt.Errorf("failed to parse frame JSON: %w", err)
	}

	frameNumber := 0
	for _, frame := range result.Frames {
		frameNumber++

		for _, sideData := range frame.SideData {
			// Look for AFD in various side data types
			if aa.containsAFDData(sideData.Type) {
				if afdValue := aa.extractAFDValue(sideData.Data); afdValue >= 0 {
					afdInfo := &AFDInfo{
						StreamIndex:        0,
						AFDValue:           afdValue,
						AFDDescription:     afdDefinitions[afdValue],
						PresentationMode:   afdPresentationModes[afdValue],
						FirstDetectedFrame: frameNumber,
						LastDetectedFrame:  frameNumber,
						Confidence:         0.9,
						IsValid:            aa.isValidAFDValue(afdValue),
					}

					// Derive aspect ratio from AFD value
					afdInfo.AspectRatio = aa.deriveAspectRatioFromAFD(afdValue)
					afdInfo.ProtectedArea = aa.deriveProtectedAreaFromAFD(afdValue)

					analysis.AFDStreams[0] = afdInfo
					analysis.HasAFD = true
				}
			}
		}
	}

	return nil
}

// detectAFDFromVideoCharacteristics infers AFD from video properties
func (aa *AFDAnalyzer) detectAFDFromVideoCharacteristics(ctx context.Context, filePath string, streams []StreamInfo, analysis *AFDAnalysis) error {
	// If no explicit AFD found, try to infer from video characteristics
	if analysis.HasAFD {
		return nil // Already have explicit AFD
	}

	for _, stream := range streams {
		if strings.ToLower(stream.CodecType) == "video" {
			// Analyze letterboxing/pillarboxing through frame analysis
			if err := aa.detectLetterboxing(ctx, filePath, analysis); err != nil {
				aa.logger.Warn().Err(err).Msg("Failed to detect letterboxing")
			}

			// Infer AFD from aspect ratio
			if analysis.AspectRatioInfo != nil {
				inferredAFD := aa.inferAFDFromAspectRatio(analysis.AspectRatioInfo.EffectiveAspectRatio)
				if inferredAFD >= 0 {
					afdInfo := &AFDInfo{
						StreamIndex:      0,
						AFDValue:         inferredAFD,
						AFDDescription:   afdDefinitions[inferredAFD],
						PresentationMode: afdPresentationModes[inferredAFD],
						AspectRatio:      analysis.AspectRatioInfo.EffectiveAspectRatio,
						Confidence:       0.6, // Lower confidence for inferred AFD
						IsValid:          true,
					}

					analysis.AFDStreams[0] = afdInfo
					analysis.HasAFD = true
				}
			}
			break
		}
	}

	return nil
}

// escapeFFmpegFilterPath escapes a file path for use in FFmpeg filter expressions.
// FFmpeg filter syntax requires special characters to be escaped.
// See: https://ffmpeg.org/ffmpeg-filters.html#Notes-on-filtergraph-escaping
func escapeFFmpegFilterPath(path string) string {
	// Escape backslashes first (must be done before other escapes)
	escaped := strings.ReplaceAll(path, `\`, `\\`)
	// Escape single quotes
	escaped = strings.ReplaceAll(escaped, `'`, `'\''`)
	// Escape colons (used as option separator)
	escaped = strings.ReplaceAll(escaped, `:`, `\:`)
	// Escape commas (used as filter separator)
	escaped = strings.ReplaceAll(escaped, `,`, `\,`)
	// Escape brackets (used in filter chains)
	escaped = strings.ReplaceAll(escaped, `[`, `\[`)
	escaped = strings.ReplaceAll(escaped, `]`, `\]`)
	// Escape semicolons (used as filterchain separator)
	escaped = strings.ReplaceAll(escaped, `;`, `\;`)
	// Wrap in single quotes for additional safety
	return `'` + escaped + `'`
}

// detectLetterboxing analyzes frames to detect letterboxing/pillarboxing
func (aa *AFDAnalyzer) detectLetterboxing(ctx context.Context, filePath string, analysis *AFDAnalysis) error {
	// Escape the file path for FFmpeg filter expression to prevent injection
	escapedPath := escapeFFmpegFilterPath(filePath)

	// Use ffprobe with cropdetect filter to detect black bars
	cmd := []string{
		aa.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_entries", "frame_tags=lavfi.cropdetect.w,lavfi.cropdetect.h,lavfi.cropdetect.x,lavfi.cropdetect.y",
		"-f", "lavfi",
		"-i", fmt.Sprintf("movie=%s,cropdetect=24:16:0", escapedPath),
		"-frames:v", "10",
	}

	output, err := aa.executeCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to detect letterboxing: %w", err)
	}

	// Parse cropdetect results to determine if content is letterboxed/pillarboxed
	var result struct {
		Frames []struct {
			Tags map[string]string `json:"tags"`
		} `json:"frames"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return fmt.Errorf("failed to parse cropdetect JSON: %w", err)
	}

	// Analyze crop detection results
	if len(result.Frames) > 0 {
		for _, frame := range result.Frames {
			if cropW := frame.Tags["lavfi.cropdetect.w"]; cropW != "" {
				if cropH := frame.Tags["lavfi.cropdetect.h"]; cropH != "" {
					// Determine if significant cropping indicates letterboxing
					if aa.isLetterboxed(cropW, cropH) {
						// Update analysis with letterbox detection
						if analysis.AspectRatioInfo != nil {
							analysis.AspectRatioInfo.AspectRatioCategory += " (letterboxed)"
						}
					}
				}
			}
		}
	}

	return nil
}

// analyzeAFDChanges detects changes in AFD signaling throughout the content
func (aa *AFDAnalyzer) analyzeAFDChanges(ctx context.Context, filePath string, analysis *AFDAnalysis) error {
	if !analysis.HasAFD {
		return nil
	}

	// Sample frames throughout the content to detect AFD changes
	cmd := []string{
		aa.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_entries", "frame=side_data,pkt_pts_time",
		"-select_streams", "v:0",
		"-read_intervals", "%+#100", // Sample 100 frames
		filePath,
	}

	output, err := aa.executeCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to analyze AFD changes: %w", err)
	}

	// Parse and detect AFD changes
	var result struct {
		Frames []struct {
			PktPtsTime string `json:"pkt_pts_time"`
			SideData   []struct {
				Type string                 `json:"type"`
				Data map[string]interface{} `json:"data,omitempty"`
			} `json:"side_data,omitempty"`
		} `json:"frames"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return fmt.Errorf("failed to parse AFD changes JSON: %w", err)
	}

	previousAFD := -1
	frameNumber := 0

	for _, frame := range result.Frames {
		frameNumber++
		timestamp, _ := strconv.ParseFloat(frame.PktPtsTime, 64)

		for _, sideData := range frame.SideData {
			if aa.containsAFDData(sideData.Type) {
				if afdValue := aa.extractAFDValue(sideData.Data); afdValue >= 0 {
					if previousAFD >= 0 && afdValue != previousAFD {
						// AFD change detected
						change := AFDChange{
							FrameNumber:    frameNumber,
							Timestamp:      timestamp,
							OldAFDValue:    previousAFD,
							NewAFDValue:    afdValue,
							OldDescription: afdDefinitions[previousAFD],
							NewDescription: afdDefinitions[afdValue],
							Reason:         "AFD signaling change",
						}
						analysis.AFDChanges = append(analysis.AFDChanges, change)
					}
					previousAFD = afdValue
				}
			}
		}
	}

	return nil
}

// determinePrimaryAFD identifies the most reliable AFD value
func (aa *AFDAnalyzer) determinePrimaryAFD(analysis *AFDAnalysis) {
	// Use the AFD with highest confidence
	var bestAFD *AFDInfo
	for _, afdInfo := range analysis.AFDStreams {
		if bestAFD == nil || afdInfo.Confidence > bestAFD.Confidence {
			bestAFD = afdInfo
		}
	}

	if bestAFD != nil {
		analysis.PrimaryAFD = bestAFD
	}
}

// validateAFD performs comprehensive AFD validation
func (aa *AFDAnalyzer) validateAFD(analysis *AFDAnalysis) *AFDValidation {
	validation := &AFDValidation{
		IsValid:              analysis.HasAFD,
		IsConsistent:         true,
		HasValidTransitions:  true,
		IsBroadcastCompliant: true,
		Issues:               []string{},
		Warnings:             []string{},
		Recommendations:      []string{},
	}

	if !analysis.HasAFD {
		validation.Issues = append(validation.Issues, "No AFD signaling detected")
		validation.Recommendations = append(validation.Recommendations, "Consider adding AFD signaling for broadcast compliance")
		return validation
	}

	// Validate primary AFD
	if analysis.PrimaryAFD != nil {
		if !analysis.PrimaryAFD.IsValid {
			validation.Issues = append(validation.Issues, "Invalid AFD value detected")
			validation.IsValid = false
		}

		// Check for reserved AFD values
		if aa.isReservedAFDValue(analysis.PrimaryAFD.AFDValue) {
			validation.Warnings = append(validation.Warnings, "Reserved AFD value in use")
		}
	}

	// Validate AFD changes
	if len(analysis.AFDChanges) > 0 {
		validation.Warnings = append(validation.Warnings, "AFD changes detected - verify intentional")

		// Check for rapid AFD changes
		for i := 1; i < len(analysis.AFDChanges); i++ {
			timeDiff := analysis.AFDChanges[i].Timestamp - analysis.AFDChanges[i-1].Timestamp
			if timeDiff < 1.0 { // Less than 1 second
				validation.Issues = append(validation.Issues, "Rapid AFD changes detected")
				validation.HasValidTransitions = false
				break
			}
		}
	}

	return validation
}

// checkBroadcastCompliance validates against broadcast standards
func (aa *AFDAnalyzer) checkBroadcastCompliance(analysis *AFDAnalysis) *BroadcastCompliance {
	compliance := &BroadcastCompliance{
		ATSC:             true,
		DVB:              true,
		ARIB:             true,
		SMPTE:            true,
		ComplianceIssues: []string{},
	}

	if !analysis.HasAFD {
		compliance.ATSC = false
		compliance.DVB = false
		compliance.ComplianceIssues = append(compliance.ComplianceIssues, "AFD signaling required for broadcast compliance")
		return compliance
	}

	if analysis.PrimaryAFD != nil {
		afdValue := analysis.PrimaryAFD.AFDValue

		// Check ATSC A/53 compliance
		if !aa.isATSCCompliantAFD(afdValue) {
			compliance.ATSC = false
			compliance.ComplianceIssues = append(compliance.ComplianceIssues, "AFD value not ATSC A/53 compliant")
		}

		// Check DVB compliance
		if !aa.isDVBCompliantAFD(afdValue) {
			compliance.DVB = false
			compliance.ComplianceIssues = append(compliance.ComplianceIssues, "AFD value not DVB compliant")
		}

		// Recommend optimal AFD value
		if analysis.AspectRatioInfo != nil {
			compliance.RecommendedAFD = aa.getRecommendedAFD(analysis.AspectRatioInfo.EffectiveAspectRatio)
		}
	}

	return compliance
}

// Helper methods

func (aa *AFDAnalyzer) parseSampleAspectRatio(sarStr string) float64 {
	if strings.Contains(sarStr, ":") {
		parts := strings.Split(sarStr, ":")
		if len(parts) == 2 {
			num, _ := strconv.ParseFloat(parts[0], 64)
			den, _ := strconv.ParseFloat(parts[1], 64)
			if den != 0 {
				return num / den
			}
		}
	}
	return 1.0
}

func (aa *AFDAnalyzer) formatAspectRatio(ratio float64) string {
	// Common aspect ratios
	ratioMap := map[float64]string{
		1.333: "4:3",
		1.777: "16:9",
		1.556: "14:9",
		2.35:  "2.35:1",
		2.39:  "2.39:1",
		1.85:  "1.85:1",
	}

	tolerance := 0.01
	for r, name := range ratioMap {
		if ratio >= r-tolerance && ratio <= r+tolerance {
			return name
		}
	}

	return fmt.Sprintf("%.3f:1", ratio)
}

func (aa *AFDAnalyzer) categorizeAspectRatio(aspectRatio string) string {
	switch aspectRatio {
	case "4:3":
		return "Standard Definition"
	case "16:9":
		return "High Definition Widescreen"
	case "14:9":
		return "Compromise Aspect Ratio"
	case "2.35:1", "2.39:1":
		return "Cinematic Widescreen"
	case "1.85:1":
		return "Theatrical Widescreen"
	default:
		return "Custom Aspect Ratio"
	}
}

func (aa *AFDAnalyzer) containsAFDData(sideDataType string) bool {
	afdTypes := []string{
		"h264_sei", "h265_sei", "afd", "user_data",
		"cea_708", "atsc_a53", "bar_data",
	}

	lowerType := strings.ToLower(sideDataType)
	for _, afdType := range afdTypes {
		if strings.Contains(lowerType, afdType) {
			return true
		}
	}
	return false
}

func (aa *AFDAnalyzer) extractAFDValue(data map[string]interface{}) int {
	// Try to extract AFD value from various data structures
	if afdVal, ok := data["afd"].(float64); ok {
		return int(afdVal)
	}
	if afdVal, ok := data["active_format"].(float64); ok {
		return int(afdVal)
	}
	if afdVal, ok := data["aspect_ratio_info"].(float64); ok {
		return int(afdVal) & 0x0F // AFD is lower 4 bits
	}
	return -1
}

func (aa *AFDAnalyzer) isValidAFDValue(afdValue int) bool {
	return afdValue >= 0 && afdValue <= 15
}

func (aa *AFDAnalyzer) isReservedAFDValue(afdValue int) bool {
	reservedValues := []int{0, 1, 5, 6, 7, 12}
	for _, reserved := range reservedValues {
		if afdValue == reserved {
			return true
		}
	}
	return false
}

func (aa *AFDAnalyzer) deriveAspectRatioFromAFD(afdValue int) string {
	switch afdValue {
	case 8, 9, 13:
		return "4:3"
	case 10, 14, 15:
		return "16:9"
	case 11:
		return "14:9"
	case 2:
		return "16:9" // Box 16:9
	case 3:
		return "14:9" // Box 14:9
	case 4:
		return ">16:9" // Ultra-wide
	default:
		return "Unknown"
	}
}

func (aa *AFDAnalyzer) deriveProtectedAreaFromAFD(afdValue int) string {
	switch afdValue {
	case 9:
		return "14:9"
	case 13, 15:
		return "4:3"
	case 14:
		return "14:9"
	default:
		return "None"
	}
}

func (aa *AFDAnalyzer) inferAFDFromAspectRatio(aspectRatio string) int {
	switch aspectRatio {
	case "4:3":
		return 8 // Full frame 4:3
	case "16:9":
		return 10 // Full frame 16:9
	case "14:9":
		return 11 // Full frame 14:9
	default:
		return -1 // Cannot infer
	}
}

func (aa *AFDAnalyzer) isLetterboxed(cropW, cropH string) bool {
	w, _ := strconv.Atoi(cropW)
	h, _ := strconv.Atoi(cropH)

	if w > 0 && h > 0 {
		ratio := float64(w) / float64(h)
		// Letterboxed if significantly wider than 16:9
		return ratio > 1.85
	}
	return false
}

func (aa *AFDAnalyzer) isATSCCompliantAFD(afdValue int) bool {
	// ATSC A/53 compliant values
	compliantValues := []int{8, 9, 10, 11, 13, 14, 15}
	for _, val := range compliantValues {
		if afdValue == val {
			return true
		}
	}
	return false
}

func (aa *AFDAnalyzer) isDVBCompliantAFD(afdValue int) bool {
	// DVB compliant values (similar to ATSC but may have differences)
	return aa.isATSCCompliantAFD(afdValue)
}

func (aa *AFDAnalyzer) getRecommendedAFD(aspectRatio string) int {
	switch aspectRatio {
	case "4:3":
		return 8 // Full frame 4:3
	case "16:9":
		return 10 // Full frame 16:9
	case "14:9":
		return 11 // Full frame 14:9
	default:
		return 10 // Default to 16:9
	}
}

func (aa *AFDAnalyzer) executeCommand(ctx context.Context, cmd []string) (string, error) {
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	output, err := executeFFprobeCommand(execCtx, cmd)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
