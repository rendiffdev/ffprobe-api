package services

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/database"
	"github.com/rendiffdev/ffprobe-api/internal/hls"
	"github.com/rendiffdev/ffprobe-api/internal/models"
)

var (
	// ErrAnalysisNotFound is returned when an analysis is not found
	ErrAnalysisNotFound = errors.New("analysis not found")
)

// HLSService handles HLS-specific analysis operations
type HLSService struct {
	db          *database.DB
	hlsAnalyzer *hls.Analyzer
	logger      zerolog.Logger
}

// NewHLSService creates a new HLS service
func NewHLSService(db *database.DB, hlsAnalyzer *hls.Analyzer, logger zerolog.Logger) *HLSService {
	return &HLSService{
		db:          db,
		hlsAnalyzer: hlsAnalyzer,
		logger:      logger,
	}
}

// HLSAnalysisOptions contains options for HLS analysis
type HLSAnalysisOptions struct {
	Source          string
	AnalyzeSegments bool
	MaxSegments     int
	IncludeQuality  bool
	FFprobeArgs     []string
	UserID          string
}

// AnalyzeHLS performs HLS manifest/folder analysis
func (s *HLSService) AnalyzeHLS(ctx context.Context, analysisID string, opts HLSAnalysisOptions) error {
	startTime := time.Now()
	
	// Parse analysis ID
	analysisUUID, err := uuid.Parse(analysisID)
	if err != nil {
		return fmt.Errorf("invalid analysis ID: %w", err)
	}

	// Create initial analysis record
	analysis := &models.Analysis{
		ID:         analysisUUID,
		UserID:     parseUserID(opts.UserID),
		FileName:   filepath.Base(opts.Source),
		FilePath:   opts.Source,
		SourceType: detectSourceType(opts.Source),
		Status:     models.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save initial analysis record
	if err := s.db.CreateAnalysis(ctx, analysis); err != nil {
		return fmt.Errorf("failed to create analysis record: %w", err)
	}

	// Update status to processing
	analysis.Status = models.StatusProcessing
	if err := s.db.UpdateAnalysisStatus(ctx, analysisID, models.StatusProcessing); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update analysis status")
	}

	// Perform HLS analysis
	hlsResult, err := s.hlsAnalyzer.AnalyzeManifest(ctx, opts.Source, hls.AnalysisOptions{
		AnalyzeSegments: opts.AnalyzeSegments,
		MaxSegments:     opts.MaxSegments,
		IncludeQuality:  opts.IncludeQuality,
		FFprobeArgs:     opts.FFprobeArgs,
	})

	if err != nil {
		// Update status to failed
		analysis.Status = models.StatusFailed
		analysis.ErrorMsg = err.Error()
		s.db.UpdateAnalysis(ctx, analysis)
		return fmt.Errorf("HLS analysis failed: %w", err)
	}

	// Create HLS analysis record
	hlsAnalysis := &models.HLSAnalysis{
		ID:             uuid.New(),
		AnalysisID:     analysisUUID,
		ManifestPath:   opts.Source,
		ManifestType:   models.HLSManifestType(hlsResult.Type),
		ManifestData:   hlsResult.Metadata,
		SegmentCount:   len(hlsResult.Segments),
		TotalDuration:  hlsResult.TotalDuration,
		BitrateVariants: extractBitrateVariants(hlsResult),
		SegmentDuration: hlsResult.SegmentDuration,
		PlaylistVersion: hlsResult.PlaylistVersion,
		Status:         models.StatusCompleted,
		ProcessingTime: time.Since(startTime).Seconds(),
		CreatedAt:      time.Now(),
		CompletedAt:    &time.Time{},
	}
	*hlsAnalysis.CompletedAt = time.Now()

	// Save HLS analysis
	if err := s.db.CreateHLSAnalysis(ctx, hlsAnalysis); err != nil {
		return fmt.Errorf("failed to save HLS analysis: %w", err)
	}

	// Save segments if analyzed
	if opts.AnalyzeSegments && len(hlsResult.Segments) > 0 {
		for i, segment := range hlsResult.Segments {
			if opts.MaxSegments > 0 && i >= opts.MaxSegments {
				break
			}

			segmentRecord := &models.HLSSegment{
				ID:             uuid.New(),
				HLSAnalysisID:  hlsAnalysis.ID,
				SegmentURI:     segment.URI,
				SequenceNumber: segment.SequenceNumber,
				Duration:       segment.Duration,
				FileSize:       segment.FileSize,
				Bitrate:        segment.Bitrate,
				Resolution:     segment.Resolution,
				FrameRate:      segment.FrameRate,
				SegmentData:    segment.Metadata,
				QualityScore:   segment.QualityScore,
				Status:         models.StatusCompleted,
				ProcessedAt:    &time.Time{},
				CreatedAt:      time.Now(),
			}
			*segmentRecord.ProcessedAt = time.Now()

			if err := s.db.CreateHLSSegment(ctx, segmentRecord); err != nil {
				s.logger.Error().Err(err).Msg("Failed to save HLS segment")
			}
		}
	}

	// Update main analysis status
	analysis.Status = models.StatusCompleted
	analysis.ProcessedAt = &time.Time{}
	*analysis.ProcessedAt = time.Now()
	
	// Add basic ffprobe data
	analysis.FFprobeData = map[string]interface{}{
		"format": map[string]interface{}{
			"format_name": "hls",
			"duration":    hlsResult.TotalDuration,
			"nb_streams":  len(hlsResult.Variants),
		},
		"hls_analysis_id": hlsAnalysis.ID.String(),
	}

	if err := s.db.UpdateAnalysis(ctx, analysis); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update analysis record")
	}

	return nil
}

// GetHLSAnalysis retrieves HLS analysis by ID
func (s *HLSService) GetHLSAnalysis(ctx context.Context, analysisID string) (*models.HLSAnalysis, error) {
	// Parse UUID
	id, err := uuid.Parse(analysisID)
	if err != nil {
		return nil, fmt.Errorf("invalid analysis ID: %w", err)
	}

	// Get analysis from database
	analysis, err := s.db.GetHLSAnalysisByAnalysisID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrAnalysisNotFound
		}
		return nil, fmt.Errorf("failed to get HLS analysis: %w", err)
	}

	return analysis, nil
}

// GetHLSSegments retrieves segments for an HLS analysis
func (s *HLSService) GetHLSSegments(ctx context.Context, hlsAnalysisID string, limit int) ([]*models.HLSSegment, error) {
	// Parse UUID
	id, err := uuid.Parse(hlsAnalysisID)
	if err != nil {
		return nil, fmt.Errorf("invalid HLS analysis ID: %w", err)
	}

	// Get segments from database
	segments, err := s.db.GetHLSSegments(ctx, id, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get HLS segments: %w", err)
	}

	return segments, nil
}

// ValidatePlaylist validates an HLS playlist
func (s *HLSService) ValidatePlaylist(ctx context.Context, source string) ([]string, error) {
	// Use HLS analyzer to validate
	issues, err := s.hlsAnalyzer.ValidatePlaylist(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return issues, nil
}

// ListHLSAnalyses lists HLS analyses for a user
func (s *HLSService) ListHLSAnalyses(ctx context.Context, userID string, limit, offset int) ([]*models.HLSAnalysis, int, error) {
	// Parse user ID if provided
	var userUUID *uuid.UUID
	if userID != "" {
		id, err := uuid.Parse(userID)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid user ID: %w", err)
		}
		userUUID = &id
	}

	// Get analyses from database
	analyses, total, err := s.db.ListHLSAnalyses(ctx, userUUID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list HLS analyses: %w", err)
	}

	return analyses, total, nil
}

// Helper functions

func parseUserID(userID string) *uuid.UUID {
	if userID == "" {
		return nil
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil
	}
	return &id
}

func detectSourceType(source string) string {
	source = strings.ToLower(source)
	switch {
	case strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://"):
		return "url"
	case strings.HasPrefix(source, "s3://"):
		return "s3"
	case strings.HasPrefix(source, "gs://"):
		return "gcs"
	case strings.HasPrefix(source, "rtmp://") || strings.HasPrefix(source, "rtsp://"):
		return "stream"
	default:
		return "file"
	}
}

func extractBitrateVariants(result *hls.AnalysisResult) []int {
	bitrates := make([]int, 0, len(result.Variants))
	seen := make(map[int]bool)
	
	for _, variant := range result.Variants {
		if !seen[variant.Bandwidth] {
			bitrates = append(bitrates, variant.Bandwidth)
			seen[variant.Bandwidth] = true
		}
	}
	
	return bitrates
}