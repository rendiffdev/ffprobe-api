package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rendiffdev/ffprobe-api/internal/database"
	"github.com/rendiffdev/ffprobe-api/internal/models"
	"github.com/rendiffdev/ffprobe-api/internal/repositories"
)

// ComparisonService handles video comparison operations
type ComparisonService struct {
	comparisonRepo repositories.ComparisonRepository
	analysisRepo   database.Repository
	llmService     *LLMService
}

// NewComparisonService creates a new comparison service
func NewComparisonService(
	comparisonRepo repositories.ComparisonRepository,
	analysisRepo database.Repository,
	llmService *LLMService,
) *ComparisonService {
	return &ComparisonService{
		comparisonRepo: comparisonRepo,
		analysisRepo:   analysisRepo,
		llmService:     llmService,
	}
}

// CreateComparison creates a new video comparison
func (s *ComparisonService) CreateComparison(ctx context.Context, req *models.CreateComparisonRequest) (*models.ComparisonResponse, error) {
	// Validate that both analyses exist
	originalAnalysis, err := s.analysisRepo.GetAnalysis(ctx, req.OriginalAnalysisID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original analysis: %w", err)
	}

	modifiedAnalysis, err := s.analysisRepo.GetAnalysis(ctx, req.ModifiedAnalysisID)
	if err != nil {
		return nil, fmt.Errorf("failed to get modified analysis: %w", err)
	}

	// Ensure both analyses are completed
	if originalAnalysis.Status != models.StatusCompleted {
		return nil, fmt.Errorf("original analysis is not completed (status: %s)", originalAnalysis.Status)
	}

	if modifiedAnalysis.Status != models.StatusCompleted {
		return nil, fmt.Errorf("modified analysis is not completed (status: %s)", modifiedAnalysis.Status)
	}

	// Create comparison record
	comparison := &models.VideoComparison{
		ID:                 uuid.New(),
		UserID:             originalAnalysis.UserID,
		OriginalAnalysisID: req.OriginalAnalysisID,
		ModifiedAnalysisID: req.ModifiedAnalysisID,
		ComparisonType:     req.ComparisonType,
		Status:             models.ComparisonStatusPending,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Save initial comparison record
	if err := s.comparisonRepo.Create(ctx, comparison); err != nil {
		return nil, fmt.Errorf("failed to create comparison: %w", err)
	}

	// Process comparison asynchronously
	go func() {
		ctx := context.Background()
		if err := s.processComparison(ctx, comparison, originalAnalysis, modifiedAnalysis, req); err != nil {
			s.markComparisonFailed(ctx, comparison.ID, err.Error())
		}
	}()

	return &models.ComparisonResponse{
		ID:                 comparison.ID,
		OriginalAnalysisID: comparison.OriginalAnalysisID,
		ModifiedAnalysisID: comparison.ModifiedAnalysisID,
		ComparisonType:     comparison.ComparisonType,
		Status:             comparison.Status,
		CreatedAt:          comparison.CreatedAt,
		UpdatedAt:          comparison.UpdatedAt,
	}, nil
}

// processComparison performs the actual comparison analysis
func (s *ComparisonService) processComparison(
	ctx context.Context,
	comparison *models.VideoComparison,
	original, modified *models.Analysis,
	req *models.CreateComparisonRequest,
) error {
	// Update status to processing
	comparison.Status = models.ComparisonStatusProcessing
	comparison.UpdatedAt = time.Now()
	if err := s.comparisonRepo.Update(ctx, comparison); err != nil {
		return fmt.Errorf("failed to update comparison status: %w", err)
	}

	// Perform comparison analysis
	comparisonData, err := s.performComparison(original, modified, req.ComparisonType)
	if err != nil {
		return fmt.Errorf("failed to perform comparison: %w", err)
	}

	// Calculate quality score
	qualityScore := s.calculateQualityScore(comparisonData)

	// Generate LLM assessment if requested
	var llmAssessment *string
	if req.IncludeLLM && s.llmService != nil {
		assessment, err := s.generateLLMAssessment(ctx, original, modified, comparisonData)
		if err != nil {
			// Log error but don't fail the comparison
			fmt.Printf("Failed to generate LLM assessment: %v\n", err)
		} else {
			llmAssessment = &assessment
		}
	}

	// Update comparison with results
	comparison.ComparisonData = *comparisonData
	comparison.QualityScore = qualityScore
	comparison.LLMAssessment = llmAssessment
	comparison.Status = models.ComparisonStatusCompleted
	comparison.UpdatedAt = time.Now()

	return s.comparisonRepo.Update(ctx, comparison)
}

// performComparison performs the detailed comparison between two analyses
func (s *ComparisonService) performComparison(
	original, modified *models.Analysis,
	comparisonType models.ComparisonType,
) (*models.ComparisonData, error) {
	comparisonData := &models.ComparisonData{}

	// Parse FFprobe data for both analyses
	originalStreams, err := s.parseStreams(original.FFprobeData.Streams)
	if err != nil {
		return nil, fmt.Errorf("failed to parse original streams: %w", err)
	}

	modifiedStreams, err := s.parseStreams(modified.FFprobeData.Streams)
	if err != nil {
		return nil, fmt.Errorf("failed to parse modified streams: %w", err)
	}

	originalFormat, err := s.parseFormat(original.FFprobeData.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to parse original format: %w", err)
	}

	modifiedFormat, err := s.parseFormat(modified.FFprobeData.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to parse modified format: %w", err)
	}

	// Compare video quality
	if videoComparison := s.compareVideoQuality(originalStreams, modifiedStreams); videoComparison != nil {
		comparisonData.VideoQuality = videoComparison
	}

	// Compare audio quality
	if audioComparison := s.compareAudioQuality(originalStreams, modifiedStreams); audioComparison != nil {
		comparisonData.AudioQuality = audioComparison
	}

	// Compare file size
	comparisonData.FileSize = s.compareFileSize(original.FileSize, modified.FileSize)

	// Compare bitrate
	comparisonData.BitrateAnalysis = s.compareBitrate(originalFormat, modifiedFormat, originalStreams, modifiedStreams)

	// Compare formats
	comparisonData.FormatChanges = s.compareFormats(originalFormat, modifiedFormat, originalStreams, modifiedStreams)

	// Analyze issues and improvements
	s.analyzeIssuesAndImprovements(comparisonData, original, modified)

	// Generate summary
	comparisonData.Summary = s.generateComparisonSummary(comparisonData)

	return comparisonData, nil
}

// compareVideoQuality compares video quality metrics between two analyses
func (s *ComparisonService) compareVideoQuality(originalStreams, modifiedStreams []map[string]interface{}) *models.VideoQualityComparison {
	originalVideo := s.getVideoStream(originalStreams)
	modifiedVideo := s.getVideoStream(modifiedStreams)

	if originalVideo == nil || modifiedVideo == nil {
		return nil
	}

	comparison := &models.VideoQualityComparison{}

	// Compare resolution
	if resComparison := s.compareResolution(originalVideo, modifiedVideo); resComparison != nil {
		comparison.Resolution = resComparison
	}

	// Compare frame rate
	if frameRateComparison := s.compareMetric(
		s.getFloatValue(originalVideo, "r_frame_rate"),
		s.getFloatValue(modifiedVideo, "r_frame_rate"),
		true, // Higher frame rate is generally better
	); frameRateComparison != nil {
		comparison.FrameRate = frameRateComparison
	}

	// Compare bit depth
	if bitDepthComparison := s.compareMetric(
		s.getFloatValue(originalVideo, "bits_per_raw_sample"),
		s.getFloatValue(modifiedVideo, "bits_per_raw_sample"),
		true, // Higher bit depth is generally better
	); bitDepthComparison != nil {
		comparison.BitDepth = bitDepthComparison
	}

	// Compare color space
	originalColorSpace := s.getStringValue(originalVideo, "color_space")
	modifiedColorSpace := s.getStringValue(modifiedVideo, "color_space")
	if originalColorSpace != "" || modifiedColorSpace != "" {
		comparison.ColorSpace = &models.FormatChange{
			Original: originalColorSpace,
			Modified: modifiedColorSpace,
			Changed:  originalColorSpace != modifiedColorSpace,
		}
	}

	// Calculate overall video quality improvement (simplified)
	improvement := 0.0
	if comparison.Resolution != nil && comparison.Resolution.ScalingFactor > 1.0 {
		improvement += 20.0 // Resolution increase
	}
	if comparison.FrameRate != nil && comparison.FrameRate.Improvement {
		improvement += 10.0 // Frame rate improvement
	}
	if comparison.BitDepth != nil && comparison.BitDepth.Improvement {
		improvement += 15.0 // Bit depth improvement
	}

	comparison.QualityImprovement = improvement

	return comparison
}

// compareAudioQuality compares audio quality metrics
func (s *ComparisonService) compareAudioQuality(originalStreams, modifiedStreams []map[string]interface{}) *models.AudioQualityComparison {
	originalAudio := s.getAudioStream(originalStreams)
	modifiedAudio := s.getAudioStream(modifiedStreams)

	if originalAudio == nil || modifiedAudio == nil {
		return nil
	}

	comparison := &models.AudioQualityComparison{}

	// Compare sample rate
	if sampleRateComparison := s.compareMetric(
		s.getFloatValue(originalAudio, "sample_rate"),
		s.getFloatValue(modifiedAudio, "sample_rate"),
		true, // Higher sample rate is generally better
	); sampleRateComparison != nil {
		comparison.SampleRate = sampleRateComparison
	}

	// Compare channels
	if channelsComparison := s.compareMetric(
		s.getFloatValue(originalAudio, "channels"),
		s.getFloatValue(modifiedAudio, "channels"),
		true, // More channels can be better (context dependent)
	); channelsComparison != nil {
		comparison.Channels = channelsComparison
	}

	// Compare bit depth
	if bitDepthComparison := s.compareMetric(
		s.getFloatValue(originalAudio, "bits_per_sample"),
		s.getFloatValue(modifiedAudio, "bits_per_sample"),
		true, // Higher bit depth is generally better
	); bitDepthComparison != nil {
		comparison.BitDepth = bitDepthComparison
	}

	// Compare codec
	originalCodec := s.getStringValue(originalAudio, "codec_name")
	modifiedCodec := s.getStringValue(modifiedAudio, "codec_name")
	if originalCodec != "" || modifiedCodec != "" {
		comparison.Codec = &models.FormatChange{
			Original: originalCodec,
			Modified: modifiedCodec,
			Changed:  originalCodec != modifiedCodec,
		}
	}

	// Calculate overall audio quality improvement (simplified)
	improvement := 0.0
	if comparison.SampleRate != nil && comparison.SampleRate.Improvement {
		improvement += 15.0
	}
	if comparison.BitDepth != nil && comparison.BitDepth.Improvement {
		improvement += 20.0
	}
	if comparison.Channels != nil && comparison.Channels.Improvement {
		improvement += 10.0
	}

	comparison.QualityImprovement = improvement

	return comparison
}

// compareFileSize compares file sizes
func (s *ComparisonService) compareFileSize(originalSize, modifiedSize int64) *models.FileSizeComparison {
	sizeChange := modifiedSize - originalSize
	percentageChange := 0.0
	compressionRatio := 0.0

	if originalSize > 0 {
		percentageChange = (float64(sizeChange) / float64(originalSize)) * 100
		compressionRatio = float64(modifiedSize) / float64(originalSize)
	}

	return &models.FileSizeComparison{
		OriginalSize:     originalSize,
		ModifiedSize:     modifiedSize,
		SizeChange:       sizeChange,
		PercentageChange: percentageChange,
		CompressionRatio: compressionRatio,
	}
}

// compareBitrate compares bitrate metrics
func (s *ComparisonService) compareBitrate(
	originalFormat, modifiedFormat map[string]interface{},
	originalStreams, modifiedStreams []map[string]interface{},
) *models.BitrateComparison {
	comparison := &models.BitrateComparison{}

	// Compare overall bitrate
	originalBitrate := s.getFloatValue(originalFormat, "bit_rate")
	modifiedBitrate := s.getFloatValue(modifiedFormat, "bit_rate")

	if originalBitrate > 0 && modifiedBitrate > 0 {
		comparison.Overall = s.compareMetric(originalBitrate, modifiedBitrate, false) // Lower bitrate can be better if quality is maintained
	}

	// Compare video bitrate
	originalVideo := s.getVideoStream(originalStreams)
	modifiedVideo := s.getVideoStream(modifiedStreams)
	if originalVideo != nil && modifiedVideo != nil {
		originalVideoBitrate := s.getFloatValue(originalVideo, "bit_rate")
		modifiedVideoBitrate := s.getFloatValue(modifiedVideo, "bit_rate")
		if originalVideoBitrate > 0 && modifiedVideoBitrate > 0 {
			comparison.Video = s.compareMetric(originalVideoBitrate, modifiedVideoBitrate, false)
		}
	}

	// Compare audio bitrate
	originalAudio := s.getAudioStream(originalStreams)
	modifiedAudio := s.getAudioStream(modifiedStreams)
	if originalAudio != nil && modifiedAudio != nil {
		originalAudioBitrate := s.getFloatValue(originalAudio, "bit_rate")
		modifiedAudioBitrate := s.getFloatValue(modifiedAudio, "bit_rate")
		if originalAudioBitrate > 0 && modifiedAudioBitrate > 0 {
			comparison.Audio = s.compareMetric(originalAudioBitrate, modifiedAudioBitrate, false)
		}
	}

	// Calculate bitrate efficiency (simplified quality per bit metric)
	if comparison.Overall != nil {
		comparison.BitrateEfficiency = s.calculateBitrateEfficiency(comparison.Overall.PercentageChange)
	}

	return comparison
}

// compareFormats compares format and codec information
func (s *ComparisonService) compareFormats(
	originalFormat, modifiedFormat map[string]interface{},
	originalStreams, modifiedStreams []map[string]interface{},
) *models.FormatComparison {
	comparison := &models.FormatComparison{}

	// Compare container format
	originalContainer := s.getStringValue(originalFormat, "format_name")
	modifiedContainer := s.getStringValue(modifiedFormat, "format_name")
	if originalContainer != "" || modifiedContainer != "" {
		comparison.Container = &models.FormatChange{
			Original: originalContainer,
			Modified: modifiedContainer,
			Changed:  originalContainer != modifiedContainer,
		}
	}

	// Compare video codec
	originalVideo := s.getVideoStream(originalStreams)
	modifiedVideo := s.getVideoStream(modifiedStreams)
	if originalVideo != nil && modifiedVideo != nil {
		originalVideoCodec := s.getStringValue(originalVideo, "codec_name")
		modifiedVideoCodec := s.getStringValue(modifiedVideo, "codec_name")
		comparison.VideoCodec = &models.FormatChange{
			Original: originalVideoCodec,
			Modified: modifiedVideoCodec,
			Changed:  originalVideoCodec != modifiedVideoCodec,
		}

		// Compare profile
		originalProfile := s.getStringValue(originalVideo, "profile")
		modifiedProfile := s.getStringValue(modifiedVideo, "profile")
		comparison.Profile = &models.FormatChange{
			Original: originalProfile,
			Modified: modifiedProfile,
			Changed:  originalProfile != modifiedProfile,
		}

		// Compare level
		originalLevel := s.getStringValue(originalVideo, "level")
		modifiedLevel := s.getStringValue(modifiedVideo, "level")
		comparison.Level = &models.FormatChange{
			Original: originalLevel,
			Modified: modifiedLevel,
			Changed:  originalLevel != modifiedLevel,
		}
	}

	// Compare audio codec
	originalAudio := s.getAudioStream(originalStreams)
	modifiedAudio := s.getAudioStream(modifiedStreams)
	if originalAudio != nil && modifiedAudio != nil {
		originalAudioCodec := s.getStringValue(originalAudio, "codec_name")
		modifiedAudioCodec := s.getStringValue(modifiedAudio, "codec_name")
		comparison.AudioCodec = &models.FormatChange{
			Original: originalAudioCodec,
			Modified: modifiedAudioCodec,
			Changed:  originalAudioCodec != modifiedAudioCodec,
		}
	}

	return comparison
}

// analyzeIssuesAndImprovements analyzes what issues were fixed and what new issues appeared
func (s *ComparisonService) analyzeIssuesAndImprovements(
	comparisonData *models.ComparisonData,
	original, modified *models.Analysis,
) {
	var issuesFixed []string
	var newIssues []string
	var recommendations []string

	// Analyze video quality changes
	if comparisonData.VideoQuality != nil {
		if comparisonData.VideoQuality.Resolution != nil {
			if comparisonData.VideoQuality.Resolution.ScalingFactor > 1.0 {
				issuesFixed = append(issuesFixed, "Resolution increased")
			} else if comparisonData.VideoQuality.Resolution.ScalingFactor < 1.0 {
				newIssues = append(newIssues, "Resolution decreased")
				recommendations = append(recommendations, "Consider maintaining original resolution")
			}
		}

		if comparisonData.VideoQuality.FrameRate != nil && comparisonData.VideoQuality.FrameRate.Improvement {
			issuesFixed = append(issuesFixed, "Frame rate improved")
		}
	}

	// Analyze file size changes
	if comparisonData.FileSize != nil {
		if comparisonData.FileSize.PercentageChange < -20 {
			issuesFixed = append(issuesFixed, "Significant file size reduction achieved")
		} else if comparisonData.FileSize.PercentageChange > 50 {
			newIssues = append(newIssues, "File size increased significantly")
			recommendations = append(recommendations, "Consider optimizing compression settings")
		}
	}

	// Analyze format changes
	if comparisonData.FormatChanges != nil {
		if comparisonData.FormatChanges.VideoCodec != nil && comparisonData.FormatChanges.VideoCodec.Changed {
			modernCodecs := []string{"h265", "hevc", "av1", "vp9"}
			if s.isModernCodec(comparisonData.FormatChanges.VideoCodec.Modified, modernCodecs) {
				issuesFixed = append(issuesFixed, "Upgraded to modern video codec")
			}
		}
	}

	comparisonData.IssuesFixed = issuesFixed
	comparisonData.NewIssues = newIssues
	comparisonData.Recommendations = recommendations
}

// generateComparisonSummary generates an overall summary of the comparison
func (s *ComparisonService) generateComparisonSummary(comparisonData *models.ComparisonData) *models.ComparisonSummary {
	summary := &models.ComparisonSummary{
		CriticalIssues:   []string{},
		ImprovementAreas: []string{},
		RegressionAreas:  []string{},
	}

	// Calculate overall improvement score
	overallImprovement := 0.0
	factors := 0

	if comparisonData.VideoQuality != nil {
		overallImprovement += comparisonData.VideoQuality.QualityImprovement
		factors++
	}

	if comparisonData.AudioQuality != nil {
		overallImprovement += comparisonData.AudioQuality.QualityImprovement
		factors++
	}

	if factors > 0 {
		overallImprovement /= float64(factors)
	}

	// Adjust for file size changes
	if comparisonData.FileSize != nil {
		if comparisonData.FileSize.PercentageChange < -10 { // File size reduction
			overallImprovement += 10.0
		} else if comparisonData.FileSize.PercentageChange > 20 { // Significant size increase
			overallImprovement -= 15.0
		}
	}

	summary.OverallImprovement = math.Max(0, math.Min(100, overallImprovement))

	// Determine quality verdict
	if summary.OverallImprovement >= 30 {
		summary.QualityVerdict = models.VerdictSignificantImprovement
	} else if summary.OverallImprovement >= 10 {
		summary.QualityVerdict = models.VerdictImprovement
	} else if summary.OverallImprovement >= -10 {
		summary.QualityVerdict = models.VerdictMinimalChange
	} else if summary.OverallImprovement >= -30 {
		summary.QualityVerdict = models.VerdictRegression
	} else {
		summary.QualityVerdict = models.VerdictSignificantRegression
	}

	// Determine recommended action
	switch summary.QualityVerdict {
	case models.VerdictSignificantImprovement, models.VerdictImprovement:
		summary.RecommendedAction = models.ActionAccept
	case models.VerdictMinimalChange:
		if len(comparisonData.NewIssues) > 0 {
			summary.RecommendedAction = models.ActionReviewManually
		} else {
			summary.RecommendedAction = models.ActionAccept
		}
	case models.VerdictRegression:
		summary.RecommendedAction = models.ActionFurtherOptimize
	case models.VerdictSignificantRegression:
		summary.RecommendedAction = models.ActionReject
	}

	// Set compliance status (simplified)
	if len(comparisonData.NewIssues) == 0 {
		summary.ComplianceStatus = models.CompliancePass
	} else if len(comparisonData.NewIssues) <= 2 {
		summary.ComplianceStatus = models.ComplianceWarning
	} else {
		summary.ComplianceStatus = models.ComplianceFail
	}

	// Populate improvement and regression areas
	if len(comparisonData.IssuesFixed) > 0 {
		summary.ImprovementAreas = comparisonData.IssuesFixed
	}
	if len(comparisonData.NewIssues) > 0 {
		summary.RegressionAreas = comparisonData.NewIssues
	}

	return summary
}

// calculateQualityScore calculates overall quality scores
func (s *ComparisonService) calculateQualityScore(comparisonData *models.ComparisonData) *models.QualityScore {
	score := &models.QualityScore{}

	// Calculate video score
	videoScore := 50.0 // Base score
	if comparisonData.VideoQuality != nil {
		videoScore += comparisonData.VideoQuality.QualityImprovement
	}
	score.VideoScore = math.Max(0, math.Min(100, videoScore))

	// Calculate audio score
	audioScore := 50.0 // Base score
	if comparisonData.AudioQuality != nil {
		audioScore += comparisonData.AudioQuality.QualityImprovement
	}
	score.AudioScore = math.Max(0, math.Min(100, audioScore))

	// Calculate compression score
	compressionScore := 50.0 // Base score
	if comparisonData.FileSize != nil {
		if comparisonData.FileSize.PercentageChange < -10 {
			compressionScore += 20.0 // Good compression
		} else if comparisonData.FileSize.PercentageChange > 20 {
			compressionScore -= 20.0 // Poor compression
		}
	}
	score.CompressionScore = math.Max(0, math.Min(100, compressionScore))

	// Calculate compliance score
	complianceScore := 100.0
	if len(comparisonData.NewIssues) > 0 {
		complianceScore -= float64(len(comparisonData.NewIssues)) * 15.0
	}
	score.ComplianceScore = math.Max(0, math.Min(100, complianceScore))

	// Calculate overall score
	score.OverallScore = (score.VideoScore + score.AudioScore + score.CompressionScore + score.ComplianceScore) / 4.0

	return score
}

// Helper functions

func (s *ComparisonService) parseStreams(streamsData json.RawMessage) ([]map[string]interface{}, error) {
	var streams []map[string]interface{}
	if err := json.Unmarshal(streamsData, &streams); err != nil {
		return nil, err
	}
	return streams, nil
}

func (s *ComparisonService) parseFormat(formatData json.RawMessage) (map[string]interface{}, error) {
	var format map[string]interface{}
	if err := json.Unmarshal(formatData, &format); err != nil {
		return nil, err
	}
	return format, nil
}

func (s *ComparisonService) getVideoStream(streams []map[string]interface{}) map[string]interface{} {
	for _, stream := range streams {
		if codecType, ok := stream["codec_type"].(string); ok && codecType == "video" {
			return stream
		}
	}
	return nil
}

func (s *ComparisonService) getAudioStream(streams []map[string]interface{}) map[string]interface{} {
	for _, stream := range streams {
		if codecType, ok := stream["codec_type"].(string); ok && codecType == "audio" {
			return stream
		}
	}
	return nil
}

func (s *ComparisonService) getFloatValue(data map[string]interface{}, key string) float64 {
	if value, ok := data[key]; ok {
		switch v := value.(type) {
		case float64:
			return v
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		case int:
			return float64(v)
		}
	}
	return 0
}

func (s *ComparisonService) getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func (s *ComparisonService) compareMetric(original, modified float64, higherIsBetter bool) *models.MetricComparison {
	if original == 0 && modified == 0 {
		return nil
	}

	change := modified - original
	percentageChange := 0.0
	if original != 0 {
		percentageChange = (change / original) * 100
	}

	improvement := false
	if higherIsBetter {
		improvement = change > 0
	} else {
		improvement = change < 0
	}

	return &models.MetricComparison{
		Original:         original,
		Modified:         modified,
		Change:           change,
		PercentageChange: percentageChange,
		Improvement:      improvement,
	}
}

func (s *ComparisonService) compareResolution(originalVideo, modifiedVideo map[string]interface{}) *models.ResolutionChange {
	originalWidth := int(s.getFloatValue(originalVideo, "width"))
	originalHeight := int(s.getFloatValue(originalVideo, "height"))
	modifiedWidth := int(s.getFloatValue(modifiedVideo, "width"))
	modifiedHeight := int(s.getFloatValue(modifiedVideo, "height"))

	if originalWidth == 0 || originalHeight == 0 || modifiedWidth == 0 || modifiedHeight == 0 {
		return nil
	}

	scalingFactor := math.Sqrt(float64(modifiedWidth*modifiedHeight) / float64(originalWidth*originalHeight))

	originalAspectRatio := float64(originalWidth) / float64(originalHeight)
	modifiedAspectRatio := float64(modifiedWidth) / float64(modifiedHeight)
	aspectRatioChange := math.Abs(originalAspectRatio-modifiedAspectRatio) > 0.01

	return &models.ResolutionChange{
		OriginalWidth:     originalWidth,
		OriginalHeight:    originalHeight,
		ModifiedWidth:     modifiedWidth,
		ModifiedHeight:    modifiedHeight,
		ScalingFactor:     scalingFactor,
		AspectRatioChange: aspectRatioChange,
	}
}

func (s *ComparisonService) calculateBitrateEfficiency(bitrateChangePercent float64) float64 {
	// Simplified bitrate efficiency calculation
	// Positive efficiency means better quality per bit
	if bitrateChangePercent < -10 {
		return 20.0 // Good compression with maintained quality
	} else if bitrateChangePercent > 20 {
		return -15.0 // Poor efficiency
	}
	return 0.0
}

func (s *ComparisonService) isModernCodec(codec string, modernCodecs []string) bool {
	codec = strings.ToLower(codec)
	for _, modern := range modernCodecs {
		if strings.Contains(codec, modern) {
			return true
		}
	}
	return false
}

func (s *ComparisonService) markComparisonFailed(ctx context.Context, comparisonID uuid.UUID, errorMsg string) {
	comparison := &models.VideoComparison{
		ID:        comparisonID,
		Status:    models.ComparisonStatusFailed,
		ErrorMsg:  &errorMsg,
		UpdatedAt: time.Now(),
	}
	s.comparisonRepo.Update(ctx, comparison)
}

// generateLLMAssessment generates an AI-powered assessment of the comparison
func (s *ComparisonService) generateLLMAssessment(
	ctx context.Context,
	original, modified *models.Analysis,
	comparisonData *models.ComparisonData,
) (string, error) {
	prompt := s.buildComparisonPrompt(original, modified, comparisonData)

	if s.llmService == nil {
		return "", fmt.Errorf("LLM service not available")
	}

	response, err := s.llmService.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate LLM assessment: %w", err)
	}

	return response, nil
}

// buildComparisonPrompt builds the prompt for LLM assessment
func (s *ComparisonService) buildComparisonPrompt(
	original, modified *models.Analysis,
	comparisonData *models.ComparisonData,
) string {
	prompt := fmt.Sprintf(`You are a senior video engineer analyzing the results of video optimization work. 

COMPARISON ANALYSIS:
Original file: %s (%.2f MB)
Modified file: %s (%.2f MB)

COMPARISON RESULTS:
`, original.FileName, float64(original.FileSize)/(1024*1024), modified.FileName, float64(modified.FileSize)/(1024*1024))

	if comparisonData.FileSize != nil {
		prompt += fmt.Sprintf("File Size Change: %.1f%% (%s %.2f MB)\n",
			comparisonData.FileSize.PercentageChange,
			func() string {
				if comparisonData.FileSize.SizeChange > 0 {
					return "increased by"
				} else {
					return "reduced by"
				}
			}(),
			math.Abs(float64(comparisonData.FileSize.SizeChange))/(1024*1024))
	}

	if comparisonData.VideoQuality != nil {
		prompt += fmt.Sprintf("Video Quality Improvement: %.1f%%\n", comparisonData.VideoQuality.QualityImprovement)
	}

	if comparisonData.AudioQuality != nil {
		prompt += fmt.Sprintf("Audio Quality Improvement: %.1f%%\n", comparisonData.AudioQuality.QualityImprovement)
	}

	if len(comparisonData.IssuesFixed) > 0 {
		prompt += fmt.Sprintf("Issues Fixed: %s\n", strings.Join(comparisonData.IssuesFixed, ", "))
	}

	if len(comparisonData.NewIssues) > 0 {
		prompt += fmt.Sprintf("New Issues: %s\n", strings.Join(comparisonData.NewIssues, ", "))
	}

	if comparisonData.Summary != nil {
		prompt += fmt.Sprintf("Overall Assessment: %s\n", comparisonData.Summary.QualityVerdict)
		prompt += fmt.Sprintf("Recommended Action: %s\n", comparisonData.Summary.RecommendedAction)
	}

	prompt += `
Please provide a professional assessment covering:

1. **Quality Assessment**: How does the modified version compare to the original in terms of visual and audio quality?

2. **Compression Efficiency**: Was the file size change justified by the quality changes?

3. **Technical Issues**: Were the originally identified issues properly addressed? Are there any new concerns?

4. **Compliance & Standards**: Does the modified version meet industry standards and best practices?

5. **Production Readiness**: Is the modified version suitable for production use?

6. **Optimization Recommendations**: What further optimizations could be considered?

7. **Business Impact**: How do these changes affect workflow efficiency and storage costs?

8. **Final Verdict**: Should this modification be accepted, rejected, or require further work?

Provide specific, actionable insights based on the technical data and focus on practical business decisions.`

	return prompt
}

// GetComparison retrieves a comparison by ID
func (s *ComparisonService) GetComparison(ctx context.Context, id uuid.UUID) (*models.ComparisonResponse, error) {
	comparison, err := s.comparisonRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comparison: %w", err)
	}

	return &models.ComparisonResponse{
		ID:                 comparison.ID,
		OriginalAnalysisID: comparison.OriginalAnalysisID,
		ModifiedAnalysisID: comparison.ModifiedAnalysisID,
		ComparisonType:     comparison.ComparisonType,
		Status:             comparison.Status,
		ComparisonData:     &comparison.ComparisonData,
		LLMAssessment:      comparison.LLMAssessment,
		QualityScore:       comparison.QualityScore,
		CreatedAt:          comparison.CreatedAt,
		UpdatedAt:          comparison.UpdatedAt,
		ErrorMsg:           comparison.ErrorMsg,
	}, nil
}

// ListComparisons lists comparisons for a user
func (s *ComparisonService) ListComparisons(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]*models.ComparisonSummaryResponse, error) {
	comparisons, err := s.comparisonRepo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list comparisons: %w", err)
	}

	responses := make([]*models.ComparisonSummaryResponse, len(comparisons))
	for i, comparison := range comparisons {
		processingTime := comparison.UpdatedAt.Sub(comparison.CreatedAt)

		summary := &models.ComparisonSummaryResponse{
			ID:             comparison.ID,
			IssuesFixed:    len(comparison.ComparisonData.IssuesFixed),
			NewIssues:      len(comparison.ComparisonData.NewIssues),
			QualityScore:   comparison.QualityScore,
			ProcessingTime: processingTime,
			CreatedAt:      comparison.CreatedAt,
		}

		if comparison.ComparisonData.Summary != nil {
			summary.OverallImprovement = comparison.ComparisonData.Summary.OverallImprovement
			summary.QualityVerdict = comparison.ComparisonData.Summary.QualityVerdict
			summary.RecommendedAction = comparison.ComparisonData.Summary.RecommendedAction
		}

		if comparison.ComparisonData.FileSize != nil {
			summary.FileSizeChange = comparison.ComparisonData.FileSize
		}

		responses[i] = summary
	}

	return responses, nil
}
