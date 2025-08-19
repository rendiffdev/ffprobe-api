package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// HDRAnalyzerOptimized provides industry-standard HDR metadata detection and validation.
// This optimized version follows ITU-R BT.2100, SMPTE ST 2084, SMPTE ST 2094, 
// and other relevant HDR standards for broadcast and streaming compliance.
//
// Key improvements over the original:
//   - Enhanced precision in metadata parsing and validation
//   - Support for additional HDR formats (P3-D65, XYZ, etc.)
//   - Industry-standard validation thresholds and tolerances
//   - Optimized FFprobe command execution with proper error handling
//   - Comprehensive compliance checking for broadcast and streaming
type HDRAnalyzerOptimized struct {
	ffprobePath string
	logger      zerolog.Logger
	timeout     time.Duration
}

// NewHDRAnalyzerOptimized creates a new optimized HDR analyzer.
//
// Parameters:
//   - ffprobePath: Path to FFprobe binary
//   - logger: Structured logger for operation tracking
//
// Returns:
//   - *HDRAnalyzerOptimized: Configured HDR analyzer instance
//
// The analyzer supports all major HDR formats and follows industry standards:
//   - HDR10 (SMPTE ST 2084 + BT.2020)
//   - HDR10+ (SMPTE ST 2094-40 dynamic metadata)
//   - Dolby Vision (all profiles 4, 5, 7, 8, 9)
//   - HLG (ITU-R BT.2100 + ARIB STD-B67)
//   - P3-D65 (DCI-P3 with D65 white point)
func NewHDRAnalyzerOptimized(ffprobePath string, logger zerolog.Logger) *HDRAnalyzerOptimized {
	if ffprobePath == "" {
		ffprobePath = "ffprobe"
	}

	return &HDRAnalyzerOptimized{
		ffprobePath: ffprobePath,
		logger:      logger,
		timeout:     30 * time.Second, // Reasonable timeout for HDR analysis
	}
}

// AnalyzeHDR performs comprehensive HDR metadata analysis with industry-standard validation.
//
// The analysis process:
//   1. Extract stream metadata with optimized FFprobe commands
//   2. Detect HDR format based on color characteristics
//   3. Parse advanced metadata (mastering display, content light level, dynamic metadata)
//   4. Validate compliance against industry standards
//   5. Provide actionable recommendations for compliance issues
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - filePath: Path to video file for HDR analysis
//
// Returns:
//   - *HDRAnalysis: Comprehensive HDR analysis results with validation
//   - error: Error if analysis fails or context is cancelled
func (ha *HDRAnalyzerOptimized) AnalyzeHDR(ctx context.Context, filePath string) (*HDRAnalysis, error) {
	ha.logger.Info().Str("file", filePath).Msg("Starting optimized HDR analysis")
	
	// Apply timeout to context
	timeoutCtx, cancel := context.WithTimeout(ctx, ha.timeout)
	defer cancel()

	analysis := &HDRAnalysis{
		IsHDR:     false,
		HDRFormat: "SDR", // Default to SDR
	}

	// Step 1: Get comprehensive stream metadata
	streamData, err := ha.getOptimizedStreamMetadata(timeoutCtx, filePath)
	if err != nil {
		return analysis, fmt.Errorf("failed to get stream metadata: %w", err)
	}

	if len(streamData.Streams) == 0 {
		return analysis, fmt.Errorf("no video streams found")
	}

	// Analyze the primary video stream
	stream := streamData.Streams[0]
	ha.analyzeBasicColorCharacteristics(stream, analysis)

	// Step 2: Determine HDR format with enhanced detection
	ha.determineHDRFormatOptimized(stream, analysis)

	// Step 3: Extract advanced HDR metadata if HDR content is detected
	if analysis.IsHDR {
		if err := ha.extractAdvancedHDRMetadata(timeoutCtx, filePath, analysis); err != nil {
			ha.logger.Warn().Err(err).Msg("Failed to extract advanced HDR metadata")
		}
	}

	// Step 4: Validate HDR compliance with industry standards
	analysis.Validation = ha.validateHDRComplianceOptimized(analysis)

	ha.logger.Info().
		Bool("is_hdr", analysis.IsHDR).
		Str("hdr_format", analysis.HDRFormat).
		Bool("compliant", analysis.Validation.IsCompliant).
		Msg("HDR analysis completed")

	return analysis, nil
}

// getOptimizedStreamMetadata retrieves comprehensive stream metadata with optimized FFprobe command.
func (ha *HDRAnalyzerOptimized) getOptimizedStreamMetadata(ctx context.Context, filePath string) (*streamMetadataResult, error) {
	// Optimized FFprobe command that gets all relevant HDR metadata in one call
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		"-select_streams", "v:0", // Only primary video stream
		"-show_entries", "stream=index,codec_type,codec_name,profile,level,pix_fmt,color_primaries,color_trc,color_space,color_range,chroma_location,field_order,refs,width,height,display_aspect_ratio,r_frame_rate,avg_frame_rate,time_base,start_pts,duration_ts,duration,bit_rate,max_bit_rate,bits_per_raw_sample",
		filePath,
	}

	output, err := executeFFprobeCommand(ctx, append([]string{ha.ffprobePath}, args...))
	if err != nil {
		return nil, fmt.Errorf("FFprobe command failed: %w", err)
	}

	var result streamMetadataResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse FFprobe output: %w", err)
	}

	return &result, nil
}

// analyzeBasicColorCharacteristics extracts and analyzes basic color characteristics.
func (ha *HDRAnalyzerOptimized) analyzeBasicColorCharacteristics(stream streamMetadata, analysis *HDRAnalysis) {
	// Normalize color metadata field names (FFprobe uses different naming)
	analysis.ColorPrimaries = ha.normalizeColorPrimaries(stream.ColorPrimaries)
	analysis.ColorTransfer = ha.normalizeColorTransfer(stream.ColorTransfer)
	analysis.ColorSpace = ha.normalizeColorSpace(stream.ColorSpace)

	// Store additional useful information
	analysis.PixelFormat = stream.PixFmt
	analysis.BitDepth = ha.extractBitDepth(stream.PixFmt)
	analysis.ColorRange = stream.ColorRange
}

// normalizeColorPrimaries normalizes color primaries to standard naming.
func (ha *HDRAnalyzerOptimized) normalizeColorPrimaries(primaries string) string {
	// Map FFprobe color primaries to standard names
	primariesMap := map[string]string{
		"bt2020":     "bt2020",
		"BT.2020":    "bt2020",
		"smpte432":   "p3-d65",     // P3-D65 (Display P3)
		"bt709":      "bt709",
		"BT.709":     "bt709",
		"smpte170m":  "smpte170m",  // NTSC
		"bt470bg":    "bt470bg",    // PAL
		"smpte240m":  "smpte240m",  // SMPTE 240M
		"film":       "film",       // Film
		"bt470m":     "bt470m",     // NTSC 1953
		"":           "unknown",
	}

	if normalized, exists := primariesMap[primaries]; exists {
		return normalized
	}
	return primaries // Return as-is if not in map
}

// normalizeColorTransfer normalizes color transfer characteristics to standard naming.
func (ha *HDRAnalyzerOptimized) normalizeColorTransfer(transfer string) string {
	transferMap := map[string]string{
		"smpte2084":     "smpte2084",     // PQ (HDR10)
		"arib-std-b67":  "arib-std-b67",  // HLG
		"bt709":         "bt709",         // SDR Gamma
		"BT.709":        "bt709",
		"smpte170m":     "smpte170m",     // NTSC
		"bt470bg":       "bt470bg",       // PAL Gamma
		"smpte240m":     "smpte240m",     // SMPTE 240M
		"linear":        "linear",        // Linear
		"log100":        "log100",        // Log 100:1
		"log316":        "log316",        // Log 316:1
		"smpte2094-40":  "smpte2094-40",  // HDR10+ ST 2094-40
		"smpte2094-10":  "smpte2094-10",  // HDR10+ ST 2094-10
		"":              "unknown",
	}

	if normalized, exists := transferMap[transfer]; exists {
		return normalized
	}
	return transfer
}

// normalizeColorSpace normalizes color space to standard naming.
func (ha *HDRAnalyzerOptimized) normalizeColorSpace(colorSpace string) string {
	colorSpaceMap := map[string]string{
		"bt2020nc":  "bt2020nc",  // BT.2020 non-constant
		"bt2020c":   "bt2020c",   // BT.2020 constant  
		"bt709":     "bt709",     // BT.709
		"BT.709":    "bt709",
		"smpte170m": "smpte170m", // NTSC
		"bt470bg":   "bt470bg",   // PAL
		"smpte240m": "smpte240m", // SMPTE 240M
		"ycgco":     "ycgco",     // YCgCo
		"rgb":       "rgb",       // RGB
		"":          "unknown",
	}

	if normalized, exists := colorSpaceMap[colorSpace]; exists {
		return normalized
	}
	return colorSpace
}

// extractBitDepth extracts bit depth from pixel format.
func (ha *HDRAnalyzerOptimized) extractBitDepth(pixFmt string) int {
	// Extract bit depth from pixel format
	if strings.Contains(pixFmt, "10") {
		return 10
	} else if strings.Contains(pixFmt, "12") {
		return 12
	} else if strings.Contains(pixFmt, "16") {
		return 16
	} else if strings.Contains(pixFmt, "8") || pixFmt == "yuv420p" || pixFmt == "yuv422p" || pixFmt == "yuv444p" {
		return 8
	}
	return 8 // Default to 8-bit
}

// determineHDRFormatOptimized determines HDR format with enhanced accuracy.
func (ha *HDRAnalyzerOptimized) determineHDRFormatOptimized(stream streamMetadata, analysis *HDRAnalysis) {
	// Enhanced HDR detection logic
	primaries := analysis.ColorPrimaries
	transfer := analysis.ColorTransfer
	colorSpace := analysis.ColorSpace
	bitDepth := analysis.BitDepth

	// HDR Format Detection Matrix (following industry standards)
	
	// HDR10: BT.2020 + SMPTE 2084 (PQ) + 10+ bit
	if primaries == "bt2020" && transfer == "smpte2084" && (colorSpace == "bt2020nc" || colorSpace == "bt2020c") {
		analysis.IsHDR = true
		analysis.HDRFormat = "HDR10"
		return
	}

	// HLG: BT.2020 + ARIB STD-B67 + 10+ bit
	if primaries == "bt2020" && transfer == "arib-std-b67" && (colorSpace == "bt2020nc" || colorSpace == "bt2020c") {
		analysis.IsHDR = true
		analysis.HDRFormat = "HLG"
		analysis.HLGCompatible = true
		return
	}

	// HDR10+: Will be detected by dynamic metadata presence
	if primaries == "bt2020" && (transfer == "smpte2094-40" || transfer == "smpte2094-10") {
		analysis.IsHDR = true
		analysis.HDRFormat = "HDR10+"
		return
	}

	// P3-D65 (Wide Color Gamut, not technically HDR but extended gamut)
	if primaries == "p3-d65" && bitDepth >= 10 {
		analysis.IsHDR = true
		analysis.HDRFormat = "P3-D65"
		return
	}

	// Check for potential HDR content with missing or incorrect metadata
	if bitDepth >= 10 && (primaries == "bt2020" || transfer == "smpte2084" || transfer == "arib-std-b67") {
		analysis.IsHDR = true
		analysis.HDRFormat = "HDR (Incomplete Metadata)"
		ha.logger.Warn().
			Str("primaries", primaries).
			Str("transfer", transfer).
			Str("color_space", colorSpace).
			Int("bit_depth", bitDepth).
			Msg("Detected HDR content with incomplete metadata")
		return
	}

	// SDR content
	analysis.IsHDR = false
	analysis.HDRFormat = "SDR"
}

// extractAdvancedHDRMetadata extracts mastering display, content light level, and dynamic metadata.
func (ha *HDRAnalyzerOptimized) extractAdvancedHDRMetadata(ctx context.Context, filePath string, analysis *HDRAnalysis) error {
	// Get side data with optimized command
	args := []string{
		"-v", "quiet",
		"-print_format", "default",
		"-show_frames",
		"-select_streams", "v:0",
		"-read_intervals", "%+#1", // Only first frame for metadata
		"-show_entries", "frame=side_data_list:side_data=side_data_type",
		filePath,
	}

	output, err := executeFFprobeCommand(ctx, append([]string{ha.ffprobePath}, args...))
	if err != nil {
		return fmt.Errorf("failed to get side data: %w", err)
	}

	ha.parseAdvancedMetadataOptimized(output, analysis)
	return nil
}

// parseAdvancedMetadataOptimized parses side data with improved accuracy and error handling.
func (ha *HDRAnalyzerOptimized) parseAdvancedMetadataOptimized(sideData string, analysis *HDRAnalysis) {
	lines := strings.Split(sideData, "\n")
	var currentSideData string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Detect side data types with improved pattern matching
		if strings.Contains(line, "side_data_type=Mastering display metadata") {
			currentSideData = "mastering"
			analysis.MasteringDisplay = &MasteringDisplayMetadata{HasMasteringDisplay: true}
			continue
		}

		if strings.Contains(line, "side_data_type=Content light level metadata") {
			currentSideData = "content_light"
			analysis.ContentLightLevel = &ContentLightLevelData{HasContentLightLevel: true}
			continue
		}

		if strings.Contains(line, "side_data_type=HDR10+") || strings.Contains(line, "HDR10+ Dynamic Metadata") {
			currentSideData = "hdr10plus"
			analysis.HDR10Plus = &HDR10PlusMetadata{Present: true}
			// Update format if it was detected as HDR10
			if analysis.HDRFormat == "HDR10" {
				analysis.HDRFormat = "HDR10+"
			}
			continue
		}

		if strings.Contains(line, "side_data_type=DOVI") || strings.Contains(line, "Dolby Vision") {
			currentSideData = "dolby_vision"
			analysis.DolbyVision = &DolbyVisionMetadata{}
			analysis.HDRFormat = "Dolby Vision"
			continue
		}

		// Parse specific metadata with enhanced precision
		switch currentSideData {
		case "mastering":
			ha.parseMasteringDisplayOptimized(line, analysis.MasteringDisplay)
		case "content_light":
			ha.parseContentLightOptimized(line, analysis.ContentLightLevel)
		case "hdr10plus":
			ha.parseHDR10PlusOptimized(line, analysis.HDR10Plus)
		case "dolby_vision":
			ha.parseDolbyVisionOptimized(line, analysis.DolbyVision)
		}
	}
}

// parseMasteringDisplayOptimized parses mastering display metadata with enhanced precision.
func (ha *HDRAnalyzerOptimized) parseMasteringDisplayOptimized(line string, metadata *MasteringDisplayMetadata) {
	// Enhanced parsing for display primaries with better regex and error handling
	if strings.Contains(line, "display_primaries=") {
		// Pattern: G(13250,34500)B(7500,3000)R(34000,16000)
		primariesRegex := regexp.MustCompile(`G\((\d+),(\d+)\)B\((\d+),(\d+)\)R\((\d+),(\d+)\)`)
		matches := primariesRegex.FindStringSubmatch(line)
		if len(matches) == 7 {
			// Convert from CIE 1931 fixed-point (divide by 50000)
			values := make([]float64, 6)
			for i, match := range matches[1:] {
				if val, err := strconv.ParseFloat(match, 64); err == nil {
					values[i] = val / 50000.0
				}
			}
			
			// Store in RGB order: [0]=Red, [1]=Green, [2]=Blue
			metadata.DisplayPrimariesX[0] = values[4] // Red X
			metadata.DisplayPrimariesY[0] = values[5] // Red Y
			metadata.DisplayPrimariesX[1] = values[0] // Green X
			metadata.DisplayPrimariesY[1] = values[1] // Green Y
			metadata.DisplayPrimariesX[2] = values[2] // Blue X
			metadata.DisplayPrimariesY[2] = values[3] // Blue Y
		}
	}

	// Enhanced white point parsing
	if strings.Contains(line, "white_point=") {
		whitePointRegex := regexp.MustCompile(`white_point=([0-9.]+)/([0-9.]+),([0-9.]+)/([0-9.]+)`)
		matches := whitePointRegex.FindStringSubmatch(line)
		if len(matches) == 5 {
			if wxNum, err1 := strconv.ParseFloat(matches[1], 64); err1 == nil {
				if wxDen, err2 := strconv.ParseFloat(matches[2], 64); err2 == nil && wxDen != 0 {
					metadata.WhitePointX = wxNum / wxDen
				}
			}
			if wyNum, err1 := strconv.ParseFloat(matches[3], 64); err1 == nil {
				if wyDen, err2 := strconv.ParseFloat(matches[4], 64); err2 == nil && wyDen != 0 {
					metadata.WhitePointY = wyNum / wyDen
				}
			}
		}
	}

	// Enhanced luminance parsing with validation
	if strings.Contains(line, "max_luminance=") {
		maxLumRegex := regexp.MustCompile(`max_luminance=([0-9.]+)`)
		matches := maxLumRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.ParseFloat(matches[1], 64); err == nil && val > 0 && val <= 10000 {
				metadata.MaxDisplayLuminance = val
			}
		}
	}

	if strings.Contains(line, "min_luminance=") {
		minLumRegex := regexp.MustCompile(`min_luminance=([0-9.]+)`)
		matches := minLumRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.ParseFloat(matches[1], 64); err == nil && val >= 0 && val <= 1 {
				metadata.MinDisplayLuminance = val
			}
		}
	}
}

// parseContentLightOptimized parses content light level metadata with validation.
func (ha *HDRAnalyzerOptimized) parseContentLightOptimized(line string, metadata *ContentLightLevelData) {
	if strings.Contains(line, "max_content=") {
		maxContentRegex := regexp.MustCompile(`max_content=(\d+)`)
		matches := maxContentRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.Atoi(matches[1]); err == nil && val >= 0 && val <= 10000 {
				metadata.MaxCLL = val
			}
		}
	}

	if strings.Contains(line, "max_average=") {
		maxAvgRegex := regexp.MustCompile(`max_average=(\d+)`)
		matches := maxAvgRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.Atoi(matches[1]); err == nil && val >= 0 && val <= 4000 {
				metadata.MaxFALL = val
			}
		}
	}
}

// parseHDR10PlusOptimized parses HDR10+ metadata with enhanced validation.
func (ha *HDRAnalyzerOptimized) parseHDR10PlusOptimized(line string, metadata *HDR10PlusMetadata) {
	if strings.Contains(line, "application_version=") {
		versionRegex := regexp.MustCompile(`application_version=(\d+)`)
		matches := versionRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.Atoi(matches[1]); err == nil && val >= 0 && val <= 255 {
				metadata.ApplicationVersion = val
			}
		}
	}

	if strings.Contains(line, "num_windows=") {
		windowsRegex := regexp.MustCompile(`num_windows=(\d+)`)
		matches := windowsRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.Atoi(matches[1]); err == nil && val >= 1 && val <= 3 {
				metadata.NumWindows = val
			}
		}
	}
}

// parseDolbyVisionOptimized parses Dolby Vision metadata with profile validation.
func (ha *HDRAnalyzerOptimized) parseDolbyVisionOptimized(line string, metadata *DolbyVisionMetadata) {
	if strings.Contains(line, "dv_profile=") {
		profileRegex := regexp.MustCompile(`dv_profile=(\d+)`)
		matches := profileRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				// Validate against known Dolby Vision profiles
				validProfiles := map[int]bool{4: true, 5: true, 7: true, 8: true, 9: true}
				if validProfiles[val] {
					metadata.Profile = val
				}
			}
		}
	}

	if strings.Contains(line, "dv_level=") {
		levelRegex := regexp.MustCompile(`dv_level=(\d+)`)
		matches := levelRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.Atoi(matches[1]); err == nil && val >= 1 && val <= 13 {
				metadata.Level = val
			}
		}
	}

	metadata.RPUPresent = strings.Contains(line, "rpu_present=1")
	metadata.ELPresent = strings.Contains(line, "el_present=1")
	metadata.BLPresent = strings.Contains(line, "bl_present=1")
}

// validateHDRComplianceOptimized validates HDR compliance with enhanced industry standards.
func (ha *HDRAnalyzerOptimized) validateHDRComplianceOptimized(analysis *HDRAnalysis) *HDRValidation {
	validation := &HDRValidation{
		Standard:        analysis.HDRFormat,
		Issues:          []string{},
		Recommendations: []string{},
		IsCompliant:     true,
	}

	if !analysis.IsHDR {
		return validation // SDR content is compliant by default
	}

	// Validate based on HDR format with industry standards
	switch analysis.HDRFormat {
	case "HDR10":
		ha.validateHDR10ComplianceOptimized(analysis, validation)
	case "HDR10+":
		ha.validateHDR10PlusComplianceOptimized(analysis, validation)
	case "Dolby Vision":
		ha.validateDolbyVisionComplianceOptimized(analysis, validation)
	case "HLG":
		ha.validateHLGComplianceOptimized(analysis, validation)
	case "P3-D65":
		ha.validateP3D65Compliance(analysis, validation)
	default:
		validation.Issues = append(validation.Issues, fmt.Sprintf("Unknown or unsupported HDR format: %s", analysis.HDRFormat))
		validation.IsCompliant = false
	}

	// General HDR validation
	ha.validateGeneralHDRCompliance(analysis, validation)

	return validation
}

// validateHDR10ComplianceOptimized validates HDR10 compliance per SMPTE ST 2084.
func (ha *HDRAnalyzerOptimized) validateHDR10ComplianceOptimized(analysis *HDRAnalysis, validation *HDRValidation) {
	// Essential HDR10 requirements per SMPTE ST 2084 and ITU-R BT.2100
	
	if analysis.ColorPrimaries != "bt2020" {
		validation.Issues = append(validation.Issues, "HDR10 requires BT.2020 color primaries per ITU-R BT.2100")
		validation.IsCompliant = false
	}

	if analysis.ColorTransfer != "smpte2084" {
		validation.Issues = append(validation.Issues, "HDR10 requires SMPTE ST 2084 (PQ) transfer function")
		validation.IsCompliant = false
	}

	if analysis.ColorSpace != "bt2020nc" && analysis.ColorSpace != "bt2020c" {
		validation.Issues = append(validation.Issues, "HDR10 requires BT.2020 non-constant or constant luminance")
		validation.IsCompliant = false
	}

	if analysis.BitDepth < 10 {
		validation.Issues = append(validation.Issues, "HDR10 requires minimum 10-bit color depth")
		validation.IsCompliant = false
	}

	// Mastering display metadata validation
	if analysis.MasteringDisplay == nil || !analysis.MasteringDisplay.HasMasteringDisplay {
		validation.Issues = append(validation.Issues, "HDR10 requires mastering display metadata per SMPTE ST 2086")
		validation.IsCompliant = false
	} else {
		ha.validateMasteringDisplayMetadata(analysis.MasteringDisplay, validation)
	}

	// Content light level recommendations
	if analysis.ContentLightLevel == nil || !analysis.ContentLightLevel.HasContentLightLevel {
		validation.Recommendations = append(validation.Recommendations, 
			"Consider adding content light level metadata (SMPTE ST 2094-10) for optimal HDR10 display adaptation")
	}
}

// validateMasteringDisplayMetadata validates mastering display metadata accuracy.
func (ha *HDRAnalyzerOptimized) validateMasteringDisplayMetadata(metadata *MasteringDisplayMetadata, validation *HDRValidation) {
	// Validate display primaries are within reasonable ranges
	for i := 0; i < 3; i++ {
		x, y := metadata.DisplayPrimariesX[i], metadata.DisplayPrimariesY[i]
		if x < 0 || x > 1 || y < 0 || y > 1 {
			validation.Issues = append(validation.Issues, 
				fmt.Sprintf("Display primary %d coordinates out of valid range [0,1]: (%.4f, %.4f)", i, x, y))
			validation.IsCompliant = false
		}
	}

	// Validate white point
	if metadata.WhitePointX < 0 || metadata.WhitePointX > 1 || metadata.WhitePointY < 0 || metadata.WhitePointY > 1 {
		validation.Issues = append(validation.Issues, 
			fmt.Sprintf("White point coordinates out of valid range: (%.4f, %.4f)", metadata.WhitePointX, metadata.WhitePointY))
		validation.IsCompliant = false
	}

	// Validate luminance ranges per SMPTE ST 2086
	if metadata.MaxDisplayLuminance <= metadata.MinDisplayLuminance {
		validation.Issues = append(validation.Issues, "Maximum luminance must be greater than minimum luminance")
		validation.IsCompliant = false
	}

	if metadata.MaxDisplayLuminance > 10000 {
		validation.Issues = append(validation.Issues, "Maximum display luminance exceeds 10,000 nits (SMPTE ST 2086 limit)")
		validation.IsCompliant = false
	}

	if metadata.MinDisplayLuminance < 0 {
		validation.Issues = append(validation.Issues, "Minimum display luminance cannot be negative")
		validation.IsCompliant = false
	}

	// Check for reasonable luminance values
	if metadata.MaxDisplayLuminance < 100 {
		validation.Recommendations = append(validation.Recommendations, 
			"Maximum luminance appears low for HDR content (< 100 nits)")
	}
}

// validateHLGComplianceOptimized validates HLG compliance per ITU-R BT.2100.
func (ha *HDRAnalyzerOptimized) validateHLGComplianceOptimized(analysis *HDRAnalysis, validation *HDRValidation) {
	if analysis.ColorTransfer != "arib-std-b67" {
		validation.Issues = append(validation.Issues, "HLG requires ARIB STD-B67 transfer function per ITU-R BT.2100")
		validation.IsCompliant = false
	}

	if analysis.ColorPrimaries != "bt2020" {
		validation.Issues = append(validation.Issues, "HLG requires BT.2020 color primaries per ITU-R BT.2100")
		validation.IsCompliant = false
	}

	if analysis.BitDepth < 10 {
		validation.Issues = append(validation.Issues, "HLG requires minimum 10-bit color depth")
		validation.IsCompliant = false
	}

	// HLG typically doesn't require mastering display metadata
	validation.Recommendations = append(validation.Recommendations, 
		"HLG content is backward compatible with SDR displays when properly tone-mapped")
}

// validateHDR10PlusComplianceOptimized validates HDR10+ compliance.
func (ha *HDRAnalyzerOptimized) validateHDR10PlusComplianceOptimized(analysis *HDRAnalysis, validation *HDRValidation) {
	// First validate base HDR10 compliance
	ha.validateHDR10ComplianceOptimized(analysis, validation)

	// HDR10+ specific requirements
	if analysis.HDR10Plus == nil || !analysis.HDR10Plus.Present {
		validation.Issues = append(validation.Issues, "HDR10+ requires dynamic metadata per SMPTE ST 2094-40")
		validation.IsCompliant = false
	}
}

// validateDolbyVisionComplianceOptimized validates Dolby Vision compliance.
func (ha *HDRAnalyzerOptimized) validateDolbyVisionComplianceOptimized(analysis *HDRAnalysis, validation *HDRValidation) {
	if analysis.DolbyVision == nil {
		validation.Issues = append(validation.Issues, "Dolby Vision metadata missing")
		validation.IsCompliant = false
		return
	}

	// Validate profile
	validProfiles := map[int]string{
		4: "Profile 4 (Base Layer + Enhancement Layer)",
		5: "Profile 5 (Base Layer Only)",
		7: "Profile 7 (Single Layer)",
		8: "Profile 8 (Cross-Compatible)",
		9: "Profile 9 (Cross-Compatible with HLG)",
	}

	if profileName, valid := validProfiles[analysis.DolbyVision.Profile]; valid {
		validation.Recommendations = append(validation.Recommendations, 
			fmt.Sprintf("Detected %s", profileName))
	} else {
		validation.Issues = append(validation.Issues, 
			fmt.Sprintf("Invalid Dolby Vision profile: %d", analysis.DolbyVision.Profile))
		validation.IsCompliant = false
	}

	// Profile-specific validation
	if analysis.DolbyVision.Profile != 5 && !analysis.DolbyVision.RPUPresent {
		validation.Issues = append(validation.Issues, "Dolby Vision RPU (Reference Processing Unit) missing for this profile")
		validation.IsCompliant = false
	}
}

// validateP3D65Compliance validates P3-D65 wide color gamut compliance.
func (ha *HDRAnalyzerOptimized) validateP3D65Compliance(analysis *HDRAnalysis, validation *HDRValidation) {
	if analysis.ColorPrimaries != "p3-d65" {
		validation.Issues = append(validation.Issues, "P3-D65 requires SMPTE 432-1 (P3-D65) color primaries")
		validation.IsCompliant = false
	}

	if analysis.BitDepth < 10 {
		validation.Recommendations = append(validation.Recommendations, 
			"P3-D65 content benefits from 10-bit or higher color depth")
	}
}

// validateGeneralHDRCompliance performs general HDR validation checks.
func (ha *HDRAnalyzerOptimized) validateGeneralHDRCompliance(analysis *HDRAnalysis, validation *HDRValidation) {
	// Check for common HDR authoring issues
	if analysis.ColorRange == "tv" {
		validation.Recommendations = append(validation.Recommendations, 
			"Consider using full range (0-255) for HDR content instead of limited range (16-235)")
	}

	// Validate content light level if present
	if analysis.ContentLightLevel != nil && analysis.ContentLightLevel.HasContentLightLevel {
		if analysis.ContentLightLevel.MaxCLL > 10000 {
			validation.Issues = append(validation.Issues, "MaxCLL exceeds 10,000 nits")
			validation.IsCompliant = false
		}
		
		if analysis.ContentLightLevel.MaxFALL > analysis.ContentLightLevel.MaxCLL {
			validation.Issues = append(validation.Issues, "MaxFALL cannot exceed MaxCLL")
			validation.IsCompliant = false
		}
	}
}