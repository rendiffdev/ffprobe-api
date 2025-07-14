package database

import (
	"context"
	"database/sql"
	"errors"
	
	"github.com/google/uuid"
	"github.com/rendiffdev/ffprobe-api/internal/models"
)

var (
	// ErrNotFound is returned when a requested resource is not found
	ErrNotFound = errors.New("not found")
)

// Repository embeds PostgreSQLRepository to provide all database methods
type repository struct {
	*PostgreSQLRepository
}

// newRepository creates a new repository instance
func newRepository(db *DB) *repository {
	return &repository{
		PostgreSQLRepository: NewPostgreSQLRepository(db.SQLX),
	}
}

// Analysis methods
func (db *DB) CreateAnalysis(ctx context.Context, analysis *models.Analysis) error {
	repo := newRepository(db)
	return repo.CreateAnalysis(ctx, analysis)
}

func (db *DB) GetAnalysis(ctx context.Context, id uuid.UUID) (*models.Analysis, error) {
	repo := newRepository(db)
	return repo.GetAnalysis(ctx, id)
}

func (db *DB) UpdateAnalysis(ctx context.Context, analysis *models.Analysis) error {
	repo := newRepository(db)
	return repo.UpdateAnalysis(ctx, analysis)
}

func (db *DB) UpdateAnalysisStatus(ctx context.Context, id string, status models.AnalysisStatus) error {
	analysisID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	repo := newRepository(db)
	return repo.UpdateAnalysisStatus(ctx, analysisID, status)
}

func (db *DB) DeleteAnalysis(ctx context.Context, id uuid.UUID) error {
	repo := newRepository(db)
	return repo.DeleteAnalysis(ctx, id)
}

func (db *DB) ListAnalyses(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]*models.Analysis, error) {
	repo := newRepository(db)
	return repo.ListAnalyses(ctx, userID, limit, offset)
}

// HLS methods
func (db *DB) CreateHLSAnalysis(ctx context.Context, hls *models.HLSAnalysis) error {
	repo := newRepository(db)
	return repo.CreateHLSAnalysis(ctx, hls)
}

func (db *DB) GetHLSAnalysis(ctx context.Context, analysisID uuid.UUID) (*models.HLSAnalysis, error) {
	repo := newRepository(db)
	return repo.GetHLSAnalysis(ctx, analysisID)
}

func (db *DB) GetHLSAnalysisByAnalysisID(ctx context.Context, analysisID uuid.UUID) (*models.HLSAnalysis, error) {
	repo := newRepository(db)
	analysis, err := repo.GetHLSAnalysisByAnalysisID(ctx, analysisID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return analysis, nil
}

func (db *DB) CreateHLSSegment(ctx context.Context, segment *models.HLSSegment) error {
	repo := newRepository(db)
	return repo.CreateHLSSegment(ctx, segment)
}

func (db *DB) GetHLSSegments(ctx context.Context, hlsAnalysisID uuid.UUID, limit int) ([]*models.HLSSegment, error) {
	repo := newRepository(db)
	return repo.GetHLSSegments(ctx, hlsAnalysisID, limit)
}

func (db *DB) ListHLSAnalyses(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]*models.HLSAnalysis, int, error) {
	repo := newRepository(db)
	return repo.ListHLSAnalyses(ctx, userID, limit, offset)
}

// Quality methods
func (db *DB) CreateQualityMetric(ctx context.Context, metric *models.QualityMetric) error {
	repo := newRepository(db)
	return repo.CreateQualityMetric(ctx, metric)
}

func (db *DB) GetQualityMetrics(ctx context.Context, analysisID uuid.UUID) ([]*models.QualityMetric, error) {
	repo := newRepository(db)
	return repo.GetQualityMetrics(ctx, analysisID)
}

func (db *DB) CreateQualityFrame(ctx context.Context, frames []*models.QualityFrame) error {
	repo := newRepository(db)
	return repo.CreateQualityFrame(ctx, frames)
}

func (db *DB) GetQualityFrames(ctx context.Context, metricID uuid.UUID, limit, offset int) ([]*models.QualityFrame, error) {
	repo := newRepository(db)
	return repo.GetQualityFrames(ctx, metricID, limit, offset)
}

func (db *DB) CreateQualityComparison(ctx context.Context, comparison *models.QualityComparison) error {
	repo := newRepository(db)
	return repo.CreateQualityComparison(ctx, comparison)
}

func (db *DB) GetQualityComparison(ctx context.Context, id uuid.UUID) (*models.QualityComparison, error) {
	repo := newRepository(db)
	return repo.GetQualityComparison(ctx, id)
}

// User methods
func (db *DB) CreateUser(ctx context.Context, user *models.User) error {
	repo := newRepository(db)
	return repo.CreateUser(ctx, user)
}

func (db *DB) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	repo := newRepository(db)
	return repo.GetUser(ctx, id)
}

func (db *DB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	repo := newRepository(db)
	return repo.GetUserByEmail(ctx, email)
}

// API Key methods
func (db *DB) CreateAPIKey(ctx context.Context, apiKey *models.APIKey) error {
	repo := newRepository(db)
	return repo.CreateAPIKey(ctx, apiKey)
}

func (db *DB) GetAPIKey(ctx context.Context, keyHash string) (*models.APIKey, error) {
	repo := newRepository(db)
	return repo.GetAPIKey(ctx, keyHash)
}

func (db *DB) UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error {
	repo := newRepository(db)
	return repo.UpdateAPIKeyLastUsed(ctx, id)
}

// Report methods
func (db *DB) CreateReport(ctx context.Context, report *models.Report) error {
	repo := newRepository(db)
	return repo.CreateReport(ctx, report)
}

func (db *DB) GetReport(ctx context.Context, id uuid.UUID) (*models.Report, error) {
	repo := newRepository(db)
	report, err := repo.GetReport(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return report, nil
}

func (db *DB) ListReports(ctx context.Context, userID *uuid.UUID, analysisID, reportType, format string, limit, offset int) ([]*models.Report, int, error) {
	repo := newRepository(db)
	return repo.ListReports(ctx, userID, analysisID, reportType, format, limit, offset)
}

func (db *DB) DeleteReport(ctx context.Context, id uuid.UUID) error {
	repo := newRepository(db)
	return repo.DeleteReport(ctx, id)
}

func (db *DB) IncrementReportDownloadCount(ctx context.Context, id uuid.UUID) error {
	repo := newRepository(db)
	return repo.IncrementReportDownloadCount(ctx, id)
}

// Quality comparison methods
func (db *DB) CreateQualityComparison(ctx context.Context, comparison *models.QualityComparison) error {
	repo := newRepository(db)
	return repo.CreateQualityComparison(ctx, comparison)
}

func (db *DB) GetQualityComparison(ctx context.Context, id uuid.UUID) (*models.QualityComparison, error) {
	repo := newRepository(db)
	comparison, err := repo.GetQualityComparison(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return comparison, nil
}

func (db *DB) UpdateQualityComparison(ctx context.Context, comparison *models.QualityComparison) error {
	repo := newRepository(db)
	return repo.UpdateQualityComparison(ctx, comparison)
}

func (db *DB) UpdateQualityComparisonStatus(ctx context.Context, id uuid.UUID, status models.AnalysisStatus) error {
	repo := newRepository(db)
	return repo.UpdateQualityComparisonStatus(ctx, id, status)
}

func (db *DB) ListQualityComparisons(ctx context.Context, userID *uuid.UUID, referenceID, distortedID, status string, limit, offset int) ([]*models.QualityComparison, int, error) {
	repo := newRepository(db)
	return repo.ListQualityComparisons(ctx, userID, referenceID, distortedID, status, limit, offset)
}

func (db *DB) DeleteQualityComparison(ctx context.Context, id uuid.UUID) error {
	repo := newRepository(db)
	return repo.DeleteQualityComparison(ctx, id)
}