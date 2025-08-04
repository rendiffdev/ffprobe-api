package ffmpeg

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

// EnhancedAnalyzer provides additional quality control analysis
type EnhancedAnalyzer struct {
	contentAnalyzer *ContentAnalyzer
}

// NewEnhancedAnalyzer creates a new enhanced analyzer
func NewEnhancedAnalyzer() *EnhancedAnalyzer {
	return &EnhancedAnalyzer{}
}

// NewEnhancedAnalyzerWithContentAnalysis creates analyzer with content analysis capability
func NewEnhancedAnalyzerWithContentAnalysis(ffmpegPath string, logger zerolog.Logger) *EnhancedAnalyzer {
	return &EnhancedAnalyzer{
		contentAnalyzer: NewContentAnalyzer(ffmpegPath, logger),
	}
}

// AnalyzeResult performs enhanced analysis on FFprobe results
func (ea *EnhancedAnalyzer) AnalyzeResult(result *FFprobeResult) error {
	if result == nil {
		return fmt.Errorf("result cannot be nil")
	}

	enhanced := &EnhancedAnalysis{}

	// Analyze stream counts
	if len(result.Streams) > 0 {
		enhanced.StreamCounts = ea.analyzeStreamCounts(result.Streams)
	}

	// Analyze video streams
	videoAnalysis := ea.analyzeVideoStreams(result.Streams)
	if videoAnalysis != nil {
		enhanced.VideoAnalysis = videoAnalysis
	}

	// Analyze audio streams
	audioAnalysis := ea.analyzeAudioStreams(result.Streams)
	if audioAnalysis != nil {
		enhanced.AudioAnalysis = audioAnalysis
	}

	// Analyze GOP structure if frames are available
	if len(result.Frames) > 0 {
		enhanced.GOPAnalysis = ea.analyzeGOPStructure(result.Frames)
		enhanced.FrameStatistics = ea.analyzeFrameStatistics(result.Frames)
	}

	result.EnhancedAnalysis = enhanced
	return nil
}

// AnalyzeResultWithContent performs enhanced analysis including content analysis
func (ea *EnhancedAnalyzer) AnalyzeResultWithContent(ctx context.Context, result *FFprobeResult, filePath string) error {
	// First run standard enhanced analysis
	if err := ea.AnalyzeResult(result); err != nil {
		return err
	}

	// Run content analysis if analyzer is available
	if ea.contentAnalyzer != nil && filePath != "" {
		contentAnalysis, err := ea.contentAnalyzer.AnalyzeContent(ctx, filePath)
		if err != nil {
			return fmt.Errorf("content analysis failed: %w", err)
		}
		result.EnhancedAnalysis.ContentAnalysis = contentAnalysis
	}

	return nil
}

// analyzeStreamCounts counts different types of streams
func (ea *EnhancedAnalyzer) analyzeStreamCounts(streams []StreamInfo) *StreamCounts {
	counts := &StreamCounts{}

	for _, stream := range streams {
		counts.TotalStreams++
		
		switch strings.ToLower(stream.CodecType) {
		case "video":
			counts.VideoStreams++
		case "audio":
			counts.AudioStreams++
		case "subtitle":
			counts.SubtitleStreams++
		case "data":
			counts.DataStreams++
		case "attachment":
			counts.AttachmentStreams++
		}
	}

	return counts
}

// analyzeVideoStreams performs enhanced video stream analysis
func (ea *EnhancedAnalyzer) analyzeVideoStreams(streams []StreamInfo) *VideoAnalysis {
	var analysis *VideoAnalysis

	for _, stream := range streams {
		if strings.ToLower(stream.CodecType) != "video" {
			continue
		}

		if analysis == nil {
			analysis = &VideoAnalysis{}
		}

		// Extract chroma subsampling from pixel format
		if stream.PixFmt != "" {
			chroma := ea.extractChromaSubsampling(stream.PixFmt)
			if chroma != "" {
				analysis.ChromaSubsampling = &chroma
			}
		}

		// Extract matrix coefficients from color space
		if stream.ColorSpace != "" {
			matrix := ea.extractMatrixCoefficients(stream.ColorSpace)
			if matrix != "" {
				analysis.MatrixCoefficients = &matrix
			}
		}

		// Check for closed captions
		if stream.ClosedCaptions > 0 {
			analysis.HasClosedCaptions = true
		}

		// Analyze bit rate mode (basic implementation)
		if stream.BitRate != "" {
			mode := ea.analyzeBitRateMode(stream.BitRate)
			if mode != "" {
				analysis.BitRateMode = &mode
			}
		}

		// Only analyze the first video stream for now
		break
	}

	return analysis
}

// analyzeAudioStreams performs enhanced audio stream analysis
func (ea *EnhancedAnalyzer) analyzeAudioStreams(streams []StreamInfo) *AudioAnalysis {
	var analysis *AudioAnalysis

	for _, stream := range streams {
		if strings.ToLower(stream.CodecType) != "audio" {
			continue
		}

		if analysis == nil {
			analysis = &AudioAnalysis{}
		}

		// Analyze audio bit rate mode
		if stream.BitRate != "" {
			mode := ea.analyzeBitRateMode(stream.BitRate)
			if mode != "" {
				analysis.BitRateMode = &mode
			}
		}

		// Only analyze the first audio stream for now
		break
	}

	return analysis
}

// analyzeGOPStructure analyzes Group of Pictures structure from frame data
func (ea *EnhancedAnalyzer) analyzeGOPStructure(frames []FrameInfo) *GOPAnalysis {
	analysis := &GOPAnalysis{
		TotalFrameCount: len(frames),
	}

	var gopSizes []int
	var currentGOPSize int
	var keyFrameCount int

	for _, frame := range frames {
		if strings.ToLower(frame.MediaType) != "video" {
			continue
		}

		currentGOPSize++

		if frame.KeyFrame == 1 {
			keyFrameCount++
			if currentGOPSize > 1 {
				gopSizes = append(gopSizes, currentGOPSize-1)
			}
			currentGOPSize = 1
		}
	}

	// Add the last GOP if it exists
	if currentGOPSize > 0 {
		gopSizes = append(gopSizes, currentGOPSize)
	}

	analysis.KeyFrameCount = keyFrameCount

	if len(gopSizes) > 0 {
		// Calculate GOP statistics
		sum := 0
		maxGOP := gopSizes[0]
		minGOP := gopSizes[0]

		for _, size := range gopSizes {
			sum += size
			if size > maxGOP {
				maxGOP = size
			}
			if size < minGOP {
				minGOP = size
			}
		}

		avgGOP := float64(sum) / float64(len(gopSizes))
		analysis.AverageGOPSize = &avgGOP
		analysis.MaxGOPSize = &maxGOP
		analysis.MinGOPSize = &minGOP

		// Determine GOP pattern (basic)
		pattern := ea.determineGOPPattern(gopSizes)
		if pattern != "" {
			analysis.GOPPattern = &pattern
		}
	}

	return analysis
}

// extractChromaSubsampling extracts chroma subsampling info from pixel format
func (ea *EnhancedAnalyzer) extractChromaSubsampling(pixfmt string) string {
	// Common pixel format to chroma subsampling mappings
	chromaMap := map[string]string{
		"yuv420p":   "4:2:0",
		"yuv422p":   "4:2:2",
		"yuv444p":   "4:4:4",
		"yuv410p":   "4:1:0",
		"yuv411p":   "4:1:1",
		"yuvj420p":  "4:2:0",
		"yuvj422p":  "4:2:2",
		"yuvj444p":  "4:4:4",
		"yuv420p10": "4:2:0",
		"yuv422p10": "4:2:2",
		"yuv444p10": "4:4:4",
	}

	if chroma, exists := chromaMap[strings.ToLower(pixfmt)]; exists {
		return chroma
	}

	// Try to extract from format name patterns
	lower := strings.ToLower(pixfmt)
	if strings.Contains(lower, "420") {
		return "4:2:0"
	} else if strings.Contains(lower, "422") {
		return "4:2:2"
	} else if strings.Contains(lower, "444") {
		return "4:4:4"
	}

	return ""
}

// extractMatrixCoefficients extracts matrix coefficients from color space
func (ea *EnhancedAnalyzer) extractMatrixCoefficients(colorSpace string) string {
	// Map color space to matrix coefficients
	matrixMap := map[string]string{
		"bt709":     "ITU-R BT.709",
		"bt601":     "ITU-R BT.601",
		"bt2020":    "ITU-R BT.2020",
		"smpte170m": "SMPTE 170M",
		"smpte240m": "SMPTE 240M",
		"fcc":       "FCC",
		"ycgco":     "YCgCo",
	}

	lower := strings.ToLower(colorSpace)
	if matrix, exists := matrixMap[lower]; exists {
		return matrix
	}

	return colorSpace
}

// analyzeBitRateMode determines if bitrate is CBR or VBR (basic heuristic)
func (ea *EnhancedAnalyzer) analyzeBitRateMode(bitrate string) string {
	// This is a basic implementation
	// In a real scenario, you'd analyze bitrate variations over time
	if bitrate == "" || bitrate == "N/A" {
		return "VBR" // Variable if not specified
	}

	// If we have a specific number, assume CBR for now
	if _, err := strconv.Atoi(bitrate); err == nil {
		return "CBR"
	}

	return "Unknown"
}

// determineGOPPattern determines the GOP pattern from sizes
func (ea *EnhancedAnalyzer) determineGOPPattern(gopSizes []int) string {
	if len(gopSizes) == 0 {
		return ""
	}

	// Check if all GOPs are the same size (regular pattern)
	firstSize := gopSizes[0]
	allSame := true
	for _, size := range gopSizes {
		if size != firstSize {
			allSame = false
			break
		}
	}

	if allSame {
		return fmt.Sprintf("Regular (GOP=%d)", firstSize)
	}

	// Check for common patterns
	if len(gopSizes) >= 2 {
		// Simple alternating pattern check
		if len(gopSizes)%2 == 0 {
			alternating := true
			for i := 0; i < len(gopSizes)-1; i += 2 {
				if i+1 < len(gopSizes) && gopSizes[i] != gopSizes[i+1] {
					alternating = false
					break
				}
			}
			if alternating {
				return fmt.Sprintf("Alternating (%d/%d)", gopSizes[0], gopSizes[1])
			}
		}
	}

	return "Irregular"
}

// analyzeFrameStatistics provides comprehensive frame-level statistics
func (ea *EnhancedAnalyzer) analyzeFrameStatistics(frames []FrameInfo) *FrameStatistics {
	stats := &FrameStatistics{
		FrameTypes: make(map[string]int),
	}

	var frameSizes []int64
	totalFrames := 0
	iFrames := 0
	pFrames := 0 
	bFrames := 0

	for _, frame := range frames {
		if strings.ToLower(frame.MediaType) != "video" {
			continue
		}

		totalFrames++

		// Count frame types
		if frame.PictType != "" {
			frameType := strings.ToUpper(frame.PictType)
			stats.FrameTypes[frameType]++

			switch frameType {
			case "I":
				iFrames++
			case "P":
				pFrames++
			case "B":
				bFrames++
			}
		}

		// Track frame sizes if available
		if frame.PktSize != "" {
			if size, err := strconv.ParseInt(frame.PktSize, 10, 64); err == nil {
				frameSizes = append(frameSizes, size)
			}
		}
	}

	stats.TotalFrames = totalFrames
	stats.IFrames = iFrames
	stats.PFrames = pFrames
	stats.BFrames = bFrames

	// Calculate frame size statistics
	if len(frameSizes) > 0 {
		var sum int64
		maxSize := frameSizes[0]
		minSize := frameSizes[0]

		for _, size := range frameSizes {
			sum += size
			if size > maxSize {
				maxSize = size
			}
			if size < minSize {
				minSize = size
			}
		}

		avgSize := float64(sum) / float64(len(frameSizes))
		stats.AverageFrameSize = &avgSize
		stats.MaxFrameSize = &maxSize
		stats.MinFrameSize = &minSize
	}

	return stats
}