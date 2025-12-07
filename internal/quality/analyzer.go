package quality

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// QualityAnalyzer handles video quality analysis operations
type QualityAnalyzer struct {
	ffmpegPath string
	tempDir    string
	logger     zerolog.Logger
	thresholds QualityThresholds
}

// NewQualityAnalyzer creates a new quality analyzer
func NewQualityAnalyzer(ffmpegPath string, logger zerolog.Logger) *QualityAnalyzer {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}

	return &QualityAnalyzer{
		ffmpegPath: ffmpegPath,
		tempDir:    "/tmp/quality",
		logger:     logger,
		thresholds: DefaultQualityThresholds(),
	}
}

// SetTempDirectory sets the temporary directory for analysis files
func (qa *QualityAnalyzer) SetTempDirectory(dir string) {
	qa.tempDir = dir
}

// SetThresholds sets custom quality thresholds
func (qa *QualityAnalyzer) SetThresholds(thresholds QualityThresholds) {
	qa.thresholds = thresholds
}

// AnalyzeQuality performs quality analysis between reference and distorted videos
func (qa *QualityAnalyzer) AnalyzeQuality(ctx context.Context, request *QualityComparisonRequest) (*QualityResult, error) {
	analysisID := uuid.New()
	startTime := time.Now()

	qa.logger.Info().
		Str("analysis_id", analysisID.String()).
		Str("reference", request.ReferenceFile).
		Str("distorted", request.DistortedFile).
		Strs("metrics", metricTypesToStrings(request.Metrics)).
		Msg("Starting quality analysis")

	// Validate input files
	if err := qa.validateInputFiles(request.ReferenceFile, request.DistortedFile); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	result := &QualityResult{
		ID:             analysisID,
		Status:         QualityStatusProcessing,
		Analysis:       make([]*QualityAnalysis, 0, len(request.Metrics)),
		ProcessingTime: 0,
	}

	// Process each requested metric
	for _, metric := range request.Metrics {
		analysis, err := qa.analyzeMetric(ctx, request, metric, analysisID)
		if err != nil {
			qa.logger.Error().
				Err(err).
				Str("analysis_id", analysisID.String()).
				Str("metric", string(metric)).
				Msg("Metric analysis failed")

			result.Status = QualityStatusFailed
			result.Error = fmt.Sprintf("Failed to analyze %s: %v", metric, err)
			return result, err
		}

		result.Analysis = append(result.Analysis, analysis)
	}

	// Generate summary and insights
	summary, err := qa.generateSummary(result.Analysis, request.ReferenceFile, request.DistortedFile)
	if err != nil {
		qa.logger.Warn().Err(err).Msg("Failed to generate summary")
	} else {
		result.Summary = summary
	}

	// Generate visualization data if requested
	if request.FrameLevel {
		visualization, err := qa.generateVisualization(result.Analysis)
		if err != nil {
			qa.logger.Warn().Err(err).Msg("Failed to generate visualization")
		} else {
			result.Visualization = visualization
		}
	}

	result.Status = QualityStatusCompleted
	result.ProcessingTime = time.Since(startTime)
	result.Message = "Quality analysis completed successfully"

	qa.logger.Info().
		Str("analysis_id", analysisID.String()).
		Dur("processing_time", result.ProcessingTime).
		Msg("Quality analysis completed")

	return result, nil
}

// analyzeMetric performs analysis for a specific quality metric
func (qa *QualityAnalyzer) analyzeMetric(ctx context.Context, request *QualityComparisonRequest, metric QualityMetricType, analysisID uuid.UUID) (*QualityAnalysis, error) {
	analysis := &QualityAnalysis{
		ID:            uuid.New(),
		AnalysisID:    analysisID,
		ReferenceFile: request.ReferenceFile,
		DistortedFile: request.DistortedFile,
		MetricType:    metric,
		Status:        QualityStatusProcessing,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	startTime := time.Now()

	switch metric {
	case MetricVMAF:
		if err := qa.analyzeVMAF(ctx, analysis, request.Configuration.VMAF); err != nil {
			return nil, err
		}
	case MetricPSNR:
		if err := qa.analyzePSNR(ctx, analysis, request.Configuration.PSNR); err != nil {
			return nil, err
		}
	case MetricSSIM:
		if err := qa.analyzeSSIM(ctx, analysis, request.Configuration.SSIM); err != nil {
			return nil, err
		}
	case MetricMSE:
		if err := qa.analyzeMSE(ctx, analysis); err != nil {
			return nil, err
		}
	case MetricMSSSIM:
		if err := qa.analyzeMSSSIM(ctx, analysis, request.Configuration.SSIM); err != nil {
			return nil, err
		}
	case MetricLPIPS:
		if err := qa.analyzeLPIPS(ctx, analysis); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", metric)
	}

	analysis.ProcessingTime = time.Since(startTime)
	analysis.Status = QualityStatusCompleted
	analysis.UpdatedAt = time.Now()
	completedAt := time.Now()
	analysis.CompletedAt = &completedAt

	return analysis, nil
}

// analyzeVMAF performs VMAF analysis
func (qa *QualityAnalyzer) analyzeVMAF(ctx context.Context, analysis *QualityAnalysis, config *VMAFConfiguration) error {
	// Set default VMAF configuration
	if config == nil {
		config = &VMAFConfiguration{
			Model:         "version=vmaf_v0.6.1",
			SubSampling:   1,
			PoolingMethod: "mean",
			NThreads:      4,
			OutputFormat:  "json",
			LogLevel:      "info",
		}
	}

	// Check if a custom model path is specified
	if config.CustomModelPath != "" {
		// Validate custom model path
		if _, err := os.Stat(config.CustomModelPath); err != nil {
			return fmt.Errorf("custom VMAF model not found: %w", err)
		}
		// Use custom model path
		config.Model = fmt.Sprintf("path=%s", config.CustomModelPath)
	}

	// Ensure temp directory exists
	if err := os.MkdirAll(qa.tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create temporary output file
	outputFile := filepath.Join(qa.tempDir, fmt.Sprintf("vmaf_%s.json", analysis.ID.String()))
	defer os.Remove(outputFile)

	// Build VMAF filter
	vmafFilter := fmt.Sprintf(
		"[0:v]scale=1920:1080:flags=bicubic[ref];[1:v]scale=1920:1080:flags=bicubic[dist];[dist][ref]libvmaf=model=%s:log_fmt=%s:log_path=%s:n_threads=%d:n_subsample=%d",
		config.Model,
		config.OutputFormat,
		outputFile,
		config.NThreads,
		config.SubSampling,
	)

	// Build FFmpeg command
	cmd := exec.CommandContext(ctx,
		qa.ffmpegPath,
		"-i", analysis.DistortedFile,
		"-i", analysis.ReferenceFile,
		"-lavfi", vmafFilter,
		"-f", "null",
		"-",
	)

	qa.logger.Debug().
		Str("analysis_id", analysis.ID.String()).
		Str("command", cmd.String()).
		Msg("Executing VMAF analysis")

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		qa.logger.Error().
			Err(err).
			Str("output", string(output)).
			Msg("VMAF analysis failed")
		return fmt.Errorf("VMAF analysis failed: %w", err)
	}

	// Parse VMAF output
	return qa.parseVMAFOutput(analysis, outputFile, config)
}

// analyzePSNR performs PSNR analysis
func (qa *QualityAnalyzer) analyzePSNR(ctx context.Context, analysis *QualityAnalysis, config *PSNRConfiguration) error {
	// Set default PSNR configuration
	if config == nil {
		config = &PSNRConfiguration{
			ComponentMask: "Y",
			Stats:         true,
			OutputFormat:  "json",
		}
	}

	// Build PSNR filter
	psrFilter := "[0:v][1:v]psnr=stats_file=-"

	// Build FFmpeg command
	cmd := exec.CommandContext(ctx,
		qa.ffmpegPath,
		"-i", analysis.ReferenceFile,
		"-i", analysis.DistortedFile,
		"-lavfi", psrFilter,
		"-f", "null",
		"-",
	)

	qa.logger.Debug().
		Str("analysis_id", analysis.ID.String()).
		Str("command", cmd.String()).
		Msg("Executing PSNR analysis")

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("PSNR analysis failed: %w", err)
	}

	// Parse PSNR output
	return qa.parsePSNROutput(analysis, string(output), config)
}

// analyzeSSIM performs SSIM analysis
func (qa *QualityAnalyzer) analyzeSSIM(ctx context.Context, analysis *QualityAnalysis, config *SSIMConfiguration) error {
	// Set default SSIM configuration
	if config == nil {
		config = &SSIMConfiguration{
			WindowSize:   11,
			K1:           0.01,
			K2:           0.03,
			Stats:        true,
			OutputFormat: "json",
		}
	}

	// Build SSIM filter
	ssimFilter := "[0:v][1:v]ssim=stats_file=-"

	// Build FFmpeg command
	cmd := exec.CommandContext(ctx,
		qa.ffmpegPath,
		"-i", analysis.ReferenceFile,
		"-i", analysis.DistortedFile,
		"-lavfi", ssimFilter,
		"-f", "null",
		"-",
	)

	qa.logger.Debug().
		Str("analysis_id", analysis.ID.String()).
		Str("command", cmd.String()).
		Msg("Executing SSIM analysis")

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("SSIM analysis failed: %w", err)
	}

	// Parse SSIM output
	return qa.parseSSIMOutput(analysis, string(output), config)
}

// parseVMAFOutput parses VMAF JSON output
func (qa *QualityAnalyzer) parseVMAFOutput(analysis *QualityAnalysis, outputFile string, config *VMAFConfiguration) error {
	data, err := os.ReadFile(outputFile)
	if err != nil {
		return fmt.Errorf("failed to read VMAF output: %w", err)
	}

	var vmafResult struct {
		PooledMetrics struct {
			VMAF struct {
				Mean     float64 `json:"mean"`
				Min      float64 `json:"min"`
				Max      float64 `json:"max"`
				Std      float64 `json:"std"`
				Harmonic float64 `json:"harmonic_mean"`
			} `json:"vmaf"`
		} `json:"pooled_metrics"`
		AggregateMetrics struct {
			VMAF struct {
				Mean     float64 `json:"mean"`
				Min      float64 `json:"min"`
				Max      float64 `json:"max"`
				Std      float64 `json:"std"`
				Harmonic float64 `json:"harmonic_mean"`
			} `json:"vmaf"`
		} `json:"aggregate_metrics"`
	}

	if err := json.Unmarshal(data, &vmafResult); err != nil {
		return fmt.Errorf("failed to parse VMAF output: %w", err)
	}

	// Use pooled metrics if available, otherwise aggregate
	vmafMetrics := vmafResult.PooledMetrics.VMAF
	if vmafMetrics.Mean == 0 {
		vmafMetrics = vmafResult.AggregateMetrics.VMAF
	}

	analysis.OverallScore = vmafMetrics.Mean
	analysis.MinScore = vmafMetrics.Min
	analysis.MaxScore = vmafMetrics.Max
	analysis.MeanScore = vmafMetrics.Mean
	analysis.StdDevScore = vmafMetrics.Std

	// Store configuration
	configData, _ := json.Marshal(config)
	analysis.Configuration = configData

	return nil
}

// parsePSNROutput parses PSNR output
func (qa *QualityAnalyzer) parsePSNROutput(analysis *QualityAnalysis, output string, config *PSNRConfiguration) error {
	lines := strings.Split(output, "\n")

	var psnrValues []float64
	for _, line := range lines {
		if strings.Contains(line, "PSNR") && strings.Contains(line, "average") {
			// Extract PSNR value from line like: "PSNR y:42.123456 u:44.123456 v:45.123456 average:43.123456"
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "average:") {
					valueStr := strings.TrimPrefix(part, "average:")
					if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
						psnrValues = append(psnrValues, value)
					}
				}
			}
		}
	}

	if len(psnrValues) == 0 {
		return fmt.Errorf("no PSNR values found in output")
	}

	// Calculate statistics
	analysis.OverallScore = calculateMean(psnrValues)
	analysis.MinScore = calculateMin(psnrValues)
	analysis.MaxScore = calculateMax(psnrValues)
	analysis.MeanScore = analysis.OverallScore
	analysis.StdDevScore = calculateStdDev(psnrValues)

	// Store configuration
	configData, _ := json.Marshal(config)
	analysis.Configuration = configData

	return nil
}

// parseSSIMOutput parses SSIM output
func (qa *QualityAnalyzer) parseSSIMOutput(analysis *QualityAnalysis, output string, config *SSIMConfiguration) error {
	lines := strings.Split(output, "\n")

	var ssimValues []float64
	for _, line := range lines {
		if strings.Contains(line, "SSIM") && strings.Contains(line, "All:") {
			// Extract SSIM value from line like: "SSIM Y:0.95 U:0.96 V:0.97 All:0.96"
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "All:") {
					valueStr := strings.TrimPrefix(part, "All:")
					if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
						ssimValues = append(ssimValues, value)
					}
				}
			}
		}
	}

	if len(ssimValues) == 0 {
		return fmt.Errorf("no SSIM values found in output")
	}

	// Calculate statistics
	analysis.OverallScore = calculateMean(ssimValues)
	analysis.MinScore = calculateMin(ssimValues)
	analysis.MaxScore = calculateMax(ssimValues)
	analysis.MeanScore = analysis.OverallScore
	analysis.StdDevScore = calculateStdDev(ssimValues)

	// Store configuration
	configData, _ := json.Marshal(config)
	analysis.Configuration = configData

	return nil
}

// generateSummary creates a human-readable summary of quality analysis
func (qa *QualityAnalyzer) generateSummary(analyses []*QualityAnalysis, referenceFile, distortedFile string) (*QualitySummary, error) {
	summary := &QualitySummary{
		ReferenceFile:   referenceFile,
		DistortedFile:   distortedFile,
		MetricSummaries: make(map[QualityMetricType]*MetricSummary),
		Recommendations: make([]string, 0),
		QualityIssues:   make([]QualityIssue, 0),
	}

	var totalRatingScore float64
	var ratingCount int

	// Process each analysis
	for _, analysis := range analyses {
		rating := qa.thresholds.GetRating(analysis.MetricType, analysis.OverallScore)

		metricSummary := &MetricSummary{
			MetricType:     analysis.MetricType,
			Score:          analysis.OverallScore,
			Rating:         rating,
			Description:    qa.getMetricDescription(analysis.MetricType),
			Interpretation: qa.getScoreInterpretation(analysis.MetricType, analysis.OverallScore, rating),
		}

		summary.MetricSummaries[analysis.MetricType] = metricSummary

		// Calculate weighted rating
		weight := qa.getMetricWeight(analysis.MetricType)
		totalRatingScore += qa.getRatingScore(rating) * weight
		ratingCount++

		// Detect quality issues
		issues := qa.detectQualityIssues(analysis)
		summary.QualityIssues = append(summary.QualityIssues, issues...)
	}

	// Calculate overall rating
	if ratingCount > 0 {
		avgRatingScore := totalRatingScore / float64(ratingCount)
		summary.OverallRating = qa.scoreToRating(avgRatingScore)
	}

	// Generate recommendations
	summary.Recommendations = qa.generateRecommendations(analyses, summary.QualityIssues)

	// Generate comparison insights
	summary.ComparisonInsights = qa.generateComparisonInsights(analyses)

	return summary, nil
}

// generateVisualization creates visualization data for quality metrics
func (qa *QualityAnalyzer) generateVisualization(analyses []*QualityAnalysis) (*QualityVisualization, error) {
	visualization := &QualityVisualization{}

	// Generate chart data for time series
	chartData := make(map[string]interface{})
	chartData["labels"] = []string{} // Time labels
	chartData["datasets"] = []map[string]interface{}{}

	for _, analysis := range analyses {
		dataset := map[string]interface{}{
			"label":           string(analysis.MetricType),
			"data":            []float64{analysis.OverallScore},
			"borderColor":     qa.getMetricColor(analysis.MetricType),
			"backgroundColor": qa.getMetricColor(analysis.MetricType),
		}
		chartData["datasets"] = append(chartData["datasets"].([]map[string]interface{}), dataset)
	}

	chartDataJSON, _ := json.Marshal(chartData)
	visualization.ChartData = chartDataJSON

	return visualization, nil
}

// Helper functions for quality analysis

func (qa *QualityAnalyzer) validateInputFiles(reference, distorted string) error {
	if _, err := os.Stat(reference); err != nil {
		return fmt.Errorf("reference file not accessible: %w", err)
	}
	if _, err := os.Stat(distorted); err != nil {
		return fmt.Errorf("distorted file not accessible: %w", err)
	}
	return nil
}

func (qa *QualityAnalyzer) getMetricDescription(metric QualityMetricType) string {
	descriptions := map[QualityMetricType]string{
		MetricVMAF: "Video Multi-Method Assessment Fusion - perceptual video quality metric",
		MetricPSNR: "Peak Signal-to-Noise Ratio - objective quality metric in dB",
		MetricSSIM: "Structural Similarity Index - perceptual quality metric",
	}
	return descriptions[metric]
}

func (qa *QualityAnalyzer) getScoreInterpretation(metric QualityMetricType, score float64, rating QualityRating) string {
	base := fmt.Sprintf("Score: %.2f - %s quality", score, rating)

	switch metric {
	case MetricVMAF:
		return fmt.Sprintf("%s. VMAF scores range from 0-100, with higher values indicating better perceptual quality.", base)
	case MetricPSNR:
		return fmt.Sprintf("%s. PSNR is measured in dB, with higher values indicating less distortion.", base)
	case MetricSSIM:
		return fmt.Sprintf("%s. SSIM ranges from 0-1, with values closer to 1 indicating better structural similarity.", base)
	}
	return base
}

func (qa *QualityAnalyzer) getMetricWeight(metric QualityMetricType) float64 {
	weights := map[QualityMetricType]float64{
		MetricVMAF: 1.0, // Highest weight as it's most perceptual
		MetricSSIM: 0.8,
		MetricPSNR: 0.6,
	}
	if weight, exists := weights[metric]; exists {
		return weight
	}
	return 1.0
}

func (qa *QualityAnalyzer) getRatingScore(rating QualityRating) float64 {
	scores := map[QualityRating]float64{
		RatingExcellent: 95.0,
		RatingGood:      85.0,
		RatingFair:      75.0,
		RatingPoor:      60.0,
		RatingBad:       40.0,
	}
	return scores[rating]
}

func (qa *QualityAnalyzer) scoreToRating(score float64) QualityRating {
	if score >= 90 {
		return RatingExcellent
	} else if score >= 80 {
		return RatingGood
	} else if score >= 70 {
		return RatingFair
	} else if score >= 50 {
		return RatingPoor
	}
	return RatingBad
}

func (qa *QualityAnalyzer) detectQualityIssues(analysis *QualityAnalysis) []QualityIssue {
	var issues []QualityIssue

	rating := qa.thresholds.GetRating(analysis.MetricType, analysis.OverallScore)

	if rating == RatingPoor || rating == RatingBad {
		severity := "medium"
		if rating == RatingBad {
			severity = "high"
		}

		issue := QualityIssue{
			Type:        qa.getIssueType(analysis.MetricType),
			Severity:    severity,
			Description: fmt.Sprintf("Low %s score detected: %.2f", analysis.MetricType, analysis.OverallScore),
			Score:       analysis.OverallScore,
		}
		issues = append(issues, issue)
	}

	return issues
}

func (qa *QualityAnalyzer) getIssueType(metric QualityMetricType) string {
	types := map[QualityMetricType]string{
		MetricVMAF: "perceptual_quality",
		MetricPSNR: "signal_distortion",
		MetricSSIM: "structural_similarity",
	}
	return types[metric]
}

func (qa *QualityAnalyzer) generateRecommendations(analyses []*QualityAnalysis, issues []QualityIssue) []string {
	var recommendations []string

	for _, analysis := range analyses {
		rating := qa.thresholds.GetRating(analysis.MetricType, analysis.OverallScore)

		if rating == RatingPoor || rating == RatingBad {
			switch analysis.MetricType {
			case MetricVMAF:
				recommendations = append(recommendations, "Consider increasing bitrate or using a higher quality encoding profile")
			case MetricPSNR:
				recommendations = append(recommendations, "Reduce quantization parameter or increase bitrate to reduce signal distortion")
			case MetricSSIM:
				recommendations = append(recommendations, "Check for preprocessing artifacts or encoding settings that may affect structural similarity")
			}
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Video quality is within acceptable ranges")
	}

	return recommendations
}

func (qa *QualityAnalyzer) generateComparisonInsights(analyses []*QualityAnalysis) string {
	if len(analyses) == 0 {
		return "No analysis data available"
	}

	var insights []string
	for _, analysis := range analyses {
		rating := qa.thresholds.GetRating(analysis.MetricType, analysis.OverallScore)
		insights = append(insights, fmt.Sprintf("%s: %.2f (%s)", analysis.MetricType, analysis.OverallScore, rating))
	}

	return "Quality comparison results: " + strings.Join(insights, ", ")
}

func (qa *QualityAnalyzer) getMetricColor(metric QualityMetricType) string {
	colors := map[QualityMetricType]string{
		MetricVMAF: "#FF6B6B",
		MetricPSNR: "#4ECDC4",
		MetricSSIM: "#45B7D1",
	}
	return colors[metric]
}

// Statistical helper functions

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func calculateStdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := calculateMean(values)
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	variance := sumSquares / float64(len(values))
	return math.Sqrt(variance)
}

// sortFloat64s sorts a slice of float64 in ascending order
func sortFloat64s(values []float64) {
	sort.Float64s(values)
}

// percentile calculates the p-th percentile of sorted values
func percentile(sortedValues []float64, p float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}
	if p <= 0 {
		return sortedValues[0]
	}
	if p >= 100 {
		return sortedValues[len(sortedValues)-1]
	}
	index := (p / 100) * float64(len(sortedValues)-1)
	lower := int(index)
	upper := lower + 1
	if upper >= len(sortedValues) {
		return sortedValues[lower]
	}
	fraction := index - float64(lower)
	return sortedValues[lower]*(1-fraction) + sortedValues[upper]*fraction
}

func metricTypesToStrings(metrics []QualityMetricType) []string {
	strings := make([]string, len(metrics))
	for i, metric := range metrics {
		strings[i] = string(metric)
	}
	return strings
}

// analyzeMSE performs Mean Squared Error analysis using FFmpeg's psnr filter
func (qa *QualityAnalyzer) analyzeMSE(ctx context.Context, analysis *QualityAnalysis) error {
	// MSE is calculated alongside PSNR using FFmpeg's psnr filter
	// The filter outputs mse_avg, mse_y, mse_u, mse_v values

	// Build PSNR/MSE filter - psnr filter outputs both PSNR and MSE
	psnrFilter := "[0:v][1:v]psnr=stats_file=-"

	// Build FFmpeg command
	cmd := exec.CommandContext(ctx,
		qa.ffmpegPath,
		"-i", analysis.ReferenceFile,
		"-i", analysis.DistortedFile,
		"-lavfi", psnrFilter,
		"-f", "null",
		"-",
	)

	qa.logger.Debug().
		Str("analysis_id", analysis.ID.String()).
		Str("command", cmd.String()).
		Msg("Executing MSE analysis")

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("MSE analysis failed: %w", err)
	}

	// Parse MSE output from PSNR filter output
	return qa.parseMSEOutput(analysis, string(output))
}

// parseMSEOutput parses MSE values from PSNR filter output
func (qa *QualityAnalyzer) parseMSEOutput(analysis *QualityAnalysis, output string) error {
	lines := strings.Split(output, "\n")

	var mseValues []float64
	for _, line := range lines {
		// Look for lines containing mse_avg or extract MSE from PSNR output
		// PSNR output format: "n:X mse_avg:Y mse_y:Z mse_u:W mse_v:V"
		if strings.Contains(line, "mse_avg:") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "mse_avg:") {
					valueStr := strings.TrimPrefix(part, "mse_avg:")
					if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
						mseValues = append(mseValues, value)
					}
				}
			}
		}
		// Alternative: Calculate MSE from PSNR (MSE = 255^2 / 10^(PSNR/10))
		if strings.Contains(line, "PSNR") && strings.Contains(line, "average:") && len(mseValues) == 0 {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "average:") {
					valueStr := strings.TrimPrefix(part, "average:")
					if psnr, err := strconv.ParseFloat(valueStr, 64); err == nil && psnr > 0 {
						// Convert PSNR to MSE: MSE = MAX^2 / 10^(PSNR/10) where MAX=255
						mse := math.Pow(255, 2) / math.Pow(10, psnr/10)
						mseValues = append(mseValues, mse)
					}
				}
			}
		}
	}

	if len(mseValues) == 0 {
		return fmt.Errorf("no MSE values found in output")
	}

	// Calculate statistics
	analysis.OverallScore = calculateMean(mseValues)
	analysis.MinScore = calculateMin(mseValues)
	analysis.MaxScore = calculateMax(mseValues)
	analysis.MeanScore = analysis.OverallScore
	analysis.StdDevScore = calculateStdDev(mseValues)

	// Calculate percentiles
	sortedValues := make([]float64, len(mseValues))
	copy(sortedValues, mseValues)
	sortFloat64s(sortedValues)
	analysis.Percentile1 = percentile(sortedValues, 1)
	analysis.Percentile5 = percentile(sortedValues, 5)
	analysis.Percentile95 = percentile(sortedValues, 95)
	analysis.Percentile99 = percentile(sortedValues, 99)

	return nil
}

// analyzeMSSSIM performs MS-SSIM (Multi-Scale SSIM) analysis
func (qa *QualityAnalyzer) analyzeMSSSIM(ctx context.Context, analysis *QualityAnalysis, config *SSIMConfiguration) error {
	// MS-SSIM (Multi-Scale SSIM) implementation using FFmpeg
	// MS-SSIM computes SSIM at multiple scales and combines them with specific weights
	// Standard scales: 1, 1/2, 1/4, 1/8, 1/16 with weights: 0.0448, 0.2856, 0.3001, 0.2363, 0.1333

	scales := []float64{1.0, 0.5, 0.25, 0.125, 0.0625}
	weights := []float64{0.0448, 0.2856, 0.3001, 0.2363, 0.1333}

	var scaleSSIMValues [][]float64
	for scaleIdx, scale := range scales {
		ssimValues, err := qa.computeSSIMAtScale(ctx, analysis.ReferenceFile, analysis.DistortedFile, scale)
		if err != nil {
			qa.logger.Warn().
				Err(err).
				Float64("scale", scale).
				Msg("Failed to compute SSIM at scale, skipping")
			continue
		}
		if len(ssimValues) > 0 {
			scaleSSIMValues = append(scaleSSIMValues, ssimValues)
		}

		qa.logger.Debug().
			Int("scale_idx", scaleIdx).
			Float64("scale", scale).
			Int("frame_count", len(ssimValues)).
			Msg("Computed SSIM at scale")
	}

	if len(scaleSSIMValues) == 0 {
		return fmt.Errorf("no SSIM values computed at any scale")
	}

	// Compute MS-SSIM by combining scales with weights
	// For each frame, compute weighted product of SSIM across scales
	minFrames := len(scaleSSIMValues[0])
	for _, sv := range scaleSSIMValues {
		if len(sv) < minFrames {
			minFrames = len(sv)
		}
	}

	msssimValues := make([]float64, minFrames)
	for frameIdx := 0; frameIdx < minFrames; frameIdx++ {
		// MS-SSIM = product of (SSIM_scale ^ weight_scale) across all scales
		msssim := 1.0
		totalWeight := 0.0
		for scaleIdx := 0; scaleIdx < len(scaleSSIMValues); scaleIdx++ {
			weight := weights[scaleIdx]
			ssimVal := scaleSSIMValues[scaleIdx][frameIdx]
			// Ensure SSIM is positive for power calculation
			if ssimVal > 0 {
				msssim *= math.Pow(ssimVal, weight)
				totalWeight += weight
			}
		}
		// Normalize if not all scales contributed
		if totalWeight > 0 && totalWeight < 1.0 {
			msssim = math.Pow(msssim, 1.0/totalWeight)
		}
		msssimValues[frameIdx] = msssim
	}

	// Calculate statistics
	analysis.OverallScore = calculateMean(msssimValues)
	analysis.MinScore = calculateMin(msssimValues)
	analysis.MaxScore = calculateMax(msssimValues)
	analysis.MeanScore = analysis.OverallScore
	analysis.StdDevScore = calculateStdDev(msssimValues)

	// Calculate percentiles
	sortedValues := make([]float64, len(msssimValues))
	copy(sortedValues, msssimValues)
	sortFloat64s(sortedValues)
	analysis.Percentile1 = percentile(sortedValues, 1)
	analysis.Percentile5 = percentile(sortedValues, 5)
	analysis.Percentile95 = percentile(sortedValues, 95)
	analysis.Percentile99 = percentile(sortedValues, 99)

	qa.logger.Info().
		Float64("ms_ssim_mean", analysis.MeanScore).
		Float64("ms_ssim_min", analysis.MinScore).
		Float64("ms_ssim_max", analysis.MaxScore).
		Int("scales_used", len(scaleSSIMValues)).
		Msg("MS-SSIM analysis completed")

	return nil
}

// computeSSIMAtScale computes SSIM at a specific scale by downscaling the videos
func (qa *QualityAnalyzer) computeSSIMAtScale(ctx context.Context, refFile, distFile string, scale float64) ([]float64, error) {
	// Build filter for scaling and SSIM computation
	var filterComplex string
	if scale < 1.0 {
		// Downscale both inputs before comparing
		scaleFilter := fmt.Sprintf("scale=iw*%.4f:ih*%.4f:flags=lanczos", scale, scale)
		filterComplex = fmt.Sprintf("[0:v]%s[ref];[1:v]%s[dist];[ref][dist]ssim=stats_file=-", scaleFilter, scaleFilter)
	} else {
		filterComplex = "[0:v][1:v]ssim=stats_file=-"
	}

	cmd := exec.CommandContext(ctx,
		qa.ffmpegPath,
		"-i", refFile,
		"-i", distFile,
		"-filter_complex", filterComplex,
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("SSIM at scale %.4f failed: %w", scale, err)
	}

	return qa.extractSSIMValues(string(output))
}

// extractSSIMValues extracts raw SSIM values from FFmpeg output
func (qa *QualityAnalyzer) extractSSIMValues(output string) ([]float64, error) {
	var ssimValues []float64

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// SSIM output format: "n:X Y:0.XX U:0.XX V:0.XX All:0.XX (XX.XX)"
		if strings.Contains(line, "All:") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "All:") {
					valueStr := strings.TrimPrefix(part, "All:")
					if ssim, err := strconv.ParseFloat(valueStr, 64); err == nil {
						ssimValues = append(ssimValues, ssim)
					}
				}
			}
		}
	}

	if len(ssimValues) == 0 {
		return nil, fmt.Errorf("no SSIM values found in output")
	}

	return ssimValues, nil
}

// analyzeLPIPS performs LPIPS (Learned Perceptual Image Patch Similarity) analysis
// LPIPS requires a pre-trained neural network model (VGG, AlexNet, or SqueezeNet)
// This cannot be implemented with FFmpeg alone - requires PyTorch/TensorFlow with LPIPS library
func (qa *QualityAnalyzer) analyzeLPIPS(ctx context.Context, analysis *QualityAnalysis) error {
	// LPIPS (Learned Perceptual Image Patch Similarity) is a deep learning-based metric
	// that requires a pre-trained neural network model to compute perceptual similarity.
	//
	// To use LPIPS, you would need:
	// 1. Python with torch and lpips packages installed
	// 2. Pre-trained model weights (VGG, AlexNet, or SqueezeNet)
	// 3. GPU acceleration (recommended for performance)
	//
	// Since FFmpeg cannot compute LPIPS, we return an error indicating this metric
	// requires external tooling.

	qa.logger.Warn().
		Str("analysis_id", analysis.ID.String()).
		Msg("LPIPS analysis requires external neural network model - not available")

	return fmt.Errorf("LPIPS analysis unavailable: requires PyTorch with lpips package and pre-trained model. " +
		"Install with: pip install lpips torch, then configure external LPIPS analyzer")
}
