package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/database"
	"github.com/rendiffdev/ffprobe-api/internal/ffmpeg"
	"github.com/rendiffdev/ffprobe-api/internal/models"
)

// AnalysisResult represents the result of an analysis
type AnalysisResult struct {
	Analysis *models.Analysis `json:"analysis"`
}

// AnalysisOptions contains options for media analysis
type AnalysisOptions struct {
	Source      string   `json:"source"`
	UserID      string   `json:"user_id"`
	FFprobeArgs []string `json:"ffprobe_args"`
}

// AnalysisService handles media file analysis operations
type AnalysisService struct {
	db           *database.DB
	repo         database.Repository
	ffprobe      *ffmpeg.FFprobe
	llmService   *LLMService
	workerClient *WorkerClient
	logger       zerolog.Logger
	tempDir      string
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(db *database.DB, ffprobePath string, logger zerolog.Logger) *AnalysisService {
	return &AnalysisService{
		db:           db,
		repo:         database.NewRepository(db),
		ffprobe:      ffmpeg.NewFFprobe(ffprobePath, logger),
		llmService:   nil, // Will be set via SetLLMService
		workerClient: nil, // Will be set via SetWorkerClient
		logger:       logger,
		tempDir:      "/tmp", // Default temp directory
	}
}

// SetLLMService sets the LLM service for automatic report generation
func (s *AnalysisService) SetLLMService(llmService *LLMService) {
	s.llmService = llmService
}

// SetWorkerClient sets the worker client for distributed processing
func (s *AnalysisService) SetWorkerClient(workerClient *WorkerClient) {
	s.workerClient = workerClient
}

// SetTempDirectory sets the temporary directory for file processing
func (s *AnalysisService) SetTempDirectory(dir string) {
	s.tempDir = dir
}

// CreateAnalysis creates a new analysis record and starts processing
func (s *AnalysisService) CreateAnalysis(ctx context.Context, request *models.CreateAnalysisRequest) (*models.Analysis, error) {
	// Generate analysis ID
	analysisID := uuid.New()

	// Create analysis record
	analysis := &models.Analysis{
		ID:          analysisID,
		FileName:    request.FileName,
		FilePath:    request.FilePath,
		FileSize:    request.FileSize,
		ContentHash: request.ContentHash,
		SourceType:  request.SourceType,
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Calculate content hash if not provided
	if analysis.ContentHash == "" {
		hash, err := s.calculateContentHash(request.FilePath)
		if err != nil {
			s.logger.Warn().Err(err).Str("file_path", request.FilePath).Msg("Failed to calculate content hash")
		} else {
			analysis.ContentHash = hash
		}
	}

	// Get file size if not provided
	if analysis.FileSize == 0 {
		if size, err := s.getFileSize(request.FilePath); err == nil {
			analysis.FileSize = size
		}
	}

	// Save to database
	if err := s.repo.CreateAnalysis(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to create analysis record: %w", err)
	}

	s.logger.Info().
		Str("analysis_id", analysisID.String()).
		Str("file_name", analysis.FileName).
		Str("source_type", analysis.SourceType).
		Msg("Analysis record created")

	return analysis, nil
}

// ProcessAnalysisWithContent processes a media file with optional content analysis
func (s *AnalysisService) ProcessAnalysisWithContent(ctx context.Context, analysisID uuid.UUID, options *ffmpeg.FFprobeOptions, enableContentAnalysis bool) error {
	// Update status to processing
	if err := s.repo.UpdateAnalysisStatus(ctx, analysisID, models.StatusProcessing, nil); err != nil {
		return fmt.Errorf("failed to update analysis status: %w", err)
	}

	// Get analysis record
	analysis, err := s.repo.GetAnalysis(ctx, analysisID)
	if err != nil {
		return fmt.Errorf("failed to get analysis record: %w", err)
	}

	// Set default options if not provided
	if options == nil {
		options = ffmpeg.NewOptionsBuilder().
			Input(analysis.FilePath).
			DetailedInfo().
			Build()
	} else if options.Input == "" {
		options.Input = analysis.FilePath
	}

	// Configure ffprobe for content analysis if enabled
	if enableContentAnalysis {
		s.ffprobe.EnableContentAnalysis()
		defer s.ffprobe.DisableContentAnalysis()
		
		// Use content analysis method
		result, err := s.ffprobe.ProbeFileWithContentAnalysis(ctx, analysis.FilePath)
		if err != nil {
			s.updateAnalysisError(ctx, analysisID, fmt.Sprintf("FFprobe with content analysis failed: %v", err))
			return fmt.Errorf("ffprobe with content analysis failed: %w", err)
		}
		
		return s.completeAnalysis(ctx, analysisID, result)
	}

	// Standard analysis
	result, err := s.ffprobe.Probe(ctx, options)
	if err != nil {
		s.updateAnalysisError(ctx, analysisID, fmt.Sprintf("FFprobe failed: %v", err))
		return fmt.Errorf("ffprobe failed: %w", err)
	}

	return s.completeAnalysis(ctx, analysisID, result)
}

// ProcessAnalysis processes a media file and updates the analysis record
func (s *AnalysisService) ProcessAnalysis(ctx context.Context, analysisID uuid.UUID, options *ffmpeg.FFprobeOptions) error {
	return s.ProcessAnalysisWithContent(ctx, analysisID, options, false)
}

// completeAnalysis completes the analysis with the given result
func (s *AnalysisService) completeAnalysis(ctx context.Context, analysisID uuid.UUID, result *ffmpeg.FFprobeResult) error {
	// Get analysis record
	analysis, err := s.repo.GetAnalysis(ctx, analysisID)
	if err != nil {
		return fmt.Errorf("failed to get analysis record: %w", err)
	}

	s.logger.Info().
		Str("analysis_id", analysisID.String()).
		Str("file_path", analysis.FilePath).
		Msg("Completing ffprobe analysis")

	// Result is already provided, no need to execute ffprobe again
	if result == nil {
		return fmt.Errorf("analysis result cannot be nil")
	}

	// Convert result to FFprobeData
	ffprobeData, err := s.convertResultToFFprobeData(result)
	if err != nil {
		s.updateAnalysisError(ctx, analysisID, fmt.Sprintf("Failed to convert result: %v", err))
		return fmt.Errorf("failed to convert result: %w", err)
	}

	// Update analysis with results
	analysis.FFprobeData = ffprobeData
	analysis.Status = models.StatusCompleted
	analysis.UpdatedAt = time.Now()
	processedAt := time.Now()
	analysis.ProcessedAt = &processedAt

	// Update in database (this would need to be implemented in repository)
	if err := s.repo.UpdateAnalysisStatus(ctx, analysisID, models.StatusCompleted, nil); err != nil {
		return fmt.Errorf("failed to update analysis status: %w", err)
	}

	s.logger.Info().
		Str("analysis_id", analysisID.String()).
		Dur("execution_time", result.ExecutionTime).
		Msg("Analysis completed successfully")

	// Generate AI analysis report as part of standard analysis (synchronous)
	if s.workerClient != nil {
		s.logger.Info().Str("analysis_id", analysis.ID.String()).Msg("Generating AI analysis report via worker")
		
		// Create timeout context for AI generation
		llmCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		
		// Convert FFprobeData to map for worker
		analysisMap := make(map[string]interface{})
		if ffprobeData.Format != nil {
			json.Unmarshal(ffprobeData.Format, &analysisMap)
		}
		
		report, err := s.workerClient.GenerateAnalysisWithLLM(llmCtx, analysisMap)
		if err != nil {
			s.logger.Warn().
				Err(err).
				Str("analysis_id", analysis.ID.String()).
				Msg("Failed to generate AI analysis report via worker - continuing without AI insights")
		} else if report != "" {
			// Store the AI report
			analysis.LLMReport = &report
			analysis.UpdatedAt = time.Now()
			
			if err := s.updateAnalysisLLMReport(llmCtx, analysis.ID, report); err != nil {
				s.logger.Error().
					Err(err).
					Str("analysis_id", analysis.ID.String()).
					Msg("Failed to save AI analysis report")
			} else {
				s.logger.Info().
					Str("analysis_id", analysis.ID.String()).
					Int("report_length", len(report)).
					Msg("AI analysis report integrated successfully via worker")
			}
		}
	} else if s.llmService != nil {
		// Fallback to local LLM service
		s.logger.Info().Str("analysis_id", analysis.ID.String()).Msg("Generating AI analysis report locally")
		
		llmCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		
		report, err := s.llmService.GenerateAnalysis(llmCtx, analysis)
		if err != nil {
			s.logger.Warn().
				Err(err).
				Str("analysis_id", analysis.ID.String()).
				Msg("Failed to generate AI analysis report - continuing without AI insights")
		} else {
			analysis.LLMReport = &report
			analysis.UpdatedAt = time.Now()
			
			if err := s.updateAnalysisLLMReport(llmCtx, analysis.ID, report); err != nil {
				s.logger.Error().
					Err(err).
					Str("analysis_id", analysis.ID.String()).
					Msg("Failed to save AI analysis report")
			} else {
				s.logger.Info().
					Str("analysis_id", analysis.ID.String()).
					Int("report_length", len(report)).
					Msg("AI analysis report integrated successfully")
			}
		}
	} else {
		s.logger.Warn().Str("analysis_id", analysis.ID.String()).Msg("AI service not available - analysis will not include professional insights")
	}

	return nil
}


// updateAnalysisLLMReport updates the analysis with the LLM report
func (s *AnalysisService) updateAnalysisLLMReport(ctx context.Context, analysisID uuid.UUID, report string) error {
	return s.repo.UpdateAnalysisLLMReport(ctx, analysisID, report)
}

// ProcessFile is a convenience method that creates and processes an analysis
func (s *AnalysisService) ProcessFile(ctx context.Context, filePath string, options *ffmpeg.FFprobeOptions) (*models.Analysis, error) {
	// Create analysis request
	fileName := filepath.Base(filePath)
	request := &models.CreateAnalysisRequest{
		FileName:   fileName,
		FilePath:   filePath,
		SourceType: "local",
	}

	// Create analysis record
	analysis, err := s.CreateAnalysis(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create analysis: %w", err)
	}

	// Process the analysis
	if err := s.ProcessAnalysis(ctx, analysis.ID, options); err != nil {
		return analysis, fmt.Errorf("failed to process analysis: %w", err)
	}

	// Return updated analysis
	return s.repo.GetAnalysis(ctx, analysis.ID)
}

// GetAnalysis retrieves an analysis by ID
func (s *AnalysisService) GetAnalysis(ctx context.Context, id uuid.UUID) (*models.Analysis, error) {
	return s.repo.GetAnalysis(ctx, id)
}

// GetAnalysesByUser retrieves analyses for a specific user
func (s *AnalysisService) GetAnalysesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Analysis, error) {
	return s.repo.GetAnalysesByUser(ctx, userID, limit, offset)
}

// DeleteAnalysis deletes an analysis record
func (s *AnalysisService) DeleteAnalysis(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteAnalysis(ctx, id)
}

// CheckFFprobeAvailability checks if ffprobe is available
func (s *AnalysisService) CheckFFprobeAvailability(ctx context.Context) error {
	return s.ffprobe.CheckBinary(ctx)
}

// GetFFprobeVersion returns the ffprobe version
func (s *AnalysisService) GetFFprobeVersion(ctx context.Context) (string, error) {
	return s.ffprobe.GetVersion(ctx)
}

// Helper methods

func (s *AnalysisService) updateAnalysisError(ctx context.Context, analysisID uuid.UUID, errorMsg string) {
	s.logger.Error().
		Str("analysis_id", analysisID.String()).
		Str("error", errorMsg).
		Msg("Analysis failed")

	if err := s.repo.UpdateAnalysisStatus(ctx, analysisID, models.StatusFailed, &errorMsg); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update analysis error status")
	}
}

func (s *AnalysisService) calculateContentHash(filePath string) (string, error) {
	// For URLs or remote files, skip hash calculation
	if filepath.IsAbs(filePath) == false || len(filePath) > 2048 {
		return "", nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (s *AnalysisService) getFileSize(filePath string) (int64, error) {
	// For URLs or remote files, return 0
	if filepath.IsAbs(filePath) == false || len(filePath) > 2048 {
		return 0, nil
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

func (s *AnalysisService) convertResultToFFprobeData(result *ffmpeg.FFprobeResult) (models.FFprobeData, error) {
	data := models.FFprobeData{}

	// Convert format information
	if result.Format != nil {
		formatJSON, err := json.Marshal(result.Format)
		if err != nil {
			return data, fmt.Errorf("failed to marshal format: %w", err)
		}
		data.Format = formatJSON
	}

	// Convert streams information
	if len(result.Streams) > 0 {
		streamsJSON, err := json.Marshal(result.Streams)
		if err != nil {
			return data, fmt.Errorf("failed to marshal streams: %w", err)
		}
		data.Streams = streamsJSON
	}

	// Convert frames information
	if len(result.Frames) > 0 {
		framesJSON, err := json.Marshal(result.Frames)
		if err != nil {
			return data, fmt.Errorf("failed to marshal frames: %w", err)
		}
		data.Frames = framesJSON
	}

	// Convert packets information
	if len(result.Packets) > 0 {
		packetsJSON, err := json.Marshal(result.Packets)
		if err != nil {
			return data, fmt.Errorf("failed to marshal packets: %w", err)
		}
		data.Packets = packetsJSON
	}

	// Convert chapters information
	if len(result.Chapters) > 0 {
		chaptersJSON, err := json.Marshal(result.Chapters)
		if err != nil {
			return data, fmt.Errorf("failed to marshal chapters: %w", err)
		}
		data.Chapters = chaptersJSON
	}

	// Convert programs information
	if len(result.Programs) > 0 {
		programsJSON, err := json.Marshal(result.Programs)
		if err != nil {
			return data, fmt.Errorf("failed to marshal programs: %w", err)
		}
		data.Programs = programsJSON
	}

	// Convert error information
	if result.Error != nil {
		errorJSON, err := json.Marshal(result.Error)
		if err != nil {
			return data, fmt.Errorf("failed to marshal error: %w", err)
		}
		data.Error = errorJSON
	}

	return data, nil
}

// ProcessWithProgress processes a file with progress reporting
func (s *AnalysisService) ProcessWithProgress(ctx context.Context, analysisID uuid.UUID, options *ffmpeg.FFprobeOptions, progressCallback func(float64)) error {
	// Update status to processing
	if err := s.repo.UpdateAnalysisStatus(ctx, analysisID, models.StatusProcessing, nil); err != nil {
		return fmt.Errorf("failed to update analysis status: %w", err)
	}

	// Get analysis record
	analysis, err := s.repo.GetAnalysis(ctx, analysisID)
	if err != nil {
		return fmt.Errorf("failed to get analysis record: %w", err)
	}

	// Set default options if not provided
	if options == nil {
		options = ffmpeg.NewOptionsBuilder().
			Input(analysis.FilePath).
			BasicInfo().
			Build()
	} else {
		options.Input = analysis.FilePath
	}

	// Execute ffprobe with progress
	result, err := s.ffprobe.ProbeWithProgress(ctx, options, progressCallback)
	if err != nil {
		s.updateAnalysisError(ctx, analysisID, fmt.Sprintf("FFprobe execution failed: %v", err))
		return fmt.Errorf("ffprobe execution failed: %w", err)
	}

	// Convert and save result (similar to ProcessAnalysis)
	ffprobeData, err := s.convertResultToFFprobeData(result)
	if err != nil {
		s.updateAnalysisError(ctx, analysisID, fmt.Sprintf("Failed to convert result: %v", err))
		return fmt.Errorf("failed to convert result: %w", err)
	}

	// Update analysis with results
	if err := s.repo.UpdateAnalysisStatus(ctx, analysisID, models.StatusCompleted, nil); err != nil {
		return fmt.Errorf("failed to update analysis status: %w", err)
	}

	return nil
}

// GetAnalysisResult retrieves an analysis by ID and returns it as AnalysisResult
func (s *AnalysisService) GetAnalysisResult(ctx context.Context, analysisID string) (*AnalysisResult, error) {
	// Parse UUID
	id, err := uuid.Parse(analysisID)
	if err != nil {
		return nil, fmt.Errorf("invalid analysis ID: %w", err)
	}

	// Get analysis from database
	analysis, err := s.repo.GetAnalysis(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}

	return &AnalysisResult{
		Analysis: analysis,
	}, nil
}

// AnalyzeMedia starts media analysis for a given source
func (s *AnalysisService) AnalyzeMedia(ctx context.Context, analysisID string, opts AnalysisOptions) error {
	// Parse UUID
	id, err := uuid.Parse(analysisID)
	if err != nil {
		return fmt.Errorf("invalid analysis ID: %w", err)
	}

	// Parse user ID if provided
	var userUUID *uuid.UUID
	if opts.UserID != "" {
		userID, err := uuid.Parse(opts.UserID)
		if err == nil {
			userUUID = &userID
		}
	}

	// Create analysis record
	analysis := &models.Analysis{
		ID:         id,
		UserID:     userUUID,
		FileName:   filepath.Base(opts.Source),
		FilePath:   opts.Source,
		SourceType: detectSourceType(opts.Source),
		Status:     models.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Get file size
	if size, err := s.getFileSize(opts.Source); err == nil {
		analysis.FileSize = size
	}

	// Calculate content hash
	if hash, err := s.calculateContentHash(opts.Source); err == nil {
		analysis.ContentHash = hash
	}

	// Save to database
	if err := s.repo.CreateAnalysis(ctx, analysis); err != nil {
		return fmt.Errorf("failed to create analysis: %w", err)
	}

	// Start processing with independent context to avoid cancellation when HTTP request ends
	go func() {
		// Create independent context with timeout for background processing
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		
		if err := s.ProcessAnalysis(bgCtx, id, &ffmpeg.FFprobeOptions{
			ShowFormat:      true,
			ShowStreams:     true,
			ShowChapters:    true,
			ShowPrograms:    true,
			ShowPrivateData: true,
			CountFrames:     true,
			CountPackets:    true,
			ProbeSize:       50 * 1024 * 1024, // 50MB probe size for better analysis
			AnalyzeDuration: 10 * 1000000,     // 10 seconds analysis duration
			OutputFormat:    ffmpeg.OutputJSON,
			PrettyPrint:     true,
			HideBanner:      true,
			Args:            opts.FFprobeArgs,
		}); err != nil {
			s.logger.Error().Err(err).Str("analysis_id", analysisID).Msg("Failed to process analysis")
		}
	}()

	return nil
}

// GetQualityMetrics retrieves quality metrics for an analysis
func (s *AnalysisService) GetQualityMetrics(ctx context.Context, analysisID uuid.UUID) ([]*models.QualityMetric, error) {
	return s.db.GetQualityMetrics(ctx, analysisID)
}

// detectSourceType detects the source type from the source string
func detectSourceType(source string) string {
	if len(source) == 0 {
		return "unknown"
	}
	
	switch {
	case source[0] == '/' || source[0] == '.' || (len(source) > 1 && source[1] == ':'):
		return "file"
	case len(source) > 4 && source[:4] == "http":
		return "url"
	case len(source) > 2 && source[:3] == "s3:":
		return "s3"
	case len(source) > 4 && source[:5] == "rtmp:":
		return "stream"
	default:
		return "file"
	}
}

// Helper functions for converting worker data to expected formats
func convertToStreams(data interface{}) []interface{} {
	if data == nil {
		return nil
	}
	if streams, ok := data.([]interface{}); ok {
		return streams
	}
	return nil
}

func convertToChapters(data interface{}) []interface{} {
	if data == nil {
		return nil
	}
	if chapters, ok := data.([]interface{}); ok {
		return chapters
	}
	return nil
}

func convertToPrograms(data interface{}) []interface{} {
	if data == nil {
		return nil
	}
	if programs, ok := data.([]interface{}); ok {
		return programs
	}
	return nil
}