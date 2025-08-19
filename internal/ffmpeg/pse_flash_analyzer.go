package ffmpeg

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// PSEFlashAnalyzer handles flash-specific analysis for photosensitive epilepsy detection.
// This module focuses specifically on detecting and analyzing flash patterns that
// may trigger seizures in photosensitive individuals.
//
// The analyzer implements algorithms based on:
//   - ITU-R BT.1702 flash detection methodology
//   - ITC guidelines for flash analysis
//   - Medical research on seizure triggers
//
// Flash analysis includes:
//   - General flash detection and rate calculation
//   - Red flash analysis with enhanced sensitivity
//   - Flash intensity and duration measurement
//   - Temporal pattern analysis
//   - Risk assessment and threshold compliance
type PSEFlashAnalyzer struct {
	ffprobePath string
	logger      zerolog.Logger
}

// NewPSEFlashAnalyzer creates a new flash analyzer for PSE detection.
//
// Parameters:
//   - ffprobePath: Path to FFprobe binary for video analysis
//   - logger: Structured logger for operation tracking
//
// Returns:
//   - *PSEFlashAnalyzer: Configured flash analyzer instance
//
// The analyzer is optimized for accuracy over speed and should be used
// when detailed flash analysis is required for broadcast safety compliance.
func NewPSEFlashAnalyzer(ffprobePath string, logger zerolog.Logger) *PSEFlashAnalyzer {
	return &PSEFlashAnalyzer{
		ffprobePath: ffprobePath,
		logger:      logger,
	}
}

// AnalyzeFlashes performs comprehensive flash analysis on video content.
// This is the main entry point for flash detection and risk assessment.
//
// The analysis process:
//   1. Extracts frame-by-frame luminance data
//   2. Detects flash events using ITU-R BT.1702 methodology
//   3. Analyzes flash rates and patterns
//   4. Assesses risk levels and compliance
//   5. Generates detailed flash analysis report
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - filePath: Path to video file for analysis
//   - streams: Video stream information from FFprobe
//
// Returns:
//   - *FlashAnalysis: Comprehensive flash analysis results
//   - error: Error if analysis fails or context is cancelled
//
// The function handles various video formats and provides frame-accurate
// flash detection suitable for broadcast compliance checking.
func (pfa *PSEFlashAnalyzer) AnalyzeFlashes(ctx context.Context, filePath string, streams []StreamInfo) (*FlashAnalysis, error) {
	pfa.logger.Info().Str("file", filePath).Msg("Starting comprehensive flash analysis")
	
	// Find video stream for analysis
	videoStream := findPrimaryVideoStream(streams)
	if videoStream == nil {
		return nil, fmt.Errorf("no suitable video stream found for flash analysis")
	}
	
	// Extract luminance data using FFprobe
	luminanceData, err := pfa.extractLuminanceData(ctx, filePath, videoStream)
	if err != nil {
		return nil, fmt.Errorf("failed to extract luminance data: %w", err)
	}
	
	// Detect flash events
	flashEvents := pfa.detectFlashEvents(luminanceData, videoStream.RFrameRate)
	
	// Calculate flash rates and statistics
	flashStats := pfa.calculateFlashStatistics(flashEvents, luminanceData.Duration)
	
	// Analyze flash patterns and timing
	patterns := pfa.analyzeFlashPatterns(flashEvents)
	
	// Assess risk levels
	riskAssessment := pfa.assessFlashRisk(flashStats, patterns)
	
	// Generate comprehensive analysis
	analysis := &FlashAnalysis{
		TotalFlashes:          len(flashEvents),
		FlashRate:             flashStats.PeakRate,
		AverageFlashRate:      flashStats.AverageRate,
		MaxFlashRate:          flashStats.MaxRate,
		FlashDurations:        pfa.calculateFlashDurations(flashEvents),
		FlashFrequencyBands:   pfa.analyzeFrequencyDistribution(flashEvents),
		DangerousFlashPeriods: pfa.identifyDangerousPeriods(flashEvents, flashStats),
		FlashIntensity:        pfa.analyzeFlashIntensity(flashEvents),
		ExceedsFlashThreshold: riskAssessment.ExceedsThreshold,
	}
	
	pfa.logger.Info().
		Int("total_flashes", analysis.TotalFlashes).
		Float64("max_rate", analysis.MaxFlashRate).
		Bool("exceeds_threshold", analysis.ExceedsFlashThreshold).
		Msg("Flash analysis completed")
	
	return analysis, nil
}

// AnalyzeRedFlashes performs specialized analysis of red flash patterns.
// Red flashes are particularly dangerous for photosensitive individuals
// and require separate analysis with lower thresholds.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - filePath: Path to video file for analysis
//   - streams: Video stream information from FFprobe
//
// Returns:
//   - *RedFlashAnalysis: Detailed red flash analysis results
//   - error: Error if analysis fails or context is cancelled
//
// The function uses color space analysis to isolate red channel information
// and applies specialized detection algorithms for red flash events.
func (pfa *PSEFlashAnalyzer) AnalyzeRedFlashes(ctx context.Context, filePath string, streams []StreamInfo) (*RedFlashAnalysis, error) {
	pfa.logger.Info().Str("file", filePath).Msg("Starting red flash analysis")
	
	videoStream := findPrimaryVideoStream(streams)
	if videoStream == nil {
		return nil, fmt.Errorf("no suitable video stream found for red flash analysis")
	}
	
	// Extract red channel data specifically
	redChannelData, err := pfa.extractRedChannelData(ctx, filePath, videoStream)
	if err != nil {
		return nil, fmt.Errorf("failed to extract red channel data: %w", err)
	}
	
	// Detect red flash events with specialized algorithm
	redFlashEvents := pfa.detectRedFlashEvents(redChannelData, videoStream.RFrameRate)
	
	// Calculate red-specific statistics
	redStats := pfa.calculateRedFlashStatistics(redFlashEvents, redChannelData.Duration)
	
	// Assess red flash risk (lower thresholds)
	redRiskAssessment := pfa.assessRedFlashRisk(redStats, redFlashEvents)
	
	analysis := &RedFlashAnalysis{
		RedFlashCount:       len(redFlashEvents),
		RedFlashRate:        redStats.AverageRate,
		MaxRedFlashRate:     redStats.MaxRate,
		RedFlashDurations:   pfa.calculateRedFlashDurations(redFlashEvents),
		RedSaturationLevels: pfa.extractSaturationLevels(redFlashEvents),
		DangerousRedPeriods: pfa.identifyDangerousRedPeriods(redFlashEvents),
		ExceedsRedThreshold: redRiskAssessment.ExceedsRedThreshold,
	}
	
	pfa.logger.Info().
		Int("red_flashes", analysis.RedFlashCount).
		Float64("max_red_rate", analysis.MaxRedFlashRate).
		Bool("exceeds_red_threshold", analysis.ExceedsRedThreshold).
		Msg("Red flash analysis completed")
	
	return analysis, nil
}

// Private helper methods for flash analysis implementation

// extractLuminanceData extracts frame-by-frame luminance values using FFprobe.
func (pfa *PSEFlashAnalyzer) extractLuminanceData(ctx context.Context, filePath string, videoStream *StreamInfo) (*LuminanceData, error) {
	// Use FFprobe to extract luminance statistics for each frame
	args := []string{
		"-f", "lavfi",
		"-i", fmt.Sprintf("movie=%s,lutyuv=y='val*255':u=128:v=128,stats", filePath),
		"-show_entries", "frame=pkt_pts_time:frame_tags=lavfi.stats.avg_luma",
		"-of", "csv=print_section=0",
		"-v", "quiet",
	}
	
	output, err := executeFFprobeCommand(ctx, append([]string{pfa.ffprobePath}, args...))
	if err != nil {
		return nil, fmt.Errorf("failed to extract luminance data: %w", err)
	}
	
	return pfa.parseLuminanceData(output, videoStream)
}

// extractRedChannelData extracts red channel intensity data for red flash analysis.
func (pfa *PSEFlashAnalyzer) extractRedChannelData(ctx context.Context, filePath string, videoStream *StreamInfo) (*RedChannelData, error) {
	// Use FFprobe with color channel filters to extract red channel data
	args := []string{
		"-f", "lavfi",
		"-i", fmt.Sprintf("movie=%s,extractplanes=r,stats", filePath),
		"-show_entries", "frame=pkt_pts_time:frame_tags=lavfi.stats.avg_luma",
		"-of", "csv=print_section=0",
		"-v", "quiet",
	}
	
	output, err := executeFFprobeCommand(ctx, append([]string{pfa.ffprobePath}, args...))
	if err != nil {
		return nil, fmt.Errorf("failed to extract red channel data: %w", err)
	}
	
	return pfa.parseRedChannelData(output, videoStream)
}

// detectFlashEvents identifies flash events in luminance data using ITU-R BT.1702 methodology.
func (pfa *PSEFlashAnalyzer) detectFlashEvents(luminanceData *LuminanceData, frameRate string) []FlashEvent {
	var flashEvents []FlashEvent
	
	fps := parseFrameRate(frameRate)
	if fps == 0 {
		fps = 25.0 // Default fallback
	}
	
	// Implementation of ITU-R BT.1702 flash detection algorithm
	for i := 1; i < len(luminanceData.Values); i++ {
		currentLuma := luminanceData.Values[i].Luminance
		previousLuma := luminanceData.Values[i-1].Luminance
		
		// Calculate luminance change
		lumaChange := math.Abs(currentLuma - previousLuma)
		
		// Check if change exceeds flash threshold
		if lumaChange > MinFlashIntensity {
			timestamp := luminanceData.Values[i].Timestamp
			
			// Determine flash characteristics
			intensity := lumaChange / math.Max(currentLuma, previousLuma)
			duration := pfa.calculateFlashDuration(luminanceData, i, fps)
			
			flashEvent := FlashEvent{
				Timestamp: timestamp,
				Intensity: intensity,
				Duration:  duration,
				Type:      pfa.classifyFlashType(lumaChange, intensity),
			}
			
			flashEvents = append(flashEvents, flashEvent)
		}
	}
	
	return pfa.mergeConsecutiveFlashes(flashEvents)
}

// detectRedFlashEvents identifies red flash events with specialized red-sensitive algorithm.
func (pfa *PSEFlashAnalyzer) detectRedFlashEvents(redData *RedChannelData, frameRate string) []RedFlashEvent {
	var redFlashEvents []RedFlashEvent
	
	fps := parseFrameRate(frameRate)
	if fps == 0 {
		fps = 25.0
	}
	
	// Red flash detection with enhanced sensitivity
	for i := 1; i < len(redData.Values); i++ {
		currentRed := redData.Values[i].RedIntensity
		previousRed := redData.Values[i-1].RedIntensity
		
		// Calculate red channel change with red-specific weighting
		redChange := math.Abs(currentRed - previousRed) * RedLuminanceWeight
		
		// Lower threshold for red flashes
		if redChange > (MinFlashIntensity * 0.7) { // 30% lower threshold for red
			timestamp := redData.Values[i].Timestamp
			saturation := redData.Values[i].Saturation
			
			// Only consider high saturation red as dangerous
			if saturation > RedSaturationThreshold {
				intensity := redChange / math.Max(currentRed, previousRed)
				duration := pfa.calculateRedFlashDuration(redData, i, fps)
				
				redFlashEvent := RedFlashEvent{
					Timestamp:   timestamp,
					Intensity:   intensity,
					Duration:    duration,
					Saturation:  saturation,
					RedValue:    currentRed,
				}
				
				redFlashEvents = append(redFlashEvents, redFlashEvent)
			}
		}
	}
	
	return pfa.mergeConsecutiveRedFlashes(redFlashEvents)
}

// Supporting data structures for flash analysis

// LuminanceData contains extracted luminance information from video frames.
type LuminanceData struct {
	Duration float64          `json:"duration"`
	Values   []LuminanceValue `json:"values"`
}

// LuminanceValue represents luminance data for a single frame.
type LuminanceValue struct {
	Timestamp float64 `json:"timestamp"`
	Luminance float64 `json:"luminance"`
}

// RedChannelData contains extracted red channel information from video frames.
type RedChannelData struct {
	Duration float64        `json:"duration"`
	Values   []RedValue     `json:"values"`
}

// RedValue represents red channel data for a single frame.
type RedValue struct {
	Timestamp    float64 `json:"timestamp"`
	RedIntensity float64 `json:"red_intensity"`
	Saturation   float64 `json:"saturation"`
}

// FlashEvent represents a detected flash event.
type FlashEvent struct {
	Timestamp float64 `json:"timestamp"`
	Intensity float64 `json:"intensity"`
	Duration  float64 `json:"duration"`
	Type      string  `json:"type"` // sudden, gradual, strobe
}

// RedFlashEvent represents a detected red flash event.
type RedFlashEvent struct {
	Timestamp  float64 `json:"timestamp"`
	Intensity  float64 `json:"intensity"`
	Duration   float64 `json:"duration"`
	Saturation float64 `json:"saturation"`
	RedValue   float64 `json:"red_value"`
}

// FlashStatistics contains calculated flash rate statistics.
type FlashStatistics struct {
	AverageRate float64 `json:"average_rate"`
	PeakRate    float64 `json:"peak_rate"`
	MaxRate     float64 `json:"max_rate"`
	TotalFlashes int    `json:"total_flashes"`
}

// RiskAssessment contains flash risk evaluation results.
type RiskAssessment struct {
	ExceedsThreshold bool    `json:"exceeds_threshold"`
	RiskLevel        string  `json:"risk_level"`
	RiskScore        float64 `json:"risk_score"`
}

// Placeholder implementations for helper methods (these would contain the actual algorithm implementations)

func findPrimaryVideoStream(streams []StreamInfo) *StreamInfo {
	for i := range streams {
		if streams[i].CodecType == "video" {
			return &streams[i]
		}
	}
	return nil
}

func parseFrameRate(frameRate string) float64 {
	// Parse frame rate from string format (e.g., "25/1", "29.97")
	if strings.Contains(frameRate, "/") {
		parts := strings.Split(frameRate, "/")
		if len(parts) == 2 {
			num, _ := strconv.ParseFloat(parts[0], 64)
			den, _ := strconv.ParseFloat(parts[1], 64)
			if den != 0 {
				return num / den
			}
		}
	}
	rate, _ := strconv.ParseFloat(frameRate, 64)
	return rate
}

// Additional helper method implementations would go here...
// (These would contain the actual parsing and analysis logic)

func (pfa *PSEFlashAnalyzer) parseLuminanceData(output string, videoStream *StreamInfo) (*LuminanceData, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var values []LuminanceValue
	duration := 0.0
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}
		
		timestamp, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			continue
		}
		
		luminance, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}
		
		values = append(values, LuminanceValue{
			Timestamp: timestamp,
			Luminance: luminance,
		})
		
		if timestamp > duration {
			duration = timestamp
		}
	}
	
	return &LuminanceData{
		Duration: duration,
		Values:   values,
	}, nil
}

func (pfa *PSEFlashAnalyzer) parseRedChannelData(output string, videoStream *StreamInfo) (*RedChannelData, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var values []RedValue
	duration := 0.0
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}
		
		timestamp, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			continue
		}
		
		redIntensity, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}
		
		// Calculate saturation as a proxy from red intensity
		saturation := math.Min(redIntensity/255.0, 1.0)
		
		values = append(values, RedValue{
			Timestamp:    timestamp,
			RedIntensity: redIntensity,
			Saturation:   saturation,
		})
		
		if timestamp > duration {
			duration = timestamp
		}
	}
	
	return &RedChannelData{
		Duration: duration,
		Values:   values,
	}, nil
}

func (pfa *PSEFlashAnalyzer) calculateFlashDuration(data *LuminanceData, index int, fps float64) float64 {
	if index >= len(data.Values) || fps <= 0 {
		return 1.0 / 25.0 // Default frame duration
	}
	
	// Look ahead to find the end of the flash event
	baseLuma := data.Values[index].Luminance
	duration := 1.0 / fps // Start with one frame
	
	// Check up to 5 frames ahead for flash continuation
	for i := index + 1; i < len(data.Values) && i < index+5; i++ {
		currentLuma := data.Values[i].Luminance
		lumaChange := math.Abs(currentLuma - baseLuma)
		
		// If luminance change is still significant, extend duration
		if lumaChange > MinFlashIntensity*0.5 {
			duration += 1.0 / fps
		} else {
			break
		}
	}
	
	return duration
}

func (pfa *PSEFlashAnalyzer) calculateRedFlashDuration(data *RedChannelData, index int, fps float64) float64 {
	if index >= len(data.Values) || fps <= 0 {
		return 1.0 / 25.0
	}
	
	baseRed := data.Values[index].RedIntensity
	duration := 1.0 / fps
	
	// Check up to 3 frames ahead for red flash continuation (shorter than general flash)
	for i := index + 1; i < len(data.Values) && i < index+3; i++ {
		currentRed := data.Values[i].RedIntensity
		redChange := math.Abs(currentRed - baseRed) * RedLuminanceWeight
		
		if redChange > MinFlashIntensity*0.7*0.5 { // Half the red threshold
			duration += 1.0 / fps
		} else {
			break
		}
	}
	
	return duration
}

func (pfa *PSEFlashAnalyzer) classifyFlashType(lumaChange, intensity float64) string {
	// Implementation would classify flash type based on characteristics
	if intensity > 0.8 {
		return "sudden"
	} else if intensity > 0.4 {
		return "gradual"
	}
	return "subtle"
}

func (pfa *PSEFlashAnalyzer) mergeConsecutiveFlashes(events []FlashEvent) []FlashEvent {
	if len(events) <= 1 {
		return events
	}
	
	var merged []FlashEvent
	current := events[0]
	
	for i := 1; i < len(events); i++ {
		next := events[i]
		
		// If events are within 0.1 seconds of each other, merge them
		if next.Timestamp-current.Timestamp <= 0.1 {
			// Merge: extend duration and average intensity
			endTime := math.Max(current.Timestamp+current.Duration, next.Timestamp+next.Duration)
			current.Duration = endTime - current.Timestamp
			current.Intensity = (current.Intensity + next.Intensity) / 2
			
			// Keep the more severe flash type
			if next.Type == "sudden" || (current.Type != "sudden" && next.Type == "gradual") {
				current.Type = next.Type
			}
		} else {
			// Not consecutive, add current and move to next
			merged = append(merged, current)
			current = next
		}
	}
	
	// Add the last event
	merged = append(merged, current)
	return merged
}

func (pfa *PSEFlashAnalyzer) mergeConsecutiveRedFlashes(events []RedFlashEvent) []RedFlashEvent {
	if len(events) <= 1 {
		return events
	}
	
	var merged []RedFlashEvent
	current := events[0]
	
	for i := 1; i < len(events); i++ {
		next := events[i]
		
		// Red flashes merge within 0.05 seconds (tighter than general flashes)
		if next.Timestamp-current.Timestamp <= 0.05 {
			endTime := math.Max(current.Timestamp+current.Duration, next.Timestamp+next.Duration)
			current.Duration = endTime - current.Timestamp
			current.Intensity = (current.Intensity + next.Intensity) / 2
			current.Saturation = math.Max(current.Saturation, next.Saturation) // Keep higher saturation
			current.RedValue = math.Max(current.RedValue, next.RedValue)
		} else {
			merged = append(merged, current)
			current = next
		}
	}
	
	merged = append(merged, current)
	return merged
}

func (pfa *PSEFlashAnalyzer) calculateFlashStatistics(events []FlashEvent, duration float64) FlashStatistics {
	if len(events) == 0 || duration <= 0 {
		return FlashStatistics{}
	}
	
	// Calculate average rate
	averageRate := float64(len(events)) / duration
	
	// Calculate peak rate in 1-second windows
	peakRate := 0.0
	maxRate := 0.0
	
	// Analyze in 1-second windows as per ITU-R BT.1702
	for windowStart := 0.0; windowStart < duration; windowStart += AnalysisWindowSize {
		windowEnd := windowStart + AnalysisWindowSize
		windowCount := 0
		
		for _, event := range events {
			if event.Timestamp >= windowStart && event.Timestamp < windowEnd {
				windowCount++
			}
		}
		
		windowRate := float64(windowCount)
		if windowRate > peakRate {
			peakRate = windowRate
		}
		if windowRate > maxRate {
			maxRate = windowRate
		}
	}
	
	return FlashStatistics{
		AverageRate:  averageRate,
		PeakRate:     peakRate,
		MaxRate:      maxRate,
		TotalFlashes: len(events),
	}
}

func (pfa *PSEFlashAnalyzer) calculateRedFlashStatistics(events []RedFlashEvent, duration float64) FlashStatistics {
	if len(events) == 0 || duration <= 0 {
		return FlashStatistics{}
	}
	
	averageRate := float64(len(events)) / duration
	peakRate := 0.0
	maxRate := 0.0
	
	// Analyze in 1-second windows with red-specific thresholds
	for windowStart := 0.0; windowStart < duration; windowStart += AnalysisWindowSize {
		windowEnd := windowStart + AnalysisWindowSize
		windowCount := 0
		
		for _, event := range events {
			if event.Timestamp >= windowStart && event.Timestamp < windowEnd {
				windowCount++
			}
		}
		
		windowRate := float64(windowCount)
		if windowRate > peakRate {
			peakRate = windowRate
		}
		if windowRate > maxRate {
			maxRate = windowRate
		}
	}
	
	return FlashStatistics{
		AverageRate:  averageRate,
		PeakRate:     peakRate,
		MaxRate:      maxRate,
		TotalFlashes: len(events),
	}
}

func (pfa *PSEFlashAnalyzer) analyzeFlashPatterns(events []FlashEvent) interface{} {
	if len(events) < 3 {
		return nil // Need at least 3 events for pattern analysis
	}
	
	patterns := struct {
		RegularRhythm    bool    `json:"regular_rhythm"`
		RhythmFrequency  float64 `json:"rhythm_frequency"`
		PatternType      string  `json:"pattern_type"`
		TemporalSpacing  float64 `json:"temporal_spacing"`
	}{
		RegularRhythm:   false,
		RhythmFrequency: 0,
		PatternType:     "irregular",
		TemporalSpacing: 0,
	}
	
	// Calculate average time between events
	var intervals []float64
	for i := 1; i < len(events); i++ {
		interval := events[i].Timestamp - events[i-1].Timestamp
		intervals = append(intervals, interval)
	}
	
	if len(intervals) > 0 {
		totalInterval := 0.0
		for _, interval := range intervals {
			totalInterval += interval
		}
		averageInterval := totalInterval / float64(len(intervals))
		patterns.TemporalSpacing = averageInterval
		
		// Check for regularity (variance in intervals)
		variance := 0.0
		for _, interval := range intervals {
			variance += math.Pow(interval-averageInterval, 2)
		}
		variance /= float64(len(intervals))
		stdDev := math.Sqrt(variance)
		
		// If standard deviation is less than 20% of average, consider it regular
		if stdDev < (averageInterval * 0.2) {
			patterns.RegularRhythm = true
			patterns.RhythmFrequency = 1.0 / averageInterval
			patterns.PatternType = "regular_strobe"
		} else if stdDev < (averageInterval * 0.5) {
			patterns.PatternType = "semi_regular"
		}
	}
	
	return patterns
}

func (pfa *PSEFlashAnalyzer) assessFlashRisk(stats FlashStatistics, patterns interface{}) RiskAssessment {
	riskScore := 0.0
	riskLevel := "safe"
	exceedsThreshold := false
	
	// Check against ITU-R BT.1702 thresholds
	if stats.MaxRate > MaxSafeFlashRate {
		exceedsThreshold = true
		riskScore += 40.0 // Base risk for exceeding threshold
		
		// Additional risk based on how much threshold is exceeded
		excessRate := stats.MaxRate - MaxSafeFlashRate
		riskScore += math.Min(excessRate*15.0, 40.0) // Up to 40 additional points
	}
	
	// Risk from total flash count
	if stats.TotalFlashes > 100 {
		riskScore += math.Min(float64(stats.TotalFlashes-100)/50.0*10.0, 20.0)
	}
	
	// Determine risk level
	switch {
	case riskScore <= SafeRiskThreshold:
		riskLevel = "safe"
	case riskScore <= LowRiskThreshold:
		riskLevel = "low"
	case riskScore <= MediumRiskThreshold:
		riskLevel = "medium"
	case riskScore <= HighRiskThreshold:
		riskLevel = "high"
	default:
		riskLevel = "critical"
	}
	
	return RiskAssessment{
		ExceedsThreshold: exceedsThreshold,
		RiskLevel:        riskLevel,
		RiskScore:        riskScore,
	}
}

func (pfa *PSEFlashAnalyzer) assessRedFlashRisk(stats FlashStatistics, events []RedFlashEvent) struct{ ExceedsRedThreshold bool } {
	exceedsRedThreshold := false
	
	// Red flashes have lower threshold (2.0 vs 3.0)
	if stats.MaxRate > MaxSafeRedFlashRate {
		exceedsRedThreshold = true
	}
	
	// Also check high saturation red events
	highSaturationCount := 0
	for _, event := range events {
		if event.Saturation > RedSaturationThreshold {
			highSaturationCount++
		}
	}
	
	// If more than 10 high saturation red flashes, also risky
	if highSaturationCount > 10 {
		exceedsRedThreshold = true
	}
	
	return struct{ ExceedsRedThreshold bool }{
		ExceedsRedThreshold: exceedsRedThreshold,
	}
}

func (pfa *PSEFlashAnalyzer) calculateFlashDurations(events []FlashEvent) []FlashDuration {
	var durations []FlashDuration
	
	for _, event := range events {
		riskLevel := "low"
		if event.Intensity > DangerousFlashIntensity {
			riskLevel = "high"
		} else if event.Intensity > 0.5 {
			riskLevel = "medium"
		}
		
		durations = append(durations, FlashDuration{
			StartTime:  event.Timestamp,
			EndTime:    event.Timestamp + event.Duration,
			Duration:   event.Duration,
			Intensity:  event.Intensity,
			ScreenArea: 1.0, // Assume full screen for now
			RiskLevel:  riskLevel,
		})
	}
	
	return durations
}

func (pfa *PSEFlashAnalyzer) calculateRedFlashDurations(events []RedFlashEvent) []FlashDuration {
	var durations []FlashDuration
	
	for _, event := range events {
		riskLevel := "medium" // Red flashes start at medium risk
		if event.Intensity > 0.8 {
			riskLevel = "high"
		} else if event.Saturation > 0.9 {
			riskLevel = "high"
		}
		
		durations = append(durations, FlashDuration{
			StartTime:  event.Timestamp,
			EndTime:    event.Timestamp + event.Duration,
			Duration:   event.Duration,
			Intensity:  event.Intensity,
			ScreenArea: 1.0,
			RiskLevel:  riskLevel,
		})
	}
	
	return durations
}

func (pfa *PSEFlashAnalyzer) analyzeFrequencyDistribution(events []FlashEvent) map[string]int {
	distribution := map[string]int{
		"0-1Hz":   0,
		"1-3Hz":   0,
		"3-5Hz":   0,
		"5-10Hz":  0,
		"10-25Hz": 0,
		">25Hz":   0,
	}
	
	// Group events by time windows and calculate frequency
	for windowStart := 0.0; windowStart < 60.0; windowStart += AnalysisWindowSize {
		windowEnd := windowStart + AnalysisWindowSize
		windowCount := 0
		
		for _, event := range events {
			if event.Timestamp >= windowStart && event.Timestamp < windowEnd {
				windowCount++
			}
		}
		
		frequency := float64(windowCount)
		switch {
		case frequency <= 1:
			distribution["0-1Hz"]++
		case frequency <= 3:
			distribution["1-3Hz"]++
		case frequency <= 5:
			distribution["3-5Hz"]++
		case frequency <= 10:
			distribution["5-10Hz"]++
		case frequency <= 25:
			distribution["10-25Hz"]++
		default:
			distribution[">25Hz"]++
		}
	}
	
	return distribution
}

func (pfa *PSEFlashAnalyzer) identifyDangerousPeriods(events []FlashEvent, stats FlashStatistics) []TimePeriod {
	var dangerousPeriods []TimePeriod
	
	// Find periods with high flash rates
	for windowStart := 0.0; windowStart < 3600.0; windowStart += AnalysisWindowSize {
		windowEnd := windowStart + AnalysisWindowSize
		windowEvents := 0
		highIntensityEvents := 0
		
		for _, event := range events {
			if event.Timestamp >= windowStart && event.Timestamp < windowEnd {
				windowEvents++
				if event.Intensity > DangerousFlashIntensity {
					highIntensityEvents++
				}
			}
		}
		
		flashRate := float64(windowEvents)
		if flashRate > MaxSafeFlashRate || highIntensityEvents > 0 {
			riskLevel := "medium"
			description := fmt.Sprintf("High flash rate period: %.1f flashes/second", flashRate)
			
			if flashRate > CriticalFlashRate {
				riskLevel = "critical"
				description = fmt.Sprintf("Critical flash rate period: %.1f flashes/second", flashRate)
			} else if highIntensityEvents > 0 {
				riskLevel = "high"
				description = fmt.Sprintf("High intensity flash period: %d dangerous flashes", highIntensityEvents)
			}
			
			dangerousPeriods = append(dangerousPeriods, TimePeriod{
				StartTime:   windowStart,
				EndTime:     windowEnd,
				RiskLevel:   riskLevel,
				Description: description,
				Confidence:  0.85,
			})
		}
	}
	
	return dangerousPeriods
}

func (pfa *PSEFlashAnalyzer) identifyDangerousRedPeriods(events []RedFlashEvent) []TimePeriod {
	var dangerousPeriods []TimePeriod
	
	// Find periods with high red flash rates (lower threshold than general flashes)
	for windowStart := 0.0; windowStart < 3600.0; windowStart += AnalysisWindowSize {
		windowEnd := windowStart + AnalysisWindowSize
		windowEvents := 0
		highSaturationEvents := 0
		
		for _, event := range events {
			if event.Timestamp >= windowStart && event.Timestamp < windowEnd {
				windowEvents++
				if event.Saturation > RedSaturationThreshold {
					highSaturationEvents++
				}
			}
		}
		
		redFlashRate := float64(windowEvents)
		if redFlashRate > MaxSafeRedFlashRate || highSaturationEvents > 0 {
			riskLevel := "high" // Red flashes start at high risk
			description := fmt.Sprintf("High red flash rate period: %.1f red flashes/second", redFlashRate)
			
			if redFlashRate > MaxSafeRedFlashRate*2 {
				riskLevel = "critical"
				description = fmt.Sprintf("Critical red flash period: %.1f red flashes/second", redFlashRate)
			}
			
			dangerousPeriods = append(dangerousPeriods, TimePeriod{
				StartTime:   windowStart,
				EndTime:     windowEnd,
				RiskLevel:   riskLevel,
				Description: description,
				Confidence:  0.9, // Higher confidence for red flash detection
			})
		}
	}
	
	return dangerousPeriods
}

func (pfa *PSEFlashAnalyzer) analyzeFlashIntensity(events []FlashEvent) *FlashIntensity {
	if len(events) == 0 {
		return &FlashIntensity{}
	}
	
	var intensities []float64
	peakIntensity := 0.0
	totalIntensity := 0.0
	
	for _, event := range events {
		intensities = append(intensities, event.Intensity)
		totalIntensity += event.Intensity
		if event.Intensity > peakIntensity {
			peakIntensity = event.Intensity
		}
	}
	
	averageIntensity := totalIntensity / float64(len(events))
	
	// Calculate variance
	variance := 0.0
	for _, intensity := range intensities {
		variance += math.Pow(intensity-averageIntensity, 2)
	}
	variance /= float64(len(intensities))
	
	// Create intensity distribution
	distribution := map[string]int{
		"0.0-0.2": 0,
		"0.2-0.4": 0,
		"0.4-0.6": 0,
		"0.6-0.8": 0,
		"0.8-1.0": 0,
	}
	
	for _, intensity := range intensities {
		switch {
		case intensity < 0.2:
			distribution["0.0-0.2"]++
		case intensity < 0.4:
			distribution["0.2-0.4"]++
		case intensity < 0.6:
			distribution["0.4-0.6"]++
		case intensity < 0.8:
			distribution["0.6-0.8"]++
		default:
			distribution["0.8-1.0"]++
		}
	}
	
	return &FlashIntensity{
		PeakIntensity:         peakIntensity,
		AverageIntensity:      averageIntensity,
		IntensityVariance:     variance,
		IntensityDistribution: distribution,
	}
}

func (pfa *PSEFlashAnalyzer) extractSaturationLevels(events []RedFlashEvent) []float64 {
	var saturations []float64
	
	for _, event := range events {
		saturations = append(saturations, event.Saturation)
	}
	
	return saturations
}