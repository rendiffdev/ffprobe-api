package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/database"
	"github.com/rendiffdev/ffprobe-api/internal/quality"
)

// QualityService handles video quality analysis operations
type QualityService struct {
	analyzer         *quality.QualityAnalyzer
	qualityRepo      database.QualityRepository
	analysisRepo     database.AnalysisRepository
	logger           zerolog.Logger
	maxConcurrentJobs int
}

// NewQualityService creates a new quality service
func NewQualityService(
	analyzer *quality.QualityAnalyzer,
	qualityRepo database.QualityRepository,
	analysisRepo database.AnalysisRepository,
	logger zerolog.Logger,
) *QualityService {
	return &QualityService{
		analyzer:         analyzer,
		qualityRepo:      qualityRepo,
		analysisRepo:     analysisRepo,
		logger:           logger,
		maxConcurrentJobs: 10, // Default concurrent jobs
	}
}

// SetMaxConcurrentJobs sets the maximum number of concurrent quality analysis jobs
func (s *QualityService) SetMaxConcurrentJobs(max int) {
	s.maxConcurrentJobs = max
}

// AnalyzeQuality performs quality analysis and stores results
func (s *QualityService) AnalyzeQuality(ctx context.Context, request *quality.QualityComparisonRequest) (*quality.QualityResult, error) {
	s.logger.Info().
		Str("reference_file", request.ReferenceFile).
		Str("distorted_file", request.DistortedFile).
		Strs("metrics", s.metricsToStrings(request.Metrics)).
		Bool("async", request.Async).
		Bool("frame_level", request.FrameLevel).
		Msg("Starting quality analysis")

	// Create base analysis record
	baseAnalysis := &database.Analysis{
		ID:            uuid.New(),
		InputSource:   request.ReferenceFile,
		Status:        database.StatusProcessing,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		RequestID:     s.getRequestID(ctx),
		ContentType:   "video/mp4", // Default, should be detected
		FileSize:      0,           // Will be populated later
		ProcessingTime: 0,
		RawData:       []byte("{}"),
	}

	// Store base analysis
	if err := s.analysisRepo.Create(ctx, baseAnalysis); err != nil {
		s.logger.Error().Err(err).Msg("Failed to create base analysis")
		return nil, fmt.Errorf("failed to create base analysis: %w", err)
	}

	// Perform quality analysis
	result, err := s.analyzer.AnalyzeQuality(ctx, request)
	if err != nil {
		s.logger.Error().Err(err).Msg("Quality analysis failed")
		
		// Update base analysis with error
		baseAnalysis.Status = database.StatusFailed
		baseAnalysis.ErrorMessage = err.Error()
		baseAnalysis.UpdatedAt = time.Now()
		s.analysisRepo.Update(ctx, baseAnalysis)
		
		return nil, fmt.Errorf("quality analysis failed: %w", err)
	}

	// Store quality analysis results
	if err := s.storeQualityResults(ctx, baseAnalysis.ID, result); err != nil {
		s.logger.Error().Err(err).Msg("Failed to store quality results")
		return nil, fmt.Errorf("failed to store quality results: %w", err)
	}

	// Update base analysis with completion
	baseAnalysis.Status = database.StatusCompleted
	baseAnalysis.ProcessingTime = result.ProcessingTime
	baseAnalysis.UpdatedAt = time.Now()
	completedAt := time.Now()
	baseAnalysis.CompletedAt = &completedAt
	
	if err := s.analysisRepo.Update(ctx, baseAnalysis); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to update base analysis")
	}

	s.logger.Info().
		Str("analysis_id", result.ID.String()).
		Dur("processing_time", result.ProcessingTime).
		Msg("Quality analysis completed successfully")

	return result, nil
}

// AnalyzeQualityAsync performs asynchronous quality analysis
func (s *QualityService) AnalyzeQualityAsync(ctx context.Context, request *quality.QualityComparisonRequest) (*quality.QualityResult, error) {
	// Create initial result with pending status
	result := &quality.QualityResult{
		ID:     uuid.New(),
		Status: quality.QualityStatusPending,
	}

	// Store initial quality comparison
	if err := s.qualityRepo.CreateQualityComparison(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to create quality comparison: %w", err)
	}

	// Launch analysis in background
	go func() {
		bgCtx := context.Background()
		
		// Update status to processing
		result.Status = quality.QualityStatusProcessing
		s.qualityRepo.UpdateQualityComparison(bgCtx, result)
		
		// Perform analysis
		analysisResult, err := s.AnalyzeQuality(bgCtx, request)
		if err != nil {
			result.Status = quality.QualityStatusFailed
			result.Error = err.Error()
			s.logger.Error().Err(err).Str("analysis_id", result.ID.String()).Msg("Async quality analysis failed")
		} else {
			result.Status = quality.QualityStatusCompleted
			result.Analysis = analysisResult.Analysis
			result.Summary = analysisResult.Summary
			result.Visualization = analysisResult.Visualization
			result.ProcessingTime = analysisResult.ProcessingTime
			result.Message = "Quality analysis completed successfully"
		}
		
		// Update final result
		if err := s.qualityRepo.UpdateQualityComparison(bgCtx, result); err != nil {
			s.logger.Error().Err(err).Str("analysis_id", result.ID.String()).Msg("Failed to update async analysis result")
		}
	}()

	return result, nil
}

// BatchAnalyzeQuality performs batch quality analysis
func (s *QualityService) BatchAnalyzeQuality(ctx context.Context, request *quality.BatchQualityRequest) (*quality.BatchQualityResult, error) {
	batchID := uuid.New()
	
	s.logger.Info().
		Str("batch_id", batchID.String()).
		Int("total_comparisons", len(request.Comparisons)).
		Int("parallel", request.Parallel).
		Bool("async", request.Async).
		Msg("Starting batch quality analysis")

	batchResult := &quality.BatchQualityResult{
		BatchID:   batchID,
		Status:    "processing",
		Total:     len(request.Comparisons),
		Completed: 0,
		Failed:    0,
		Results:   make([]*quality.QualityResult, 0, len(request.Comparisons)),
	}

	// Set parallel processing limit
	parallel := request.Parallel
	if parallel == 0 {
		parallel = s.maxConcurrentJobs
	}
	if parallel > s.maxConcurrentJobs {
		parallel = s.maxConcurrentJobs
	}

	// Process comparisons with controlled concurrency
	semaphore := make(chan struct{}, parallel)
	results := make(chan *quality.QualityResult, len(request.Comparisons))
	errors := make(chan error, len(request.Comparisons))

	startTime := time.Now()

	for i, comparison := range request.Comparisons {
		go func(idx int, comp quality.QualityComparisonRequest) {
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			result, err := s.AnalyzeQuality(ctx, &comp)
			if err != nil {
				s.logger.Error().Err(err).Int("comparison_index", idx).Msg("Batch comparison failed")
				errors <- err
				return
			}

			results <- result
		}(i, comparison)
	}

	// Collect results
	for i := 0; i < len(request.Comparisons); i++ {
		select {
		case result := <-results:
			batchResult.Results = append(batchResult.Results, result)
			batchResult.Completed++
		case err := <-errors:
			batchResult.Failed++
			s.logger.Error().Err(err).Msg("Batch analysis error")
		case <-ctx.Done():
			batchResult.Status = "cancelled"
			return batchResult, ctx.Err()
		}
	}

	// Generate batch summary
	batchResult.Status = "completed"
	batchResult.Summary = s.generateBatchSummary(batchResult.Results, time.Since(startTime))

	s.logger.Info().
		Str("batch_id", batchID.String()).
		Int("completed", batchResult.Completed).
		Int("failed", batchResult.Failed).
		Dur("processing_time", time.Since(startTime)).
		Msg("Batch quality analysis completed")

	return batchResult, nil
}

// GetQualityAnalysis retrieves a quality analysis by ID
func (s *QualityService) GetQualityAnalysis(ctx context.Context, id uuid.UUID) (*quality.QualityAnalysis, error) {
	return s.qualityRepo.GetQualityAnalysis(ctx, id)
}

// GetQualityComparison retrieves a quality comparison by ID
func (s *QualityService) GetQualityComparison(ctx context.Context, id uuid.UUID) (*quality.QualityResult, error) {
	return s.qualityRepo.GetQualityComparison(ctx, id)
}

// GetQualityFrames retrieves quality frames for analysis
func (s *QualityService) GetQualityFrames(ctx context.Context, qualityID uuid.UUID, limit, offset int) ([]*quality.QualityFrameMetric, error) {
	return s.qualityRepo.GetQualityFrames(ctx, qualityID, limit, offset)
}

// GetQualityIssues retrieves quality issues for analysis
func (s *QualityService) GetQualityIssues(ctx context.Context, qualityID uuid.UUID) ([]*quality.QualityIssue, error) {
	return s.qualityRepo.GetQualityIssues(ctx, qualityID)
}

// GetQualityStatistics retrieves quality statistics
func (s *QualityService) GetQualityStatistics(ctx context.Context, filters database.QualityStatisticsFilters) (*database.QualityStatistics, error) {
	return s.qualityRepo.GetQualityStatistics(ctx, filters)
}

// GetQualityThresholds retrieves quality thresholds
func (s *QualityService) GetQualityThresholds(ctx context.Context, metricType quality.QualityMetricType) (*quality.QualityThresholds, error) {
	return s.qualityRepo.GetQualityThresholds(ctx, metricType)
}

// DeleteQualityAnalysis deletes a quality analysis and related data
func (s *QualityService) DeleteQualityAnalysis(ctx context.Context, id uuid.UUID) error {
	// Delete frames and issues first (due to foreign key constraints)
	if err := s.qualityRepo.DeleteQualityFrames(ctx, id); err != nil {
		s.logger.Warn().Err(err).Str("quality_id", id.String()).Msg("Failed to delete quality frames")
	}

	if err := s.qualityRepo.DeleteQualityIssues(ctx, id); err != nil {
		s.logger.Warn().Err(err).Str("quality_id", id.String()).Msg("Failed to delete quality issues")
	}

	// Delete main analysis
	return s.qualityRepo.DeleteQualityAnalysis(ctx, id)
}

// storeQualityResults stores quality analysis results in the database
func (s *QualityService) storeQualityResults(ctx context.Context, analysisID uuid.UUID, result *quality.QualityResult) error {
	// Store each quality analysis
	for _, analysis := range result.Analysis {
		analysis.AnalysisID = analysisID
		if err := s.qualityRepo.CreateQualityAnalysis(ctx, analysis); err != nil {
			return fmt.Errorf("failed to store quality analysis: %w", err)
		}

		// Store frame metrics if available
		if result.FrameMetrics != nil {
			for _, frameMetric := range result.FrameMetrics {
				if frameMetric.QualityID == analysis.ID {
					frameMetric.QualityID = analysis.ID
				}
			}
			
			if err := s.qualityRepo.CreateQualityFrames(ctx, result.FrameMetrics); err != nil {
				s.logger.Warn().Err(err).Msg("Failed to store frame metrics")
			}
		}

		// Store quality issues if available
		if result.Summary != nil && len(result.Summary.QualityIssues) > 0 {
			for _, issue := range result.Summary.QualityIssues {
				issue.QualityID = analysis.ID
			}
			
			if err := s.qualityRepo.CreateQualityIssues(ctx, result.Summary.QualityIssues); err != nil {
				s.logger.Warn().Err(err).Msg("Failed to store quality issues")
			}
		}
	}

	// Store quality comparison
	if err := s.qualityRepo.CreateQualityComparison(ctx, result); err != nil {
		return fmt.Errorf("failed to store quality comparison: %w", err)
	}

	return nil
}

// generateBatchSummary generates a summary for batch quality analysis
func (s *QualityService) generateBatchSummary(results []*quality.QualityResult, processingTime time.Duration) *quality.BatchQualitySummary {
	if len(results) == 0 {
		return &quality.BatchQualitySummary{
			ProcessingTime: processingTime,
		}
	}

	summary := &quality.BatchQualitySummary{
		AverageScores:   make(map[quality.QualityMetricType]float64),
		Recommendations: make([]string, 0),
		ProcessingTime:  processingTime,
	}

	// Calculate average scores by metric type
	metricCounts := make(map[quality.QualityMetricType]int)
	metricSums := make(map[quality.QualityMetricType]float64)

	var bestScore float64 = -1
	var worstScore float64 = 101
	var bestFile, worstFile string

	for _, result := range results {
		if result.Analysis != nil {
			for _, analysis := range result.Analysis {
				metricSums[analysis.MetricType] += analysis.OverallScore
				metricCounts[analysis.MetricType]++

				// Track best and worst performing files
				if analysis.OverallScore > bestScore {
					bestScore = analysis.OverallScore
					bestFile = analysis.DistortedFile
				}
				if analysis.OverallScore < worstScore {
					worstScore = analysis.OverallScore
					worstFile = analysis.DistortedFile
				}
			}
		}
	}

	// Calculate averages
	for metric, sum := range metricSums {
		if metricCounts[metric] > 0 {
			summary.AverageScores[metric] = sum / float64(metricCounts[metric])
		}
	}

	summary.BestPerforming = bestFile
	summary.WorstPerforming = worstFile

	// Calculate overall rating based on average scores
	var totalScore float64
	var totalCount int
	for _, score := range summary.AverageScores {
		totalScore += score
		totalCount++
	}

	if totalCount > 0 {
		avgScore := totalScore / float64(totalCount)
		summary.OverallRating = s.scoreToRating(avgScore)
	}

	// Generate recommendations
	summary.Recommendations = s.generateBatchRecommendations(summary.AverageScores, summary.OverallRating)

	return summary
}

// generateBatchRecommendations generates recommendations for batch analysis
func (s *QualityService) generateBatchRecommendations(averageScores map[quality.QualityMetricType]float64, overallRating quality.QualityRating) []string {
	var recommendations []string

	for metric, score := range averageScores {
		rating := s.scoreToRating(score)
		
		if rating == quality.RatingPoor || rating == quality.RatingBad {
			switch metric {
			case quality.MetricVMAF:
				recommendations = append(recommendations, "Consider increasing bitrate or using higher quality encoding profiles to improve VMAF scores")
			case quality.MetricPSNR:
				recommendations = append(recommendations, "Reduce quantization parameters or increase bitrate to improve PSNR values")
			case quality.MetricSSIM:
				recommendations = append(recommendations, "Review preprocessing and encoding settings to improve structural similarity")
			}
		}
	}

	if overallRating == quality.RatingExcellent {
		recommendations = append(recommendations, "Excellent quality across all metrics. Current encoding settings are optimal.")
	} else if overallRating == quality.RatingGood {
		recommendations = append(recommendations, "Good quality overall. Minor optimizations may further improve results.")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Quality analysis completed successfully with acceptable results.")
	}

	return recommendations
}

// scoreToRating converts a numeric score to quality rating
func (s *QualityService) scoreToRating(score float64) quality.QualityRating {
	if score >= 90 {
		return quality.RatingExcellent
	} else if score >= 80 {
		return quality.RatingGood
	} else if score >= 70 {
		return quality.RatingFair
	} else if score >= 50 {
		return quality.RatingPoor
	}
	return quality.RatingBad
}

// getRequestID extracts request ID from context
func (s *QualityService) getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// metricsToStrings converts metric types to strings for logging
func (s *QualityService) metricsToStrings(metrics []quality.QualityMetricType) []string {
	strings := make([]string, len(metrics))
	for i, metric := range metrics {
		strings[i] = string(metric)
	}
	return strings
}

// QualityComparisonOptions contains options for quality comparison
type QualityComparisonOptions struct {
	ReferenceID    uuid.UUID
	DistortedID    uuid.UUID
	ComparisonType models.ComparisonType
	Metrics        []string
	UserID         string
	Options        map[string]interface{}
}

// ComparisonFilters contains filters for listing comparisons
type ComparisonFilters struct {
	ReferenceID string
	DistortedID string
	Status      string
}

var (
	// ErrComparisonNotFound is returned when a comparison is not found
	ErrComparisonNotFound = errors.New("comparison not found")
)

// CompareQuality performs quality comparison between reference and distorted videos
func (s *QualityService) CompareQuality(ctx context.Context, comparisonID string, opts QualityComparisonOptions) error {
	startTime := time.Now()

	// Parse comparison ID
	comparisonUUID, err := uuid.Parse(comparisonID)
	if err != nil {
		return fmt.Errorf("invalid comparison ID: %w", err)
	}

	// Parse user ID
	var userUUID *uuid.UUID
	if opts.UserID != "" {
		id, err := uuid.Parse(opts.UserID)
		if err == nil {
			userUUID = &id
		}
	}

	// Create comparison record
	comparison := &models.QualityComparison{
		ID:             comparisonUUID,
		ReferenceID:    opts.ReferenceID,
		DistortedID:    opts.DistortedID,
		ComparisonType: opts.ComparisonType,
		Status:         models.StatusPending,
		CreatedAt:      time.Now(),
	}

	// Save initial comparison record
	if err := s.db.CreateQualityComparison(ctx, comparison); err != nil {
		return fmt.Errorf("failed to create comparison record: %w", err)
	}

	// Update status to processing
	comparison.Status = models.StatusProcessing
	if err := s.db.UpdateQualityComparisonStatus(ctx, comparisonUUID, models.StatusProcessing); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update comparison status")
	}

	// Get reference and distorted analyses
	refAnalysis, err := s.db.GetAnalysis(ctx, opts.ReferenceID)
	if err != nil {
		s.updateComparisonError(ctx, comparisonUUID, "Reference analysis not found")
		return fmt.Errorf("reference analysis not found: %w", err)
	}

	distAnalysis, err := s.db.GetAnalysis(ctx, opts.DistortedID)
	if err != nil {
		s.updateComparisonError(ctx, comparisonUUID, "Distorted analysis not found")
		return fmt.Errorf("distorted analysis not found: %w", err)
	}

	// Perform quality comparison using analyzer
	result, err := s.analyzer.CompareQuality(ctx, quality.ComparisonRequest{
		ReferenceFile: refAnalysis.FilePath,
		DistortedFile: distAnalysis.FilePath,
		Metrics:       opts.Metrics,
		ComparisonType: string(opts.ComparisonType),
		Options:       opts.Options,
	})

	if err != nil {
		s.updateComparisonError(ctx, comparisonUUID, fmt.Sprintf("Quality comparison failed: %v", err))
		return fmt.Errorf("quality comparison failed: %w", err)
	}

	// Update comparison with results
	comparison.Status = models.StatusCompleted
	comparison.ResultSummary = result.Summary
	comparison.ProcessingTime = time.Since(startTime).Seconds()
	comparison.CompletedAt = &time.Time{}
	*comparison.CompletedAt = time.Now()

	if err := s.db.UpdateQualityComparison(ctx, comparison); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update comparison record")
	}

	// Store detailed quality metrics
	for metricType, metricData := range result.Metrics {
		metric := &models.QualityMetric{
			ID:             uuid.New(),
			AnalysisID:     opts.DistortedID, // Associate with distorted file
			ReferenceFileID: &opts.ReferenceID,
			MetricType:     models.QualityMetricType(metricType),
			ProcessingTime: time.Since(startTime).Seconds(),
			CreatedAt:      time.Now(),
		}

		if scores, ok := metricData.(map[string]interface{}); ok {
			if overall, ok := scores["overall"].(float64); ok {
				metric.OverallScore = overall
			}
			if min, ok := scores["min"].(float64); ok {
				metric.MinScore = min
			}
			if max, ok := scores["max"].(float64); ok {
				metric.MaxScore = max
			}
			if mean, ok := scores["mean"].(float64); ok {
				metric.MeanScore = mean
			}
		}

		if err := s.db.CreateQualityMetric(ctx, metric); err != nil {
			s.logger.Error().Err(err).Msg("Failed to save quality metric")
		}
	}

	return nil
}

// GetQualityComparison retrieves a quality comparison by ID
func (s *QualityService) GetQualityComparison(ctx context.Context, comparisonID uuid.UUID) (*models.QualityComparison, error) {
	comparison, err := s.db.GetQualityComparison(ctx, comparisonID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrComparisonNotFound
		}
		return nil, fmt.Errorf("failed to get quality comparison: %w", err)
	}
	return comparison, nil
}

// ListQualityComparisons lists quality comparisons for a user
func (s *QualityService) ListQualityComparisons(ctx context.Context, userID string, filters ComparisonFilters, limit, offset int) ([]*models.QualityComparison, int, error) {
	// Parse user ID if provided
	var userUUID *uuid.UUID
	if userID != "" {
		id, err := uuid.Parse(userID)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid user ID: %w", err)
		}
		userUUID = &id
	}

	// Get comparisons from database
	comparisons, total, err := s.db.ListQualityComparisons(ctx, userUUID, filters.ReferenceID, filters.DistortedID, filters.Status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list quality comparisons: %w", err)
	}

	return comparisons, total, nil
}

// DeleteQualityComparison deletes a quality comparison
func (s *QualityService) DeleteQualityComparison(ctx context.Context, comparisonID uuid.UUID) error {
	err := s.db.DeleteQualityComparison(ctx, comparisonID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return ErrComparisonNotFound
		}
		return fmt.Errorf("failed to delete quality comparison: %w", err)
	}
	return nil
}

// Helper methods

func (s *QualityService) updateComparisonError(ctx context.Context, comparisonID uuid.UUID, errorMsg string) {
	s.logger.Error().
		Str("comparison_id", comparisonID.String()).
		Str("error", errorMsg).
		Msg("Quality comparison failed")

	comparison := &models.QualityComparison{
		ID:       comparisonID,
		Status:   models.StatusFailed,
		ErrorMsg: errorMsg,
	}

	if err := s.db.UpdateQualityComparison(ctx, comparison); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update comparison error status")
	}
}