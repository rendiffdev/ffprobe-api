package ffmpeg

import (
	"context"
	"strings"

	"github.com/rs/zerolog"
)

// StreamDispositionAnalyzer analyzes stream disposition for accessibility and compliance
type StreamDispositionAnalyzer struct {
	ffprobePath string
	logger      zerolog.Logger
}

// NewStreamDispositionAnalyzer creates a new stream disposition analyzer
func NewStreamDispositionAnalyzer(ffprobePath string, logger zerolog.Logger) *StreamDispositionAnalyzer {
	return &StreamDispositionAnalyzer{
		ffprobePath: ffprobePath,
		logger:      logger.With().Str("analyzer", "stream_disposition").Logger(),
	}
}

// AnalyzeStreamDisposition performs comprehensive stream disposition analysis
func (a *StreamDispositionAnalyzer) AnalyzeStreamDisposition(ctx context.Context, filePath string, streams []StreamInfo) (*StreamDispositionAnalysis, error) {
	analysis := &StreamDispositionAnalysis{
		VideoStreams:         make(map[int]*StreamDisposition),
		AudioStreams:         make(map[int]*StreamDisposition),
		SubtitleStreams:      make(map[int]*StreamDisposition),
		LanguageDistribution: make(map[string]int),
	}

	var accessibilityScore int
	maxAccessibilityScore := 100

	// Analyze each stream's disposition
	for _, stream := range streams {
		disposition := a.analyzeStreamDisposition(&stream)

		// Categorize by stream type
		switch strings.ToLower(stream.CodecType) {
		case "video":
			analysis.VideoStreams[stream.Index] = disposition
		case "audio":
			analysis.AudioStreams[stream.Index] = disposition
			a.analyzeAudioDisposition(analysis, disposition)
		case "subtitle":
			analysis.SubtitleStreams[stream.Index] = disposition
			a.analyzeSubtitleDisposition(analysis, disposition)
		}

		// Track language distribution
		if language := a.extractLanguage(&stream); language != "" {
			analysis.LanguageDistribution[language]++
		}
	}

	// Calculate accessibility features
	if analysis.HasForcedSubtitles {
		accessibilityScore += 20
	}
	if analysis.HasSDHSubtitles {
		accessibilityScore += 30
	}
	if analysis.HasDescriptiveAudio {
		accessibilityScore += 25
	}
	if analysis.HasAlternateStreams {
		accessibilityScore += 15
	}
	if len(analysis.LanguageDistribution) > 1 {
		accessibilityScore += 10 // Multi-language support
	}

	analysis.AccessibilityScore = (accessibilityScore * 100) / maxAccessibilityScore

	// Validate disposition compliance
	validation := a.validateDisposition(analysis)
	analysis.Validation = validation

	return analysis, nil
}

// analyzeStreamDisposition analyzes disposition for a single stream
func (a *StreamDispositionAnalyzer) analyzeStreamDisposition(stream *StreamInfo) *StreamDisposition {
	disposition := &StreamDisposition{
		StreamIndex:      stream.Index,
		DispositionFlags: make(map[string]bool),
		Language:         a.extractLanguage(stream),
	}

	// Extract disposition flags from FFprobe data
	if stream.Disposition != nil {
		disposition.Default = stream.Disposition["default"] > 0
		disposition.Dub = stream.Disposition["dub"] > 0
		disposition.Original = stream.Disposition["original"] > 0
		disposition.Comment = stream.Disposition["comment"] > 0
		disposition.Lyrics = stream.Disposition["lyrics"] > 0
		disposition.Karaoke = stream.Disposition["karaoke"] > 0
		disposition.Forced = stream.Disposition["forced"] > 0
		disposition.HearingImpaired = stream.Disposition["hearing_impaired"] > 0
		disposition.VisualImpaired = stream.Disposition["visual_impaired"] > 0
		disposition.CleanEffects = stream.Disposition["clean_effects"] > 0
		disposition.AttachedPic = stream.Disposition["attached_pic"] > 0
		disposition.TimedThumbnails = stream.Disposition["timed_thumbnails"] > 0
		disposition.CaptionService = stream.Disposition["captions"] > 0

		// Store all disposition flags
		for key, value := range stream.Disposition {
			disposition.DispositionFlags[key] = value > 0
		}
	}

	// Extract title from tags
	if stream.Tags != nil {
		if title, exists := stream.Tags["title"]; exists {
			disposition.Title = title
		}
	}

	// Determine role and accessibility type
	disposition.Role = a.determineStreamRole(disposition, stream)
	disposition.AccessibilityType = a.determineAccessibilityType(disposition, stream)
	disposition.IsCompliant = a.isDispositionCompliant(disposition, stream)

	return disposition
}

// analyzeAudioDisposition analyzes audio-specific disposition features
func (a *StreamDispositionAnalyzer) analyzeAudioDisposition(analysis *StreamDispositionAnalysis, disposition *StreamDisposition) {
	if disposition.Default {
		analysis.HasMainStreams = true
	}

	if disposition.Comment {
		analysis.HasCommentary = true
	}

	if disposition.VisualImpaired {
		analysis.HasDescriptiveAudio = true
	}

	if disposition.Dub || (!disposition.Default && !disposition.Comment) {
		analysis.HasAlternateStreams = true
	}
}

// analyzeSubtitleDisposition analyzes subtitle-specific disposition features
func (a *StreamDispositionAnalyzer) analyzeSubtitleDisposition(analysis *StreamDispositionAnalysis, disposition *StreamDisposition) {
	if disposition.Forced {
		analysis.HasForcedSubtitles = true
	}

	if disposition.HearingImpaired {
		analysis.HasSDHSubtitles = true
	}

	if !disposition.Default && !disposition.Forced {
		analysis.HasAlternateStreams = true
	}
}

// extractLanguage extracts language information from stream
func (a *StreamDispositionAnalyzer) extractLanguage(stream *StreamInfo) string {
	if stream.Tags != nil {
		// Try common language tag names
		for _, langTag := range []string{"language", "lang", "LANGUAGE", "LANG"} {
			if lang, exists := stream.Tags[langTag]; exists {
				return strings.ToLower(lang)
			}
		}
	}
	return "und" // undefined
}

// determineStreamRole determines the role of the stream
func (a *StreamDispositionAnalyzer) determineStreamRole(disposition *StreamDisposition, stream *StreamInfo) string {
	if disposition.Default {
		return "main"
	}
	if disposition.Comment {
		return "commentary"
	}
	if disposition.Dub {
		return "dub"
	}
	if disposition.Original {
		return "original"
	}
	if disposition.Forced {
		return "forced"
	}
	return "alternate"
}

// determineAccessibilityType determines the accessibility type of the stream
func (a *StreamDispositionAnalyzer) determineAccessibilityType(disposition *StreamDisposition, stream *StreamInfo) string {
	if disposition.HearingImpaired {
		return "sdh" // Subtitles for Deaf/Hard-of-hearing
	}
	if disposition.VisualImpaired {
		return "audio_description"
	}
	if disposition.CaptionService {
		return "captions"
	}
	if disposition.Forced {
		return "forced_narrative"
	}
	return "none"
}

// isDispositionCompliant checks if stream disposition follows accessibility standards
func (a *StreamDispositionAnalyzer) isDispositionCompliant(disposition *StreamDisposition, stream *StreamInfo) bool {
	// Basic compliance checks

	// For subtitle streams, check accessibility requirements
	if strings.ToLower(stream.CodecType) == "subtitle" {
		// If hearing impaired content, should be properly flagged
		if disposition.AccessibilityType == "sdh" && !disposition.HearingImpaired {
			return false
		}

		// Forced subtitles should be flagged as forced
		if strings.Contains(strings.ToLower(disposition.Title), "forced") && !disposition.Forced {
			return false
		}
	}

	// For audio streams, check descriptive audio compliance
	if strings.ToLower(stream.CodecType) == "audio" {
		if disposition.AccessibilityType == "audio_description" && !disposition.VisualImpaired {
			return false
		}
	}

	// Language should be specified for non-default streams
	if !disposition.Default && disposition.Language == "und" {
		return false
	}

	return true
}

// validateDisposition validates overall stream disposition compliance
func (a *StreamDispositionAnalyzer) validateDisposition(analysis *StreamDispositionAnalysis) *DispositionValidation {
	validation := &DispositionValidation{
		IsValid:                true,
		AccessibilityCompliant: true,
		Standards:              []string{"ADA", "Section 508", "WCAG 2.1"},
	}

	// Check for main streams
	if !analysis.HasMainStreams {
		validation.Issues = append(validation.Issues, "No default/main streams found")
		validation.IsValid = false
	}

	// Check accessibility compliance
	totalStreams := len(analysis.VideoStreams) + len(analysis.AudioStreams) + len(analysis.SubtitleStreams)
	if totalStreams > 1 {
		// Multi-stream content should have accessibility features
		if !analysis.HasSDHSubtitles && !analysis.HasForcedSubtitles {
			validation.Issues = append(validation.Issues, "No accessibility subtitles found")
			validation.AccessibilityCompliant = false
		}

		// Check for audio descriptions if multiple audio tracks
		if len(analysis.AudioStreams) > 1 && !analysis.HasDescriptiveAudio {
			validation.Recommendations = append(validation.Recommendations, "Consider adding audio descriptions for visually impaired users")
		}
	}

	// Check language distribution
	if len(analysis.LanguageDistribution) > 1 {
		undefinedCount := analysis.LanguageDistribution["und"]
		totalLangStreams := 0
		for _, count := range analysis.LanguageDistribution {
			totalLangStreams += count
		}

		if float64(undefinedCount)/float64(totalLangStreams) > 0.3 {
			validation.Issues = append(validation.Issues, "High percentage of streams with undefined language")
			validation.IsValid = false
		}
	}

	// Accessibility score validation
	if analysis.AccessibilityScore < 40 {
		validation.AccessibilityCompliant = false
		validation.Recommendations = append(validation.Recommendations,
			"Improve accessibility features: add SDH subtitles, audio descriptions, or forced subtitles")
	}

	return validation
}
