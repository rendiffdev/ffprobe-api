package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rendiffdev/ffprobe-api/internal/models"
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
	GetHLSSegments(ctx context.Context, hlsAnalysisID uuid.UUID) ([]models.HLSSegment, error)

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

	// Report operations
	CreateReport(ctx context.Context, report *models.Report) error
	GetReport(ctx context.Context, id uuid.UUID) (*models.Report, error)
	GetReportsByAnalysis(ctx context.Context, analysisID uuid.UUID) ([]models.Report, error)
	UpdateReportDownload(ctx context.Context, id uuid.UUID) error
}

// PostgreSQLRepository implements Repository interface for PostgreSQL
type PostgreSQLRepository struct {
	db *DB
}

// NewRepository creates a new PostgreSQL repository
func NewRepository(db *DB) Repository {
	return &PostgreSQLRepository{db: db}
}

// CreateAnalysis creates a new analysis record
func (r *PostgreSQLRepository) CreateAnalysis(ctx context.Context, analysis *models.Analysis) error {
	query := `
		INSERT INTO analyses (id, user_id, file_name, file_path, file_size, content_hash, 
			source_type, status, ffprobe_data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.Pool.Exec(ctx, query,
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
func (r *PostgreSQLRepository) GetAnalysis(ctx context.Context, id uuid.UUID) (*models.Analysis, error) {
	query := `
		SELECT id, user_id, file_name, file_path, file_size, content_hash, source_type,
			status, ffprobe_data, llm_report, processed_at, created_at, updated_at, error_msg
		FROM analyses WHERE id = $1`

	var analysis models.Analysis
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
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
func (r *PostgreSQLRepository) GetAnalysesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Analysis, error) {
	query := `
		SELECT id, user_id, file_name, file_path, file_size, content_hash, source_type,
			status, ffprobe_data, llm_report, processed_at, created_at, updated_at, error_msg
		FROM analyses 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
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
func (r *PostgreSQLRepository) UpdateAnalysisStatus(ctx context.Context, id uuid.UUID, status models.AnalysisStatus, errorMsg *string) error {
	query := `
		UPDATE analyses 
		SET status = $2, error_msg = $3, updated_at = $4
		WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id, status, errorMsg, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update analysis status: %w", err)
	}

	return nil
}

// UpdateAnalysisLLMReport updates the LLM report of an analysis
func (r *PostgreSQLRepository) UpdateAnalysisLLMReport(ctx context.Context, id uuid.UUID, report string) error {
	query := `
		UPDATE analyses 
		SET llm_report = $2, updated_at = $3
		WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id, report, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update analysis LLM report: %w", err)
	}

	return nil
}

// DeleteAnalysis deletes an analysis record
func (r *PostgreSQLRepository) DeleteAnalysis(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM analyses WHERE id = $1`
	
	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete analysis: %w", err)
	}

	return nil
}

// Quality metrics operations are implemented in quality_repository.go

func (r *PostgreSQLRepository) CreateHLSAnalysis(ctx context.Context, hls *models.HLSAnalysis) error {
	query := `
		INSERT INTO hls_analyses (
			id, analysis_id, manifest_path, manifest_type, manifest_data,
			segment_count, total_duration, bitrate_variants, segment_duration,
			playlist_version, status, processing_time, created_at, completed_at, error_msg
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
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

func (r *PostgreSQLRepository) GetHLSAnalysis(ctx context.Context, analysisID uuid.UUID) (*models.HLSAnalysis, error) {
	query := `
		SELECT id, analysis_id, manifest_path, manifest_type, manifest_data,
			segment_count, total_duration, bitrate_variants, segment_duration,
			playlist_version, status, processing_time, created_at, completed_at, error_msg
		FROM hls_analyses
		WHERE analysis_id = $1
	`
	
	var hls models.HLSAnalysis
	err := r.db.DB.GetContext(ctx, &hls, query, analysisID)
	if err != nil {
		return nil, err
	}
	return &hls, nil
}

func (r *PostgreSQLRepository) CreateHLSSegment(ctx context.Context, segment *models.HLSSegment) error {
	query := `
		INSERT INTO hls_segments (
			id, hls_analysis_id, segment_uri, sequence_number, duration,
			file_size, bitrate, resolution, frame_rate, segment_data,
			quality_score, status, processed_at, created_at, error_msg
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
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

func (r *PostgreSQLRepository) GetHLSSegments(ctx context.Context, hlsAnalysisID uuid.UUID, limit int) ([]*models.HLSSegment, error) {
	query := `
		SELECT id, hls_analysis_id, segment_uri, sequence_number, duration,
			file_size, bitrate, resolution, frame_rate, segment_data,
			quality_score, status, processed_at, created_at, error_msg
		FROM hls_segments
		WHERE hls_analysis_id = $1
		ORDER BY sequence_number
		LIMIT $2
	`
	
	var segments []*models.HLSSegment
	err := r.db.DB.SelectContext(ctx, &segments, query, hlsAnalysisID, limit)
	if err != nil {
		return nil, err
	}
	return segments, nil
}

func (r *PostgreSQLRepository) GetHLSAnalysisByAnalysisID(ctx context.Context, analysisID uuid.UUID) (*models.HLSAnalysis, error) {
	query := `
		SELECT id, analysis_id, manifest_path, manifest_type, manifest_data,
			segment_count, total_duration, bitrate_variants, segment_duration,
			playlist_version, status, processing_time, created_at, completed_at, error_msg
		FROM hls_analyses
		WHERE analysis_id = $1
	`
	
	var hls models.HLSAnalysis
	err := r.db.DB.GetContext(ctx, &hls, query, analysisID)
	if err != nil {
		return nil, err
	}
	return &hls, nil
}

func (r *PostgreSQLRepository) ListHLSAnalyses(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]*models.HLSAnalysis, int, error) {
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
func (r *PostgreSQLRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err := r.db.Pool.Exec(ctx, query,
		user.ID, user.Email, user.Username, user.PasswordHash,
		user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *PostgreSQLRepository) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE id = $1`
	
	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgreSQLRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE email = $1`
	
	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgreSQLRepository) CreateAPIKey(ctx context.Context, apiKey *models.APIKey) error {
	query := `
		INSERT INTO api_keys (id, user_id, key_hash, name, permissions, is_active, expires_at, created_at, last_used)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	_, err := r.db.Pool.Exec(ctx, query,
		apiKey.ID, apiKey.UserID, apiKey.KeyHash, apiKey.Name,
		apiKey.Permissions, apiKey.IsActive, apiKey.ExpiresAt,
		apiKey.CreatedAt, apiKey.LastUsed)
	return err
}

func (r *PostgreSQLRepository) GetAPIKey(ctx context.Context, keyHash string) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, name, permissions, is_active, expires_at, created_at, last_used
		FROM api_keys WHERE key_hash = $1 AND is_active = true`
	
	var apiKey models.APIKey
	err := r.db.Pool.QueryRow(ctx, query, keyHash).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.KeyHash, &apiKey.Name,
		&apiKey.Permissions, &apiKey.IsActive, &apiKey.ExpiresAt,
		&apiKey.CreatedAt, &apiKey.LastUsed)
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *PostgreSQLRepository) UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE api_keys SET last_used = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

// Report methods
func (r *PostgreSQLRepository) CreateReport(ctx context.Context, report *models.Report) error {
	query := `
		INSERT INTO reports (
			id, analysis_id, user_id, report_type, format, title, description,
			file_path, file_size, download_count, is_public, expires_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
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

func (r *PostgreSQLRepository) GetReport(ctx context.Context, id uuid.UUID) (*models.Report, error) {
	query := `
		SELECT id, analysis_id, user_id, report_type, format, title, description,
			file_path, file_size, download_count, is_public, expires_at, created_at, last_download
		FROM reports
		WHERE id = $1
	`
	
	var report models.Report
	err := r.db.DB.GetContext(ctx, &report, query, id)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *PostgreSQLRepository) ListReports(ctx context.Context, userID *uuid.UUID, analysisID, reportType, format string, limit, offset int) ([]*models.Report, int, error) {
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

func (r *PostgreSQLRepository) DeleteReport(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM reports WHERE id = $1"
	_, err := r.db.DB.ExecContext(ctx, query, id)
	return err
}

func (r *PostgreSQLRepository) IncrementReportDownloadCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE reports 
		SET download_count = download_count + 1, last_download = NOW()
		WHERE id = $1
	`
	_, err := r.db.DB.ExecContext(ctx, query, id)
	return err
}

// Quality comparison methods
func (r *PostgreSQLRepository) CreateQualityComparison(ctx context.Context, comparison *models.QualityComparison) error {
	query := `
		INSERT INTO quality_comparisons (
			id, reference_id, distorted_id, comparison_type, status,
			result_summary, processing_time, created_at, completed_at, error_msg
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
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

func (r *PostgreSQLRepository) GetQualityComparison(ctx context.Context, id uuid.UUID) (*models.QualityComparison, error) {
	query := `
		SELECT id, reference_id, distorted_id, comparison_type, status,
			result_summary, processing_time, created_at, completed_at, error_msg
		FROM quality_comparisons
		WHERE id = $1
	`
	
	var comparison models.QualityComparison
	err := r.db.DB.GetContext(ctx, &comparison, query, id)
	if err != nil {
		return nil, err
	}
	return &comparison, nil
}

func (r *PostgreSQLRepository) UpdateQualityComparison(ctx context.Context, comparison *models.QualityComparison) error {
	query := `
		UPDATE quality_comparisons 
		SET status = $2, result_summary = $3, processing_time = $4,
			completed_at = $5, error_msg = $6, updated_at = NOW()
		WHERE id = $1
	`
	
	_, err := r.db.DB.ExecContext(
		ctx, query,
		comparison.ID, comparison.Status, comparison.ResultSummary, comparison.ProcessingTime,
		comparison.CompletedAt, comparison.ErrorMsg,
	)
	return err
}

func (r *PostgreSQLRepository) UpdateQualityComparisonStatus(ctx context.Context, id uuid.UUID, status models.AnalysisStatus) error {
	query := "UPDATE quality_comparisons SET status = $2, updated_at = NOW() WHERE id = $1"
	_, err := r.db.DB.ExecContext(ctx, query, id, status)
	return err
}

func (r *PostgreSQLRepository) ListQualityComparisons(ctx context.Context, userID *uuid.UUID, referenceID, distortedID, status string, limit, offset int) ([]*models.QualityComparison, int, error) {
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

func (r *PostgreSQLRepository) DeleteQualityComparison(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM quality_comparisons WHERE id = $1"
	result, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return database.ErrNotFound
	}
	
	return nil
}

// Processing job operations - basic implementations
func (r *PostgreSQLRepository) CreateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	query := `
		INSERT INTO processing_jobs (id, analysis_id, job_type, status, priority, 
			scheduled_at, started_at, completed_at, error_msg, retry_count, max_retries)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	
	_, err := r.db.Pool.Exec(ctx, query,
		job.ID, job.AnalysisID, job.JobType, job.Status, job.Priority,
		job.ScheduledAt, job.StartedAt, job.CompletedAt, job.ErrorMsg,
		job.RetryCount, job.MaxRetries)
	return err
}

func (r *PostgreSQLRepository) GetProcessingJob(ctx context.Context, id uuid.UUID) (*models.ProcessingJob, error) {
	query := `
		SELECT id, analysis_id, job_type, status, priority, scheduled_at, 
			started_at, completed_at, error_msg, retry_count, max_retries, created_at
		FROM processing_jobs WHERE id = $1`
	
	var job models.ProcessingJob
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&job.ID, &job.AnalysisID, &job.JobType, &job.Status, &job.Priority,
		&job.ScheduledAt, &job.StartedAt, &job.CompletedAt, &job.ErrorMsg,
		&job.RetryCount, &job.MaxRetries, &job.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *PostgreSQLRepository) UpdateProcessingJob(ctx context.Context, job *models.ProcessingJob) error {
	query := `
		UPDATE processing_jobs 
		SET status = $2, started_at = $3, completed_at = $4, error_msg = $5, retry_count = $6
		WHERE id = $1`
	
	_, err := r.db.Pool.Exec(ctx, query,
		job.ID, job.Status, job.StartedAt, job.CompletedAt, job.ErrorMsg, job.RetryCount)
	return err
}

func (r *PostgreSQLRepository) GetPendingJobs(ctx context.Context, jobType models.JobType, limit int) ([]models.ProcessingJob, error) {
	query := `
		SELECT id, analysis_id, job_type, status, priority, scheduled_at, 
			started_at, completed_at, error_msg, retry_count, max_retries, created_at
		FROM processing_jobs 
		WHERE job_type = $1 AND status = 'pending' 
		ORDER BY priority DESC, scheduled_at ASC 
		LIMIT $2`
	
	rows, err := r.db.Pool.Query(ctx, query, jobType, limit)
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
func (r *PostgreSQLRepository) CreateCacheEntry(ctx context.Context, entry *models.CacheEntry) error {
	query := `
		INSERT INTO cache_entries (id, content_hash, cache_type, file_path, 
			hit_count, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	_, err := r.db.Pool.Exec(ctx, query,
		entry.ID, entry.ContentHash, entry.CacheType, entry.FilePath,
		entry.HitCount, entry.ExpiresAt, entry.CreatedAt)
	return err
}

func (r *PostgreSQLRepository) GetCacheEntry(ctx context.Context, contentHash string, cacheType models.CacheType) (*models.CacheEntry, error) {
	query := `
		SELECT id, content_hash, cache_type, file_path, hit_count, expires_at, created_at
		FROM cache_entries 
		WHERE content_hash = $1 AND cache_type = $2 AND expires_at > NOW()`
	
	var entry models.CacheEntry
	err := r.db.Pool.QueryRow(ctx, query, contentHash, cacheType).Scan(
		&entry.ID, &entry.ContentHash, &entry.CacheType, &entry.FilePath,
		&entry.HitCount, &entry.ExpiresAt, &entry.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *PostgreSQLRepository) UpdateCacheHit(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE cache_entries SET hit_count = hit_count + 1 WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *PostgreSQLRepository) CleanupExpiredCache(ctx context.Context) error {
	query := `DELETE FROM cache_entries WHERE expires_at <= NOW()`
	_, err := r.db.Pool.Exec(ctx, query)
	return err
}