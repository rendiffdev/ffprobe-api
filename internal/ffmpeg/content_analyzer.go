package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// ContentAnalyzer handles content-based quality analysis using FFmpeg filters
type ContentAnalyzer struct {
	ffmpegPath string
	logger     zerolog.Logger
	tempDir    string
	hdrAnalyzer *HDRAnalyzer
}

// NewContentAnalyzer creates a new content analyzer
func NewContentAnalyzer(ffmpegPath string, logger zerolog.Logger) *ContentAnalyzer {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}

	return &ContentAnalyzer{
		ffmpegPath: ffmpegPath,
		logger:     logger,
		tempDir:    "/tmp/content_analysis",
		hdrAnalyzer: NewHDRAnalyzer("ffprobe", logger),
	}
}

// AnalyzeContent performs content-based analysis on a video file
func (ca *ContentAnalyzer) AnalyzeContent(ctx context.Context, filePath string) (*ContentAnalysis, error) {
	analysis := &ContentAnalysis{}

	// Run analyses in parallel for efficiency
	resultChan := make(chan func())
	errorChan := make(chan error, 8)

	// Launch blackness detection
	go func() {
		if result, err := ca.analyzeBlackFrames(ctx, filePath); err != nil {
			errorChan <- fmt.Errorf("blackness analysis failed: %w", err)
		} else {
			resultChan <- func() { analysis.BlackFrames = result }
		}
	}()

	// Launch freeze frame detection
	go func() {
		if result, err := ca.analyzeFreezeFrames(ctx, filePath); err != nil {
			errorChan <- fmt.Errorf("freeze frame analysis failed: %w", err)
		} else {
			resultChan <- func() { analysis.FreezeFrames = result }
		}
	}()

	// Launch audio clipping detection
	go func() {
		if result, err := ca.analyzeAudioClipping(ctx, filePath); err != nil {
			errorChan <- fmt.Errorf("audio clipping analysis failed: %w", err)
		} else {
			resultChan <- func() { analysis.AudioClipping = result }
		}
	}()

	// Launch blockiness detection
	go func() {
		if result, err := ca.analyzeBlockiness(ctx, filePath); err != nil {
			errorChan <- fmt.Errorf("blockiness analysis failed: %w", err)
		} else {
			resultChan <- func() { analysis.Blockiness = result }
		}
	}()

	// Launch blur detection
	go func() {
		if result, err := ca.analyzeBlurriness(ctx, filePath); err != nil {
			errorChan <- fmt.Errorf("blurriness analysis failed: %w", err)
		} else {
			resultChan <- func() { analysis.Blurriness = result }
		}
	}()

	// Launch interlace detection
	go func() {
		if result, err := ca.analyzeInterlacing(ctx, filePath); err != nil {
			errorChan <- fmt.Errorf("interlace analysis failed: %w", err)
		} else {
			resultChan <- func() { analysis.InterlaceInfo = result }
		}
	}()

	// Launch noise detection
	go func() {
		if result, err := ca.analyzeNoise(ctx, filePath); err != nil {
			errorChan <- fmt.Errorf("noise analysis failed: %w", err)
		} else {
			resultChan <- func() { analysis.NoiseLevel = result }
		}
	}()

	// Launch loudness measurement
	go func() {
		if result, err := ca.analyzeLoudness(ctx, filePath); err != nil {
			errorChan <- fmt.Errorf("loudness analysis failed: %w", err)
		} else {
			resultChan <- func() { analysis.LoudnessMeter = result }
		}
	}()

	// Launch HDR analysis
	go func() {
		if result, err := ca.hdrAnalyzer.AnalyzeHDR(ctx, filePath); err != nil {
			errorChan <- fmt.Errorf("HDR analysis failed: %w", err)
		} else {
			resultChan <- func() { analysis.HDRAnalysis = result }
		}
	}()

	// Collect results with timeout
	completed := 0
	target := 9
	timeout := time.After(60 * time.Second) // 60 second timeout for all analyses

	for completed < target {
		select {
		case applyResult := <-resultChan:
			applyResult()
			completed++
		case err := <-errorChan:
			ca.logger.Warn().Err(err).Msg("Content analysis error")
			completed++
		case <-timeout:
			ca.logger.Warn().Msg("Content analysis timed out")
			return analysis, nil
		case <-ctx.Done():
			return analysis, ctx.Err()
		}
	}

	return analysis, nil
}

// analyzeBlackFrames detects black or nearly black frames
func (ca *ContentAnalyzer) analyzeBlackFrames(ctx context.Context, filePath string) (*BlackFrameAnalysis, error) {
	threshold := 0.1 // 10% threshold for blackness

	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", fmt.Sprintf("blackdetect=d=0.5:pix_th=%f", threshold),
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("blackdetect failed: %w", err)
	}

	// Parse output for black frame detections
	lines := strings.Split(string(output), "\n")
	detectedFrames := 0

	for _, line := range lines {
		if strings.Contains(line, "blackdetect") && strings.Contains(line, "black_start") {
			detectedFrames++
		}
	}

	return &BlackFrameAnalysis{
		DetectedFrames: detectedFrames,
		Percentage:     0.0, // Would need total frame count to calculate
		Threshold:      threshold,
	}, nil
}

// analyzeFreezeFrames detects static/frozen frames
func (ca *ContentAnalyzer) analyzeFreezeFrames(ctx context.Context, filePath string) (*FreezeFrameAnalysis, error) {
	threshold := 0.001 // Very low threshold for freeze detection

	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", fmt.Sprintf("freezedetect=n=%f:d=2", threshold),
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("freezedetect failed: %w", err)
	}

	// Parse output for freeze detections
	lines := strings.Split(string(output), "\n")
	detectedFrames := 0

	for _, line := range lines {
		if strings.Contains(line, "freezedetect") && strings.Contains(line, "freeze_start") {
			detectedFrames++
		}
	}

	return &FreezeFrameAnalysis{
		DetectedFrames: detectedFrames,
		Percentage:     0.0, // Would need total frame count to calculate
		Threshold:      threshold,
	}, nil
}

// analyzeAudioClipping detects audio clipping
func (ca *ContentAnalyzer) analyzeAudioClipping(ctx context.Context, filePath string) (*AudioClippingAnalysis, error) {
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-af", "astats=metadata=1:reset=1",
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("audio clipping analysis failed: %w", err)
	}

	// Parse output for peak levels
	lines := strings.Split(string(output), "\n")
	var peakLevel float64 = -96.0 // Default very low level

	for _, line := range lines {
		if strings.Contains(line, "Peak level") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				if level, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
					if level > peakLevel {
						peakLevel = level
					}
				}
			}
		}
	}

	// Determine if clipping occurred (above -1dB is likely clipping)
	clippedSamples := 0
	if peakLevel > -1.0 {
		clippedSamples = 1 // Simplified detection
	}

	return &AudioClippingAnalysis{
		ClippedSamples: clippedSamples,
		Percentage:     0.0, // Would need total sample count
		PeakLevel:      peakLevel,
	}, nil
}

// analyzeBlockiness measures compression blockiness
func (ca *ContentAnalyzer) analyzeBlockiness(ctx context.Context, filePath string) (*BlockinessAnalysis, error) {
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", "blockdetect",
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("blockiness analysis failed: %w", err)
	}

	// Parse blockdetect output
	lines := strings.Split(string(output), "\n")
	var totalBlockiness float64
	measurements := 0

	for _, line := range lines {
		if strings.Contains(line, "blockdetect") {
			// Extract blockiness value if available
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "avg:") {
					if val, err := strconv.ParseFloat(strings.TrimPrefix(part, "avg:"), 64); err == nil {
						totalBlockiness += val
						measurements++
					}
				}
			}
		}
	}

	avgBlockiness := 0.0
	if measurements > 0 {
		avgBlockiness = totalBlockiness / float64(measurements)
	}

	return &BlockinessAnalysis{
		AverageBlockiness: avgBlockiness,
		MaxBlockiness:     avgBlockiness, // Simplified
		Threshold:         0.1,
	}, nil
}

// analyzeBlurriness measures image sharpness
func (ca *ContentAnalyzer) analyzeBlurriness(ctx context.Context, filePath string) (*BlurrinessAnalysis, error) {
	// Use a simple edge detection approach for blur measurement
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", "convolution='0 -1 0:-1 5 -1:0 -1 0:0 -1 0:-1 5 -1:0 -1 0',signalstats",
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("blurriness analysis failed: %w", err)
	}

	// Parse signalstats output for edge information
	lines := strings.Split(string(output), "\n")
	var totalSharpness float64
	measurements := 0

	for _, line := range lines {
		if strings.Contains(line, "YAVG") {
			// Extract average Y value as sharpness metric
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "YAVG:") {
					if val, err := strconv.ParseFloat(strings.TrimPrefix(part, "YAVG:"), 64); err == nil {
						totalSharpness += val
						measurements++
					}
				}
			}
		}
	}

	avgSharpness := 0.0
	if measurements > 0 {
		avgSharpness = totalSharpness / float64(measurements)
	}

	blurThreshold := 50.0 // Threshold for blur detection
	blurDetected := avgSharpness < blurThreshold

	return &BlurrinessAnalysis{
		AverageSharpness: avgSharpness,
		MinSharpness:     avgSharpness, // Simplified
		BlurDetected:     blurDetected,
	}, nil
}

// analyzeInterlacing detects interlacing artifacts
func (ca *ContentAnalyzer) analyzeInterlacing(ctx context.Context, filePath string) (*InterlaceAnalysis, error) {
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", "idet",
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("interlace detection failed: %w", err)
	}

	// Parse idet output
	lines := strings.Split(string(output), "\n")
	var progressiveFrames, interlacedFrames int
	var confidence float64

	for _, line := range lines {
		if strings.Contains(line, "Multi frame detection:") {
			// Parse detection results
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "Progressive:" && i+1 < len(parts) {
					if val, err := strconv.Atoi(parts[i+1]); err == nil {
						progressiveFrames = val
					}
				}
				if part == "Interlaced:" && i+1 < len(parts) {
					if val, err := strconv.Atoi(parts[i+1]); err == nil {
						interlacedFrames = val
					}
				}
			}
		}
	}

	totalFrames := progressiveFrames + interlacedFrames
	interlaceDetected := interlacedFrames > progressiveFrames

	if totalFrames > 0 {
		confidence = float64(interlacedFrames) / float64(totalFrames)
	}

	return &InterlaceAnalysis{
		InterlaceDetected: interlaceDetected,
		ProgressiveFrames: progressiveFrames,
		InterlacedFrames:  interlacedFrames,
		Confidence:        confidence,
	}, nil
}

// analyzeNoise measures video noise levels
func (ca *ContentAnalyzer) analyzeNoise(ctx context.Context, filePath string) (*NoiseAnalysis, error) {
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", "signalstats",
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("noise analysis failed: %w", err)
	}

	// Parse signalstats for noise indicators
	lines := strings.Split(string(output), "\n")
	var totalNoise float64
	measurements := 0

	for _, line := range lines {
		if strings.Contains(line, "YDIF") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "YDIF:") {
					if val, err := strconv.ParseFloat(strings.TrimPrefix(part, "YDIF:"), 64); err == nil {
						totalNoise += val
						measurements++
					}
				}
			}
		}
	}

	avgNoise := 0.0
	if measurements > 0 {
		avgNoise = totalNoise / float64(measurements)
	}

	return &NoiseAnalysis{
		AverageNoise: avgNoise,
		MaxNoise:     avgNoise, // Simplified
		NoiseProfile: "detected",
	}, nil
}

// analyzeLoudness provides broadcast loudness compliance
func (ca *ContentAnalyzer) analyzeLoudness(ctx context.Context, filePath string) (*LoudnessAnalysis, error) {
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-af", "ebur128=metadata=1",
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("loudness analysis failed: %w", err)
	}

	// Parse EBU R128 output
	lines := strings.Split(string(output), "\n")
	var integratedLoudness, loudnessRange, truePeak float64

	for _, line := range lines {
		if strings.Contains(line, "Integrated loudness:") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "I:" && i+1 < len(parts) {
					val := strings.TrimSuffix(parts[i+1], " LUFS")
					if lufs, err := strconv.ParseFloat(val, 64); err == nil {
						integratedLoudness = lufs
					}
				}
			}
		}
		if strings.Contains(line, "Loudness range:") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "LRA:" && i+1 < len(parts) {
					val := strings.TrimSuffix(parts[i+1], " LU")
					if lu, err := strconv.ParseFloat(val, 64); err == nil {
						loudnessRange = lu
					}
				}
			}
		}
		if strings.Contains(line, "True peak:") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "Peak:" && i+1 < len(parts) {
					val := strings.TrimSuffix(parts[i+1], " dBTP")
					if dbtp, err := strconv.ParseFloat(val, 64); err == nil {
						truePeak = dbtp
					}
				}
			}
		}
	}

	// Check compliance with broadcast standards (EBU R128)
	compliant := integratedLoudness >= -25.0 && integratedLoudness <= -21.0 && truePeak <= -1.0

	return &LoudnessAnalysis{
		IntegratedLoudness: integratedLoudness,
		LoudnessRange:      loudnessRange,
		TruePeak:           truePeak,
		Compliant:          compliant,
		Standard:           "EBU R128",
	}, nil
}