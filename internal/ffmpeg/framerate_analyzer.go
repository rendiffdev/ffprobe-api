package ffmpeg

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// FrameRateAnalyzer handles frame rate analysis and validation
type FrameRateAnalyzer struct{}

// NewFrameRateAnalyzer creates a new frame rate analyzer
func NewFrameRateAnalyzer() *FrameRateAnalyzer {
	return &FrameRateAnalyzer{}
}

// AnalyzeFrameRate analyzes frame rate from stream information
func (fra *FrameRateAnalyzer) AnalyzeFrameRate(streams []StreamInfo) *FrameRateAnalysis {
	analysis := &FrameRateAnalysis{
		VideoStreams: make(map[int]*VideoFrameRate),
	}

	for _, stream := range streams {
		if strings.ToLower(stream.CodecType) == "video" {
			videoFrameRate := fra.analyzeVideoFrameRate(stream)
			if videoFrameRate != nil {
				analysis.VideoStreams[stream.Index] = videoFrameRate

				// Update overall analysis
				if videoFrameRate.EffectiveFrameRate > analysis.MaxFrameRate {
					analysis.MaxFrameRate = videoFrameRate.EffectiveFrameRate
				}
				if analysis.MinFrameRate == 0 || videoFrameRate.EffectiveFrameRate < analysis.MinFrameRate {
					analysis.MinFrameRate = videoFrameRate.EffectiveFrameRate
				}
				if analysis.PrimaryFrameRateStandard == "" {
					analysis.PrimaryFrameRateStandard = videoFrameRate.Standard
				}
			}
		}
	}

	// Determine overall characteristics
	analysis.IsVariableFrameRate = fra.hasVariableFrameRate(analysis.VideoStreams)
	analysis.IsHighFrameRate = analysis.MaxFrameRate >= 60.0
	analysis.HasMultipleFrameRates = fra.hasMultipleFrameRates(analysis.VideoStreams)
	analysis.IsInterlaced = fra.hasInterlacedContent(analysis.VideoStreams)
	analysis.Validation = fra.validateFrameRate(analysis)

	return analysis
}

// analyzeVideoFrameRate extracts frame rate information from a video stream
func (fra *FrameRateAnalyzer) analyzeVideoFrameRate(stream StreamInfo) *VideoFrameRate {
	frameRate := &VideoFrameRate{}

	// Parse real frame rate (r_frame_rate)
	frameRate.RealFrameRate = fra.parseFrameRate(stream.RFrameRate)

	// Parse average frame rate (avg_frame_rate)
	frameRate.AverageFrameRate = fra.parseFrameRate(stream.AvgFrameRate)

	// Determine effective frame rate (prefer average, fallback to real)
	if frameRate.AverageFrameRate > 0 {
		frameRate.EffectiveFrameRate = frameRate.AverageFrameRate
		frameRate.Source = "average_frame_rate"
	} else if frameRate.RealFrameRate > 0 {
		frameRate.EffectiveFrameRate = frameRate.RealFrameRate
		frameRate.Source = "real_frame_rate"
	}

	// Determine if variable frame rate
	frameRate.IsVariableFrameRate = fra.isVariableFrameRate(frameRate.RealFrameRate, frameRate.AverageFrameRate)

	// Categorize frame rate
	frameRate.Standard = fra.categorizeFrameRate(frameRate.EffectiveFrameRate)
	frameRate.Category = fra.getFrameRateCategory(frameRate.EffectiveFrameRate)

	// Check for interlaced content indicators
	frameRate.IsInterlaced = fra.detectInterlacing(stream)

	// Validate consistency
	frameRate.IsConsistent = fra.validateFrameRateConsistency(frameRate, stream)

	// Calculate temporal characteristics
	if frameRate.EffectiveFrameRate > 0 {
		frameRate.FrameDuration = 1000.0 / frameRate.EffectiveFrameRate // milliseconds
	}

	return frameRate
}

// parseFrameRate parses frame rate string to float64
func (fra *FrameRateAnalyzer) parseFrameRate(frameRateStr string) float64 {
	if frameRateStr == "" || frameRateStr == "N/A" || frameRateStr == "0/0" {
		return 0.0
	}

	// Handle fraction format like "30000/1001" or "25/1"
	if strings.Contains(frameRateStr, "/") {
		parts := strings.Split(frameRateStr, "/")
		if len(parts) == 2 {
			num, err1 := strconv.ParseFloat(parts[0], 64)
			den, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil && den != 0 {
				return num / den
			}
		}
	}

	// Handle direct decimal values
	if val, err := strconv.ParseFloat(frameRateStr, 64); err == nil {
		return val
	}

	return 0.0
}

// categorizeFrameRate determines the standard frame rate category
func (fra *FrameRateAnalyzer) categorizeFrameRate(frameRate float64) string {
	if frameRate == 0.0 {
		return "Unknown"
	}

	// Standard frame rates with tolerance
	tolerance := 0.1
	standards := map[string]float64{
		"23.976p": 23.976,
		"24p":     24.0,
		"25p":     25.0,
		"29.97p":  29.97,
		"30p":     30.0,
		"48p":     48.0,
		"50p":     50.0,
		"59.94p":  59.94,
		"60p":     60.0,
		"96p":     96.0,
		"100p":    100.0,
		"120p":    120.0,
		"240p":    240.0,
		"480p":    480.0,
		"1000p":   1000.0,
	}

	for standard, rate := range standards {
		if math.Abs(frameRate-rate) <= tolerance {
			return standard
		}
	}

	// For interlaced content, check common interlaced rates
	interlacedStandards := map[string]float64{
		"25i":    25.0,
		"29.97i": 29.97,
		"30i":    30.0,
		"50i":    50.0,
		"59.94i": 59.94,
		"60i":    60.0,
	}

	for standard, rate := range interlacedStandards {
		if math.Abs(frameRate-rate) <= tolerance {
			return standard
		}
	}

	// Custom frame rate
	return fmt.Sprintf("%.3fp", frameRate)
}

// getFrameRateCategory determines the frame rate category
func (fra *FrameRateAnalyzer) getFrameRateCategory(frameRate float64) string {
	switch {
	case frameRate == 0.0:
		return "Unknown"
	case frameRate < 20.0:
		return "Very Low Frame Rate"
	case frameRate < 30.0:
		return "Cinema Frame Rate"
	case frameRate < 50.0:
		return "Standard Frame Rate"
	case frameRate < 100.0:
		return "High Frame Rate"
	case frameRate < 250.0:
		return "Very High Frame Rate"
	default:
		return "Ultra High Frame Rate"
	}
}

// isVariableFrameRate determines if content has variable frame rate
func (fra *FrameRateAnalyzer) isVariableFrameRate(realFR, avgFR float64) bool {
	if realFR == 0.0 || avgFR == 0.0 {
		return false
	}

	// If real and average frame rates differ significantly, it's likely VFR
	tolerance := 0.1
	return math.Abs(realFR-avgFR) > tolerance
}

// detectInterlacing detects interlaced content from stream metadata
func (fra *FrameRateAnalyzer) detectInterlacing(stream StreamInfo) bool {
	// Check field order
	if stream.FieldOrder != "" && stream.FieldOrder != "progressive" {
		return true
	}

	// Check for interlaced frame indicators
	// Note: This is basic detection, actual interlacing detection would require frame analysis
	return false
}

// validateFrameRateConsistency validates frame rate metadata consistency
func (fra *FrameRateAnalyzer) validateFrameRateConsistency(frameRate *VideoFrameRate, stream StreamInfo) bool {
	issues := 0

	// Check if real and average frame rates are reasonable
	if frameRate.RealFrameRate > 0 && frameRate.AverageFrameRate > 0 {
		ratio := frameRate.RealFrameRate / frameRate.AverageFrameRate
		// If the ratio is very different from 1, there might be an issue
		if ratio > 2.0 || ratio < 0.5 {
			issues++
		}
	}

	// Check for unreasonable frame rates
	if frameRate.EffectiveFrameRate > 1000.0 || frameRate.EffectiveFrameRate < 0.1 {
		issues++
	}

	return issues == 0
}

// hasVariableFrameRate checks if any stream has variable frame rate
func (fra *FrameRateAnalyzer) hasVariableFrameRate(videoStreams map[int]*VideoFrameRate) bool {
	for _, frameRate := range videoStreams {
		if frameRate.IsVariableFrameRate {
			return true
		}
	}
	return false
}

// hasMultipleFrameRates checks if there are multiple different frame rates
func (fra *FrameRateAnalyzer) hasMultipleFrameRates(videoStreams map[int]*VideoFrameRate) bool {
	if len(videoStreams) <= 1 {
		return false
	}

	var firstFrameRate float64
	first := true

	for _, frameRate := range videoStreams {
		if first {
			firstFrameRate = frameRate.EffectiveFrameRate
			first = false
		} else {
			tolerance := 0.1
			if math.Abs(frameRate.EffectiveFrameRate-firstFrameRate) > tolerance {
				return true
			}
		}
	}

	return false
}

// hasInterlacedContent checks if any stream is interlaced
func (fra *FrameRateAnalyzer) hasInterlacedContent(videoStreams map[int]*VideoFrameRate) bool {
	for _, frameRate := range videoStreams {
		if frameRate.IsInterlaced {
			return true
		}
	}
	return false
}

// validateFrameRate validates overall frame rate characteristics
func (fra *FrameRateAnalyzer) validateFrameRate(analysis *FrameRateAnalysis) *FrameRateValidation {
	validation := &FrameRateValidation{
		IsValid:         true,
		Issues:          []string{},
		Recommendations: []string{},
	}

	// Check for inconsistencies in video streams
	for streamIndex, frameRate := range analysis.VideoStreams {
		if !frameRate.IsConsistent {
			validation.Issues = append(validation.Issues,
				fmt.Sprintf("Video stream %d has inconsistent frame rate metadata", streamIndex))
			validation.IsValid = false
		}

		// Check for unusual frame rates
		if frameRate.EffectiveFrameRate > 0 {
			if frameRate.EffectiveFrameRate < 5.0 {
				validation.Issues = append(validation.Issues,
					fmt.Sprintf("Video stream %d has very low frame rate: %.3f fps", streamIndex, frameRate.EffectiveFrameRate))
			} else if frameRate.EffectiveFrameRate > 500.0 {
				validation.Issues = append(validation.Issues,
					fmt.Sprintf("Video stream %d has extremely high frame rate: %.3f fps", streamIndex, frameRate.EffectiveFrameRate))
			}
		}

		// Check for variable frame rate issues
		if frameRate.IsVariableFrameRate {
			validation.Recommendations = append(validation.Recommendations,
				fmt.Sprintf("Video stream %d uses variable frame rate - consider converting to constant frame rate for better compatibility", streamIndex))
		}
	}

	// Provide recommendations based on frame rate characteristics
	if analysis.IsHighFrameRate {
		validation.Recommendations = append(validation.Recommendations,
			"High frame rate content detected - ensure delivery infrastructure supports HFR playback")
	}

	if analysis.HasMultipleFrameRates {
		validation.Recommendations = append(validation.Recommendations,
			"Multiple frame rates detected - verify this is intentional for adaptive streaming")
	}

	if analysis.IsInterlaced {
		validation.Recommendations = append(validation.Recommendations,
			"Interlaced content detected - consider deinterlacing for modern viewing devices")
	}

	// Check for common problematic frame rates
	for streamIndex, frameRate := range analysis.VideoStreams {
		if frameRate.EffectiveFrameRate > 0 {
			// Frame rates that might cause issues
			if math.Abs(frameRate.EffectiveFrameRate-23.976) < 0.1 {
				validation.Recommendations = append(validation.Recommendations,
					fmt.Sprintf("Video stream %d uses 23.976 fps - ensure proper pulldown handling for broadcast", streamIndex))
			}
			if math.Abs(frameRate.EffectiveFrameRate-29.97) < 0.1 {
				validation.Recommendations = append(validation.Recommendations,
					fmt.Sprintf("Video stream %d uses 29.97 fps - verify NTSC compatibility", streamIndex))
			}
		}
	}

	return validation
}
