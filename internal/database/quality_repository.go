package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rendiffdev/ffprobe-api/internal/quality"
)

// QualityRepository handles quality metrics database operations
type QualityRepository interface {
	// Quality Analysis operations
	CreateQualityAnalysis(ctx context.Context, analysis *quality.QualityAnalysis) error
	GetQualityAnalysis(ctx context.Context, id uuid.UUID) (*quality.QualityAnalysis, error)
	GetQualityAnalysisByAnalysisID(ctx context.Context, analysisID uuid.UUID) ([]*quality.QualityAnalysis, error)
	UpdateQualityAnalysis(ctx context.Context, analysis *quality.QualityAnalysis) error
	DeleteQualityAnalysis(ctx context.Context, id uuid.UUID) error
	ListQualityAnalyses(ctx context.Context, limit, offset int) ([]*quality.QualityAnalysis, error)

	// Frame metrics operations
	CreateQualityFrames(ctx context.Context, frames []*quality.QualityFrameMetric) error
	GetQualityFrames(ctx context.Context, qualityID uuid.UUID, limit, offset int) ([]*quality.QualityFrameMetric, error)
	GetQualityFramesByTimeRange(ctx context.Context, qualityID uuid.UUID, startTime, endTime float64) ([]*quality.QualityFrameMetric, error)
	DeleteQualityFrames(ctx context.Context, qualityID uuid.UUID) error

	// Quality comparisons operations
	CreateQualityComparison(ctx context.Context, comparison *quality.QualityResult) error
	GetQualityComparison(ctx context.Context, id uuid.UUID) (*quality.QualityResult, error)
	GetQualityComparisonsByBatch(ctx context.Context, batchID uuid.UUID) ([]*quality.QualityResult, error)
	UpdateQualityComparison(ctx context.Context, comparison *quality.QualityResult) error
	DeleteQualityComparison(ctx context.Context, id uuid.UUID) error

	// Quality thresholds operations
	GetQualityThresholds(ctx context.Context, metricType quality.QualityMetricType) (*quality.QualityThresholds, error)
	CreateQualityThreshold(ctx context.Context, threshold *QualityThreshold) error
	UpdateQualityThreshold(ctx context.Context, threshold *QualityThreshold) error
	GetDefaultQualityThresholds(ctx context.Context) (*quality.QualityThresholds, error)

	// Quality issues operations
	CreateQualityIssues(ctx context.Context, issues []*quality.QualityIssue) error
	GetQualityIssues(ctx context.Context, qualityID uuid.UUID) ([]*quality.QualityIssue, error)
	GetQualityIssuesBySeverity(ctx context.Context, qualityID uuid.UUID, severity string) ([]*quality.QualityIssue, error)
	DeleteQualityIssues(ctx context.Context, qualityID uuid.UUID) error

	// Statistics and aggregation
	GetQualityStatistics(ctx context.Context, filters QualityStatisticsFilters) (*QualityStatistics, error)
	GetQualityTrends(ctx context.Context, metricType quality.QualityMetricType, days int) ([]*QualityTrend, error)
}

// QualityThreshold represents a quality threshold configuration
type QualityThreshold struct {
	ID                 uuid.UUID              `json:"id" db:"id"`
	MetricType         quality.QualityMetricType `json:"metric_type" db:"metric_type"`
	ExcellentThreshold float64                `json:"excellent_threshold" db:"excellent_threshold"`
	GoodThreshold      float64                `json:"good_threshold" db:"good_threshold"`
	FairThreshold      float64                `json:"fair_threshold" db:"fair_threshold"`
	PoorThreshold      float64                `json:"poor_threshold" db:"poor_threshold"`
	IsDefault          bool                   `json:"is_default" db:"is_default"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
}

// QualityStatisticsFilters represents filters for quality statistics
type QualityStatisticsFilters struct {
	MetricType quality.QualityMetricType `json:"metric_type"`
	StartDate  *time.Time                `json:"start_date"`
	EndDate    *time.Time                `json:"end_date"`
	MinScore   *float64                  `json:"min_score"`
	MaxScore   *float64                  `json:"max_score"`
	Status     *quality.QualityAnalysisStatus `json:"status"`
}

// QualityStatistics represents aggregated quality statistics
type QualityStatistics struct {
	TotalAnalyses    int     `json:"total_analyses"`
	AverageScore     float64 `json:"average_score"`
	MedianScore      float64 `json:"median_score"`
	MinScore         float64 `json:"min_score"`
	MaxScore         float64 `json:"max_score"`
	StdDevScore      float64 `json:"std_dev_score"`
	ExcellentCount   int     `json:"excellent_count"`
	GoodCount        int     `json:"good_count"`
	FairCount        int     `json:"fair_count"`
	PoorCount        int     `json:"poor_count"`
	BadCount         int     `json:"bad_count"`
	AverageFrameCount int    `json:"average_frame_count"`
	TotalFrames      int     `json:"total_frames"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
}

// QualityTrend represents quality trend data over time
type QualityTrend struct {
	Date         time.Time `json:"date"`
	AverageScore float64   `json:"average_score"`
	SampleCount  int       `json:"sample_count"`
}

// qualityRepository implements QualityRepository
type qualityRepository struct {
	db *sqlx.DB
}

// NewQualityRepository creates a new quality repository
func NewQualityRepository(db *sqlx.DB) QualityRepository {
	return &qualityRepository{db: db}
}

// CreateQualityAnalysis creates a new quality analysis record
func (r *qualityRepository) CreateQualityAnalysis(ctx context.Context, analysis *quality.QualityAnalysis) error {
	query := `
		INSERT INTO quality_metrics (
			id, analysis_id, reference_file, distorted_file, metric_type,
			overall_score, min_score, max_score, mean_score, median_score,
			std_dev_score, percentile_1, percentile_5, percentile_10, percentile_25,
			percentile_75, percentile_90, percentile_95, percentile_99,
			frame_count, duration, width, height, frame_rate, bit_rate,
			configuration, processing_time, status, error_message,
			created_at, updated_at, completed_at
		) VALUES (
			:id, :analysis_id, :reference_file, :distorted_file, :metric_type,
			:overall_score, :min_score, :max_score, :mean_score, :median_score,
			:std_dev_score, :percentile_1, :percentile_5, :percentile_10, :percentile_25,
			:percentile_75, :percentile_90, :percentile_95, :percentile_99,
			:frame_count, :duration, :width, :height, :frame_rate, :bit_rate,
			:configuration, :processing_time, :status, :error_message,
			:created_at, :updated_at, :completed_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, analysis)
	return err
}

// GetQualityAnalysis retrieves a quality analysis by ID
func (r *qualityRepository) GetQualityAnalysis(ctx context.Context, id uuid.UUID) (*quality.QualityAnalysis, error) {
	query := `
		SELECT * FROM quality_metrics 
		WHERE id = $1`

	var analysis quality.QualityAnalysis
	err := r.db.GetContext(ctx, &analysis, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAnalysisNotFound
		}
		return nil, err
	}

	return &analysis, nil
}

// GetQualityAnalysisByAnalysisID retrieves quality analyses by analysis ID
func (r *qualityRepository) GetQualityAnalysisByAnalysisID(ctx context.Context, analysisID uuid.UUID) ([]*quality.QualityAnalysis, error) {
	query := `
		SELECT * FROM quality_metrics 
		WHERE analysis_id = $1 
		ORDER BY created_at DESC`

	var analyses []*quality.QualityAnalysis
	err := r.db.SelectContext(ctx, &analyses, query, analysisID)
	if err != nil {
		return nil, err
	}

	return analyses, nil
}

// UpdateQualityAnalysis updates an existing quality analysis
func (r *qualityRepository) UpdateQualityAnalysis(ctx context.Context, analysis *quality.QualityAnalysis) error {
	query := `
		UPDATE quality_metrics SET
			overall_score = :overall_score,
			min_score = :min_score,
			max_score = :max_score,
			mean_score = :mean_score,
			median_score = :median_score,
			std_dev_score = :std_dev_score,
			percentile_1 = :percentile_1,
			percentile_5 = :percentile_5,
			percentile_10 = :percentile_10,
			percentile_25 = :percentile_25,
			percentile_75 = :percentile_75,
			percentile_90 = :percentile_90,
			percentile_95 = :percentile_95,
			percentile_99 = :percentile_99,
			frame_count = :frame_count,
			duration = :duration,
			width = :width,
			height = :height,
			frame_rate = :frame_rate,
			bit_rate = :bit_rate,
			configuration = :configuration,
			processing_time = :processing_time,
			status = :status,
			error_message = :error_message,
			updated_at = :updated_at,
			completed_at = :completed_at
		WHERE id = :id`

	_, err := r.db.NamedExecContext(ctx, query, analysis)
	return err
}

// DeleteQualityAnalysis deletes a quality analysis
func (r *qualityRepository) DeleteQualityAnalysis(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM quality_metrics WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListQualityAnalyses lists quality analyses with pagination
func (r *qualityRepository) ListQualityAnalyses(ctx context.Context, limit, offset int) ([]*quality.QualityAnalysis, error) {
	query := `
		SELECT * FROM quality_metrics 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`

	var analyses []*quality.QualityAnalysis
	err := r.db.SelectContext(ctx, &analyses, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return analyses, nil
}

// CreateQualityFrames creates quality frame metrics in batch
func (r *qualityRepository) CreateQualityFrames(ctx context.Context, frames []*quality.QualityFrameMetric) error {
	if len(frames) == 0 {
		return nil
	}

	query := `
		INSERT INTO quality_frames (
			id, quality_id, frame_number, timestamp, score,
			component_y, component_u, component_v, additional_data, created_at
		) VALUES (
			:id, :quality_id, :frame_number, :timestamp, :score,
			:component_y, :component_u, :component_v, :additional_data, :created_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, frames)
	return err
}

// GetQualityFrames retrieves quality frames with pagination
func (r *qualityRepository) GetQualityFrames(ctx context.Context, qualityID uuid.UUID, limit, offset int) ([]*quality.QualityFrameMetric, error) {
	query := `
		SELECT * FROM quality_frames 
		WHERE quality_id = $1 
		ORDER BY frame_number 
		LIMIT $2 OFFSET $3`

	var frames []*quality.QualityFrameMetric
	err := r.db.SelectContext(ctx, &frames, query, qualityID, limit, offset)
	if err != nil {
		return nil, err
	}

	return frames, nil
}

// GetQualityFramesByTimeRange retrieves quality frames within a time range
func (r *qualityRepository) GetQualityFramesByTimeRange(ctx context.Context, qualityID uuid.UUID, startTime, endTime float64) ([]*quality.QualityFrameMetric, error) {
	query := `
		SELECT * FROM quality_frames 
		WHERE quality_id = $1 AND timestamp >= $2 AND timestamp <= $3 
		ORDER BY timestamp`

	var frames []*quality.QualityFrameMetric
	err := r.db.SelectContext(ctx, &frames, query, qualityID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	return frames, nil
}

// DeleteQualityFrames deletes all frames for a quality analysis
func (r *qualityRepository) DeleteQualityFrames(ctx context.Context, qualityID uuid.UUID) error {
	query := `DELETE FROM quality_frames WHERE quality_id = $1`
	_, err := r.db.ExecContext(ctx, query, qualityID)
	return err
}

// CreateQualityComparison creates a quality comparison record
func (r *qualityRepository) CreateQualityComparison(ctx context.Context, comparison *quality.QualityResult) error {
	// Marshal complex fields to JSON
	summaryJSON, err := json.Marshal(comparison.Summary)
	if err != nil {
		return err
	}

	visualizationJSON, err := json.Marshal(comparison.Visualization)
	if err != nil {
		return err
	}

	// Extract metrics from analysis
	var metrics []string
	for _, analysis := range comparison.Analysis {
		metrics = append(metrics, string(analysis.MetricType))
	}

	query := `
		INSERT INTO quality_comparisons (
			id, batch_id, reference_file, distorted_file, metrics,
			status, overall_rating, summary, visualization,
			processing_time, error_message, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	// Extract reference and distorted files from first analysis
	var referenceFile, distortedFile string
	if len(comparison.Analysis) > 0 {
		referenceFile = comparison.Analysis[0].ReferenceFile
		distortedFile = comparison.Analysis[0].DistortedFile
	}

	_, err = r.db.ExecContext(ctx, query,
		comparison.ID,
		comparison.ID, // Use same ID as batch_id for single comparisons
		referenceFile,
		distortedFile,
		metrics,
		comparison.Status,
		comparison.Summary.OverallRating,
		summaryJSON,
		visualizationJSON,
		comparison.ProcessingTime,
		comparison.Error,
		time.Now(),
		time.Now(),
	)

	return err
}

// GetQualityComparison retrieves a quality comparison by ID
func (r *qualityRepository) GetQualityComparison(ctx context.Context, id uuid.UUID) (*quality.QualityResult, error) {
	query := `
		SELECT * FROM quality_comparisons 
		WHERE id = $1`

	var comparison struct {
		ID             uuid.UUID                      `db:"id"`
		BatchID        uuid.UUID                      `db:"batch_id"`
		ReferenceFile  string                         `db:"reference_file"`
		DistortedFile  string                         `db:"distorted_file"`
		Metrics        []string                       `db:"metrics"`
		Status         quality.QualityAnalysisStatus  `db:"status"`
		OverallRating  quality.QualityRating          `db:"overall_rating"`
		Summary        json.RawMessage                `db:"summary"`
		Visualization  json.RawMessage                `db:"visualization"`
		ProcessingTime time.Duration                  `db:"processing_time"`
		ErrorMessage   string                         `db:"error_message"`
		CreatedAt      time.Time                      `db:"created_at"`
		UpdatedAt      time.Time                      `db:"updated_at"`
		CompletedAt    *time.Time                     `db:"completed_at"`
	}

	err := r.db.GetContext(ctx, &comparison, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAnalysisNotFound
		}
		return nil, err
	}

	// Unmarshal JSON fields
	var summary quality.QualitySummary
	var visualization quality.QualityVisualization

	if len(comparison.Summary) > 0 {
		if err := json.Unmarshal(comparison.Summary, &summary); err != nil {
			return nil, err
		}
	}

	if len(comparison.Visualization) > 0 {
		if err := json.Unmarshal(comparison.Visualization, &visualization); err != nil {
			return nil, err
		}
	}

	result := &quality.QualityResult{
		ID:             comparison.ID,
		Status:         comparison.Status,
		Summary:        &summary,
		Visualization:  &visualization,
		ProcessingTime: comparison.ProcessingTime,
		Error:          comparison.ErrorMessage,
	}

	return result, nil
}

// GetQualityComparisonsByBatch retrieves quality comparisons by batch ID
func (r *qualityRepository) GetQualityComparisonsByBatch(ctx context.Context, batchID uuid.UUID) ([]*quality.QualityResult, error) {
	query := `
		SELECT * FROM quality_comparisons 
		WHERE batch_id = $1 
		ORDER BY created_at DESC`

	var comparisons []*quality.QualityResult
	err := r.db.SelectContext(ctx, &comparisons, query, batchID)
	if err != nil {
		return nil, err
	}

	return comparisons, nil
}

// UpdateQualityComparison updates a quality comparison
func (r *qualityRepository) UpdateQualityComparison(ctx context.Context, comparison *quality.QualityResult) error {
	// Marshal complex fields to JSON
	summaryJSON, err := json.Marshal(comparison.Summary)
	if err != nil {
		return err
	}

	visualizationJSON, err := json.Marshal(comparison.Visualization)
	if err != nil {
		return err
	}

	query := `
		UPDATE quality_comparisons SET
			status = $2,
			overall_rating = $3,
			summary = $4,
			visualization = $5,
			processing_time = $6,
			error_message = $7,
			updated_at = $8
		WHERE id = $1`

	_, err = r.db.ExecContext(ctx, query,
		comparison.ID,
		comparison.Status,
		comparison.Summary.OverallRating,
		summaryJSON,
		visualizationJSON,
		comparison.ProcessingTime,
		comparison.Error,
		time.Now(),
	)

	return err
}

// DeleteQualityComparison deletes a quality comparison
func (r *qualityRepository) DeleteQualityComparison(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM quality_comparisons WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetQualityThresholds retrieves quality thresholds for a metric type
func (r *qualityRepository) GetQualityThresholds(ctx context.Context, metricType quality.QualityMetricType) (*quality.QualityThresholds, error) {
	query := `
		SELECT * FROM quality_thresholds 
		WHERE metric_type = $1 AND is_default = true
		ORDER BY created_at DESC
		LIMIT 1`

	var threshold QualityThreshold
	err := r.db.GetContext(ctx, &threshold, query, metricType)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default thresholds if none found
			thresholds := quality.DefaultQualityThresholds()
			return &thresholds, nil
		}
		return nil, err
	}

	// Convert to quality thresholds format
	thresholds := &quality.QualityThresholds{}
	switch metricType {
	case quality.MetricVMAF:
		thresholds.VMAF = quality.VMAFThresholds{
			Excellent: threshold.ExcellentThreshold,
			Good:      threshold.GoodThreshold,
			Fair:      threshold.FairThreshold,
			Poor:      threshold.PoorThreshold,
		}
	case quality.MetricPSNR:
		thresholds.PSNR = quality.PSNRThresholds{
			Excellent: threshold.ExcellentThreshold,
			Good:      threshold.GoodThreshold,
			Fair:      threshold.FairThreshold,
			Poor:      threshold.PoorThreshold,
		}
	case quality.MetricSSIM:
		thresholds.SSIM = quality.SSIMThresholds{
			Excellent: threshold.ExcellentThreshold,
			Good:      threshold.GoodThreshold,
			Fair:      threshold.FairThreshold,
			Poor:      threshold.PoorThreshold,
		}
	}

	return thresholds, nil
}

// GetDefaultQualityThresholds retrieves all default quality thresholds
func (r *qualityRepository) GetDefaultQualityThresholds(ctx context.Context) (*quality.QualityThresholds, error) {
	query := `
		SELECT * FROM quality_thresholds 
		WHERE is_default = true
		ORDER BY metric_type`

	var thresholds []QualityThreshold
	err := r.db.SelectContext(ctx, &thresholds, query)
	if err != nil {
		return nil, err
	}

	result := &quality.QualityThresholds{}
	for _, threshold := range thresholds {
		switch threshold.MetricType {
		case quality.MetricVMAF:
			result.VMAF = quality.VMAFThresholds{
				Excellent: threshold.ExcellentThreshold,
				Good:      threshold.GoodThreshold,
				Fair:      threshold.FairThreshold,
				Poor:      threshold.PoorThreshold,
			}
		case quality.MetricPSNR:
			result.PSNR = quality.PSNRThresholds{
				Excellent: threshold.ExcellentThreshold,
				Good:      threshold.GoodThreshold,
				Fair:      threshold.FairThreshold,
				Poor:      threshold.PoorThreshold,
			}
		case quality.MetricSSIM:
			result.SSIM = quality.SSIMThresholds{
				Excellent: threshold.ExcellentThreshold,
				Good:      threshold.GoodThreshold,
				Fair:      threshold.FairThreshold,
				Poor:      threshold.PoorThreshold,
			}
		}
	}

	return result, nil
}

// CreateQualityThreshold creates a new quality threshold
func (r *qualityRepository) CreateQualityThreshold(ctx context.Context, threshold *QualityThreshold) error {
	query := `
		INSERT INTO quality_thresholds (
			id, metric_type, excellent_threshold, good_threshold, 
			fair_threshold, poor_threshold, is_default, created_at, updated_at
		) VALUES (
			:id, :metric_type, :excellent_threshold, :good_threshold,
			:fair_threshold, :poor_threshold, :is_default, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, threshold)
	return err
}

// UpdateQualityThreshold updates an existing quality threshold
func (r *qualityRepository) UpdateQualityThreshold(ctx context.Context, threshold *QualityThreshold) error {
	query := `
		UPDATE quality_thresholds SET
			excellent_threshold = :excellent_threshold,
			good_threshold = :good_threshold,
			fair_threshold = :fair_threshold,
			poor_threshold = :poor_threshold,
			is_default = :is_default,
			updated_at = :updated_at
		WHERE id = :id`

	_, err := r.db.NamedExecContext(ctx, query, threshold)
	return err
}

// CreateQualityIssues creates quality issues in batch
func (r *qualityRepository) CreateQualityIssues(ctx context.Context, issues []*quality.QualityIssue) error {
	if len(issues) == 0 {
		return nil
	}

	query := `
		INSERT INTO quality_issues (
			id, quality_id, issue_type, severity, description,
			frame_range_start, frame_range_end, timestamp_start, timestamp_end,
			score, additional_data, created_at
		) VALUES `

	// Build batch insert query
	var values []interface{}
	var placeholders []string
	
	for i, issue := range issues {
		base := i * 12
		placeholders = append(placeholders, 
			fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10, base+11, base+12))
		
		// Convert issue to values
		id := uuid.New()
		var frameRangeStart, frameRangeEnd *int
		var timestampStart, timestampEnd *float64
		
		if issue.FrameRange != nil {
			frameRangeStart = &issue.FrameRange.Start
			frameRangeEnd = &issue.FrameRange.End
		}
		
		if issue.Timestamp != nil {
			timestampStart = &issue.Timestamp.Start
			timestampEnd = &issue.Timestamp.End
		}
		
		additionalData, _ := json.Marshal(map[string]interface{}{})
		
		values = append(values, 
			id, issue.QualityID, issue.Type, issue.Severity, issue.Description,
			frameRangeStart, frameRangeEnd, timestampStart, timestampEnd,
			issue.Score, additionalData, time.Now())
	}
	
	query += strings.Join(placeholders, ", ")
	_, err := r.db.ExecContext(ctx, query, values...)
	return err
}

// GetQualityIssues retrieves quality issues for a quality analysis
func (r *qualityRepository) GetQualityIssues(ctx context.Context, qualityID uuid.UUID) ([]*quality.QualityIssue, error) {
	query := `
		SELECT * FROM quality_issues 
		WHERE quality_id = $1 
		ORDER BY severity DESC, created_at DESC`

	var dbIssues []struct {
		ID                uuid.UUID       `db:"id"`
		QualityID         uuid.UUID       `db:"quality_id"`
		IssueType         string          `db:"issue_type"`
		Severity          string          `db:"severity"`
		Description       string          `db:"description"`
		FrameRangeStart   *int            `db:"frame_range_start"`
		FrameRangeEnd     *int            `db:"frame_range_end"`
		TimestampStart    *float64        `db:"timestamp_start"`
		TimestampEnd      *float64        `db:"timestamp_end"`
		Score             float64         `db:"score"`
		AdditionalData    json.RawMessage `db:"additional_data"`
		CreatedAt         time.Time       `db:"created_at"`
	}

	err := r.db.SelectContext(ctx, &dbIssues, query, qualityID)
	if err != nil {
		return nil, err
	}

	// Convert to quality issues
	var issues []*quality.QualityIssue
	for _, dbIssue := range dbIssues {
		issue := &quality.QualityIssue{
			Type:        dbIssue.IssueType,
			Severity:    dbIssue.Severity,
			Description: dbIssue.Description,
			Score:       dbIssue.Score,
		}

		// Convert frame range
		if dbIssue.FrameRangeStart != nil && dbIssue.FrameRangeEnd != nil {
			issue.FrameRange = &quality.FrameRange{
				Start: *dbIssue.FrameRangeStart,
				End:   *dbIssue.FrameRangeEnd,
			}
		}

		// Convert timestamp
		if dbIssue.TimestampStart != nil && dbIssue.TimestampEnd != nil {
			issue.Timestamp = &quality.TimeRange{
				Start: *dbIssue.TimestampStart,
				End:   *dbIssue.TimestampEnd,
			}
		}

		issues = append(issues, issue)
	}

	return issues, nil
}

// GetQualityIssuesBySeverity retrieves quality issues by severity
func (r *qualityRepository) GetQualityIssuesBySeverity(ctx context.Context, qualityID uuid.UUID, severity string) ([]*quality.QualityIssue, error) {
	query := `
		SELECT * FROM quality_issues 
		WHERE quality_id = $1 AND severity = $2 
		ORDER BY created_at DESC`

	var issues []*quality.QualityIssue
	err := r.db.SelectContext(ctx, &issues, query, qualityID, severity)
	if err != nil {
		return nil, err
	}

	return issues, nil
}

// DeleteQualityIssues deletes all issues for a quality analysis
func (r *qualityRepository) DeleteQualityIssues(ctx context.Context, qualityID uuid.UUID) error {
	query := `DELETE FROM quality_issues WHERE quality_id = $1`
	_, err := r.db.ExecContext(ctx, query, qualityID)
	return err
}

// GetQualityStatistics retrieves aggregated quality statistics
func (r *qualityRepository) GetQualityStatistics(ctx context.Context, filters QualityStatisticsFilters) (*QualityStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_analyses,
			AVG(overall_score) as average_score,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY overall_score) as median_score,
			MIN(overall_score) as min_score,
			MAX(overall_score) as max_score,
			STDDEV(overall_score) as std_dev_score,
			AVG(frame_count) as average_frame_count,
			SUM(frame_count) as total_frames,
			AVG(EXTRACT(EPOCH FROM processing_time)) as average_processing_time
		FROM quality_metrics
		WHERE 1=1`

	var args []interface{}
	argIndex := 1

	if filters.MetricType != "" {
		query += fmt.Sprintf(" AND metric_type = $%d", argIndex)
		args = append(args, filters.MetricType)
		argIndex++
	}

	if filters.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filters.EndDate)
		argIndex++
	}

	if filters.MinScore != nil {
		query += fmt.Sprintf(" AND overall_score >= $%d", argIndex)
		args = append(args, *filters.MinScore)
		argIndex++
	}

	if filters.MaxScore != nil {
		query += fmt.Sprintf(" AND overall_score <= $%d", argIndex)
		args = append(args, *filters.MaxScore)
		argIndex++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}

	var stats QualityStatistics
	var avgProcessingSeconds float64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&stats.TotalAnalyses,
		&stats.AverageScore,
		&stats.MedianScore,
		&stats.MinScore,
		&stats.MaxScore,
		&stats.StdDevScore,
		&stats.AverageFrameCount,
		&stats.TotalFrames,
		&avgProcessingSeconds,
	)
	if err != nil {
		return nil, err
	}

	stats.AverageProcessingTime = time.Duration(avgProcessingSeconds) * time.Second

	return &stats, nil
}

// GetQualityTrends retrieves quality trends over time
func (r *qualityRepository) GetQualityTrends(ctx context.Context, metricType quality.QualityMetricType, days int) ([]*QualityTrend, error) {
	query := `
		SELECT 
			DATE(created_at) as date,
			AVG(overall_score) as average_score,
			COUNT(*) as sample_count
		FROM quality_metrics
		WHERE metric_type = $1 
		AND created_at >= CURRENT_DATE - INTERVAL '%d days'
		GROUP BY DATE(created_at)
		ORDER BY date DESC`

	query = fmt.Sprintf(query, days)

	var trends []*QualityTrend
	err := r.db.SelectContext(ctx, &trends, query, metricType)
	if err != nil {
		return nil, err
	}

	return trends, nil
}