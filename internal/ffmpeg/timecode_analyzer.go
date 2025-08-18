package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// TimecodeAnalyzer handles SMPTE timecode analysis and drop frame detection
type TimecodeAnalyzer struct {
	ffprobePath string
	logger      zerolog.Logger
}

// NewTimecodeAnalyzer creates a new timecode analyzer
func NewTimecodeAnalyzer(ffprobePath string, logger zerolog.Logger) *TimecodeAnalyzer {
	return &TimecodeAnalyzer{
		ffprobePath: ffprobePath,
		logger:      logger,
	}
}

// TimecodeAnalysis contains comprehensive timecode analysis
type TimecodeAnalysis struct {
	HasTimecode          bool                    `json:"has_timecode"`
	TimecodeStreams      map[int]*TimecodeInfo   `json:"timecode_streams,omitempty"`
	PrimaryTimecode      *TimecodeInfo           `json:"primary_timecode,omitempty"`
	IsDropFrame          bool                    `json:"is_drop_frame"`
	FrameRateCompatible  bool                    `json:"frame_rate_compatible"`
	TimecodeValidation   *TimecodeValidation     `json:"timecode_validation,omitempty"`
	EmbeddedTimecodes    []EmbeddedTimecode      `json:"embedded_timecodes,omitempty"`
	UserDataTimecodes    []UserDataTimecode      `json:"user_data_timecodes,omitempty"`
}

// TimecodeInfo contains detailed timecode information
type TimecodeInfo struct {
	StreamIndex      int                `json:"stream_index"`
	StartTimecode    string             `json:"start_timecode,omitempty"`
	Format           string             `json:"format"`               // "SMPTE", "VITC", "LTC", "user_data"
	DropFrame        bool               `json:"drop_frame"`
	FrameRate        float64            `json:"frame_rate"`
	ColorFrame       bool               `json:"color_frame,omitempty"`
	FieldMark        bool               `json:"field_mark,omitempty"`
	BGF0             bool               `json:"bgf0,omitempty"`       // Binary Group Flag 0
	BGF1             bool               `json:"bgf1,omitempty"`       // Binary Group Flag 1
	BGF2             bool               `json:"bgf2,omitempty"`       // Binary Group Flag 2
	BinaryGroups     string             `json:"binary_groups,omitempty"`
	IsValid          bool               `json:"is_valid"`
	ValidationIssues []string           `json:"validation_issues,omitempty"`
}

// EmbeddedTimecode represents timecode found in video frames
type EmbeddedTimecode struct {
	FrameNumber    int     `json:"frame_number"`
	PTS            float64 `json:"pts"`
	Timecode       string  `json:"timecode"`
	Format         string  `json:"format"`
	Confidence     float64 `json:"confidence"`
}

// UserDataTimecode represents timecode in user data (e.g., H.264 SEI)
type UserDataTimecode struct {
	Type           string  `json:"type"`           // "sei_timecode", "gop_timecode", "aux_data"
	Timecode       string  `json:"timecode"`
	DropFrame      bool    `json:"drop_frame"`
	FrameNumber    int     `json:"frame_number"`
	Confidence     float64 `json:"confidence"`
}

// TimecodeValidation contains timecode validation results
type TimecodeValidation struct {
	IsValid              bool     `json:"is_valid"`
	IsContinuous         bool     `json:"is_continuous"`
	HasDiscontinuities   bool     `json:"has_discontinuities"`
	Issues               []string `json:"issues,omitempty"`
	Recommendations      []string `json:"recommendations,omitempty"`
	DropFrameCompliance  bool     `json:"drop_frame_compliance"`
	FrameRateConsistency bool     `json:"frame_rate_consistency"`
}

// AnalyzeTimecode performs comprehensive timecode analysis
func (ta *TimecodeAnalyzer) AnalyzeTimecode(ctx context.Context, filePath string, streams []StreamInfo) (*TimecodeAnalysis, error) {
	analysis := &TimecodeAnalysis{
		TimecodeStreams:   make(map[int]*TimecodeInfo),
		EmbeddedTimecodes: []EmbeddedTimecode{},
		UserDataTimecodes: []UserDataTimecode{},
	}

	// Step 1: Analyze dedicated timecode streams
	ta.analyzeTimecodeStreams(streams, analysis)

	// Step 2: Extract timecode from metadata and user data
	if err := ta.extractMetadataTimecodes(ctx, filePath, analysis); err != nil {
		ta.logger.Warn().Err(err).Msg("Failed to extract metadata timecodes")
	}

	// Step 3: Detect embedded timecodes in video frames (sample-based)
	if err := ta.detectEmbeddedTimecodes(ctx, filePath, analysis); err != nil {
		ta.logger.Warn().Err(err).Msg("Failed to detect embedded timecodes")
	}

	// Step 4: Analyze user data timecodes (H.264 SEI, GOP headers)
	if err := ta.analyzeUserDataTimecodes(ctx, filePath, analysis); err != nil {
		ta.logger.Warn().Err(err).Msg("Failed to analyze user data timecodes")
	}

	// Step 5: Determine primary timecode and drop frame status
	ta.determinePrimaryTimecode(analysis)

	// Step 6: Validate timecode consistency and compliance
	analysis.TimecodeValidation = ta.validateTimecode(analysis)

	return analysis, nil
}

// analyzeTimecodeStreams identifies dedicated timecode streams
func (ta *TimecodeAnalyzer) analyzeTimecodeStreams(streams []StreamInfo, analysis *TimecodeAnalysis) {
	for _, stream := range streams {
		if strings.ToLower(stream.CodecType) == "data" {
			// Check for timecode codecs
			codecName := strings.ToLower(stream.CodecName)
			if ta.isTimecodeCodec(codecName) {
				timecodeInfo := &TimecodeInfo{
					StreamIndex: stream.Index,
					Format:      ta.getTimecodeFormat(codecName),
					IsValid:     true,
				}

				// Extract frame rate
				if stream.RFrameRate != "" {
					if frameRate := ta.parseFrameRate(stream.RFrameRate); frameRate > 0 {
						timecodeInfo.FrameRate = frameRate
					}
				}

				// Determine drop frame based on frame rate
				timecodeInfo.DropFrame = ta.isDropFrameRate(timecodeInfo.FrameRate)

				// Extract start timecode from tags
				if startTC, exists := stream.Tags["timecode"]; exists {
					timecodeInfo.StartTimecode = startTC
				}

				analysis.TimecodeStreams[stream.Index] = timecodeInfo
				analysis.HasTimecode = true
			}
		}
	}
}

// extractMetadataTimecodes extracts timecode information from file metadata
func (ta *TimecodeAnalyzer) extractMetadataTimecodes(ctx context.Context, filePath string, analysis *TimecodeAnalysis) error {
	// Use ffprobe to extract detailed metadata including timecode information
	cmd := []string{
		ta.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_entries", "format_tags:stream_tags:frame_tags",
		"-select_streams", "v:0",
		"-read_intervals", "%+#1", // Read first frame only for efficiency
		filePath,
	}

	output, err := ta.executeCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to extract metadata: %w", err)
	}

	// Parse output for timecode information
	var result struct {
		Format struct {
			Tags map[string]string `json:"tags"`
		} `json:"format"`
		Streams []struct {
			Tags map[string]string `json:"tags"`
		} `json:"streams"`
		Frames []struct {
			Tags map[string]string `json:"tags"`
		} `json:"frames"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	// Check format-level timecode
	if timecode, exists := result.Format.Tags["timecode"]; exists {
		analysis.UserDataTimecodes = append(analysis.UserDataTimecodes, UserDataTimecode{
			Type:       "format_metadata",
			Timecode:   timecode,
			DropFrame:  ta.detectDropFrameFromTimecode(timecode),
			Confidence: 0.9,
		})
		analysis.HasTimecode = true
	}

	// Check stream-level timecode
	for _, stream := range result.Streams {
		if timecode, exists := stream.Tags["timecode"]; exists {
			analysis.UserDataTimecodes = append(analysis.UserDataTimecodes, UserDataTimecode{
				Type:       "stream_metadata",
				Timecode:   timecode,
				DropFrame:  ta.detectDropFrameFromTimecode(timecode),
				Confidence: 0.9,
			})
			analysis.HasTimecode = true
		}
	}

	return nil
}

// detectEmbeddedTimecodes looks for timecodes embedded in video frames
func (ta *TimecodeAnalyzer) detectEmbeddedTimecodes(ctx context.Context, filePath string, analysis *TimecodeAnalysis) error {
	// Use ffprobe to examine a sample of frames for embedded timecode
	cmd := []string{
		ta.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_entries", "frame=pkt_pts_time,pict_type,side_data",
		"-select_streams", "v:0",
		"-read_intervals", "%+#10", // Sample first 10 frames
		filePath,
	}

	output, err := ta.executeCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to analyze frames: %w", err)
	}

	var result struct {
		Frames []struct {
			PktPtsTime string `json:"pkt_pts_time"`
			PictType   string `json:"pict_type"`
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
		
		// Check side data for timecode information
		for _, sideData := range frame.SideData {
			if strings.Contains(sideData.Type, "timecode") || strings.Contains(sideData.Type, "h264_sei") {
				if timecodeStr := ta.extractTimecodeFromSideData(sideData.Data); timecodeStr != "" {
					pts, _ := strconv.ParseFloat(frame.PktPtsTime, 64)
					
					analysis.EmbeddedTimecodes = append(analysis.EmbeddedTimecodes, EmbeddedTimecode{
						FrameNumber: frameNumber,
						PTS:         pts,
						Timecode:    timecodeStr,
						Format:      "embedded_" + sideData.Type,
						Confidence:  0.8,
					})
					analysis.HasTimecode = true
				}
			}
		}
	}

	return nil
}

// analyzeUserDataTimecodes analyzes timecode in user data sections
func (ta *TimecodeAnalyzer) analyzeUserDataTimecodes(ctx context.Context, filePath string, analysis *TimecodeAnalysis) error {
	// Use ffprobe with specific filters to extract user data
	cmd := []string{
		ta.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_entries", "packet=data",
		"-select_streams", "v:0",
		"-read_intervals", "%+#5", // Sample first 5 packets
		filePath,
	}

	output, err := ta.executeCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to analyze user data: %w", err)
	}

	// Parse for common timecode patterns in user data
	timecodePatterns := []struct {
		pattern string
		format  string
	}{
		{`(\d{2}):(\d{2}):(\d{2})[;:](\d{2})`, "smpte_timecode"},
		{`TC:\s*(\d{2}):(\d{2}):(\d{2})[;:](\d{2})`, "prefixed_timecode"},
		{`timecode["\s]*[:=]\s*["\s]*(\d{2}):(\d{2}):(\d{2})[;:](\d{2})`, "json_timecode"},
	}

	for _, pattern := range timecodePatterns {
		if matches := regexp.MustCompile(pattern.pattern).FindAllStringSubmatch(output, -1); len(matches) > 0 {
			for i, match := range matches {
				if len(match) >= 5 {
					timecode := fmt.Sprintf("%s:%s:%s:%s", match[1], match[2], match[3], match[4])
					
					analysis.UserDataTimecodes = append(analysis.UserDataTimecodes, UserDataTimecode{
						Type:       pattern.format,
						Timecode:   timecode,
						DropFrame:  strings.Contains(match[0], ";"),
						FrameNumber: i,
						Confidence: 0.7,
					})
					analysis.HasTimecode = true
				}
			}
		}
	}

	return nil
}

// determinePrimaryTimecode identifies the most reliable timecode source
func (ta *TimecodeAnalyzer) determinePrimaryTimecode(analysis *TimecodeAnalysis) {
	// Priority order: Dedicated timecode streams > User data > Embedded > Metadata
	
	// Check dedicated timecode streams first
	for _, timecodeInfo := range analysis.TimecodeStreams {
		if timecodeInfo.IsValid {
			analysis.PrimaryTimecode = timecodeInfo
			analysis.IsDropFrame = timecodeInfo.DropFrame
			analysis.FrameRateCompatible = true
			return
		}
	}

	// Check user data timecodes
	for _, userTC := range analysis.UserDataTimecodes {
		if userTC.Confidence > 0.8 {
			analysis.PrimaryTimecode = &TimecodeInfo{
				StartTimecode: userTC.Timecode,
				Format:        userTC.Type,
				DropFrame:     userTC.DropFrame,
				IsValid:       true,
			}
			analysis.IsDropFrame = userTC.DropFrame
			analysis.FrameRateCompatible = true
			return
		}
	}

	// Check embedded timecodes
	if len(analysis.EmbeddedTimecodes) > 0 {
		embedded := analysis.EmbeddedTimecodes[0] // Use first detected
		analysis.PrimaryTimecode = &TimecodeInfo{
			StartTimecode: embedded.Timecode,
			Format:        embedded.Format,
			DropFrame:     ta.detectDropFrameFromTimecode(embedded.Timecode),
			IsValid:       embedded.Confidence > 0.7,
		}
		analysis.IsDropFrame = analysis.PrimaryTimecode.DropFrame
		analysis.FrameRateCompatible = true
	}
}

// validateTimecode performs comprehensive timecode validation
func (ta *TimecodeAnalyzer) validateTimecode(analysis *TimecodeAnalysis) *TimecodeValidation {
	validation := &TimecodeValidation{
		IsValid:              analysis.HasTimecode,
		IsContinuous:         true,
		HasDiscontinuities:   false,
		Issues:               []string{},
		Recommendations:      []string{},
		DropFrameCompliance:  true,
		FrameRateConsistency: true,
	}

	if !analysis.HasTimecode {
		validation.Issues = append(validation.Issues, "No timecode found in media file")
		validation.Recommendations = append(validation.Recommendations, "Consider adding SMPTE timecode for professional workflows")
		return validation
	}

	// Validate primary timecode format
	if analysis.PrimaryTimecode != nil {
		if !ta.isValidTimecodeFormat(analysis.PrimaryTimecode.StartTimecode) {
			validation.Issues = append(validation.Issues, "Invalid timecode format detected")
			validation.IsValid = false
		}

		// Validate drop frame compliance
		if analysis.PrimaryTimecode.DropFrame && !ta.isDropFrameRate(analysis.PrimaryTimecode.FrameRate) {
			validation.Issues = append(validation.Issues, "Drop frame timecode used with incompatible frame rate")
			validation.DropFrameCompliance = false
		}
	}

	// Check for multiple conflicting timecodes
	if len(analysis.TimecodeStreams) > 1 || len(analysis.UserDataTimecodes) > 1 {
		validation.Recommendations = append(validation.Recommendations, "Multiple timecode sources detected - verify consistency")
	}

	// Validate drop frame usage
	if analysis.IsDropFrame {
		validation.Recommendations = append(validation.Recommendations, "Drop frame timecode detected - ensure broadcast compatibility")
	}

	return validation
}

// Helper methods

func (ta *TimecodeAnalyzer) isTimecodeCodec(codecName string) bool {
	timecodeCodecs := []string{
		"timecode", "smpte_timecode", "vitc", "ltc",
		"m2v_timecode", "dvd_nav_packet",
	}
	
	for _, codec := range timecodeCodecs {
		if strings.Contains(codecName, codec) {
			return true
		}
	}
	return false
}

func (ta *TimecodeAnalyzer) getTimecodeFormat(codecName string) string {
	formatMap := map[string]string{
		"timecode":       "SMPTE",
		"smpte_timecode": "SMPTE",
		"vitc":           "VITC",
		"ltc":            "LTC",
	}
	
	for key, format := range formatMap {
		if strings.Contains(codecName, key) {
			return format
		}
	}
	return "Unknown"
}

func (ta *TimecodeAnalyzer) parseFrameRate(frameRateStr string) float64 {
	// Parse frame rate strings like "30000/1001" or "29.97"
	if strings.Contains(frameRateStr, "/") {
		parts := strings.Split(frameRateStr, "/")
		if len(parts) == 2 {
			num, _ := strconv.ParseFloat(parts[0], 64)
			den, _ := strconv.ParseFloat(parts[1], 64)
			if den != 0 {
				return num / den
			}
		}
	}
	
	frameRate, _ := strconv.ParseFloat(frameRateStr, 64)
	return frameRate
}

func (ta *TimecodeAnalyzer) isDropFrameRate(frameRate float64) bool {
	// Drop frame is typically used with 29.97 fps and 59.94 fps
	dropFrameRates := []float64{29.97, 59.94}
	tolerance := 0.01
	
	for _, dfRate := range dropFrameRates {
		if frameRate >= dfRate-tolerance && frameRate <= dfRate+tolerance {
			return true
		}
	}
	return false
}

func (ta *TimecodeAnalyzer) detectDropFrameFromTimecode(timecode string) bool {
	// Drop frame timecodes use semicolon separator for frames
	return strings.Contains(timecode, ";")
}

func (ta *TimecodeAnalyzer) isValidTimecodeFormat(timecode string) bool {
	// Validate SMPTE timecode format: HH:MM:SS:FF or HH:MM:SS;FF
	pattern := `^\d{2}:\d{2}:\d{2}[;:]\d{2}$`
	matched, _ := regexp.MatchString(pattern, timecode)
	return matched
}

func (ta *TimecodeAnalyzer) extractTimecodeFromSideData(data map[string]interface{}) string {
	// Extract timecode from various side data structures
	if timecodeStr, ok := data["timecode"].(string); ok {
		return timecodeStr
	}
	
	// Check for numeric timecode components
	if hours, ok := data["hours"].(float64); ok {
		if minutes, ok := data["minutes"].(float64); ok {
			if seconds, ok := data["seconds"].(float64); ok {
				if frames, ok := data["frames"].(float64); ok {
					return fmt.Sprintf("%02d:%02d:%02d:%02d", 
						int(hours), int(minutes), int(seconds), int(frames))
				}
			}
		}
	}
	
	return ""
}

func (ta *TimecodeAnalyzer) executeCommand(ctx context.Context, cmd []string) (string, error) {
	// Execute command with timeout
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	output, err := executeFFprobeCommand(execCtx, cmd)
	if err != nil {
		return "", err
	}
	
	return string(output), nil
}