package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rendiffdev/rendiff-probe/internal/models"
)

// Repository defines the interface for database operations
type Repository interface {
	// Analysis operations
	CreateAnalysis(ctx context.Context, analysis *models.Analysis) error
	GetAnalysis(ctx context.Context, id uuid.UUID) (*models.Analysis, error)
	GetAnalysesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Analysis, error)
	UpdateAnalysisStatus(ctx context.Context, id uuid.UUID, status models.AnalysisStatus, errorMsg *string) error
	UpdateAnalysisLLMReport(ctx context.Context, id uuid.UUID, report string) error
	DeleteAnalysis(ctx context.Context, id uuid.UUID) error

	// Quality metrics operations
	CreateQualityMetrics(ctx context.Context, metrics *models.QualityMetrics) error
	GetQualityMetrics(ctx context.Context, analysisID uuid.UUID) ([]models.QualityMetrics, error)
	CreateQualityFrame(ctx context.Context, frame *models.QualityFrame) error
	GetQualityFrames(ctx context.Context, metricID uuid.UUID, limit, offset int) ([]models.QualityFrame, error)

	// HLS operations
	CreateHLSAnalysis(ctx context.Context, hls *models.HLSAnalysis) error
	GetHLSAnalysis(ctx context.Context, analysisID uuid.UUID) (*models.HLSAnalysis, error)
	CreateHLSSegment(ctx context.Context, segment *models.HLSSegment) error
	GetHLSSegments(ctx context.Context, hlsAnalysisID uuid.UUID, limit int) ([]*models.HLSSegment, error)

	// User operations
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)

	// API Key operations
	CreateAPIKey(ctx context.Context, apiKey *models.APIKey) error
	GetAPIKey(ctx context.Context, keyHash string) (*models.APIKey, error)
	UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error

	// Processing job operations
	CreateProcessingJob(ctx context.Context, job *models.ProcessingJob) error
	GetProcessingJob(ctx context.Context, id uuid.UUID) (*models.ProcessingJob, error)
	UpdateProcessingJob(ctx context.Context, job *models.ProcessingJob) error
	GetPendingJobs(ctx context.Context, jobType models.JobType, limit int) ([]models.ProcessingJob, error)

	// Cache operations
	CreateCacheEntry(ctx context.Context, entry *models.CacheEntry) error
	GetCacheEntry(ctx context.Context, contentHash string, cacheType models.CacheType) (*models.CacheEntry, error)
	UpdateCacheHit(ctx context.Context, id uuid.UUID) error
	CleanupExpiredCache(ctx context.Context) error

	// Report operations (placeholder)
	// CreateReport(ctx context.Context, report *models.Report) error
	// GetReport(ctx context.Context, id uuid.UUID) (*models.Report, error)
}

// SQLiteRepository implements Repository interface for SQLite
type SQLiteRepository struct {
	db *DB
}

// NewRepository creates a new SQLite repository
func NewRepository(db *DB) Repository {
	return &SQLiteRepository{db: db}
}

// CreateAnalysis creates a new analysis record
func (r *SQLiteRepository) CreateAnalysis(ctx context.Context, analysis *models.Analysis) error {
	query := `
		INSERT INTO analyses (id, user_id, file_name, file_path, file_size, content_hash, 
			source_type, status, ffprobe_data, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.DB.ExecContext(ctx, query,
		analysis.ID,
		analysis.UserID,
		analysis.FileName,
		analysis.FilePath,
		analysis.FileSize,
		analysis.ContentHash,
		analysis.SourceType,
		analysis.Status,
		analysis.FFprobeData,
		analysis.CreatedAt,
		analysis.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create analysis: %w", err)
	}

	return nil
}

// GetAnalysis retrieves an analysis by ID
func (r *SQLiteRepository) GetAnalysis(ctx context.Context, id uuid.UUID) (*models.Analysis, error) {
	query := `
		SELECT id, user_id, file_name, file_path, file_size, content_hash, source_type,
			status, ffprobe_data, llm_report, processed_at, created_at, updated_at, error_msg
		FROM analyses WHERE id = ?`

	var analysis models.Analysis
	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&analysis.ID,
		&analysis.UserID,
		&analysis.FileName,
		&analysis.FilePath,
		&analysis.FileSize,
		&analysis.ContentHash,
		&analysis.SourceType,
		&analysis.Status,
		&analysis.FFprobeData,
		&analysis.LLMReport,
		&analysis.ProcessedAt,
		&analysis.CreatedAt,
		&analysis.UpdatedAt,
		&analysis.ErrorMsg,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}

	return &analysis, nil
}

// GetAnalysesByUser retrieves analyses for a specific user with pagination
func (r *SQLiteRepository) GetAnalysesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Analysis, error) {
	query := `
		SELECT id, user_id, file_name, file_path, file_size, content_hash, source_type,
			status, ffprobe_data, llm_report, processed_at, created_at, updated_at, error_msg
		FROM analyses 
		WHERE user_id = ? 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?`

	rows, err := r.db.DB.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get analyses by user: %w", err)
	}
	defer rows.Close()

	var analyses []models.Analysis
	for rows.Next() {
		var analysis models.Analysis
		err := rows.Scan(
			&analysis.ID,
			&analysis.UserID,
			&analysis.FileName,
			&analysis.FilePath,
			&analysis.FileSize,
			&analysis.ContentHash,
			&analysis.SourceType,
			&analysis.Status,
			&analysis.FFprobeData,
			&analysis.LLMReport,
			&analysis.ProcessedAt,
			&analysis.CreatedAt,
			&analysis.UpdatedAt,
			&analysis.ErrorMsg,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analysis: %w", err)
		}
		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

// UpdateAnalysisStatus updates the status of an analysis
func (r *SQLiteRepository) UpdateAnalysisStatus(ctx context.Context, id uuid.UUID, status models.AnalysisStatus, errorMsg *string) error {
	query := `
		UPDATE analyses 
		SET status = ?, error_msg = ?, updated_at = ?
		WHERE id = ?`

	_, err := r.db.DB.ExecContext(ctx, query, status, errorMsg, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update analysis status: %w", err)
	}

	return nil
}

// UpdateAnalysisLLMReport updates the LLM report of an analysis
func (r *SQLiteRepository) UpdateAnalysisLLMReport(ctx context.Context, id uuid.UUID, report string) error {
	query := `
		UPDATE analyses 
		SET llm_report = ?, updated_at = ?
		WHERE id = ?`

	_, err := r.db.DB.ExecContext(ctx, query, report, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update analysis LLM report: %w", err)
	}

	return nil
}

// DeleteAnalysis deletes an analysis record
func (r *SQLiteRepository) DeleteAnalysis(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM analyses WHERE id = ?`

	_, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete analysis: %w", err)
	}

	return nil
}

// CreateQualityFrame creates a single quality frame record
func (r *SQLiteRepository) CreateQualityFrame(ctx context.Context, frame *models.QualityFrame) error {
	query := `
		INSERT INTO quality_frames (
			id, quality_metric_id, frame_number, timestamp, score, 
			component_scores, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.DB.ExecContext(
		ctx, query,
		frame.ID, frame.QualityMetricID, frame.FrameNumber, frame.Timestamp, frame.Score,
		frame.ComponentScores, frame.CreatedAt,
	)
	return err
}

// CreateQualityMetrics creates a quality metrics record
func (r *SQLiteRepository) CreateQualityMetrics(ctx context.Context, metrics *models.QualityMetrics) error {
	query := `
		INSERT INTO quality_metrics (
			id, analysis_id, reference_file_id, metric_type, overall_score,
			min_score, max_score, mean_score, std_deviation, percentile_data,
			frame_count, processing_time, model_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.DB.ExecContext(
		ctx, query,
		metrics.ID, metrics.AnalysisID, metrics.ReferenceFileID, metrics.MetricType, metrics.OverallScore,
		metrics.MinScore, metrics.MaxScore, metrics.MeanScore, metrics.StdDeviation, metrics.PercentileData,
		metrics.FrameCount, metrics.ProcessingTime, metrics.ModelVersion, metrics.CreatedAt,
	)
	return err
}

// GetQualityMetrics retrieves quality metrics for an analysis
func (r *SQLiteRepository) GetQualityMetrics(ctx context.Context, analysisID uuid.UUID) ([]models.QualityMetrics, error) {
	query := `
		SELECT id, analysis_id, reference_file_id, metric_type, overall_score,
			   min_score, max_score, mean_score, std_deviation, percentile_data,
			   frame_count, processing_time, model_version, created_at
		FROM quality_metrics WHERE analysis_id = ?
	`

	rows, err := r.db.DB.QueryContext(ctx, query, analysisID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []models.QualityMetrics
	for rows.Next() {
		var m models.QualityMetrics
		err := rows.Scan(
			&m.ID, &m.AnalysisID, &m.ReferenceFileID, &m.MetricType, &m.OverallScore,
			&m.MinScore, &m.MaxScore, &m.MeanScore, &m.StdDeviation, &m.PercentileData,
			&m.FrameCount, &m.ProcessingTime, &m.ModelVersion, &m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

// GetQualityFrames retrieves quality frames for a metric
func (r *SQLiteRepository) GetQualityFrames(ctx context.Context, metricID uuid.UUID, limit, offset int) ([]models.QualityFrame, error) {
	query := `
		SELECT id, quality_metric_id, frame_number, timestamp, score,
			   component_scores, created_at
		FROM quality_frames WHERE quality_metric_id = ?
		ORDER BY frame_number LIMIT ? OFFSET ?
	`

	rows, err := r.db.DB.QueryContext(ctx, query, metricID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var frames []models.QualityFrame
	for rows.Next() {
		var f models.QualityFrame
		err := rows.Scan(
			&f.ID, &f.QualityMetricID, &f.FrameNumber, &f.Timestamp, &f.Score,
			&f.ComponentScores, &f.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		frames = append(frames, f)
	}

	return frames, rows.Err()
}

// Quality metrics operations are implemented in quality_repository.go

func (r *SQLiteRepository) CreateHLSAnalysis(ctx context.Context, hls *models.HLSAnalysis) error {
	query := `
		INSERT INTO hls_analyses (
			id, analysis_id, manifest_path, manifest_type, manifest_data,
			segment_count, total_duration, bitrate_variants, segment_duration,
			playlist_version, status, processing_time, created_at, completed_at, error_msg
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	_, err := r.db.DB.ExecContext(
		ctx, query,
		hls.ID, hls.AnalysisID, hls.ManifestPath, hls.ManifestType, hls.ManifestData,
		hls.SegmentCount, hls.TotalDuration, hls.BitrateVariants, hls.SegmentDuration,
		hls.PlaylistVersion, hls.Status, hls.ProcessingTime, hls.CreatedAt, hls.CompletedAt, hls.ErrorMsg,
	)
	return err
}

func (r *SQLiteRepository) GetHLSAnalysis(ctx context.Context, analysisID uuid.UUID) (*models.HLSAnalysis, error) {
	query := `
		SELECT id, analysis_id, manifest_path, manifest_type, manifest_data,
			segment_count, total_duration, bitrate_variants, segment_duration,
			playlist_version, status, processing_time, created_at, completed_at, error_msg
		FROM hls_analyses
		WHERE analysis_id = ?
	`

	var hls models.HLSAnalysis
	err := r.db.DB.GetContext(ctx, &hls, query, analysisID)
	if err != nil {
		return nil, err
	}
	return &hls, nil
}

func (r *SQLiteRepository) CreateHLSSegment(ctx context.Context, segment *models.HLSSegment) error {
	query := `
		INSERT INTO hls_segments (
			id, hls_analysis_id, segment_uri, sequence_number, duration,
			file_size, bitrate, resolution, frame_rate, segment_data,
			quality_score, status, processed_at, created_at, error_msg
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	_, err := r.db.DB.ExecContext(
		ctx, query,
		segment.ID, segment.HLSAnalysisID, segment.SegmentURI, segment.SequenceNumber, segment.Duration,
		segment.FileSize, segment.Bitrate, segment.Resolution, segment.FrameRate, segment.SegmentData,
		segment.QualityScore, segment.Status, segment.ProcessedAt, segment.CreatedAt, segment.ErrorMsg,
	)
	return err
}

func (r *SQLiteRepository) GetHLSSegments(ctx context.Context, hlsAnalysisID uuid.UUID, limit int) ([]*models.HLSSegment, error) {
	query := `
		SELECT id, hls_analysis_id, segment_uri, sequence_number, duration,
			file_size, bitrate, resolution, frame_rate, segment_data,
			quality_score, status, processed_at, created_at, error_msg
		FROM hls_segments
		WHERE hls_analysis_id = ?
		ORDER BY sequence_number
		LIMIT ?
	`

	var segments []*models.HLSSegment
	err := r.db.DB.SelectContext(ctx, &segments, query, hlsAnalysisID, limit)
	if err != nil {
		return nil, err
	}
	return segments, nil
}

func (r *SQLiteRepository) GetHLSAnalysisByAnalysisID(ctx context.Context, analysisID uuid.UUID) (*models.HLSAnalysis, error) {
	query := `
		SELECT id, analysis_id, manifest_path, manifest_type, manifest_data,
			segment_count, total_duration, bitrate_variants, segment_duration,
			playlist_version, status, processing_time, created_at, completed_at, error_msg
		FROM hls_analyses
		WHERE analysis_id = ?
	`

	var hls models.HLSAnalysis
	err := r.db.DB.GetContext(ctx, &hls, query, analysisID)
	if err != nil {
		return nil, err
	}
	return &hls, nil
}

func (r *SQLiteRepository) ListHLSAnalyses(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]*models.HLSAnalysis, int, error) {
	var analyses []*models.HLSAnalysis
	var total int

	baseQuery := `FROM hls_analyses h JOIN analyses a ON h.analysis_id = a.id`
	whereClause := ""
	args := []interface{}{}
	argCount := 0

	if userID != nil {
		argCount++
		whereClause = fmt.Sprintf(" WHERE a.user_id = $%d", argCount)
		args = append(args, *userID)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery + whereClause
	err := r.db.DB.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	argCount++
	limitArg := argCount
	args = append(args, limit)
	argCount++
	offsetArg := argCount
	args = append(args, offset)

	query := fmt.Sprintf(`
		SELECT h.id, h.analysis_id, h.manifest_path, h.manifest_type, h.manifest_data,
			h.segment_count, h.total_duration, h.bitrate_variants, h.segment_duration,
			h.playlist_version, h.status, h.processing_time, h.created_at, h.completed_at, h.error_msg
		%s %s
		ORDER BY h.created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, limitArg, offsetArg)

	err = r.db.DB.SelectContext(ctx, &analyses, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return analyses, total, nil
}

// User and API Key operations - stub implementations for basic functionality
func (r *SQLiteRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, role, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.DB.ExecContext(ctx, query,
		user.ID, user.Email, user.Username, user.PasswordHash,
		user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *SQLiteRepository) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE id = ?`

	var user models.User
	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SQLiteRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE email = ?`

	var user models.User
	err := r.db.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SQLiteRepository) CreateAPIKey(ctx context.Context, apiKey *models.APIKey) error {
	query := `
		INSERT INTO api_keys (id, user_id, key_hash, name, permissions, is_active, expires_at, created_at, last_used)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.DB.ExecContext(ctx, query,
		apiKey.ID, apiKey.UserID, apiKey.KeyHash, apiKey.Name,
		apiKey.Permissions, apiKey.IsActive, apiKey.ExpiresAt,
		apiKey.CreatedAt, apiKey.LastUsed)
	return err
}

func (r *SQLiteRepository) GetAPIKey(ctx context.Context, keyHash string) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, name, permissions, is_active, expires_at, created_at, last_used
		FROM api_keys WHERE key_hash = ? AND is_active = true`

	var apiKey models.APIKey
	err := r.db.DB.QueryRowContext(ctx, query, keyHash).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.KeyHash, &apiKey.Name,
		&apiKey.Permissions, &apiKey.IsActive, &apiKey.ExpiresAt,
		&apiKey.CreatedAt, &apiKey.LastUsed)
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *SQLiteRepository) UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE api_keys SET last_used = datetime('now') WHERE id = ?`
	_, err := r.db.DB.ExecContext(ctx, query, id)
	return err
}

// Report methods
func (r *SQLiteRepository) CreateReport(ctx context.Context, report *models.Report) error {
	query := `
		INSERT INTO reports (
			id, analysis_id, user_id, report_type, format, title, description,
			file_path, file_size, download_count, is_public, expires_at, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	_, err := r.db.DB.ExecContext(
		ctx, query,
		report.ID, report.AnalysisID, report.UserID, report.ReportType, report.Format,
		report.Title, report.Description, report.FilePath, report.FileSize,
		report.DownloadCount, report.IsPublic, report.ExpiresAt, report.CreatedAt,
	)
	return err
}

func (r *SQLiteRepository) GetReport(ctx context.Context, id uuid.UUID) (*models.Report, error) {
	query := `
		SELECT id, analysis_id, user_id, report_type, format, title, description,
			file_path, file_size, download_count, is_public, expires_at, created_at, last_download
		FROM reports
		WHERE id = ?
	`

	var report models.Report
	err := r.db.DB.GetContext(ctx, &report, query, id)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *SQLiteRepository) ListReports(ctx context.Context, userID *uuid.UUID, analysisID, reportType, format string, limit, offset int) ([]*models.Report, int, error) {
	var reports []*models.Report
	var total int

	baseQuery := "FROM reports"
	whereConditions := []string{}
	args := []interface{}{}
	argCount := 0

	if userID != nil {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("user_id = $%d", argCount))
		args = append(args, *userID)
	}

	if analysisID != "" {
		if analysisUUID, err := uuid.Parse(analysisID); err == nil {
			argCount++
			whereConditions = append(whereConditions, fmt.Sprintf("analysis_id = $%d", argCount))
			args = append(args, analysisUUID)
		}
	}

	if reportType != "" {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("report_type = $%d", argCount))
		args = append(args, reportType)
	}

	if format != "" {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("format = $%d", argCount))
		args = append(args, format)
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery + whereClause
	err := r.db.DB.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	argCount++
	limitArg := argCount
	args = append(args, limit)
	argCount++
	offsetArg := argCount
	args = append(args, offset)

	query := fmt.Sprintf(`
		SELECT id, analysis_id, user_id, report_type, format, title, description,
			file_path, file_size, download_count, is_public, expires_at, created_at, last_download
		%s %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, limitArg, offsetArg)

	err = r.db.DB.SelectContext(ctx, &reports, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

func (r *SQLiteRepository) DeleteReport(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM reports WHERE id = ?"
	_, err := r.db.DB.ExecContext(ctx, query, id)
	return err
}

func (r *SQLiteRepository) IncrementReportDownloadCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE reports 
		SET download_count = download_count + 1, last_download = datetime('now')
		WHERE id = ?
	`
	_, err := r.db.DB.ExecContext(ctx, query, id)
	return err
}

// Quality comparison methods
func (r *SQLiteRepository) CreateQualityComparison(ctx context.Context, comparison *models.QualityComparison) error {
	query := `
		INSERT INTO quality_comparisons (
			id, reference_id, distorted_id, comparison_type, status,
			result_summary, processing_time, created_at, completed_at, error_msg
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	_, err := r.db.DB.ExecContext(
		ctx, query,
		comparison.ID, comparison.ReferenceID, comparison.DistortedID, comparison.ComparisonType,
		comparison.Status, comparison.ResultSummary, comparison.ProcessingTime,
		comparison.CreatedAt, comparison.CompletedAt, comparison.ErrorMsg,
	)
	return err
}

func (r *SQLiteRepository) GetQualityComparison(ctx context.Context, id uuid.UUID) (*models.QualityComparison, error) {
	query := `
		SELECT id, reference_id, distorted_id, comparison_type, status,
			result_summary, processing_time, created_at, completed_at, error_msg
		FROM quality_comparisons
		WHERE id = ?
	`

	var comparison models.QualityComparison
	err := r.db.DB.GetContext(ctx, &comparison, query, id)
	if err != nil {
		return nil, err
	}
	return &comparison, nil
}

func (r *SQLiteRepository) UpdateQualityComparison(ctx context.Context, comparison *models.QualityComparison) error {
	query := `
		UPDATE quality_comparisons 
		SET status = ?, result_summary = ?, processing_time = ?,
			completed_at = ?, error_msg = ?, updated_at = datetime('now')
		WHERE id = ?
	`

	_, err := r.db.DB.ExecContext(
		ctx, query,
		comparison.ID, comparison.Status, comparison.ResultSummary, comparison.ProcessingTime,
		comparison.CompletedAt, comparison.ErrorMsg,
	)
	return err
}

func (r *SQLiteRepository) UpdateQualityComparisonStatus(ctx context.Context, id uuid.UUID, status models.AnalysisStatus) error {
	query := "UPDATE quality_comparisons SET status = ?, updated_at = datetime('now') WHERE id = ?"
	_, err := r.db.DB.ExecContext(ctx, query, id, status)
	return err
}

func (r *SQLiteRepository) ListQualityComparisons(ctx context.Context, userID *uuid.UUID, referenceID, distortedID, status string, limit, offset int) ([]*models.QualityComparison, int, error) {
	var comparisons []*models.QualityComparison
	var total int

	baseQuery := `FROM quality_comparisons qc 
		LEFT JOIN analyses ref ON qc.reference_id = ref.id
		LEFT JOIN analyses dist ON qc.distorted_id = dist.id`
	whereConditions := []string{}
	args := []interface{}{}
	argCount := 0

	if userID != nil {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("(ref.user_id = $%d OR dist.user_id = $%d)", argCount, argCount))
		args = append(args, *userID)
	}

	if referenceID != "" {
		if refUUID, err := uuid.Parse(referenceID); err == nil {
			argCount++
			whereConditions = append(whereConditions, fmt.Sprintf("qc.reference_id = $%d", argCount))
			args = append(args, refUUID)
		}
	}

	if distortedID != "" {
		if distUUID, err := uuid.Parse(distortedID); err == nil {
			argCount++
			whereConditions = append(whereConditions, fmt.Sprintf("qc.distorted_id = $%d", argCount))
			args = append(args, distUUID)
		}
	}

	if status != "" {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("qc.status = $%d", argCount))
		args = append(args, status)
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery + whereClause
	err := r.db.DB.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	argCount++
	limitArg := argCount
	args = append(args, limit)
	argCount++
	offsetArg := argCount
	args = append(args, offset)

	query := fmt.Sprintf(`
		SELECT qc.id, qc.reference_id, qc.distorted_id, qc.comparison_type, qc.status,
			qc.result_summary, qc.processing_time, qc.created_at, qc.completed_at, qc.error_msg
		%s %s
		ORDER BY qc.created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, limitArg, offsetArg)

	err = r.db.DB.SelectContext(ctx, &comparisons, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return comparisons, total, nil
}

func (r *SQLiteRepository) DeleteQualityComparison(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM quality_comparisons WHERE id = ?"
	result, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("quality comparison not found")
	}

	return nil
}

// Processing job operations - basic implementations
func (r *SQLiteRepository) CreateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	query := `
		INSERT INTO processing_jobs (id, analysis_id, job_type, status, priority, 
			scheduled_at, started_at, completed_at, error_msg, retry_count, max_retries)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.DB.ExecContext(ctx, query,
		job.ID, job.AnalysisID, job.JobType, job.Status, job.Priority,
		job.ScheduledAt, job.StartedAt, job.CompletedAt, job.ErrorMsg,
		job.RetryCount, job.MaxRetries)
	return err
}

func (r *SQLiteRepository) GetProcessingJob(ctx context.Context, id uuid.UUID) (*models.ProcessingJob, error) {
	query := `
		SELECT id, analysis_id, job_type, status, priority, scheduled_at, 
			started_at, completed_at, error_msg, retry_count, max_retries, created_at
		FROM processing_jobs WHERE id = ?`

	var job models.ProcessingJob
	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&job.ID, &job.AnalysisID, &job.JobType, &job.Status, &job.Priority,
		&job.ScheduledAt, &job.StartedAt, &job.CompletedAt, &job.ErrorMsg,
		&job.RetryCount, &job.MaxRetries, &job.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *SQLiteRepository) UpdateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	query := `
		UPDATE processing_jobs 
		SET status = ?, started_at = ?, completed_at = ?, error_msg = ?, retry_count = ?
		WHERE id = ?`

	_, err := r.db.DB.ExecContext(ctx, query,
		job.Status, job.StartedAt, job.CompletedAt, job.ErrorMsg, job.RetryCount, job.ID)
	return err
}

func (r *SQLiteRepository) GetPendingJobs(ctx context.Context, jobType models.JobType, limit int) ([]models.ProcessingJob, error) {
	query := `
		SELECT id, analysis_id, job_type, status, priority, scheduled_at, 
			started_at, completed_at, error_msg, retry_count, max_retries, created_at
		FROM processing_jobs 
		WHERE job_type = ? AND status = 'pending' 
		ORDER BY priority DESC, scheduled_at ASC 
		LIMIT ?`

	rows, err := r.db.DB.QueryContext(ctx, query, jobType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.ProcessingJob
	for rows.Next() {
		var job models.ProcessingJob
		err := rows.Scan(
			&job.ID, &job.AnalysisID, &job.JobType, &job.Status, &job.Priority,
			&job.ScheduledAt, &job.StartedAt, &job.CompletedAt, &job.ErrorMsg,
			&job.RetryCount, &job.MaxRetries, &job.CreatedAt)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// Cache operations - basic implementations
func (r *SQLiteRepository) CreateCacheEntry(ctx context.Context, entry *models.CacheEntry) error {
	query := `
		INSERT INTO cache_entries (id, content_hash, cache_type, file_path, 
			hit_count, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.DB.ExecContext(ctx, query,
		entry.ID, entry.ContentHash, entry.CacheType, entry.FilePath,
		entry.HitCount, entry.ExpiresAt, entry.CreatedAt)
	return err
}

func (r *SQLiteRepository) GetCacheEntry(ctx context.Context, contentHash string, cacheType models.CacheType) (*models.CacheEntry, error) {
	query := `
		SELECT id, content_hash, cache_type, file_path, hit_count, expires_at, created_at
		FROM cache_entries 
		WHERE content_hash = ? AND cache_type = ? AND expires_at > datetime('now')`

	var entry models.CacheEntry
	err := r.db.DB.QueryRowContext(ctx, query, contentHash, cacheType).Scan(
		&entry.ID, &entry.ContentHash, &entry.CacheType, &entry.FilePath,
		&entry.HitCount, &entry.ExpiresAt, &entry.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *SQLiteRepository) UpdateCacheHit(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE cache_entries SET hit_count = hit_count + 1 WHERE id = ?`
	_, err := r.db.DB.ExecContext(ctx, query, id)
	return err
}

func (r *SQLiteRepository) CleanupExpiredCache(ctx context.Context) error {
	query := `DELETE FROM cache_entries WHERE expires_at <= datetime('now')`
	_, err := r.db.DB.ExecContext(ctx, query)
	return err
}
