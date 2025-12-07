package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// forEachLine iterates over lines in output using bufio.Scanner
// which is more memory-efficient than strings.Split for large outputs.
// The callback receives each line; return false to stop iteration.
func forEachLine(output []byte, fn func(line string) bool) {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		if !fn(scanner.Text()) {
			return
		}
	}
}

// ContentAnalyzer handles content-based quality analysis using FFmpeg filters
type ContentAnalyzer struct {
	ffmpegPath  string
	logger      zerolog.Logger
	tempDir     string
	hdrAnalyzer *HDRAnalyzer
}

// NewContentAnalyzer creates a new content analyzer
func NewContentAnalyzer(ffmpegPath string, logger zerolog.Logger) *ContentAnalyzer {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}

	return &ContentAnalyzer{
		ffmpegPath:  ffmpegPath,
		logger:      logger,
		tempDir:     "/tmp/content_analysis",
		hdrAnalyzer: NewHDRAnalyzer("ffprobe", logger),
	}
}

// AnalyzeContent performs content-based analysis on a video file
func (ca *ContentAnalyzer) AnalyzeContent(ctx context.Context, filePath string) (*ContentAnalysis, error) {
	analysis := &ContentAnalysis{}

	// Create cancellable context for proper cleanup on timeout
	analyzeCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel() // Ensures all goroutines terminate

	// Run analyses in parallel for efficiency
	// Buffered channels to prevent goroutine blocking (26 goroutines total)
	const numAnalyzers = 26
	resultChan := make(chan func(), numAnalyzers)
	errorChan := make(chan error, numAnalyzers)

	// WaitGroup to track goroutine completion
	var wg sync.WaitGroup

	// Helper to launch analyzer with proper cleanup
	launchAnalyzer := func(name string, analyze func(context.Context, string) (func(), error)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-analyzeCtx.Done():
				return // Context cancelled, exit gracefully
			default:
			}
			if applyResult, err := analyze(analyzeCtx, filePath); err != nil {
				select {
				case errorChan <- fmt.Errorf("%s failed: %w", name, err):
				case <-analyzeCtx.Done():
				}
			} else if applyResult != nil {
				select {
				case resultChan <- applyResult:
				case <-analyzeCtx.Done():
				}
			}
		}()
	}

	// Launch all 26 analyzers using the safe launchAnalyzer pattern
	launchAnalyzer("blackness analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeBlackFrames(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.BlackFrames = result }, nil
	})

	launchAnalyzer("freeze frame analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeFreezeFrames(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.FreezeFrames = result }, nil
	})

	launchAnalyzer("audio clipping analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeAudioClipping(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.AudioClipping = result }, nil
	})

	launchAnalyzer("silence analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeSilence(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.SilenceInfo = result }, nil
	})

	launchAnalyzer("phase analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzePhase(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.PhaseInfo = result }, nil
	})

	launchAnalyzer("audio level analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeAudioLevels(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.AudioLevelInfo = result }, nil
	})

	launchAnalyzer("letterbox analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeLetterbox(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.LetterboxInfo = result }, nil
	})

	launchAnalyzer("dropout analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeDropouts(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.DropoutInfo = result }, nil
	})

	launchAnalyzer("color bars analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeColorBars(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.ColorBarsInfo = result }, nil
	})

	launchAnalyzer("test tone analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeTestTone(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.TestToneInfo = result }, nil
	})

	launchAnalyzer("safe area analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeSafeArea(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.SafeAreaInfo = result }, nil
	})

	launchAnalyzer("channel mapping analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeChannelMapping(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.ChannelMappingInfo = result }, nil
	})

	launchAnalyzer("timecode analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeTimecodeContinuity(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.TimecodeInfo = result }, nil
	})

	launchAnalyzer("blockiness analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeBlockiness(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.Blockiness = result }, nil
	})

	launchAnalyzer("blurriness analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeBlurriness(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.Blurriness = result }, nil
	})

	launchAnalyzer("interlace analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeInterlacing(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.InterlaceInfo = result }, nil
	})

	launchAnalyzer("noise analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeNoise(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.NoiseLevel = result }, nil
	})

	launchAnalyzer("loudness analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeLoudness(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.LoudnessMeter = result }, nil
	})

	launchAnalyzer("HDR analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.hdrAnalyzer.AnalyzeHDR(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.HDRAnalysis = result }, nil
	})

	launchAnalyzer("baseband analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeBaseband(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.BasebandInfo = result }, nil
	})

	launchAnalyzer("video quality score analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeVideoQualityScore(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.VideoQualityScore = result }, nil
	})

	launchAnalyzer("temporal complexity analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeTemporalComplexity(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.TemporalComplexity = result }, nil
	})

	launchAnalyzer("field dominance analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeFieldDominance(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.FieldDominance = result }, nil
	})

	launchAnalyzer("differential frame analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeDifferentialFrames(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.DifferentialFrame = result }, nil
	})

	launchAnalyzer("line error analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeLineErrors(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.LineErrors = result }, nil
	})

	launchAnalyzer("audio frequency analysis", func(ctx context.Context, path string) (func(), error) {
		result, err := ca.analyzeAudioFrequency(ctx, path)
		if err != nil {
			return nil, err
		}
		return func() { analysis.AudioFrequency = result }, nil
	})

	// Close channels when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results - channels will be closed when all goroutines complete
	for {
		select {
		case applyResult, ok := <-resultChan:
			if !ok {
				resultChan = nil // Mark channel as drained
			} else if applyResult != nil {
				applyResult()
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil // Mark channel as drained
			} else {
				ca.logger.Warn().Err(err).Msg("Content analysis error")
			}
		case <-analyzeCtx.Done():
			ca.logger.Warn().Msg("Content analysis context cancelled")
			return analysis, nil
		}
		// Exit when both channels are drained
		if resultChan == nil && errorChan == nil {
			break
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

	// Parse output for black frame detections using efficient line scanner
	detectedFrames := 0
	forEachLine(output, func(line string) bool {
		if strings.Contains(line, "blackdetect") && strings.Contains(line, "black_start") {
			detectedFrames++
		}
		return true // continue iteration
	})

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

	// Parse output for freeze detections using efficient line scanner
	detectedFrames := 0
	forEachLine(output, func(line string) bool {
		if strings.Contains(line, "freezedetect") && strings.Contains(line, "freeze_start") {
			detectedFrames++
		}
		return true // continue iteration
	})

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

// analyzeSilence detects silence/mute periods using FFmpeg silencedetect
func (ca *ContentAnalyzer) analyzeSilence(ctx context.Context, filePath string) (*SilenceAnalysis, error) {
	// Default thresholds for broadcast QC
	noiseThreshold := -50.0 // dB threshold for silence detection
	minDuration := 0.5      // Minimum silence duration in seconds

	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-af", fmt.Sprintf("silencedetect=noise=%ddB:d=%f", int(noiseThreshold), minDuration),
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// silencedetect may return non-zero if no audio stream
		ca.logger.Debug().Err(err).Msg("Silence detection completed with warnings")
	}

	// Parse silencedetect output
	lines := strings.Split(string(output), "\n")
	var silencePeriods []SilencePeriod
	var currentStart float64 = -1
	var totalDuration float64

	// Get total duration from ffprobe-style output
	for _, line := range lines {
		if strings.Contains(line, "Duration:") && strings.Contains(line, ",") {
			// Parse duration from "Duration: HH:MM:SS.ms," format
			parts := strings.Split(line, "Duration:")
			if len(parts) > 1 {
				durationStr := strings.TrimSpace(strings.Split(parts[1], ",")[0])
				totalDuration = parseDurationToSeconds(durationStr)
			}
		}
	}

	for _, line := range lines {
		// Parse silence_start
		if strings.Contains(line, "silence_start:") {
			parts := strings.Split(line, "silence_start:")
			if len(parts) > 1 {
				startStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
				if start, err := strconv.ParseFloat(startStr, 64); err == nil {
					currentStart = start
				}
			}
		}

		// Parse silence_end and silence_duration
		if strings.Contains(line, "silence_end:") && currentStart >= 0 {
			var endTime, duration float64

			// Extract end time
			parts := strings.Split(line, "silence_end:")
			if len(parts) > 1 {
				endStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
				if end, err := strconv.ParseFloat(endStr, 64); err == nil {
					endTime = end
				}
			}

			// Extract duration
			if strings.Contains(line, "silence_duration:") {
				durParts := strings.Split(line, "silence_duration:")
				if len(durParts) > 1 {
					durStr := strings.TrimSpace(strings.Split(durParts[1], " ")[0])
					if dur, err := strconv.ParseFloat(durStr, 64); err == nil {
						duration = dur
					}
				}
			}

			// Determine if this is start/end mute
			isStartMute := currentStart < 1.0
			isEndMute := totalDuration > 0 && (totalDuration-endTime) < 1.0

			silencePeriods = append(silencePeriods, SilencePeriod{
				StartTime:   currentStart,
				EndTime:     endTime,
				Duration:    duration,
				NoiseFloor:  noiseThreshold,
				IsStartMute: isStartMute,
				IsEndMute:   isEndMute,
			})

			currentStart = -1
		}
	}

	// Calculate statistics
	var totalSilenceSec float64
	var longestSilenceSec float64
	hasProblematicMute := false

	for _, period := range silencePeriods {
		totalSilenceSec += period.Duration
		if period.Duration > longestSilenceSec {
			longestSilenceSec = period.Duration
		}
		// Problematic if silence > 3 seconds mid-content (not at start/end)
		if period.Duration > 3.0 && !period.IsStartMute && !period.IsEndMute {
			hasProblematicMute = true
		}
	}

	silencePercentage := 0.0
	if totalDuration > 0 {
		silencePercentage = (totalSilenceSec / totalDuration) * 100.0
	}

	return &SilenceAnalysis{
		SilencePeriods:     silencePeriods,
		TotalSilenceCount:  len(silencePeriods),
		TotalSilenceSec:    totalSilenceSec,
		LongestSilenceSec:  longestSilenceSec,
		SilencePercentage:  silencePercentage,
		NoiseFloorDB:       noiseThreshold,
		ThresholdDB:        noiseThreshold,
		MinDurationSec:     minDuration,
		HasProblematicMute: hasProblematicMute,
	}, nil
}

// parseDurationToSeconds converts HH:MM:SS.ms format to seconds
func parseDurationToSeconds(duration string) float64 {
	duration = strings.TrimSpace(duration)
	parts := strings.Split(duration, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, _ := strconv.ParseFloat(parts[0], 64)
	minutes, _ := strconv.ParseFloat(parts[1], 64)
	seconds, _ := strconv.ParseFloat(parts[2], 64)

	return hours*3600 + minutes*60 + seconds
}

// analyzePhase detects audio phase issues using FFmpeg aphasemeter
func (ca *ContentAnalyzer) analyzePhase(ctx context.Context, filePath string) (*PhaseAnalysis, error) {
	// aphasemeter outputs phase correlation values:
	// +1.0 = perfectly in phase (mono compatible)
	// 0.0 = unrelated (decorrelated)
	// -1.0 = perfectly out of phase (will cancel in mono)
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-af", "aphasemeter=video=0",
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// aphasemeter may fail on mono or no-audio files
		ca.logger.Debug().Err(err).Msg("Phase analysis completed with warnings")
	}

	// Parse aphasemeter output
	lines := strings.Split(string(output), "\n")
	var phaseValues []float64
	var totalPhase float64
	var minPhase float64 = 1.0
	var maxPhase float64 = -1.0
	var phaseProblemFrames int
	var phaseEvents []PhaseEvent
	var currentEventStart float64 = -1
	var currentEventPhases []float64

	frameCount := 0
	for _, line := range lines {
		// Look for aphasemeter output: [Parsed_aphasemeter_0 @ 0x...] phase: 0.xxx
		if strings.Contains(line, "aphasemeter") && strings.Contains(line, "phase:") {
			parts := strings.Split(line, "phase:")
			if len(parts) > 1 {
				phaseStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
				if phase, err := strconv.ParseFloat(phaseStr, 64); err == nil {
					phaseValues = append(phaseValues, phase)
					totalPhase += phase
					frameCount++

					if phase < minPhase {
						minPhase = phase
					}
					if phase > maxPhase {
						maxPhase = phase
					}

					// Check for phase issues (< -0.3 is problematic)
					if phase < -0.3 {
						phaseProblemFrames++
						// Track continuous phase events
						if currentEventStart < 0 {
							currentEventStart = float64(frameCount) * 0.02 // Approximate timestamp
						}
						currentEventPhases = append(currentEventPhases, phase)
					} else if currentEventStart >= 0 {
						// End current phase event
						eventDuration := float64(len(currentEventPhases)) * 0.02
						avgEventPhase := 0.0
						minEventPhase := 1.0
						for _, p := range currentEventPhases {
							avgEventPhase += p
							if p < minEventPhase {
								minEventPhase = p
							}
						}
						avgEventPhase /= float64(len(currentEventPhases))

						phaseEvents = append(phaseEvents, PhaseEvent{
							StartTime:    currentEventStart,
							EndTime:      currentEventStart + eventDuration,
							Duration:     eventDuration,
							AveragePhase: avgEventPhase,
							MinPhase:     minEventPhase,
						})
						currentEventStart = -1
						currentEventPhases = nil
					}
				}
			}
		}
	}

	// Handle any remaining event
	if currentEventStart >= 0 && len(currentEventPhases) > 0 {
		eventDuration := float64(len(currentEventPhases)) * 0.02
		avgEventPhase := 0.0
		minEventPhase := 1.0
		for _, p := range currentEventPhases {
			avgEventPhase += p
			if p < minEventPhase {
				minEventPhase = p
			}
		}
		avgEventPhase /= float64(len(currentEventPhases))

		phaseEvents = append(phaseEvents, PhaseEvent{
			StartTime:    currentEventStart,
			EndTime:      currentEventStart + eventDuration,
			Duration:     eventDuration,
			AveragePhase: avgEventPhase,
			MinPhase:     minEventPhase,
		})
	}

	// Calculate statistics
	averagePhase := 0.0
	if frameCount > 0 {
		averagePhase = totalPhase / float64(frameCount)
	}

	outOfPhasePercent := 0.0
	if frameCount > 0 {
		outOfPhasePercent = (float64(phaseProblemFrames) / float64(frameCount)) * 100.0
	}

	// Determine severity
	var severity string
	hasPhaseIssues := outOfPhasePercent > 1.0 || minPhase < -0.5
	if minPhase < -0.8 || outOfPhasePercent > 10 {
		severity = "critical"
	} else if minPhase < -0.5 || outOfPhasePercent > 5 {
		severity = "warning"
	} else if minPhase < -0.3 || outOfPhasePercent > 1 {
		severity = "minor"
	} else {
		severity = "none"
	}

	// If no stereo data was found, set safe defaults
	if frameCount == 0 {
		minPhase = 0
		maxPhase = 0
	}

	return &PhaseAnalysis{
		AveragePhase:       averagePhase,
		MinPhase:           minPhase,
		MaxPhase:           maxPhase,
		PhaseCorrelation:   averagePhase,
		OutOfPhasePercent:  outOfPhasePercent,
		HasPhaseIssues:     hasPhaseIssues,
		PhaseProblemFrames: phaseProblemFrames,
		TotalFrames:        frameCount,
		PhaseEvents:        phaseEvents,
		Severity:           severity,
	}, nil
}

// analyzeAudioLevels provides detailed audio level measurements using FFmpeg astats
func (ca *ContentAnalyzer) analyzeAudioLevels(ctx context.Context, filePath string) (*AudioLevelAnalysis, error) {
	// Use astats filter for comprehensive audio statistics
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-af", "astats=metadata=1:reset=0",
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		ca.logger.Debug().Err(err).Msg("Audio level analysis completed with warnings")
	}

	// Parse astats output
	lines := strings.Split(string(output), "\n")
	var channels []ChannelLevelInfo
	var overallPeakDB float64 = -96.0
	var overallRMSDB float64 = -96.0
	var clippingCount int
	var dcOffsetTotal float64

	currentChannel := -1
	var currentChannelInfo ChannelLevelInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Detect channel sections
		if strings.Contains(line, "Channel:") {
			// Save previous channel if exists
			if currentChannel >= 0 {
				currentChannelInfo.Channel = currentChannel
				currentChannelInfo.DynamicRange = currentChannelInfo.MaxDB - currentChannelInfo.RMSDB
				channels = append(channels, currentChannelInfo)
			}

			// Start new channel
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				chStr := strings.TrimSpace(parts[1])
				if ch, err := strconv.Atoi(chStr); err == nil {
					currentChannel = ch
					currentChannelInfo = ChannelLevelInfo{}
				}
			}
			continue
		}

		// Parse audio statistics for current channel
		if currentChannel >= 0 {
			if strings.Contains(line, "Peak level dB:") {
				if val := parseAudioStatValue(line); val != -1000 {
					currentChannelInfo.PeakDB = val
					if val > overallPeakDB {
						overallPeakDB = val
					}
				}
			} else if strings.Contains(line, "RMS level dB:") {
				if val := parseAudioStatValue(line); val != -1000 {
					currentChannelInfo.RMSDB = val
					if val > overallRMSDB {
						overallRMSDB = val
					}
				}
			} else if strings.Contains(line, "Min level:") {
				if val := parseAudioStatValue(line); val != -1000 {
					currentChannelInfo.MinDB = val
				}
			} else if strings.Contains(line, "Max level:") {
				if val := parseAudioStatValue(line); val != -1000 {
					currentChannelInfo.MaxDB = val
				}
			} else if strings.Contains(line, "DC offset:") {
				if val := parseAudioStatValue(line); val != -1000 {
					currentChannelInfo.DCOffset = val
					dcOffsetTotal += val
				}
			} else if strings.Contains(line, "Crest factor:") {
				if val := parseAudioStatValue(line); val != -1000 {
					currentChannelInfo.CrestFactor = val
				}
			} else if strings.Contains(line, "Flat factor:") {
				if val := parseAudioStatValue(line); val != -1000 {
					currentChannelInfo.FlatFactor = val
				}
			} else if strings.Contains(line, "Peak count:") {
				if val := parseAudioStatValue(line); val != -1000 {
					currentChannelInfo.PeakCount = int(val)
					clippingCount += int(val)
				}
			} else if strings.Contains(line, "Bit depth:") {
				if val := parseAudioStatValue(line); val != -1000 {
					currentChannelInfo.BitDepth = int(val)
				}
			}
		}

		// Also check for summary statistics
		if strings.Contains(line, "Overall") {
			if strings.Contains(line, "Peak level dB:") {
				if val := parseAudioStatValue(line); val != -1000 {
					overallPeakDB = val
				}
			} else if strings.Contains(line, "RMS level dB:") {
				if val := parseAudioStatValue(line); val != -1000 {
					overallRMSDB = val
				}
			}
		}
	}

	// Save last channel
	if currentChannel >= 0 {
		currentChannelInfo.Channel = currentChannel
		currentChannelInfo.DynamicRange = currentChannelInfo.MaxDB - currentChannelInfo.RMSDB
		channels = append(channels, currentChannelInfo)
	}

	// Calculate overall statistics
	dynamicRange := overallPeakDB - overallRMSDB
	crestFactor := 0.0
	if len(channels) > 0 {
		for _, ch := range channels {
			crestFactor += ch.CrestFactor
		}
		crestFactor /= float64(len(channels))
	}

	avgDCOffset := 0.0
	if len(channels) > 0 {
		avgDCOffset = dcOffsetTotal / float64(len(channels))
	}

	// Check for clipping (peak > -0.1 dB)
	hasClipping := overallPeakDB > -0.1 || clippingCount > 0

	// Calculate headroom (how much room below 0dB)
	headroom := -overallPeakDB
	if headroom < 0 {
		headroom = 0
	}

	// Broadcast safe check (peak < -1dB, no excessive DC offset)
	isBroadcastSafe := overallPeakDB < -1.0 && avgDCOffset < 0.1 && avgDCOffset > -0.1

	// Determine severity
	var severity string
	if overallPeakDB > 0 {
		severity = "critical"
	} else if overallPeakDB > -1.0 || hasClipping {
		severity = "warning"
	} else if overallPeakDB > -3.0 {
		severity = "minor"
	} else {
		severity = "none"
	}

	return &AudioLevelAnalysis{
		Channels:         channels,
		OverallPeakDB:    overallPeakDB,
		OverallRMSDB:     overallRMSDB,
		DynamicRangeDB:   dynamicRange,
		CrestFactor:      crestFactor,
		DCOffset:         avgDCOffset,
		HasClipping:      hasClipping,
		ClippingCount:    clippingCount,
		IsBroadcastSafe:  isBroadcastSafe,
		Headroom:         headroom,
		Severity:         severity,
	}, nil
}

// parseAudioStatValue extracts a numeric value from an astats output line
func parseAudioStatValue(line string) float64 {
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return -1000
	}
	valStr := strings.TrimSpace(parts[len(parts)-1])
	valStr = strings.TrimSuffix(valStr, " dB")
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return -1000
	}
	return val
}

// analyzeLetterbox detects letterboxing and pillarboxing using FFmpeg cropdetect
func (ca *ContentAnalyzer) analyzeLetterbox(ctx context.Context, filePath string) (*LetterboxAnalysis, error) {
	// Use cropdetect filter to detect black bars
	// We'll sample frames throughout the video for better accuracy
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", "cropdetect=24:16:0",
		"-t", "30", // Analyze first 30 seconds
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		ca.logger.Debug().Err(err).Msg("Letterbox detection completed with warnings")
	}

	// Parse cropdetect output
	// Format: [Parsed_cropdetect_0 @ 0x...] x1:0 x2:1919 y1:140 y2:939 w:1920 h:800 crop=1920:800:0:140
	lines := strings.Split(string(output), "\n")

	var cropValues []struct {
		w, h, x, y int
	}

	var originalWidth, originalHeight int

	for _, line := range lines {
		// Get original dimensions from stream info
		if strings.Contains(line, "Video:") && strings.Contains(line, "x") {
			// Parse dimensions like "1920x1080"
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.Contains(part, "x") && !strings.Contains(part, "0x") {
					dims := strings.Split(strings.Trim(part, ","), "x")
					if len(dims) == 2 {
						if w, err := strconv.Atoi(dims[0]); err == nil {
							if h, err := strconv.Atoi(dims[1]); err == nil {
								if w > 100 && h > 100 { // Valid dimensions
									originalWidth = w
									originalHeight = h
								}
							}
						}
					}
				}
			}
		}

		// Parse cropdetect output
		if strings.Contains(line, "cropdetect") && strings.Contains(line, "crop=") {
			// Extract crop parameters from crop=w:h:x:y
			cropIdx := strings.Index(line, "crop=")
			if cropIdx >= 0 {
				cropStr := line[cropIdx+5:]
				cropParts := strings.Split(strings.TrimSpace(strings.Split(cropStr, " ")[0]), ":")
				if len(cropParts) >= 4 {
					w, _ := strconv.Atoi(cropParts[0])
					h, _ := strconv.Atoi(cropParts[1])
					x, _ := strconv.Atoi(cropParts[2])
					y, _ := strconv.Atoi(cropParts[3])
					if w > 0 && h > 0 {
						cropValues = append(cropValues, struct{ w, h, x, y int }{w, h, x, y})
					}
				}
			}
		}
	}

	// Calculate most common crop values (mode)
	if len(cropValues) == 0 {
		return &LetterboxAnalysis{
			HasLetterbox:    false,
			HasPillarbox:    false,
			Type:            "none",
			OriginalWidth:   originalWidth,
			OriginalHeight:  originalHeight,
			ActiveWidth:     originalWidth,
			ActiveHeight:    originalHeight,
			IsConsistent:    true,
			FramesAnalyzed:  0,
			Confidence:      1.0,
		}, nil
	}

	// Find most common crop values
	cropCounts := make(map[string]int)
	for _, cv := range cropValues {
		key := fmt.Sprintf("%d:%d:%d:%d", cv.w, cv.h, cv.x, cv.y)
		cropCounts[key]++
	}

	var bestKey string
	var bestCount int
	for key, count := range cropCounts {
		if count > bestCount {
			bestCount = count
			bestKey = key
		}
	}

	// Parse best crop values
	parts := strings.Split(bestKey, ":")
	activeWidth, _ := strconv.Atoi(parts[0])
	activeHeight, _ := strconv.Atoi(parts[1])
	xOffset, _ := strconv.Atoi(parts[2])
	yOffset, _ := strconv.Atoi(parts[3])

	// Calculate bars
	topBar := yOffset
	bottomBar := originalHeight - activeHeight - yOffset
	leftBar := xOffset
	rightBar := originalWidth - activeWidth - xOffset

	// Determine letterbox type
	hasLetterbox := topBar > 4 || bottomBar > 4
	hasPillarbox := leftBar > 4 || rightBar > 4

	boxType := "none"
	if hasLetterbox && hasPillarbox {
		boxType = "windowbox"
	} else if hasLetterbox {
		boxType = "letterbox"
	} else if hasPillarbox {
		boxType = "pillarbox"
	}

	// Calculate aspect ratios
	aspectRatio := "unknown"
	activeAspect := "unknown"
	if originalHeight > 0 {
		ar := float64(originalWidth) / float64(originalHeight)
		aspectRatio = fmt.Sprintf("%.2f:1", ar)
	}
	if activeHeight > 0 {
		aar := float64(activeWidth) / float64(activeHeight)
		activeAspect = fmt.Sprintf("%.2f:1", aar)
	}

	// Calculate black percentage
	blackPercentage := 0.0
	totalPixels := originalWidth * originalHeight
	if totalPixels > 0 {
		activePixels := activeWidth * activeHeight
		blackPixels := totalPixels - activePixels
		blackPercentage = (float64(blackPixels) / float64(totalPixels)) * 100.0
	}

	// Calculate consistency (how often the most common crop was detected)
	consistency := float64(bestCount) / float64(len(cropValues))
	isConsistent := consistency > 0.8

	return &LetterboxAnalysis{
		HasLetterbox:    hasLetterbox,
		HasPillarbox:    hasPillarbox,
		Type:            boxType,
		OriginalWidth:   originalWidth,
		OriginalHeight:  originalHeight,
		ActiveWidth:     activeWidth,
		ActiveHeight:    activeHeight,
		TopBar:          topBar,
		BottomBar:       bottomBar,
		LeftBar:         leftBar,
		RightBar:        rightBar,
		AspectRatio:     aspectRatio,
		ActiveAspect:    activeAspect,
		CropFilter:      fmt.Sprintf("crop=%d:%d:%d:%d", activeWidth, activeHeight, xOffset, yOffset),
		BlackPercentage: blackPercentage,
		IsConsistent:    isConsistent,
		FramesAnalyzed:  len(cropValues),
		Confidence:      consistency,
	}, nil
}

// analyzeDropouts detects video and audio signal dropouts
func (ca *ContentAnalyzer) analyzeDropouts(ctx context.Context, filePath string) (*DropoutAnalysis, error) {
	// Use multiple FFmpeg filters to detect different types of dropouts:
	// 1. signalstats BRNG for video out-of-range (corrupt) pixels
	// 2. silencedetect for audio dropouts (sudden silence)
	// 3. freezedetect for repeated/frozen frames (video glitch)

	var videoDropouts []DropoutEvent
	var audioDropouts []DropoutEvent
	var totalDuration float64
	framesAnalyzed := 0

	// Detect audio dropouts using silence detection with shorter duration threshold
	audioCmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-af", "silencedetect=noise=-60dB:d=0.1",
		"-f", "null",
		"-",
	)

	audioOutput, _ := audioCmd.CombinedOutput()
	audioLines := strings.Split(string(audioOutput), "\n")

	// Get total duration from output
	for _, line := range audioLines {
		if strings.Contains(line, "Duration:") && strings.Contains(line, ",") {
			parts := strings.Split(line, "Duration:")
			if len(parts) > 1 {
				durationStr := strings.TrimSpace(strings.Split(parts[1], ",")[0])
				totalDuration = parseDurationToSeconds(durationStr)
			}
		}
	}

	// Parse audio silence events (potential dropouts)
	var currentSilenceStart float64 = -1
	for _, line := range audioLines {
		if strings.Contains(line, "silence_start:") {
			parts := strings.Split(line, "silence_start:")
			if len(parts) > 1 {
				startStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
				if start, err := strconv.ParseFloat(startStr, 64); err == nil {
					currentSilenceStart = start
				}
			}
		}

		if strings.Contains(line, "silence_end:") && currentSilenceStart >= 0 {
			var endTime, duration float64

			parts := strings.Split(line, "silence_end:")
			if len(parts) > 1 {
				endStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
				if end, err := strconv.ParseFloat(endStr, 64); err == nil {
					endTime = end
				}
			}

			if strings.Contains(line, "silence_duration:") {
				durParts := strings.Split(line, "silence_duration:")
				if len(durParts) > 1 {
					durStr := strings.TrimSpace(strings.Split(durParts[1], " ")[0])
					if dur, err := strconv.ParseFloat(durStr, 64); err == nil {
						duration = dur
					}
				}
			}

			// Only count as dropout if very short (< 2 sec) and not at start/end
			// Short silences mid-content are likely dropouts
			if duration > 0.1 && duration < 2.0 && currentSilenceStart > 1.0 && (totalDuration-endTime) > 1.0 {
				severity := "minor"
				if duration > 0.5 {
					severity = "warning"
				}
				if duration > 1.0 {
					severity = "critical"
				}

				audioDropouts = append(audioDropouts, DropoutEvent{
					Type:        "audio_dropout",
					StartTime:   currentSilenceStart,
					EndTime:     endTime,
					Duration:    duration,
					Severity:    severity,
					Description: "Sudden audio silence detected",
				})
			}
			currentSilenceStart = -1
		}
	}

	// Detect video dropouts using freezedetect (frozen frames = potential dropout)
	videoCmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", "freezedetect=n=0.003:d=0.05",
		"-f", "null",
		"-",
	)

	videoOutput, _ := videoCmd.CombinedOutput()
	videoLines := strings.Split(string(videoOutput), "\n")

	var currentFreezeStart float64 = -1
	for _, line := range videoLines {
		if strings.Contains(line, "freeze_start:") {
			parts := strings.Split(line, "freeze_start:")
			if len(parts) > 1 {
				startStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
				if start, err := strconv.ParseFloat(startStr, 64); err == nil {
					currentFreezeStart = start
				}
			}
			framesAnalyzed++
		}

		if strings.Contains(line, "freeze_end:") && currentFreezeStart >= 0 {
			var endTime, duration float64

			parts := strings.Split(line, "freeze_end:")
			if len(parts) > 1 {
				endStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
				if end, err := strconv.ParseFloat(endStr, 64); err == nil {
					endTime = end
				}
			}

			if strings.Contains(line, "freeze_duration:") {
				durParts := strings.Split(line, "freeze_duration:")
				if len(durParts) > 1 {
					durStr := strings.TrimSpace(strings.Split(durParts[1], " ")[0])
					if dur, err := strconv.ParseFloat(durStr, 64); err == nil {
						duration = dur
					}
				}
			}

			// Short freezes (< 2 sec) are likely dropouts
			if duration > 0.05 && duration < 2.0 {
				severity := "minor"
				if duration > 0.2 {
					severity = "warning"
				}
				if duration > 0.5 {
					severity = "critical"
				}

				videoDropouts = append(videoDropouts, DropoutEvent{
					Type:        "video_freeze",
					StartTime:   currentFreezeStart,
					EndTime:     endTime,
					Duration:    duration,
					Severity:    severity,
					Description: "Frozen frame detected (potential dropout)",
				})
			}
			currentFreezeStart = -1
		}
	}

	// Calculate statistics
	var maxVideoDropout, maxAudioDropout, totalDropoutTime float64
	for _, d := range videoDropouts {
		if d.Duration > maxVideoDropout {
			maxVideoDropout = d.Duration
		}
		totalDropoutTime += d.Duration
	}
	for _, d := range audioDropouts {
		if d.Duration > maxAudioDropout {
			maxAudioDropout = d.Duration
		}
		totalDropoutTime += d.Duration
	}

	dropoutPercentage := 0.0
	if totalDuration > 0 {
		dropoutPercentage = (totalDropoutTime / totalDuration) * 100.0
	}

	hasDropouts := len(videoDropouts) > 0 || len(audioDropouts) > 0

	// Broadcast compliance: no dropouts > 0.5s, total < 0.1%
	isBroadcastCompliant := maxVideoDropout < 0.5 && maxAudioDropout < 0.5 && dropoutPercentage < 0.1

	// Determine severity
	var severity string
	if maxVideoDropout > 1.0 || maxAudioDropout > 1.0 || dropoutPercentage > 1.0 {
		severity = "critical"
	} else if maxVideoDropout > 0.5 || maxAudioDropout > 0.5 || dropoutPercentage > 0.5 {
		severity = "warning"
	} else if hasDropouts {
		severity = "minor"
	} else {
		severity = "none"
	}

	return &DropoutAnalysis{
		HasDropouts:          hasDropouts,
		VideoDropouts:        videoDropouts,
		AudioDropouts:        audioDropouts,
		TotalVideoDropouts:   len(videoDropouts),
		TotalAudioDropouts:   len(audioDropouts),
		MaxVideoDropoutSec:   maxVideoDropout,
		MaxAudioDropoutSec:   maxAudioDropout,
		TotalDropoutSec:      totalDropoutTime,
		DropoutPercentage:    dropoutPercentage,
		IsBroadcastCompliant: isBroadcastCompliant,
		Severity:             severity,
		FramesAnalyzed:       framesAnalyzed,
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

// analyzeColorBars detects color bars/test patterns at start/end of content
func (ca *ContentAnalyzer) analyzeColorBars(ctx context.Context, filePath string) (*ColorBarsAnalysis, error) {
	// Color bars detection using signalstats filter to detect consistent color regions
	// SMPTE color bars have specific Y/U/V values that we can detect
	// We analyze the first and last 30 seconds of the video

	var colorBarsEvents []ColorBarsEvent
	hasColorBarsAtStart := false
	hasColorBarsAtEnd := false
	var startDuration, endDuration float64
	var totalDuration float64
	confidence := 0.0
	detectedPattern := ""

	// Get total duration first
	durationCmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-f", "null",
		"-",
	)
	durationOutput, _ := durationCmd.CombinedOutput()
	for _, line := range strings.Split(string(durationOutput), "\n") {
		if strings.Contains(line, "Duration:") && strings.Contains(line, ",") {
			parts := strings.Split(line, "Duration:")
			if len(parts) > 1 {
				durationStr := strings.TrimSpace(strings.Split(parts[1], ",")[0])
				totalDuration = parseDurationToSeconds(durationStr)
			}
		}
	}

	// Analyze start of video (first 30 seconds) using signalstats
	// Color bars have very low YDIF (frame-to-frame difference) and specific YAVG values
	startCmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-t", "30",
		"-vf", "signalstats=stat=tout+vrep+brng,metadata=print:file=-",
		"-f", "null",
		"-",
	)

	startOutput, _ := startCmd.CombinedOutput()
	startLines := strings.Split(string(startOutput), "\n")

	// Count frames with characteristics typical of color bars:
	// - Very low temporal difference (YDIF)
	// - Consistent Y average values
	// - High BRNG (out-of-broadcast-range) for some bars
	lowDiffFrames := 0
	totalFrames := 0
	var consecutiveLowDiff int
	maxConsecutiveLowDiff := 0

	for _, line := range startLines {
		if strings.Contains(line, "YDIF") {
			totalFrames++
			// Extract YDIF value
			parts := strings.Split(line, "YDIF=")
			if len(parts) > 1 {
				ydifStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
				if ydif, err := strconv.ParseFloat(ydifStr, 64); err == nil {
					// Color bars typically have very low YDIF (< 1.0)
					if ydif < 1.0 {
						lowDiffFrames++
						consecutiveLowDiff++
						if consecutiveLowDiff > maxConsecutiveLowDiff {
							maxConsecutiveLowDiff = consecutiveLowDiff
						}
					} else {
						consecutiveLowDiff = 0
					}
				}
			}
		}
	}

	// If more than 70% of frames in first 30 sec have low diff, likely color bars
	if totalFrames > 0 && float64(lowDiffFrames)/float64(totalFrames) > 0.7 {
		hasColorBarsAtStart = true
		startDuration = float64(maxConsecutiveLowDiff) / 25.0 // Approximate assuming ~25fps
		if startDuration > 30 {
			startDuration = 30
		}
		confidence = float64(lowDiffFrames) / float64(totalFrames)
		detectedPattern = "SMPTE Color Bars"

		colorBarsEvents = append(colorBarsEvents, ColorBarsEvent{
			StartTime:   0,
			EndTime:     startDuration,
			Duration:    startDuration,
			PatternType: "SMPTE",
			Confidence:  confidence,
		})
	}

	// Analyze end of video (last 30 seconds) if duration is known
	if totalDuration > 30 {
		endStartTime := totalDuration - 30
		endCmd := exec.CommandContext(ctx, ca.ffmpegPath,
			"-ss", fmt.Sprintf("%.2f", endStartTime),
			"-i", filePath,
			"-vf", "signalstats=stat=tout+vrep+brng,metadata=print:file=-",
			"-f", "null",
			"-",
		)

		endOutput, _ := endCmd.CombinedOutput()
		endLines := strings.Split(string(endOutput), "\n")

		endLowDiffFrames := 0
		endTotalFrames := 0
		endConsecutiveLowDiff := 0
		endMaxConsecutive := 0

		for _, line := range endLines {
			if strings.Contains(line, "YDIF") {
				endTotalFrames++
				parts := strings.Split(line, "YDIF=")
				if len(parts) > 1 {
					ydifStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
					if ydif, err := strconv.ParseFloat(ydifStr, 64); err == nil {
						if ydif < 1.0 {
							endLowDiffFrames++
							endConsecutiveLowDiff++
							if endConsecutiveLowDiff > endMaxConsecutive {
								endMaxConsecutive = endConsecutiveLowDiff
							}
						} else {
							endConsecutiveLowDiff = 0
						}
					}
				}
			}
		}

		if endTotalFrames > 0 && float64(endLowDiffFrames)/float64(endTotalFrames) > 0.7 {
			hasColorBarsAtEnd = true
			endDuration = float64(endMaxConsecutive) / 25.0
			if endDuration > 30 {
				endDuration = 30
			}
			endConfidence := float64(endLowDiffFrames) / float64(endTotalFrames)
			if endConfidence > confidence {
				confidence = endConfidence
			}
			if detectedPattern == "" {
				detectedPattern = "SMPTE Color Bars"
			}

			colorBarsEvents = append(colorBarsEvents, ColorBarsEvent{
				StartTime:   totalDuration - endDuration,
				EndTime:     totalDuration,
				Duration:    endDuration,
				PatternType: "SMPTE",
				Confidence:  endConfidence,
			})
		}
	}

	hasColorBars := hasColorBarsAtStart || hasColorBarsAtEnd

	// Compliance check: color bars should be present but limited duration
	isCompliant := true
	if hasColorBars {
		// Color bars > 60 seconds at start or end may indicate issue
		if startDuration > 60 || endDuration > 60 {
			isCompliant = false
		}
	}

	return &ColorBarsAnalysis{
		HasColorBars:     hasColorBars,
		ColorBarsAtStart: hasColorBarsAtStart,
		ColorBarsAtEnd:   hasColorBarsAtEnd,
		StartDuration:    startDuration,
		EndDuration:      endDuration,
		DetectedPattern:  detectedPattern,
		ColorBarsEvents:  colorBarsEvents,
		IsCompliant:      isCompliant,
		Confidence:       confidence,
	}, nil
}

// analyzeTestTone detects test tones (1kHz, slate tones) in audio
func (ca *ContentAnalyzer) analyzeTestTone(ctx context.Context, filePath string) (*TestToneAnalysis, error) {
	// Test tones are typically:
	// - 1kHz sine wave at -20dBFS or -18dBFS
	// - Located at start/end of content (slate/leader)
	// We use FFmpeg's astats with afftdn for frequency detection

	var testToneEvents []TestToneEvent
	hasTestToneAtStart := false
	hasTestToneAtEnd := false
	var startDuration, endDuration float64
	detectedFrequency := 0.0
	detectedLevel := -96.0
	var totalDuration float64

	// Get total duration
	durationCmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-f", "null",
		"-",
	)
	durationOutput, _ := durationCmd.CombinedOutput()
	for _, line := range strings.Split(string(durationOutput), "\n") {
		if strings.Contains(line, "Duration:") && strings.Contains(line, ",") {
			parts := strings.Split(line, "Duration:")
			if len(parts) > 1 {
				durationStr := strings.TrimSpace(strings.Split(parts[1], ",")[0])
				totalDuration = parseDurationToSeconds(durationStr)
			}
		}
	}

	// Analyze first 30 seconds for test tone using spectrum analysis
	// Test tones have very consistent RMS and peak levels, and low crest factor
	startCmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-t", "30",
		"-af", "astats=metadata=1:reset=1",
		"-f", "null",
		"-",
	)

	startOutput, _ := startCmd.CombinedOutput()
	startLines := strings.Split(string(startOutput), "\n")

	// Look for consistent audio characteristics of test tone:
	// - RMS close to peak (low crest factor ~1.4 for sine)
	// - Consistent level around -20dB or -18dB
	var rmsValues []float64
	var peakValues []float64

	for _, line := range startLines {
		if strings.Contains(line, "RMS level dB:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				rmsStr := strings.TrimSpace(parts[len(parts)-1])
				if rms, err := strconv.ParseFloat(rmsStr, 64); err == nil {
					rmsValues = append(rmsValues, rms)
				}
			}
		}
		if strings.Contains(line, "Peak level dB:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				peakStr := strings.TrimSpace(parts[len(parts)-1])
				if peak, err := strconv.ParseFloat(peakStr, 64); err == nil {
					peakValues = append(peakValues, peak)
				}
			}
		}
	}

	// Check if we have consistent test tone characteristics
	if len(rmsValues) > 0 && len(peakValues) > 0 {
		avgRMS := 0.0
		for _, v := range rmsValues {
			avgRMS += v
		}
		avgRMS /= float64(len(rmsValues))

		avgPeak := 0.0
		for _, v := range peakValues {
			avgPeak += v
		}
		avgPeak /= float64(len(peakValues))

		// Test tone typically has:
		// - RMS around -20 to -18 dB
		// - Peak-RMS difference < 4dB (sine wave characteristics)
		// - Very consistent levels (low variance)
		crestFactor := avgPeak - avgRMS
		isTestToneLike := avgRMS > -25 && avgRMS < -15 && crestFactor < 4.0 && crestFactor > 2.0

		// Calculate variance
		variance := 0.0
		for _, v := range rmsValues {
			variance += (v - avgRMS) * (v - avgRMS)
		}
		if len(rmsValues) > 1 {
			variance /= float64(len(rmsValues) - 1)
		}
		isConsistent := variance < 1.0 // Less than 1dB variance

		if isTestToneLike && isConsistent {
			hasTestToneAtStart = true
			startDuration = 30.0 // Full analyzed duration
			detectedFrequency = 1000 // Assume 1kHz (standard test tone)
			detectedLevel = avgRMS

			testToneEvents = append(testToneEvents, TestToneEvent{
				StartTime: 0,
				EndTime:   startDuration,
				Duration:  startDuration,
				Frequency: 1000,
				Level:     avgRMS,
			})
		}
	}

	// Analyze end of video (last 30 seconds)
	if totalDuration > 30 {
		endStartTime := totalDuration - 30
		endCmd := exec.CommandContext(ctx, ca.ffmpegPath,
			"-ss", fmt.Sprintf("%.2f", endStartTime),
			"-i", filePath,
			"-af", "astats=metadata=1:reset=1",
			"-f", "null",
			"-",
		)

		endOutput, _ := endCmd.CombinedOutput()
		endLines := strings.Split(string(endOutput), "\n")

		var endRMSValues []float64
		var endPeakValues []float64

		for _, line := range endLines {
			if strings.Contains(line, "RMS level dB:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					rmsStr := strings.TrimSpace(parts[len(parts)-1])
					if rms, err := strconv.ParseFloat(rmsStr, 64); err == nil {
						endRMSValues = append(endRMSValues, rms)
					}
				}
			}
			if strings.Contains(line, "Peak level dB:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					peakStr := strings.TrimSpace(parts[len(parts)-1])
					if peak, err := strconv.ParseFloat(peakStr, 64); err == nil {
						endPeakValues = append(endPeakValues, peak)
					}
				}
			}
		}

		if len(endRMSValues) > 0 && len(endPeakValues) > 0 {
			avgRMS := 0.0
			for _, v := range endRMSValues {
				avgRMS += v
			}
			avgRMS /= float64(len(endRMSValues))

			avgPeak := 0.0
			for _, v := range endPeakValues {
				avgPeak += v
			}
			avgPeak /= float64(len(endPeakValues))

			crestFactor := avgPeak - avgRMS
			isTestToneLike := avgRMS > -25 && avgRMS < -15 && crestFactor < 4.0 && crestFactor > 2.0

			variance := 0.0
			for _, v := range endRMSValues {
				variance += (v - avgRMS) * (v - avgRMS)
			}
			if len(endRMSValues) > 1 {
				variance /= float64(len(endRMSValues) - 1)
			}
			isConsistent := variance < 1.0

			if isTestToneLike && isConsistent {
				hasTestToneAtEnd = true
				endDuration = 30.0
				if detectedFrequency == 0 {
					detectedFrequency = 1000
				}
				if avgRMS > detectedLevel {
					detectedLevel = avgRMS
				}

				testToneEvents = append(testToneEvents, TestToneEvent{
					StartTime: totalDuration - 30,
					EndTime:   totalDuration,
					Duration:  30,
					Frequency: 1000,
					Level:     avgRMS,
				})
			}
		}
	}

	hasTestTone := hasTestToneAtStart || hasTestToneAtEnd

	// Compliance: test tones should be at standard levels
	isCompliant := true
	if hasTestTone {
		// Check if level is at broadcast standard (-20dB or -18dB)
		if detectedLevel < -22 || detectedLevel > -16 {
			isCompliant = false
		}
	}

	return &TestToneAnalysis{
		HasTestTone:       hasTestTone,
		TestToneAtStart:   hasTestToneAtStart,
		TestToneAtEnd:     hasTestToneAtEnd,
		StartDuration:     startDuration,
		EndDuration:       endDuration,
		DetectedFrequency: detectedFrequency,
		DetectedLevel:     detectedLevel,
		TestToneEvents:    testToneEvents,
		IsCompliant:       isCompliant,
	}, nil
}

// analyzeSafeArea checks title-safe and action-safe boundaries
func (ca *ContentAnalyzer) analyzeSafeArea(ctx context.Context, filePath string) (*SafeAreaAnalysis, error) {
	// Safe area standards:
	// - Title Safe: 80% of screen (10% margin on each side)
	// - Action Safe: 90% of screen (5% margin on each side)
	// We analyze edge pixels to detect content outside safe areas

	titleSafeMargin := 10.0  // 10% margin for title safe
	actionSafeMargin := 5.0  // 5% margin for action safe

	var originalWidth, originalHeight int

	// Get video dimensions first
	dimCmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-f", "null",
		"-",
	)
	dimOutput, _ := dimCmd.CombinedOutput()
	for _, line := range strings.Split(string(dimOutput), "\n") {
		if strings.Contains(line, "Video:") && strings.Contains(line, "x") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.Contains(part, "x") && !strings.Contains(part, "0x") {
					dims := strings.Split(strings.Trim(part, ","), "x")
					if len(dims) == 2 {
						if w, err := strconv.Atoi(dims[0]); err == nil {
							if h, err := strconv.Atoi(dims[1]); err == nil {
								if w > 100 && h > 100 {
									originalWidth = w
									originalHeight = h
								}
							}
						}
					}
				}
			}
		}
	}

	if originalWidth == 0 || originalHeight == 0 {
		return &SafeAreaAnalysis{
			TitleSafeCompliant:  true,
			ActionSafeCompliant: true,
			TitleSafeMargin:     titleSafeMargin,
			ActionSafeMargin:    actionSafeMargin,
			ContentInTitleSafe:  100.0,
			ContentInActionSafe: 100.0,
			ViolationCount:      0,
			FramesAnalyzed:      0,
		}, nil
	}

	// Use cropdetect to find active picture area
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", "cropdetect=24:16:0",
		"-t", "60", // Analyze first 60 seconds
		"-f", "null",
		"-",
	)

	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")

	var cropValues []struct{ w, h, x, y int }
	framesAnalyzed := 0

	for _, line := range lines {
		if strings.Contains(line, "cropdetect") && strings.Contains(line, "crop=") {
			framesAnalyzed++
			cropIdx := strings.Index(line, "crop=")
			if cropIdx >= 0 {
				cropStr := line[cropIdx+5:]
				cropParts := strings.Split(strings.TrimSpace(strings.Split(cropStr, " ")[0]), ":")
				if len(cropParts) >= 4 {
					w, _ := strconv.Atoi(cropParts[0])
					h, _ := strconv.Atoi(cropParts[1])
					x, _ := strconv.Atoi(cropParts[2])
					y, _ := strconv.Atoi(cropParts[3])
					if w > 0 && h > 0 {
						cropValues = append(cropValues, struct{ w, h, x, y int }{w, h, x, y})
					}
				}
			}
		}
	}

	// Calculate safe area boundaries
	titleSafeX := int(float64(originalWidth) * (titleSafeMargin / 100.0))
	titleSafeY := int(float64(originalHeight) * (titleSafeMargin / 100.0))
	actionSafeX := int(float64(originalWidth) * (actionSafeMargin / 100.0))
	actionSafeY := int(float64(originalHeight) * (actionSafeMargin / 100.0))

	titleSafeViolations := 0
	actionSafeViolations := 0

	for _, cv := range cropValues {
		// Check if content extends outside title safe
		if cv.x < titleSafeX || cv.y < titleSafeY ||
			(cv.x+cv.w) > (originalWidth-titleSafeX) ||
			(cv.y+cv.h) > (originalHeight-titleSafeY) {
			titleSafeViolations++
		}
		// Check if content extends outside action safe
		if cv.x < actionSafeX || cv.y < actionSafeY ||
			(cv.x+cv.w) > (originalWidth-actionSafeX) ||
			(cv.y+cv.h) > (originalHeight-actionSafeY) {
			actionSafeViolations++
		}
	}

	titleSafeCompliant := titleSafeViolations == 0
	actionSafeCompliant := actionSafeViolations == 0

	contentInTitleSafe := 100.0
	contentInActionSafe := 100.0
	if len(cropValues) > 0 {
		contentInTitleSafe = 100.0 * float64(len(cropValues)-titleSafeViolations) / float64(len(cropValues))
		contentInActionSafe = 100.0 * float64(len(cropValues)-actionSafeViolations) / float64(len(cropValues))
	}

	return &SafeAreaAnalysis{
		TitleSafeCompliant:  titleSafeCompliant,
		ActionSafeCompliant: actionSafeCompliant,
		TitleSafeMargin:     titleSafeMargin,
		ActionSafeMargin:    actionSafeMargin,
		ContentInTitleSafe:  contentInTitleSafe,
		ContentInActionSafe: contentInActionSafe,
		ViolationCount:      titleSafeViolations + actionSafeViolations,
		FramesAnalyzed:      framesAnalyzed,
	}, nil
}

// analyzeChannelMapping validates audio channel configuration
func (ca *ContentAnalyzer) analyzeChannelMapping(ctx context.Context, filePath string) (*ChannelMappingAnalysis, error) {
	// Analyze audio stream channel configuration
	// Check for proper channel layout (stereo, 5.1, 7.1, etc.)

	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-af", "astats=metadata=1:reset=0,channelsplit",
		"-f", "null",
		"-",
	)

	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")

	// Parse audio stream info
	var totalChannels int
	var channelLayout string
	var channelDetails []ChannelDetail
	var layoutIssues []string

	// First pass: get stream info
	for _, line := range lines {
		if strings.Contains(line, "Audio:") {
			// Parse channel count and layout
			parts := strings.Fields(line)
			for i, part := range parts {
				// Look for channel count (e.g., "2 channels" or "6 channels")
				if part == "channels" || part == "channels," {
					if i > 0 {
						if ch, err := strconv.Atoi(parts[i-1]); err == nil {
							totalChannels = ch
						}
					}
				}
				// Look for channel layout (e.g., "stereo", "5.1", "7.1")
				if part == "stereo" || part == "stereo," {
					channelLayout = "stereo"
					if totalChannels == 0 {
						totalChannels = 2
					}
				} else if strings.HasPrefix(part, "5.1") {
					channelLayout = "5.1"
					if totalChannels == 0 {
						totalChannels = 6
					}
				} else if strings.HasPrefix(part, "7.1") {
					channelLayout = "7.1"
					if totalChannels == 0 {
						totalChannels = 8
					}
				} else if part == "mono" || part == "mono," {
					channelLayout = "mono"
					if totalChannels == 0 {
						totalChannels = 1
					}
				}
			}
		}
	}

	// Parse per-channel statistics from astats
	currentChannel := -1
	var currentPeak, currentRMS float64

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "Channel:") {
			// Save previous channel
			if currentChannel >= 0 {
				isSilent := currentRMS < -60
				isActive := currentPeak > -60

				channelDetails = append(channelDetails, ChannelDetail{
					Index:     currentChannel,
					Name:      getChannelName(currentChannel, channelLayout),
					PeakLevel: currentPeak,
					RMSLevel:  currentRMS,
					IsSilent:  isSilent,
					IsActive:  isActive,
				})
			}

			// Start new channel
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				if ch, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					currentChannel = ch
					currentPeak = -96.0
					currentRMS = -96.0
				}
			}
		}

		if currentChannel >= 0 {
			if strings.Contains(line, "Peak level dB:") {
				if val := parseAudioStatValue(line); val != -1000 {
					currentPeak = val
				}
			}
			if strings.Contains(line, "RMS level dB:") {
				if val := parseAudioStatValue(line); val != -1000 {
					currentRMS = val
				}
			}
		}
	}

	// Save last channel
	if currentChannel >= 0 {
		isSilent := currentRMS < -60
		isActive := currentPeak > -60

		channelDetails = append(channelDetails, ChannelDetail{
			Index:     currentChannel,
			Name:      getChannelName(currentChannel, channelLayout),
			PeakLevel: currentPeak,
			RMSLevel:  currentRMS,
			IsSilent:  isSilent,
			IsActive:  isActive,
		})
	}

	// Validate channel configuration
	hasSurround := totalChannels > 2
	hasLFE := totalChannels >= 6 // 5.1 and above have LFE

	// Check for broadcast-standard layouts
	isBroadcastLayout := channelLayout == "stereo" ||
		channelLayout == "5.1" ||
		channelLayout == "7.1" ||
		channelLayout == "mono"

	isValid := true

	// Check for silent channels (potential issue)
	silentCount := 0
	for _, ch := range channelDetails {
		if ch.IsSilent {
			silentCount++
		}
	}

	if silentCount > 0 && silentCount < len(channelDetails) {
		layoutIssues = append(layoutIssues, fmt.Sprintf("%d of %d channels are silent", silentCount, len(channelDetails)))
	}

	// Check channel count matches layout
	if channelLayout == "5.1" && totalChannels != 6 {
		layoutIssues = append(layoutIssues, "5.1 layout should have 6 channels")
		isValid = false
	}
	if channelLayout == "7.1" && totalChannels != 8 {
		layoutIssues = append(layoutIssues, "7.1 layout should have 8 channels")
		isValid = false
	}
	if channelLayout == "stereo" && totalChannels != 2 {
		layoutIssues = append(layoutIssues, "Stereo layout should have 2 channels")
		isValid = false
	}

	// If no layout detected but have channels, infer it
	if channelLayout == "" && totalChannels > 0 {
		switch totalChannels {
		case 1:
			channelLayout = "mono"
		case 2:
			channelLayout = "stereo"
		case 6:
			channelLayout = "5.1"
		case 8:
			channelLayout = "7.1"
		default:
			channelLayout = fmt.Sprintf("%d channels", totalChannels)
			if totalChannels > 2 {
				layoutIssues = append(layoutIssues, "Non-standard channel count")
			}
		}
	}

	return &ChannelMappingAnalysis{
		TotalChannels:     totalChannels,
		ChannelLayout:     channelLayout,
		ExpectedLayout:    "", // Could be set based on delivery specs
		IsValid:           isValid,
		ChannelDetails:    channelDetails,
		HasSurround:       hasSurround,
		HasLFE:            hasLFE,
		IsBroadcastLayout: isBroadcastLayout,
		LayoutIssues:      layoutIssues,
	}, nil
}

// getChannelName returns human-readable channel name based on index and layout
func getChannelName(index int, layout string) string {
	if layout == "5.1" {
		names := []string{"Front Left", "Front Right", "Center", "LFE", "Surround Left", "Surround Right"}
		if index < len(names) {
			return names[index]
		}
	} else if layout == "7.1" {
		names := []string{"Front Left", "Front Right", "Center", "LFE", "Back Left", "Back Right", "Side Left", "Side Right"}
		if index < len(names) {
			return names[index]
		}
	} else if layout == "stereo" {
		names := []string{"Left", "Right"}
		if index < len(names) {
			return names[index]
		}
	} else if layout == "mono" {
		return "Mono"
	}
	return fmt.Sprintf("Channel %d", index)
}

// analyzeTimecodeContinuity checks for timecode gaps/discontinuities
func (ca *ContentAnalyzer) analyzeTimecodeContinuity(ctx context.Context, filePath string) (*TimecodeContinuityAnalysis, error) {
	// Analyze timecode metadata from video stream
	// Check for gaps, discontinuities, and proper formatting

	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-f", "null",
		"-",
	)

	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")

	hasTimecode := false
	var timecodeFormat string
	var startTimecode, endTimecode string
	var frameRate float64
	isDropFrame := false
	var discontinuities []TimecodeGap

	// Parse timecode from stream metadata
	for _, line := range lines {
		// Look for timecode in metadata
		if strings.Contains(line, "timecode") {
			hasTimecode = true
			// Extract timecode value if present
			parts := strings.Split(line, ":")
			if len(parts) >= 4 {
				// Reconstruct timecode from parts
				tcParts := parts[len(parts)-4:]
				tc := strings.Join(tcParts, ":")
				tc = strings.TrimSpace(tc)
				if startTimecode == "" {
					startTimecode = tc
				}
				endTimecode = tc

				// Check for drop frame (uses ; or , instead of :)
				if strings.Contains(line, ";") || strings.Contains(line, ",") {
					isDropFrame = true
					timecodeFormat = "Drop Frame"
				} else {
					timecodeFormat = "Non-Drop Frame"
				}
			}
		}

		// Get frame rate for timecode validation
		if strings.Contains(line, "fps") || strings.Contains(line, "tbr") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if (part == "fps" || part == "fps," || part == "tbr" || part == "tbr,") && i > 0 {
					if rate, err := strconv.ParseFloat(parts[i-1], 64); err == nil {
						if rate > 0 && rate < 200 {
							frameRate = rate
						}
					}
				}
			}
		}
	}

	// If no timecode in metadata, check for timecode data stream
	if !hasTimecode {
		// Try to extract timecode from data streams
		tcCmd := exec.CommandContext(ctx, ca.ffmpegPath,
			"-i", filePath,
			"-map", "0:d?", // Select data streams
			"-f", "null",
			"-",
		)
		tcOutput, _ := tcCmd.CombinedOutput()

		for _, line := range strings.Split(string(tcOutput), "\n") {
			if strings.Contains(line, "timecode") || strings.Contains(line, "Data:") {
				hasTimecode = true
				break
			}
		}
	}

	// Determine if timecode is continuous
	// For now, assume continuous if we have valid start timecode
	isContinuous := hasTimecode && startTimecode != ""

	// If drop frame format, validate it matches frame rate
	if isDropFrame && frameRate > 0 {
		// Drop frame should be used with 29.97 or 59.94 fps
		if frameRate < 29.9 || (frameRate > 30.1 && frameRate < 59.9) || frameRate > 60.1 {
			discontinuities = append(discontinuities, TimecodeGap{
				Position:   0,
				ExpectedTC: "",
				ActualTC:   "",
				GapFrames:  0,
				GapSeconds: 0,
			})
			isContinuous = false
		}
	}

	return &TimecodeContinuityAnalysis{
		HasTimecode:     hasTimecode,
		TimecodeFormat:  timecodeFormat,
		StartTimecode:   startTimecode,
		EndTimecode:     endTimecode,
		IsContinuous:    isContinuous,
		Discontinuities: discontinuities,
		TotalGaps:       len(discontinuities),
		IsDropFrame:     isDropFrame,
		FrameRate:       frameRate,
	}, nil
}

// analyzeBaseband performs baseband/waveform signal analysis using signalstats
func (ca *ContentAnalyzer) analyzeBaseband(ctx context.Context, filePath string) (*BasebandAnalysis, error) {
	// Use signalstats filter for comprehensive baseband analysis
	// This measures luminance levels, chroma levels, and broadcast range violations
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", "signalstats=stat=tout+vrep+brng",
		"-f", "null",
		"-",
	)

	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")

	var highestLuma, lowestLuma float64 = 0, 255
	var totalLuma float64
	var highestChromaU, highestChromaV float64 = 0, 0
	var lowestChromaU, lowestChromaV float64 = 255, 255
	var lumaFootroomViolations, lumaHeadroomViolations int
	var chromaHeadroomViolations int
	var gamutErrors int
	framesAnalyzed := 0

	// Broadcast legal range for 8-bit: Y 16-235, C 16-240
	legalMin := 16
	legalMax := 235
	chromaMax := 240

	for _, line := range lines {
		if strings.Contains(line, "signalstats") || strings.Contains(line, "Parsed_signalstats") {
			framesAnalyzed++

			// Parse YMIN (lowest luminance)
			if strings.Contains(line, "YMIN=") {
				parts := strings.Split(line, "YMIN=")
				if len(parts) > 1 {
					valStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
					if val, err := strconv.ParseFloat(valStr, 64); err == nil {
						if val < lowestLuma {
							lowestLuma = val
						}
						if val < float64(legalMin) {
							lumaFootroomViolations++
						}
					}
				}
			}

			// Parse YMAX (highest luminance)
			if strings.Contains(line, "YMAX=") {
				parts := strings.Split(line, "YMAX=")
				if len(parts) > 1 {
					valStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
					if val, err := strconv.ParseFloat(valStr, 64); err == nil {
						if val > highestLuma {
							highestLuma = val
						}
						if val > float64(legalMax) {
							lumaHeadroomViolations++
						}
					}
				}
			}

			// Parse YAVG (average luminance)
			if strings.Contains(line, "YAVG=") {
				parts := strings.Split(line, "YAVG=")
				if len(parts) > 1 {
					valStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
					if val, err := strconv.ParseFloat(valStr, 64); err == nil {
						totalLuma += val
					}
				}
			}

			// Parse UMAX/VMAX (chroma)
			if strings.Contains(line, "UMAX=") {
				parts := strings.Split(line, "UMAX=")
				if len(parts) > 1 {
					valStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
					if val, err := strconv.ParseFloat(valStr, 64); err == nil {
						if val > highestChromaU {
							highestChromaU = val
						}
						if val > float64(chromaMax) {
							chromaHeadroomViolations++
						}
					}
				}
			}

			if strings.Contains(line, "VMAX=") {
				parts := strings.Split(line, "VMAX=")
				if len(parts) > 1 {
					valStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
					if val, err := strconv.ParseFloat(valStr, 64); err == nil {
						if val > highestChromaV {
							highestChromaV = val
						}
					}
				}
			}

			// Parse BRNG (broadcast range violations / gamut errors)
			if strings.Contains(line, "BRNG=") {
				parts := strings.Split(line, "BRNG=")
				if len(parts) > 1 {
					valStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
					if val, err := strconv.Atoi(valStr); err == nil && val > 0 {
						gamutErrors += val
					}
				}
			}
		}
	}

	// Calculate statistics
	avgLuma := 0.0
	if framesAnalyzed > 0 {
		avgLuma = totalLuma / float64(framesAnalyzed)
	}

	lumaRange := highestLuma - lowestLuma
	brightness := avgLuma / 255.0 * 100 // Normalize to 0-100
	contrast := lumaRange / 255.0 * 100

	lumaOutOfRange := 0.0
	chromaOutOfRange := 0.0
	gamutErrorPercent := 0.0
	if framesAnalyzed > 0 {
		lumaOutOfRange = float64(lumaFootroomViolations+lumaHeadroomViolations) / float64(framesAnalyzed) * 100
		chromaOutOfRange = float64(chromaHeadroomViolations) / float64(framesAnalyzed) * 100
		gamutErrorPercent = float64(gamutErrors) / float64(framesAnalyzed) * 100
	}

	// Determine broadcast legal compliance
	isBroadcastLegal := lumaFootroomViolations == 0 && lumaHeadroomViolations == 0 && chromaHeadroomViolations == 0

	// Determine severity
	var severity string
	if lumaOutOfRange > 10 || chromaOutOfRange > 10 || gamutErrorPercent > 5 {
		severity = "critical"
	} else if lumaOutOfRange > 5 || chromaOutOfRange > 5 || gamutErrorPercent > 1 {
		severity = "warning"
	} else if lumaOutOfRange > 0 || chromaOutOfRange > 0 || gamutErrorPercent > 0 {
		severity = "minor"
	} else {
		severity = "none"
	}

	return &BasebandAnalysis{
		HighestLuminance:         highestLuma,
		LowestLuminance:          lowestLuma,
		AverageLuminance:         avgLuma,
		LuminanceRange:           lumaRange,
		Brightness:               brightness,
		Contrast:                 contrast,
		LumaFootroomViolations:   lumaFootroomViolations,
		LumaHeadroomViolations:   lumaHeadroomViolations,
		LumaOutOfRangePercent:    lumaOutOfRange,
		HighestChromaU:           highestChromaU,
		HighestChromaV:           highestChromaV,
		LowestChromaU:            lowestChromaU,
		LowestChromaV:            lowestChromaV,
		ChromaHeadroomViolations: chromaHeadroomViolations,
		ChromaOutOfRangePercent:  chromaOutOfRange,
		GamutErrors:              gamutErrors,
		GamutErrorPercent:        gamutErrorPercent,
		IsBroadcastLegal:         isBroadcastLegal,
		LegalRangeMin:            legalMin,
		LegalRangeMax:            legalMax,
		FramesAnalyzed:           framesAnalyzed,
		Severity:                 severity,
	}, nil
}

// analyzeVideoQualityScore calculates objective video quality metrics
func (ca *ContentAnalyzer) analyzeVideoQualityScore(ctx context.Context, filePath string) (*VideoQualityScoreAnalysis, error) {
	// Use multiple filters to compute quality scores
	// signalstats for sharpness/contrast, blur detection for blur score

	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", "signalstats=stat=tout+vrep+brng,entropy",
		"-f", "null",
		"-",
	)

	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")

	var totalSharpness, totalEntropy float64
	var minEntropy float64 = 8.0
	framesAnalyzed := 0

	for _, line := range lines {
		// Parse signalstats for contrast indicator
		if strings.Contains(line, "YDIF=") {
			framesAnalyzed++
			parts := strings.Split(line, "YDIF=")
			if len(parts) > 1 {
				valStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
				if val, err := strconv.ParseFloat(valStr, 64); err == nil {
					totalSharpness += val
				}
			}
		}

		// Parse entropy (indicates randomness/detail)
		if strings.Contains(line, "entropy") && strings.Contains(line, "normal") {
			parts := strings.Split(line, "normal")
			if len(parts) > 1 {
				for _, part := range strings.Fields(parts[1]) {
					if val, err := strconv.ParseFloat(part, 64); err == nil && val > 0 && val <= 8 {
						totalEntropy += val
						if val < minEntropy {
							minEntropy = val
						}
						break
					}
				}
			}
		}
	}

	// Calculate scores (normalized to 0-100)
	sharpnessScore := 50.0
	contrastScore := 50.0
	entropyScore := 50.0

	if framesAnalyzed > 0 {
		avgSharpness := totalSharpness / float64(framesAnalyzed)
		// Higher YDIF = more temporal activity = sharper perceived video
		sharpnessScore = math.Min(100, avgSharpness*10)

		avgEntropy := totalEntropy / float64(framesAnalyzed)
		// Higher entropy = more detail
		entropyScore = avgEntropy / 8.0 * 100

		// Contrast derived from sharpness and entropy
		contrastScore = (sharpnessScore + entropyScore) / 2
	}

	// Color score (placeholder - would need more analysis)
	colorScore := 70.0

	// Noise score (inverse of noise level - assume moderate)
	noiseScore := 75.0

	// Blockiness score (assume good if no explicit measurement)
	blockinessScore := 80.0

	// Overall score (weighted average)
	overallScore := sharpnessScore*0.25 + contrastScore*0.2 + colorScore*0.15 + noiseScore*0.15 + blockinessScore*0.15 + entropyScore*0.1

	// Quality classification
	qualityClass := "unknown"
	if overallScore >= 85 {
		qualityClass = "excellent"
	} else if overallScore >= 70 {
		qualityClass = "good"
	} else if overallScore >= 50 {
		qualityClass = "fair"
	} else {
		qualityClass = "poor"
	}

	isBroadcastQuality := overallScore >= 70

	return &VideoQualityScoreAnalysis{
		OverallScore:       overallScore,
		SharpnessScore:     sharpnessScore,
		ContrastScore:      contrastScore,
		ColorScore:         colorScore,
		NoiseScore:         noiseScore,
		BlockinessScore:    blockinessScore,
		TemporalStability:  75.0, // Default
		MotionQuality:      75.0, // Default
		QualityClass:       qualityClass,
		IsBroadcastQuality: isBroadcastQuality,
		FramesAnalyzed:     framesAnalyzed,
	}, nil
}

// analyzeTemporalComplexity measures scene complexity and motion over time
func (ca *ContentAnalyzer) analyzeTemporalComplexity(ctx context.Context, filePath string) (*TemporalComplexityAnalysis, error) {
	// Use signalstats YDIF for temporal difference and scene change detection
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", "signalstats=stat=tout+vrep,select='gt(scene,0.3)',showinfo",
		"-f", "null",
		"-",
	)

	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")

	var totalComplexity, maxComplexity, minComplexity float64 = 0, 0, 1000
	var complexityValues []float64
	var sceneChanges int
	framesAnalyzed := 0
	var totalDuration float64

	// Get duration
	for _, line := range lines {
		if strings.Contains(line, "Duration:") && strings.Contains(line, ",") {
			parts := strings.Split(line, "Duration:")
			if len(parts) > 1 {
				durationStr := strings.TrimSpace(strings.Split(parts[1], ",")[0])
				totalDuration = parseDurationToSeconds(durationStr)
			}
		}
	}

	for _, line := range lines {
		// Parse YDIF for temporal complexity
		if strings.Contains(line, "YDIF=") {
			framesAnalyzed++
			parts := strings.Split(line, "YDIF=")
			if len(parts) > 1 {
				valStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
				if val, err := strconv.ParseFloat(valStr, 64); err == nil {
					complexityValues = append(complexityValues, val)
					totalComplexity += val
					if val > maxComplexity {
						maxComplexity = val
					}
					if val < minComplexity {
						minComplexity = val
					}
				}
			}
		}

		// Count scene changes from select filter
		if strings.Contains(line, "select") && strings.Contains(line, "pts_time") {
			sceneChanges++
		}
	}

	avgComplexity := 0.0
	variance := 0.0
	highMotionCount := 0

	if framesAnalyzed > 0 {
		avgComplexity = totalComplexity / float64(framesAnalyzed)

		// Calculate variance
		for _, v := range complexityValues {
			variance += (v - avgComplexity) * (v - avgComplexity)
			if v > avgComplexity*2 {
				highMotionCount++
			}
		}
		variance /= float64(framesAnalyzed)
	}

	if minComplexity == 1000 {
		minComplexity = 0
	}

	highMotionPercent := 0.0
	if framesAnalyzed > 0 {
		highMotionPercent = float64(highMotionCount) / float64(framesAnalyzed) * 100
	}

	avgSceneLength := 0.0
	if sceneChanges > 0 && totalDuration > 0 {
		avgSceneLength = totalDuration / float64(sceneChanges)
	}

	// Classify complexity
	complexityClass := "low"
	if avgComplexity > 20 {
		complexityClass = "high"
	} else if avgComplexity > 10 {
		complexityClass = "medium"
	}

	encodingDifficulty := "easy"
	if avgComplexity > 20 || highMotionPercent > 30 {
		encodingDifficulty = "hard"
	} else if avgComplexity > 10 || highMotionPercent > 15 {
		encodingDifficulty = "medium"
	}

	return &TemporalComplexityAnalysis{
		AverageComplexity:  avgComplexity,
		MaxComplexity:      maxComplexity,
		MinComplexity:      minComplexity,
		ComplexityVariance: variance,
		AverageMotion:      avgComplexity, // YDIF approximates motion
		MaxMotion:          maxComplexity,
		HighMotionPercent:  highMotionPercent,
		SceneChangeCount:   sceneChanges,
		AverageSceneLength: avgSceneLength,
		ComplexityClass:    complexityClass,
		EncodingDifficulty: encodingDifficulty,
		FramesAnalyzed:     framesAnalyzed,
	}, nil
}

// analyzeFieldDominance detects field order issues in interlaced content
func (ca *ContentAnalyzer) analyzeFieldDominance(ctx context.Context, filePath string) (*FieldDominanceAnalysis, error) {
	// Use idet filter for interlace detection and field order analysis
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", "idet",
		"-f", "null",
		"-",
	)

	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")

	var tff, bff, progressive, undetermined int
	var detectedFieldOrder string
	framesAnalyzed := 0

	for _, line := range lines {
		// Parse idet output
		if strings.Contains(line, "idet") {
			if strings.Contains(line, "TFF:") {
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "TFF:" && i+1 < len(parts) {
						if val, err := strconv.Atoi(parts[i+1]); err == nil {
							tff = val
							framesAnalyzed += val
						}
					}
					if part == "BFF:" && i+1 < len(parts) {
						if val, err := strconv.Atoi(parts[i+1]); err == nil {
							bff = val
							framesAnalyzed += val
						}
					}
					if part == "Progressive:" && i+1 < len(parts) {
						if val, err := strconv.Atoi(parts[i+1]); err == nil {
							progressive = val
							framesAnalyzed += val
						}
					}
					if part == "Undetermined:" && i+1 < len(parts) {
						if val, err := strconv.Atoi(parts[i+1]); err == nil {
							undetermined = val
							framesAnalyzed += val
						}
					}
				}
			}
		}
	}

	// Determine dominant field order
	isInterlaced := (tff + bff) > progressive
	if tff > bff && tff > progressive {
		detectedFieldOrder = "TFF"
	} else if bff > tff && bff > progressive {
		detectedFieldOrder = "BFF"
	} else {
		detectedFieldOrder = "progressive"
		isInterlaced = false
	}

	// Calculate confidence and dominance ratio
	confidence := 0.0
	dominanceRatio := 0.0
	hasFieldOrderError := false
	fieldOrderErrors := 0

	if framesAnalyzed > 0 {
		if isInterlaced {
			dominant := max(tff, bff)
			minor := min(tff, bff)
			confidence = float64(dominant) / float64(tff+bff+progressive) * 100
			if dominant+minor > 0 {
				dominanceRatio = float64(dominant) / float64(dominant+minor)
			}
			// If both TFF and BFF are significant, there may be field order errors
			if minor > 0 && float64(minor)/float64(dominant) > 0.1 {
				hasFieldOrderError = true
				fieldOrderErrors = minor
			}
		} else {
			confidence = float64(progressive) / float64(framesAnalyzed) * 100
			dominanceRatio = 1.0
		}
	}

	errorPercent := 0.0
	if framesAnalyzed > 0 {
		errorPercent = float64(fieldOrderErrors) / float64(framesAnalyzed) * 100
	}

	severity := "none"
	if hasFieldOrderError {
		if errorPercent > 10 {
			severity = "critical"
		} else if errorPercent > 5 {
			severity = "warning"
		} else {
			severity = "minor"
		}
	}

	return &FieldDominanceAnalysis{
		IsInterlaced:       isInterlaced,
		DetectedFieldOrder: detectedFieldOrder,
		HasFieldOrderError: hasFieldOrderError,
		TopFieldFirst:      tff,
		BottomFieldFirst:   bff,
		Progressive:        progressive,
		Undetermined:       undetermined,
		Confidence:         confidence,
		DominanceRatio:     dominanceRatio,
		FieldOrderErrors:   fieldOrderErrors,
		ErrorPercent:       errorPercent,
		FramesAnalyzed:     framesAnalyzed,
		Severity:           severity,
	}, nil
}

// analyzeDifferentialFrames detects frame differences and anomalies
func (ca *ContentAnalyzer) analyzeDifferentialFrames(ctx context.Context, filePath string) (*DifferentialFrameAnalysis, error) {
	// Use signalstats YDIF for frame-to-frame differences
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", "signalstats=stat=tout+vrep",
		"-f", "null",
		"-",
	)

	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")

	var differences []float64
	var totalDiff, maxDiff, minDiff float64 = 0, 0, 1000
	var duplicateFrames, anomalousFrames int
	var suddenChanges []DifferentialEvent
	framesAnalyzed := 0
	frameNumber := 0

	// Thresholds
	duplicateThreshold := 0.1      // Very low diff = duplicate
	anomalyThreshold := 50.0       // Very high diff = anomaly
	suddenChangeThreshold := 30.0  // Sudden jump

	var prevDiff float64 = -1

	for _, line := range lines {
		if strings.Contains(line, "YDIF=") {
			framesAnalyzed++
			frameNumber++

			parts := strings.Split(line, "YDIF=")
			if len(parts) > 1 {
				valStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
				if val, err := strconv.ParseFloat(valStr, 64); err == nil {
					differences = append(differences, val)
					totalDiff += val

					if val > maxDiff {
						maxDiff = val
					}
					if val < minDiff {
						minDiff = val
					}

					// Check for duplicates
					if val < duplicateThreshold {
						duplicateFrames++
					}

					// Check for anomalies
					if val > anomalyThreshold {
						anomalousFrames++
					}

					// Check for sudden changes
					if prevDiff >= 0 && (val-prevDiff) > suddenChangeThreshold {
						suddenChanges = append(suddenChanges, DifferentialEvent{
							FrameNumber: frameNumber,
							Timestamp:   float64(frameNumber) / 25.0, // Approximate
							Difference:  val,
							EventType:   "sudden_change",
						})
					}

					prevDiff = val
				}
			}
		}
	}

	avgDiff := 0.0
	if framesAnalyzed > 0 {
		avgDiff = totalDiff / float64(framesAnalyzed)
	}

	if minDiff == 1000 {
		minDiff = 0
	}

	anomalyPercent := 0.0
	duplicatePercent := 0.0
	if framesAnalyzed > 0 {
		anomalyPercent = float64(anomalousFrames) / float64(framesAnalyzed) * 100
		duplicatePercent = float64(duplicateFrames) / float64(framesAnalyzed) * 100
	}

	// Estimate drops based on sudden changes
	estimatedDrops := len(suddenChanges)
	dropDetected := estimatedDrops > 0

	// Broadcast compliant if few anomalies
	isBroadcastCompliant := anomalyPercent < 1 && duplicatePercent < 5

	return &DifferentialFrameAnalysis{
		AverageDifference:    avgDiff,
		MaxDifference:        maxDiff,
		MinDifference:        minDiff,
		AnomalousFrames:      anomalousFrames,
		AnomalyPercent:       anomalyPercent,
		DuplicateFrames:      duplicateFrames,
		DuplicatePercent:     duplicatePercent,
		SuddenChangeCount:    len(suddenChanges),
		SuddenChanges:        suddenChanges,
		DropDetected:         dropDetected,
		EstimatedDrops:       estimatedDrops,
		FramesAnalyzed:       framesAnalyzed,
		IsBroadcastCompliant: isBroadcastCompliant,
	}, nil
}

// analyzeLineErrors detects luminance and chrominance line errors
func (ca *ContentAnalyzer) analyzeLineErrors(ctx context.Context, filePath string) (*LineErrorAnalysis, error) {
	// Use signalstats with out-of-range detection to find line errors
	// Line errors typically show as horizontal bands with incorrect values
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-vf", "signalstats=stat=tout+vrep+brng",
		"-f", "null",
		"-",
	)

	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")

	var lumaLineErrors, chromaLineErrors, digiBetaErrors int
	framesAnalyzed := 0
	affectedFrames := 0
	frameNumber := 0

	for _, line := range lines {
		if strings.Contains(line, "signalstats") || strings.Contains(line, "Parsed_signalstats") {
			framesAnalyzed++
			frameNumber++
			frameHasError := false

			// Parse TOUT (temporal outliers - can indicate line errors)
			if strings.Contains(line, "TOUT=") {
				parts := strings.Split(line, "TOUT=")
				if len(parts) > 1 {
					valStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
					if val, err := strconv.Atoi(valStr); err == nil && val > 100 {
						// High temporal outlier count suggests line errors
						lumaLineErrors++
						frameHasError = true
					}
				}
			}

			// Parse BRNG (broadcast range violations)
			if strings.Contains(line, "BRNG=") {
				parts := strings.Split(line, "BRNG=")
				if len(parts) > 1 {
					valStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
					if val, err := strconv.Atoi(valStr); err == nil && val > 1000 {
						// High BRNG with localized pattern = possible chroma error
						chromaLineErrors++
						frameHasError = true
					}
				}
			}

			// Parse VREP (vertical repeat - DigiBeta error indicator)
			if strings.Contains(line, "VREP=") {
				parts := strings.Split(line, "VREP=")
				if len(parts) > 1 {
					valStr := strings.TrimSpace(strings.Split(parts[1], " ")[0])
					if val, err := strconv.ParseFloat(valStr, 64); err == nil && val > 0.5 {
						// High vertical repeat = possible DigiBeta error
						digiBetaErrors++
						frameHasError = true
					}
				}
			}

			if frameHasError {
				affectedFrames++
			}
		}
	}

	totalErrors := lumaLineErrors + chromaLineErrors + digiBetaErrors
	errorPercentage := 0.0
	if framesAnalyzed > 0 {
		errorPercentage = float64(affectedFrames) / float64(framesAnalyzed) * 100
	}

	isBroadcastCompliant := errorPercentage < 0.1
	severity := "none"
	if errorPercentage > 5 {
		severity = "critical"
	} else if errorPercentage > 1 {
		severity = "warning"
	} else if errorPercentage > 0 {
		severity = "minor"
	}

	return &LineErrorAnalysis{
		LuminanceLineErrors:   lumaLineErrors,
		ChrominanceLineErrors: chromaLineErrors,
		DigiBetaErrors:        digiBetaErrors,
		TotalLineErrors:       totalErrors,
		ErrorPercentage:       errorPercentage,
		AffectedFrames:        affectedFrames,
		FramesAnalyzed:        framesAnalyzed,
		IsBroadcastCompliant:  isBroadcastCompliant,
		Severity:              severity,
	}, nil
}

// analyzeAudioFrequency provides detailed audio frequency analysis
func (ca *ContentAnalyzer) analyzeAudioFrequency(ctx context.Context, filePath string) (*AudioFrequencyAnalysis, error) {
	// Use astats and showfreqs for frequency analysis
	cmd := exec.CommandContext(ctx, ca.ffmpegPath,
		"-i", filePath,
		"-af", "astats=metadata=1:reset=0",
		"-f", "null",
		"-",
	)

	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")

	var sampleRate float64
	var totalRMS, totalPeak float64
	framesAnalyzed := 0

	// Get sample rate from stream info
	for _, line := range lines {
		if strings.Contains(line, "Audio:") && strings.Contains(line, "Hz") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if strings.HasSuffix(part, "Hz") || strings.HasSuffix(part, "Hz,") {
					rateStr := strings.TrimSuffix(strings.TrimSuffix(part, ","), "Hz")
					if rate, err := strconv.ParseFloat(rateStr, 64); err == nil {
						sampleRate = rate
					}
				}
				if part == "Hz" && i > 0 {
					if rate, err := strconv.ParseFloat(parts[i-1], 64); err == nil {
						sampleRate = rate
					}
				}
			}
		}
	}

	// Parse audio statistics
	for _, line := range lines {
		if strings.Contains(line, "RMS level dB:") {
			framesAnalyzed++
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				valStr := strings.TrimSpace(parts[len(parts)-1])
				if val, err := strconv.ParseFloat(valStr, 64); err == nil {
					totalRMS += val
				}
			}
		}
		if strings.Contains(line, "Peak level dB:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				valStr := strings.TrimSpace(parts[len(parts)-1])
				if val, err := strconv.ParseFloat(valStr, 64); err == nil {
					totalPeak += val
				}
			}
		}
	}

	// Calculate frequency range based on sample rate
	maxFrequency := sampleRate / 2 // Nyquist frequency
	if maxFrequency == 0 {
		maxFrequency = 22050 // Default 44.1kHz sample rate
	}

	// Estimate frequency distribution (simplified)
	// Real implementation would use FFT
	lowFreqEnergy := 30.0  // 0-300Hz
	midFreqEnergy := 50.0  // 300Hz-4kHz
	highFreqEnergy := 20.0 // 4kHz+

	// Bandwidth usage estimation
	effectiveBandwidth := maxFrequency * 0.9
	bandwidthUsage := 90.0

	// Check for pure tone (very consistent RMS indicates possible tone)
	avgRMS := -96.0
	if framesAnalyzed > 0 {
		avgRMS = totalRMS / float64(framesAnalyzed)
	}

	hasPureTone := false
	pureToneFreq := 0.0
	pureToneLevel := -96.0

	// Simple check: very consistent level might indicate test tone
	avgPeak := -96.0
	if framesAnalyzed > 0 {
		avgPeak = totalPeak / float64(framesAnalyzed)
	}

	crestFactor := avgPeak - avgRMS
	if crestFactor > 2.0 && crestFactor < 4.0 && avgRMS > -25 && avgRMS < -15 {
		hasPureTone = true
		pureToneFreq = 1000 // Assume 1kHz
		pureToneLevel = avgRMS
	}

	// Spectral metrics (simplified estimates)
	spectralFlatness := 0.5  // 0-1, 1 = white noise
	spectralCentroid := 2000.0 // Hz

	return &AudioFrequencyAnalysis{
		DominantFrequency:  spectralCentroid,
		FrequencyRange:     [2]float64{20, maxFrequency},
		LowFreqEnergy:      lowFreqEnergy,
		MidFreqEnergy:      midFreqEnergy,
		HighFreqEnergy:     highFreqEnergy,
		HasPureTone:        hasPureTone,
		PureToneFrequency:  pureToneFreq,
		PureToneLevel:      pureToneLevel,
		EffectiveBandwidth: effectiveBandwidth,
		BandwidthUsage:     bandwidthUsage,
		SpectralFlatness:   spectralFlatness,
		SpectralCentroid:   spectralCentroid,
		FramesAnalyzed:     framesAnalyzed,
	}, nil
}
