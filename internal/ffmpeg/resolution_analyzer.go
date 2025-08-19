package ffmpeg

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ResolutionAnalyzer handles resolution and aspect ratio analysis
type ResolutionAnalyzer struct{}

// NewResolutionAnalyzer creates a new resolution analyzer
func NewResolutionAnalyzer() *ResolutionAnalyzer {
	return &ResolutionAnalyzer{}
}

// AnalyzeResolution analyzes resolution and aspect ratio from stream information
func (ra *ResolutionAnalyzer) AnalyzeResolution(streams []StreamInfo) *ResolutionAnalysis {
	analysis := &ResolutionAnalysis{
		VideoStreams: make(map[int]*VideoResolution),
	}

	for _, stream := range streams {
		if strings.ToLower(stream.CodecType) == "video" {
			videoRes := ra.analyzeVideoResolution(stream)
			if videoRes != nil {
				analysis.VideoStreams[stream.Index] = videoRes

				// Update overall analysis
				if videoRes.Width > analysis.MaxWidth {
					analysis.MaxWidth = videoRes.Width
				}
				if videoRes.Height > analysis.MaxHeight {
					analysis.MaxHeight = videoRes.Height
				}
				if analysis.PrimaryResolution == "" {
					analysis.PrimaryResolution = videoRes.StandardResolution
				}
			}
		}
	}

	// Determine overall characteristics
	analysis.IsHighDefinition = analysis.MaxHeight >= 720
	analysis.IsUltraHighDefinition = analysis.MaxHeight >= 2160
	analysis.IsWidescreen = ra.isWidescreenContent(analysis.VideoStreams)
	analysis.HasMultipleResolutions = ra.hasMultipleResolutions(analysis.VideoStreams)
	analysis.Validation = ra.validateResolution(analysis)

	return analysis
}

// analyzeVideoResolution extracts resolution information from a video stream
func (ra *ResolutionAnalyzer) analyzeVideoResolution(stream StreamInfo) *VideoResolution {
	resolution := &VideoResolution{
		Width:  stream.Width,
		Height: stream.Height,
	}

	// Calculate pixel count
	if resolution.Width > 0 && resolution.Height > 0 {
		resolution.PixelCount = resolution.Width * resolution.Height
	}

	// Determine standard resolution category
	resolution.StandardResolution = ra.getStandardResolution(resolution.Width, resolution.Height)
	resolution.ResolutionClass = ra.getResolutionClass(resolution.Height)

	// Analyze aspect ratios
	resolution.SampleAspectRatio = ra.parseAspectRatio(stream.SampleAspectRatio)
	resolution.DisplayAspectRatio = ra.parseAspectRatio(stream.DisplayAspectRatio)

	// Calculate pixel aspect ratio
	if resolution.Width > 0 && resolution.Height > 0 {
		resolution.PixelAspectRatio = float64(resolution.Width) / float64(resolution.Height)
	}

	// Determine if anamorphic
	resolution.IsAnamorphic = ra.isAnamorphic(resolution.SampleAspectRatio, resolution.DisplayAspectRatio)

	// Check for common aspect ratios
	resolution.AspectRatioCategory = ra.categorizeAspectRatio(resolution.DisplayAspectRatio)

	// Detect orientation
	resolution.Orientation = ra.getOrientation(resolution.Width, resolution.Height)

	// Validate consistency
	resolution.IsConsistent = ra.validateResolutionConsistency(resolution, stream)

	return resolution
}

// getStandardResolution determines the standard resolution name
func (ra *ResolutionAnalyzer) getStandardResolution(width, height int) string {
	// Common resolution standards
	resolutions := map[string][2]int{
		"8K UHD":   {7680, 4320},
		"4K UHD":   {3840, 2160},
		"4K DCI":   {4096, 2160},
		"QHD":      {2560, 1440},
		"Full HD":  {1920, 1080},
		"HD Ready": {1280, 720},
		"WSXGA+":   {1680, 1050},
		"UXGA":     {1600, 1200},
		"SXGA":     {1280, 1024},
		"XGA":      {1024, 768},
		"SVGA":     {800, 600},
		"VGA":      {640, 480},
		"QVGA":     {320, 240},
	}

	// Exact match first
	for name, res := range resolutions {
		if width == res[0] && height == res[1] {
			return name
		}
	}

	// Close match (within 5% tolerance)
	tolerance := 0.05
	for name, res := range resolutions {
		widthDiff := math.Abs(float64(width-res[0])) / float64(res[0])
		heightDiff := math.Abs(float64(height-res[1])) / float64(res[1])
		if widthDiff <= tolerance && heightDiff <= tolerance {
			return name + " (approx)"
		}
	}

	// Custom resolution
	return fmt.Sprintf("%dx%d", width, height)
}

// getResolutionClass determines the resolution class
func (ra *ResolutionAnalyzer) getResolutionClass(height int) string {
	switch {
	case height >= 4320:
		return "8K"
	case height >= 2160:
		return "4K"
	case height >= 1440:
		return "QHD"
	case height >= 1080:
		return "Full HD"
	case height >= 720:
		return "HD"
	case height >= 480:
		return "SD"
	default:
		return "Low Resolution"
	}
}

// parseAspectRatio parses aspect ratio string to float64
func (ra *ResolutionAnalyzer) parseAspectRatio(aspectRatio string) float64 {
	if aspectRatio == "" || aspectRatio == "N/A" {
		return 0.0
	}

	// Handle ratios like "16:9" or "4:3"
	if strings.Contains(aspectRatio, ":") {
		parts := strings.Split(aspectRatio, ":")
		if len(parts) == 2 {
			num, err1 := strconv.ParseFloat(parts[0], 64)
			den, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil && den != 0 {
				return num / den
			}
		}
	}

	// Handle direct decimal values
	if val, err := strconv.ParseFloat(aspectRatio, 64); err == nil {
		return val
	}

	return 0.0
}

// isAnamorphic determines if content is anamorphic
func (ra *ResolutionAnalyzer) isAnamorphic(sampleAR, displayAR float64) bool {
	if sampleAR == 0.0 || displayAR == 0.0 {
		return false
	}

	// If sample and display aspect ratios differ significantly, it's likely anamorphic
	tolerance := 0.01
	return math.Abs(sampleAR-displayAR) > tolerance
}

// categorizeAspectRatio categorizes the aspect ratio
func (ra *ResolutionAnalyzer) categorizeAspectRatio(aspectRatio float64) string {
	if aspectRatio == 0.0 {
		return "Unknown"
	}

	// Common aspect ratio categories with tolerance
	tolerance := 0.05
	categories := map[string]float64{
		"4:3 (Standard)":       4.0 / 3.0,   // 1.333
		"16:9 (Widescreen)":    16.0 / 9.0,  // 1.778
		"16:10 (Widescreen)":   16.0 / 10.0, // 1.6
		"21:9 (Ultrawide)":     21.0 / 9.0,  // 2.333
		"2.35:1 (Cinemascope)": 2.35,
		"2.39:1 (Anamorphic)":  2.39,
		"1.85:1 (Academy)":     1.85,
		"1:1 (Square)":         1.0,
		"9:16 (Portrait)":      9.0 / 16.0, // 0.5625
	}

	for category, ratio := range categories {
		if math.Abs(aspectRatio-ratio) <= tolerance {
			return category
		}
	}

	// Determine general category
	if aspectRatio < 1.0 {
		return "Portrait"
	} else if aspectRatio > 2.0 {
		return "Ultra-widescreen"
	} else if aspectRatio > 1.5 {
		return "Widescreen"
	} else {
		return "Standard"
	}
}

// getOrientation determines content orientation
func (ra *ResolutionAnalyzer) getOrientation(width, height int) string {
	if width == height {
		return "Square"
	} else if width > height {
		return "Landscape"
	} else {
		return "Portrait"
	}
}

// validateResolutionConsistency validates resolution metadata consistency
func (ra *ResolutionAnalyzer) validateResolutionConsistency(resolution *VideoResolution, stream StreamInfo) bool {
	issues := 0

	// Check if coded dimensions match display dimensions
	if stream.CodedWidth > 0 && stream.CodedHeight > 0 {
		if stream.CodedWidth != resolution.Width || stream.CodedHeight != resolution.Height {
			// This is not necessarily an error - coded dimensions can differ
			// from display dimensions due to cropping
		}
	}

	// Check aspect ratio consistency
	if resolution.PixelAspectRatio > 0 && resolution.DisplayAspectRatio > 0 {
		// Calculate expected display AR from pixel AR and sample AR
		expectedDisplayAR := resolution.PixelAspectRatio
		if resolution.SampleAspectRatio > 0 {
			expectedDisplayAR *= resolution.SampleAspectRatio
		}

		tolerance := 0.1
		if math.Abs(expectedDisplayAR-resolution.DisplayAspectRatio) > tolerance {
			issues++
		}
	}

	return issues == 0
}

// isWidescreenContent determines if the content is primarily widescreen
func (ra *ResolutionAnalyzer) isWidescreenContent(videoStreams map[int]*VideoResolution) bool {
	if len(videoStreams) == 0 {
		return false
	}

	widescreenCount := 0
	for _, res := range videoStreams {
		if res.DisplayAspectRatio > 1.5 || res.PixelAspectRatio > 1.5 {
			widescreenCount++
		}
	}

	return float64(widescreenCount)/float64(len(videoStreams)) >= 0.5
}

// hasMultipleResolutions checks if there are multiple different resolutions
func (ra *ResolutionAnalyzer) hasMultipleResolutions(videoStreams map[int]*VideoResolution) bool {
	if len(videoStreams) <= 1 {
		return false
	}

	var firstWidth, firstHeight int
	first := true

	for _, res := range videoStreams {
		if first {
			firstWidth = res.Width
			firstHeight = res.Height
			first = false
		} else {
			if res.Width != firstWidth || res.Height != firstHeight {
				return true
			}
		}
	}

	return false
}

// validateResolution validates overall resolution characteristics
func (ra *ResolutionAnalyzer) validateResolution(analysis *ResolutionAnalysis) *ResolutionValidation {
	validation := &ResolutionValidation{
		IsValid:         true,
		Issues:          []string{},
		Recommendations: []string{},
	}

	// Check for inconsistencies in video streams
	for streamIndex, videoRes := range analysis.VideoStreams {
		if !videoRes.IsConsistent {
			validation.Issues = append(validation.Issues,
				fmt.Sprintf("Video stream %d has inconsistent resolution metadata", streamIndex))
			validation.IsValid = false
		}

		// Check for unusual aspect ratios
		if videoRes.DisplayAspectRatio > 0 {
			if videoRes.DisplayAspectRatio < 0.5 || videoRes.DisplayAspectRatio > 4.0 {
				validation.Issues = append(validation.Issues,
					fmt.Sprintf("Video stream %d has unusual aspect ratio: %.3f", streamIndex, videoRes.DisplayAspectRatio))
			}
		}

		// Check for very low resolutions
		if videoRes.PixelCount > 0 && videoRes.PixelCount < 76800 { // Less than 320x240
			validation.Issues = append(validation.Issues,
				fmt.Sprintf("Video stream %d has very low resolution: %dx%d", streamIndex, videoRes.Width, videoRes.Height))
		}
	}

	// Provide recommendations
	if !analysis.IsHighDefinition {
		validation.Recommendations = append(validation.Recommendations,
			"Consider upgrading to HD resolution (720p or higher) for better viewer experience")
	}

	if analysis.HasMultipleResolutions {
		validation.Recommendations = append(validation.Recommendations,
			"Multiple resolutions detected - ensure this is intentional for adaptive streaming")
	}

	// Check for portrait orientation in what might be landscape content
	for streamIndex, videoRes := range analysis.VideoStreams {
		if videoRes.Orientation == "Portrait" && videoRes.Height > 1000 {
			validation.Recommendations = append(validation.Recommendations,
				fmt.Sprintf("Video stream %d appears to be portrait orientation - verify this is intentional", streamIndex))
		}
	}

	return validation
}
