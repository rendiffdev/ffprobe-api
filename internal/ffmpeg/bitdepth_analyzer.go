package ffmpeg

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// BitDepthAnalyzer handles bit depth analysis
type BitDepthAnalyzer struct{}

// NewBitDepthAnalyzer creates a new bit depth analyzer
func NewBitDepthAnalyzer() *BitDepthAnalyzer {
	return &BitDepthAnalyzer{}
}

// AnalyzeBitDepth analyzes bit depth from stream information
func (bda *BitDepthAnalyzer) AnalyzeBitDepth(streams []StreamInfo) *BitDepthAnalysis {
	analysis := &BitDepthAnalysis{
		VideoStreams: make(map[int]*VideoBitDepth),
		AudioStreams: make(map[int]*AudioBitDepth),
	}

	for _, stream := range streams {
		switch strings.ToLower(stream.CodecType) {
		case "video":
			videoBitDepth := bda.analyzeVideoBitDepth(stream)
			if videoBitDepth != nil {
				analysis.VideoStreams[stream.Index] = videoBitDepth
				if analysis.MaxVideoBitDepth < videoBitDepth.BitDepth {
					analysis.MaxVideoBitDepth = videoBitDepth.BitDepth
				}
			}
		case "audio":
			audioBitDepth := bda.analyzeAudioBitDepth(stream)
			if audioBitDepth != nil {
				analysis.AudioStreams[stream.Index] = audioBitDepth
				if analysis.MaxAudioBitDepth < audioBitDepth.BitDepth {
					analysis.MaxAudioBitDepth = audioBitDepth.BitDepth
				}
			}
		}
	}

	// Determine overall characteristics
	analysis.IsHDR = analysis.MaxVideoBitDepth >= 10
	analysis.IsHighBitDepth = analysis.MaxVideoBitDepth > 8 || analysis.MaxAudioBitDepth > 16
	analysis.Validation = bda.validateBitDepth(analysis)

	return analysis
}

// analyzeVideoBitDepth extracts video bit depth information
func (bda *BitDepthAnalyzer) analyzeVideoBitDepth(stream StreamInfo) *VideoBitDepth {
	bitDepth := &VideoBitDepth{
		BitDepth: 8, // Default to 8-bit
		Source:   "default",
	}

	// Extract from pixel format
	if stream.PixFmt != "" {
		if depth := bda.extractBitDepthFromPixelFormat(stream.PixFmt); depth > 0 {
			bitDepth.BitDepth = depth
			bitDepth.Source = "pixel_format"
			bitDepth.PixelFormat = stream.PixFmt
		}
	}

	// Extract from bits per raw sample
	if stream.BitsPerRawSample != "" {
		if depth, err := strconv.Atoi(stream.BitsPerRawSample); err == nil && depth > 0 {
			bitDepth.BitDepth = depth
			bitDepth.Source = "bits_per_raw_sample"
		}
	}

	// Extract from codec profile
	if stream.Profile != "" {
		if depth := bda.extractBitDepthFromProfile(stream.Profile); depth > 0 {
			if bitDepth.Source == "default" {
				bitDepth.BitDepth = depth
				bitDepth.Source = "codec_profile"
			}
			bitDepth.ProfileIndicatedDepth = depth
		}
	}

	// Validate consistency
	bitDepth.IsConsistent = bda.validateVideoBitDepthConsistency(bitDepth, stream)

	return bitDepth
}

// analyzeAudioBitDepth extracts audio bit depth information
func (bda *BitDepthAnalyzer) analyzeAudioBitDepth(stream StreamInfo) *AudioBitDepth {
	bitDepth := &AudioBitDepth{
		BitDepth: 16, // Default to 16-bit for audio
		Source:   "default",
	}

	// Extract from sample format
	if stream.SampleFmt != "" {
		if depth := bda.extractBitDepthFromSampleFormat(stream.SampleFmt); depth > 0 {
			bitDepth.BitDepth = depth
			bitDepth.Source = "sample_format"
			bitDepth.SampleFormat = stream.SampleFmt
		}
	}

	// Extract from bits per sample
	if stream.BitsPerSample > 0 {
		bitDepth.BitDepth = stream.BitsPerSample
		bitDepth.Source = "bits_per_sample"
	}

	// Extract from bits per raw sample
	if stream.BitsPerRawSample != "" {
		if depth, err := strconv.Atoi(stream.BitsPerRawSample); err == nil && depth > 0 {
			if bitDepth.Source == "default" {
				bitDepth.BitDepth = depth
				bitDepth.Source = "bits_per_raw_sample"
			}
		}
	}

	// Validate consistency
	bitDepth.IsConsistent = bda.validateAudioBitDepthConsistency(bitDepth, stream)

	return bitDepth
}

// extractBitDepthFromPixelFormat extracts bit depth from pixel format name
func (bda *BitDepthAnalyzer) extractBitDepthFromPixelFormat(pixfmt string) int {
	// Common pixel format to bit depth mappings
	bitDepthMap := map[string]int{
		// 8-bit formats
		"yuv420p":  8,
		"yuv422p":  8,
		"yuv444p":  8,
		"yuvj420p": 8,
		"yuvj422p": 8,
		"yuvj444p": 8,
		"rgb24":    8,
		"bgr24":    8,
		"rgba":     8,
		"bgra":     8,

		// 10-bit formats
		"yuv420p10le": 10,
		"yuv420p10be": 10,
		"yuv422p10le": 10,
		"yuv422p10be": 10,
		"yuv444p10le": 10,
		"yuv444p10be": 10,
		"yuv420p10":   10,
		"yuv422p10":   10,
		"yuv444p10":   10,
		"p010le":      10,
		"p010be":      10,

		// 12-bit formats
		"yuv420p12le": 12,
		"yuv420p12be": 12,
		"yuv422p12le": 12,
		"yuv422p12be": 12,
		"yuv444p12le": 12,
		"yuv444p12be": 12,
		"yuv420p12":   12,
		"yuv422p12":   12,
		"yuv444p12":   12,

		// 16-bit formats
		"yuv420p16le": 16,
		"yuv420p16be": 16,
		"yuv422p16le": 16,
		"yuv422p16be": 16,
		"yuv444p16le": 16,
		"yuv444p16be": 16,
		"yuv420p16":   16,
		"yuv422p16":   16,
		"yuv444p16":   16,
		"rgb48le":     16,
		"rgb48be":     16,
		"rgba64le":    16,
		"rgba64be":    16,
	}

	if depth, exists := bitDepthMap[strings.ToLower(pixfmt)]; exists {
		return depth
	}

	// Extract using regex patterns
	patterns := []struct {
		regex    string
		bitDepth int
	}{
		{`p(\d+)le$`, 0}, // p010le, p016le etc.
		{`p(\d+)be$`, 0}, // p010be, p016be etc.
		{`(\d+)bit`, 0},  // 10bit, 12bit etc.
		{`(\d+)le$`, 0},  // 10le, 12le etc.
		{`(\d+)be$`, 0},  // 10be, 12be etc.
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern.regex)
		matches := re.FindStringSubmatch(strings.ToLower(pixfmt))
		if len(matches) >= 2 {
			if depth, err := strconv.Atoi(matches[1]); err == nil && depth > 0 {
				return depth
			}
		}
	}

	return 0
}

// extractBitDepthFromSampleFormat extracts bit depth from audio sample format
func (bda *BitDepthAnalyzer) extractBitDepthFromSampleFormat(sampleFmt string) int {
	// Common sample format to bit depth mappings
	bitDepthMap := map[string]int{
		"u8":   8,
		"u8p":  8,
		"s16":  16,
		"s16p": 16,
		"s32":  32,
		"s32p": 32,
		"s64":  64,
		"s64p": 64,
		"flt":  32, // 32-bit float
		"fltp": 32, // 32-bit float planar
		"dbl":  64, // 64-bit double
		"dblp": 64, // 64-bit double planar
	}

	if depth, exists := bitDepthMap[strings.ToLower(sampleFmt)]; exists {
		return depth
	}

	// Extract using regex
	re := regexp.MustCompile(`s(\d+)p?$`)
	matches := re.FindStringSubmatch(strings.ToLower(sampleFmt))
	if len(matches) >= 2 {
		if depth, err := strconv.Atoi(matches[1]); err == nil && depth > 0 {
			return depth
		}
	}

	return 0
}

// extractBitDepthFromProfile extracts bit depth hints from codec profile
func (bda *BitDepthAnalyzer) extractBitDepthFromProfile(profile string) int {
	profileLower := strings.ToLower(profile)

	// HEVC profiles
	if strings.Contains(profileLower, "main 10") || strings.Contains(profileLower, "main10") {
		return 10
	}
	if strings.Contains(profileLower, "main 12") || strings.Contains(profileLower, "main12") {
		return 12
	}

	// AV1 profiles
	if strings.Contains(profileLower, "professional") {
		return 12 // AV1 Professional typically uses 12-bit
	}

	// VP9 profiles
	if strings.Contains(profileLower, "profile 2") || strings.Contains(profileLower, "profile 3") {
		return 10 // VP9 Profile 2/3 support 10/12-bit
	}

	// Extract numeric bit depth from profile string
	re := regexp.MustCompile(`(\d+)\s*bit`)
	matches := re.FindStringSubmatch(profileLower)
	if len(matches) >= 2 {
		if depth, err := strconv.Atoi(matches[1]); err == nil && depth > 0 {
			return depth
		}
	}

	return 0
}

// validateVideoBitDepthConsistency validates consistency between different bit depth indicators
func (bda *BitDepthAnalyzer) validateVideoBitDepthConsistency(bitDepth *VideoBitDepth, stream StreamInfo) bool {
	indicators := []int{}

	// Collect all bit depth indicators
	if stream.BitsPerRawSample != "" {
		if depth, err := strconv.Atoi(stream.BitsPerRawSample); err == nil && depth > 0 {
			indicators = append(indicators, depth)
		}
	}

	if pixelDepth := bda.extractBitDepthFromPixelFormat(stream.PixFmt); pixelDepth > 0 {
		indicators = append(indicators, pixelDepth)
	}

	if profileDepth := bda.extractBitDepthFromProfile(stream.Profile); profileDepth > 0 {
		indicators = append(indicators, profileDepth)
	}

	// Check consistency
	if len(indicators) <= 1 {
		return true // Not enough data to check consistency
	}

	first := indicators[0]
	for _, depth := range indicators[1:] {
		if depth != first {
			return false
		}
	}

	return true
}

// validateAudioBitDepthConsistency validates consistency between audio bit depth indicators
func (bda *BitDepthAnalyzer) validateAudioBitDepthConsistency(bitDepth *AudioBitDepth, stream StreamInfo) bool {
	indicators := []int{}

	// Collect all bit depth indicators
	if stream.BitsPerSample > 0 {
		indicators = append(indicators, stream.BitsPerSample)
	}

	if stream.BitsPerRawSample != "" {
		if depth, err := strconv.Atoi(stream.BitsPerRawSample); err == nil && depth > 0 {
			indicators = append(indicators, depth)
		}
	}

	if sampleDepth := bda.extractBitDepthFromSampleFormat(stream.SampleFmt); sampleDepth > 0 {
		indicators = append(indicators, sampleDepth)
	}

	// Check consistency
	if len(indicators) <= 1 {
		return true
	}

	first := indicators[0]
	for _, depth := range indicators[1:] {
		if depth != first {
			return false
		}
	}

	return true
}

// validateBitDepth validates overall bit depth characteristics
func (bda *BitDepthAnalyzer) validateBitDepth(analysis *BitDepthAnalysis) *BitDepthValidation {
	validation := &BitDepthValidation{
		IsValid:         true,
		Issues:          []string{},
		Recommendations: []string{},
	}

	// Check for inconsistencies in video streams
	for streamIndex, videoBitDepth := range analysis.VideoStreams {
		if !videoBitDepth.IsConsistent {
			validation.Issues = append(validation.Issues,
				fmt.Sprintf("Video stream %d has inconsistent bit depth indicators", streamIndex))
			validation.IsValid = false
		}
	}

	// Check for inconsistencies in audio streams
	for streamIndex, audioBitDepth := range analysis.AudioStreams {
		if !audioBitDepth.IsConsistent {
			validation.Issues = append(validation.Issues,
				fmt.Sprintf("Audio stream %d has inconsistent bit depth indicators", streamIndex))
			validation.IsValid = false
		}
	}

	// Provide recommendations
	if analysis.MaxVideoBitDepth >= 10 {
		validation.Recommendations = append(validation.Recommendations,
			"High bit depth video detected - ensure delivery pipeline supports 10+ bit content")
	}

	if analysis.MaxAudioBitDepth > 24 {
		validation.Recommendations = append(validation.Recommendations,
			"High resolution audio detected - verify compatibility with target playback devices")
	}

	// Check for common issues
	if analysis.MaxVideoBitDepth == 8 && analysis.IsHDR {
		validation.Issues = append(validation.Issues,
			"HDR content detected with 8-bit depth - HDR typically requires 10+ bit depth")
		validation.IsValid = false
	}

	return validation
}
