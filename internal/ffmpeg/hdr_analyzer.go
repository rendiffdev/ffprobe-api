package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

// HDRAnalyzer handles HDR metadata detection and analysis
type HDRAnalyzer struct {
	ffprobePath string
	logger      zerolog.Logger
}

// NewHDRAnalyzer creates a new HDR analyzer
func NewHDRAnalyzer(ffprobePath string, logger zerolog.Logger) *HDRAnalyzer {
	if ffprobePath == "" {
		ffprobePath = "ffprobe"
	}

	return &HDRAnalyzer{
		ffprobePath: ffprobePath,
		logger:      logger,
	}
}

// AnalyzeHDR performs comprehensive HDR metadata analysis
func (ha *HDRAnalyzer) AnalyzeHDR(ctx context.Context, filePath string) (*HDRAnalysis, error) {
	analysis := &HDRAnalysis{}

	// Get stream metadata for HDR indicators
	streamMetadata, err := ha.getStreamMetadata(ctx, filePath)
	if err != nil {
		return analysis, fmt.Errorf("failed to get stream metadata: %w", err)
	}

	// Analyze video streams for HDR characteristics
	for _, stream := range streamMetadata {
		if stream.CodecType != "video" {
			continue
		}

		// Check basic HDR indicators
		analysis.ColorPrimaries = stream.ColorPrimaries
		analysis.ColorTransfer = stream.ColorTransfer
		analysis.ColorSpace = stream.ColorSpace

		// Determine if content is HDR based on color characteristics
		analysis.IsHDR = ha.isHDRContent(stream.ColorPrimaries, stream.ColorTransfer, stream.ColorSpace)

		if analysis.IsHDR {
			// Determine HDR format
			analysis.HDRFormat = ha.determineHDRFormat(stream)
			
			// Check for HLG compatibility
			analysis.HLGCompatible = ha.isHLGCompatible(stream.ColorTransfer)
		}

		// Only analyze the first video stream
		break
	}

	// Get side data for advanced HDR metadata
	if analysis.IsHDR {
		sideData, err := ha.getSideDataMetadata(ctx, filePath)
		if err != nil {
			ha.logger.Warn().Err(err).Msg("Failed to get side data, continuing without advanced metadata")
		} else {
			ha.parseSideDataMetadata(sideData, analysis)
		}
	}

	// Validate HDR compliance
	analysis.Validation = ha.validateHDRCompliance(analysis)

	return analysis, nil
}

// streamMetadataResult represents FFprobe stream output
type streamMetadataResult struct {
	Streams []streamMetadata `json:"streams"`
}

type streamMetadata struct {
	Index          int    `json:"index"`
	CodecType      string `json:"codec_type"`
	ColorPrimaries string `json:"color_primaries,omitempty"`
	ColorTransfer  string `json:"color_transfer,omitempty"`
	ColorSpace     string `json:"color_space,omitempty"`
	PixFmt         string `json:"pix_fmt,omitempty"`
	Profile        string `json:"profile,omitempty"`
	Level          int    `json:"level,omitempty"`
}

// getStreamMetadata retrieves basic stream metadata
func (ha *HDRAnalyzer) getStreamMetadata(ctx context.Context, filePath string) ([]streamMetadata, error) {
	cmd := exec.CommandContext(ctx, ha.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-select_streams", "v:0",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe command failed: %w", err)
	}

	var result streamMetadataResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	return result.Streams, nil
}

// getSideDataMetadata retrieves side data metadata for advanced HDR info
func (ha *HDRAnalyzer) getSideDataMetadata(ctx context.Context, filePath string) (string, error) {
	cmd := exec.CommandContext(ctx, ha.ffprobePath,
		"-v", "quiet",
		"-print_format", "default",
		"-show_frames",
		"-select_streams", "v:0",
		"-read_intervals", "%+#1",
		"-show_entries", "frame=side_data_list",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ffprobe side data command failed: %w", err)
	}

	return string(output), nil
}

// isHDRContent determines if content is HDR based on color characteristics
func (ha *HDRAnalyzer) isHDRContent(colorPrimaries, colorTransfer, colorSpace string) bool {
	// HDR indicators:
	// Color primaries: bt2020
	// Color transfer: smpte2084 (HDR10), arib-std-b67 (HLG), smpte2094-40 (HDR10+)
	// Color space: bt2020nc, bt2020c

	hdrPrimaries := map[string]bool{
		"bt2020": true,
	}

	hdrTransfers := map[string]bool{
		"smpte2084":     true, // HDR10/HDR10+
		"arib-std-b67":  true, // HLG
		"smpte2094-40":  true, // HDR10+
		"smpte2094-10":  true, // HDR10+
	}

	hdrColorSpaces := map[string]bool{
		"bt2020nc": true,
		"bt2020c":  true,
	}

	return hdrPrimaries[colorPrimaries] && 
		   hdrTransfers[colorTransfer] && 
		   hdrColorSpaces[colorSpace]
}

// determineHDRFormat determines the specific HDR format
func (ha *HDRAnalyzer) determineHDRFormat(stream streamMetadata) string {
	switch stream.ColorTransfer {
	case "smpte2084":
		return "HDR10" // Could also be HDR10+ but need side data to confirm
	case "arib-std-b67":
		return "HLG"
	case "smpte2094-40", "smpte2094-10":
		return "HDR10+"
	default:
		if stream.ColorPrimaries == "bt2020" {
			return "HDR10" // Best guess
		}
		return "Unknown"
	}
}

// isHLGCompatible checks if content is HLG compatible
func (ha *HDRAnalyzer) isHLGCompatible(colorTransfer string) bool {
	return colorTransfer == "arib-std-b67"
}

// parseSideDataMetadata parses side data for advanced HDR metadata
func (ha *HDRAnalyzer) parseSideDataMetadata(sideData string, analysis *HDRAnalysis) {
	lines := strings.Split(sideData, "\n")
	
	var currentSideData string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Detect side data types
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
		
		if strings.Contains(line, "side_data_type=HDR10+") {
			currentSideData = "hdr10plus"
			analysis.HDR10Plus = &HDR10PlusMetadata{Present: true}
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

		// Reset flags when leaving side data block
		if line == "" || strings.HasPrefix(line, "[") {
			currentSideData = ""
			continue
		}

		// Parse specific metadata based on current side data type
		switch currentSideData {
		case "mastering":
			ha.parseMasteringDisplayData(line, analysis.MasteringDisplay)
		case "content_light":
			ha.parseContentLightData(line, analysis.ContentLightLevel)
		case "hdr10plus":
			ha.parseHDR10PlusData(line, analysis.HDR10Plus)
		case "dolby_vision":
			ha.parseDolbyVisionData(line, analysis.DolbyVision)
		}
	}
}

// parseMasteringDisplayData parses mastering display metadata
func (ha *HDRAnalyzer) parseMasteringDisplayData(line string, metadata *MasteringDisplayMetadata) {
	// Parse display primaries: display_primaries=G(13250,34500)B(7500,3000)R(34000,16000)
	if strings.Contains(line, "display_primaries=") {
		primariesRegex := regexp.MustCompile(`G\((\d+),(\d+)\)B\((\d+),(\d+)\)R\((\d+),(\d+)\)`)
		matches := primariesRegex.FindStringSubmatch(line)
		if len(matches) == 7 {
			// Convert from fixed-point representation (divide by 50000)
			if gx, err := strconv.ParseFloat(matches[1], 64); err == nil {
				metadata.DisplayPrimariesX[1] = gx / 50000.0 // Green X
			}
			if gy, err := strconv.ParseFloat(matches[2], 64); err == nil {
				metadata.DisplayPrimariesY[1] = gy / 50000.0 // Green Y
			}
			if bx, err := strconv.ParseFloat(matches[3], 64); err == nil {
				metadata.DisplayPrimariesX[2] = bx / 50000.0 // Blue X
			}
			if by, err := strconv.ParseFloat(matches[4], 64); err == nil {
				metadata.DisplayPrimariesY[2] = by / 50000.0 // Blue Y
			}
			if rx, err := strconv.ParseFloat(matches[5], 64); err == nil {
				metadata.DisplayPrimariesX[0] = rx / 50000.0 // Red X
			}
			if ry, err := strconv.ParseFloat(matches[6], 64); err == nil {
				metadata.DisplayPrimariesY[0] = ry / 50000.0 // Red Y
			}
		}
	}

	// Parse white point: white_point=15635/50000,16450/50000
	if strings.Contains(line, "white_point=") {
		whitePointRegex := regexp.MustCompile(`white_point=(\d+)/(\d+),(\d+)/(\d+)`)
		matches := whitePointRegex.FindStringSubmatch(line)
		if len(matches) == 5 {
			if wxNum, err1 := strconv.ParseFloat(matches[1], 64); err1 == nil {
				if wxDen, err2 := strconv.ParseFloat(matches[2], 64); err2 == nil {
					metadata.WhitePointX = wxNum / wxDen
				}
			}
			if wyNum, err1 := strconv.ParseFloat(matches[3], 64); err1 == nil {
				if wyDen, err2 := strconv.ParseFloat(matches[4], 64); err2 == nil {
					metadata.WhitePointY = wyNum / wyDen
				}
			}
		}
	}

	// Parse luminance: max_luminance=1000.000000, min_luminance=0.005000
	if strings.Contains(line, "max_luminance=") {
		maxLumRegex := regexp.MustCompile(`max_luminance=([0-9.]+)`)
		matches := maxLumRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
				metadata.MaxDisplayLuminance = val
			}
		}
	}

	if strings.Contains(line, "min_luminance=") {
		minLumRegex := regexp.MustCompile(`min_luminance=([0-9.]+)`)
		matches := minLumRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
				metadata.MinDisplayLuminance = val
			}
		}
	}
}

// parseContentLightData parses content light level metadata
func (ha *HDRAnalyzer) parseContentLightData(line string, metadata *ContentLightLevelData) {
	// Parse max_content=1000, max_average=400
	if strings.Contains(line, "max_content=") {
		maxContentRegex := regexp.MustCompile(`max_content=(\d+)`)
		matches := maxContentRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				metadata.MaxCLL = val
			}
		}
	}

	if strings.Contains(line, "max_average=") {
		maxAvgRegex := regexp.MustCompile(`max_average=(\d+)`)
		matches := maxAvgRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				metadata.MaxFALL = val
			}
		}
	}
}

// parseHDR10PlusData parses HDR10+ metadata
func (ha *HDRAnalyzer) parseHDR10PlusData(line string, metadata *HDR10PlusMetadata) {
	if strings.Contains(line, "application_version=") {
		versionRegex := regexp.MustCompile(`application_version=(\d+)`)
		matches := versionRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				metadata.ApplicationVersion = val
			}
		}
	}

	if strings.Contains(line, "num_windows=") {
		windowsRegex := regexp.MustCompile(`num_windows=(\d+)`)
		matches := windowsRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				metadata.NumWindows = val
			}
		}
	}
}

// parseDolbyVisionData parses Dolby Vision metadata
func (ha *HDRAnalyzer) parseDolbyVisionData(line string, metadata *DolbyVisionMetadata) {
	if strings.Contains(line, "dv_profile=") {
		profileRegex := regexp.MustCompile(`dv_profile=(\d+)`)
		matches := profileRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				metadata.Profile = val
			}
		}
	}

	if strings.Contains(line, "dv_level=") {
		levelRegex := regexp.MustCompile(`dv_level=(\d+)`)
		matches := levelRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				metadata.Level = val
			}
		}
	}

	if strings.Contains(line, "rpu_present=") {
		metadata.RPUPresent = strings.Contains(line, "rpu_present=1")
	}

	if strings.Contains(line, "el_present=") {
		metadata.ELPresent = strings.Contains(line, "el_present=1")
	}

	if strings.Contains(line, "bl_present=") {
		metadata.BLPresent = strings.Contains(line, "bl_present=1")
	}
}

// validateHDRCompliance validates HDR compliance and provides recommendations
func (ha *HDRAnalyzer) validateHDRCompliance(analysis *HDRAnalysis) *HDRValidation {
	validation := &HDRValidation{
		Standard: analysis.HDRFormat,
		Issues:   []string{},
		Recommendations: []string{},
	}

	if !analysis.IsHDR {
		validation.IsCompliant = true
		return validation
	}

	// Validate based on HDR format
	switch analysis.HDRFormat {
	case "HDR10":
		validation.IsCompliant = ha.validateHDR10Compliance(analysis, validation)
	case "HDR10+":
		validation.IsCompliant = ha.validateHDR10PlusCompliance(analysis, validation)
	case "Dolby Vision":
		validation.IsCompliant = ha.validateDolbyVisionCompliance(analysis, validation)
	case "HLG":
		validation.IsCompliant = ha.validateHLGCompliance(analysis, validation)
	default:
		validation.Issues = append(validation.Issues, "Unknown HDR format")
		validation.IsCompliant = false
	}

	return validation
}

// validateHDR10Compliance validates HDR10 compliance
func (ha *HDRAnalyzer) validateHDR10Compliance(analysis *HDRAnalysis, validation *HDRValidation) bool {
	compliant := true

	// Check required color characteristics
	if analysis.ColorPrimaries != "bt2020" {
		validation.Issues = append(validation.Issues, "HDR10 requires BT.2020 color primaries")
		compliant = false
	}

	if analysis.ColorTransfer != "smpte2084" {
		validation.Issues = append(validation.Issues, "HDR10 requires SMPTE 2084 (PQ) transfer function")
		compliant = false
	}

	// Check mastering display metadata
	if analysis.MasteringDisplay == nil || !analysis.MasteringDisplay.HasMasteringDisplay {
		validation.Issues = append(validation.Issues, "HDR10 should include mastering display metadata")
		validation.Recommendations = append(validation.Recommendations, "Add mastering display metadata for better HDR10 compliance")
		compliant = false
	}

	// Check content light level
	if analysis.ContentLightLevel == nil || !analysis.ContentLightLevel.HasContentLightLevel {
		validation.Recommendations = append(validation.Recommendations, "Consider adding content light level metadata for optimal HDR10 playback")
	}

	return compliant
}

// validateHDR10PlusCompliance validates HDR10+ compliance
func (ha *HDRAnalyzer) validateHDR10PlusCompliance(analysis *HDRAnalysis, validation *HDRValidation) bool {
	// First validate base HDR10 compliance
	compliant := ha.validateHDR10Compliance(analysis, validation)

	// Check HDR10+ specific metadata
	if analysis.HDR10Plus == nil || !analysis.HDR10Plus.Present {
		validation.Issues = append(validation.Issues, "HDR10+ requires dynamic metadata")
		compliant = false
	}

	return compliant
}

// validateDolbyVisionCompliance validates Dolby Vision compliance
func (ha *HDRAnalyzer) validateDolbyVisionCompliance(analysis *HDRAnalysis, validation *HDRValidation) bool {
	compliant := true

	if analysis.DolbyVision == nil {
		validation.Issues = append(validation.Issues, "Dolby Vision metadata missing")
		return false
	}

	// Check for valid profile
	validProfiles := map[int]bool{4: true, 5: true, 7: true, 8: true, 9: true}
	if !validProfiles[analysis.DolbyVision.Profile] {
		validation.Issues = append(validation.Issues, fmt.Sprintf("Invalid Dolby Vision profile: %d", analysis.DolbyVision.Profile))
		compliant = false
	}

	// Check RPU presence for most profiles
	if !analysis.DolbyVision.RPUPresent && analysis.DolbyVision.Profile != 5 {
		validation.Issues = append(validation.Issues, "Dolby Vision RPU (Reference Processing Unit) missing")
		compliant = false
	}

	return compliant
}

// validateHLGCompliance validates HLG compliance
func (ha *HDRAnalyzer) validateHLGCompliance(analysis *HDRAnalysis, validation *HDRValidation) bool {
	compliant := true

	if analysis.ColorTransfer != "arib-std-b67" {
		validation.Issues = append(validation.Issues, "HLG requires ARIB STD-B67 transfer function")
		compliant = false
	}

	if analysis.ColorPrimaries != "bt2020" {
		validation.Issues = append(validation.Issues, "HLG requires BT.2020 color primaries")
		compliant = false
	}

	return compliant
}