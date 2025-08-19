package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rendiffdev/ffprobe-api/internal/models"
)

// ComparisonRepository defines the interface for comparison data operations
type ComparisonRepository interface {
	Create(ctx context.Context, comparison *models.VideoComparison) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.VideoComparison, error)
	Update(ctx context.Context, comparison *models.VideoComparison) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByUser(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]*models.VideoComparison, error)
	ListByAnalysis(ctx context.Context, analysisID uuid.UUID) ([]*models.VideoComparison, error)
	GetComparisonHistory(ctx context.Context, originalID, modifiedID uuid.UUID) ([]*models.VideoComparison, error)
}

// SQLiteComparisonRepository implements ComparisonRepository using SQLite
type SQLiteComparisonRepository struct {
	db *sqlx.DB
}

// NewSQLiteComparisonRepository creates a new SQLite comparison repository
func NewSQLiteComparisonRepository(db *sqlx.DB) ComparisonRepository {
	return &SQLiteComparisonRepository{db: db}
}

// Create creates a new comparison record
func (r *SQLiteComparisonRepository) Create(ctx context.Context, comparison *models.VideoComparison) error {
	query := `
		INSERT INTO video_comparisons (
			id, user_id, original_analysis_id, modified_analysis_id, 
			comparison_type, status, comparison_data, llm_assessment, 
			quality_score, created_at, updated_at, error_msg
		) VALUES (
			:id, :user_id, :original_analysis_id, :modified_analysis_id,
			:comparison_type, :status, :comparison_data, :llm_assessment,
			:quality_score, :created_at, :updated_at, :error_msg
		)`

	_, err := r.db.NamedExecContext(ctx, query, comparison)
	if err != nil {
		return fmt.Errorf("failed to create comparison: %w", err)
	}

	return nil
}

// GetByID retrieves a comparison by its ID
func (r *SQLiteComparisonRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.VideoComparison, error) {
	query := `
		SELECT id, user_id, original_analysis_id, modified_analysis_id,
			   comparison_type, status, comparison_data, llm_assessment,
			   quality_score, created_at, updated_at, error_msg
		FROM video_comparisons 
		WHERE id = $1`

	var comparison models.VideoComparison
	err := r.db.GetContext(ctx, &comparison, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("comparison not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get comparison: %w", err)
	}

	return &comparison, nil
}

// Update updates an existing comparison
func (r *SQLiteComparisonRepository) Update(ctx context.Context, comparison *models.VideoComparison) error {
	query := `
		UPDATE video_comparisons SET
			comparison_type = :comparison_type,
			status = :status,
			comparison_data = :comparison_data,
			llm_assessment = :llm_assessment,
			quality_score = :quality_score,
			updated_at = :updated_at,
			error_msg = :error_msg
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, comparison)
	if err != nil {
		return fmt.Errorf("failed to update comparison: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comparison not found")
	}

	return nil
}

// Delete deletes a comparison by ID
func (r *SQLiteComparisonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM video_comparisons WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete comparison: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comparison not found")
	}

	return nil
}

// ListByUser lists comparisons for a specific user
func (r *SQLiteComparisonRepository) ListByUser(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]*models.VideoComparison, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT id, user_id, original_analysis_id, modified_analysis_id,
				   comparison_type, status, comparison_data, llm_assessment,
				   quality_score, created_at, updated_at, error_msg
			FROM video_comparisons 
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		args = []interface{}{*userID, limit, offset}
	} else {
		query = `
			SELECT id, user_id, original_analysis_id, modified_analysis_id,
				   comparison_type, status, comparison_data, llm_assessment,
				   quality_score, created_at, updated_at, error_msg
			FROM video_comparisons 
			WHERE user_id IS NULL
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	var comparisons []*models.VideoComparison
	err := r.db.SelectContext(ctx, &comparisons, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list comparisons: %w", err)
	}

	return comparisons, nil
}

// ListByAnalysis lists all comparisons involving a specific analysis
func (r *SQLiteComparisonRepository) ListByAnalysis(ctx context.Context, analysisID uuid.UUID) ([]*models.VideoComparison, error) {
	query := `
		SELECT id, user_id, original_analysis_id, modified_analysis_id,
			   comparison_type, status, comparison_data, llm_assessment,
			   quality_score, created_at, updated_at, error_msg
		FROM video_comparisons 
		WHERE original_analysis_id = $1 OR modified_analysis_id = $1
		ORDER BY created_at DESC`

	var comparisons []*models.VideoComparison
	err := r.db.SelectContext(ctx, &comparisons, query, analysisID)
	if err != nil {
		return nil, fmt.Errorf("failed to list comparisons by analysis: %w", err)
	}

	return comparisons, nil
}

// GetComparisonHistory gets the comparison history between two specific analyses
func (r *SQLiteComparisonRepository) GetComparisonHistory(ctx context.Context, originalID, modifiedID uuid.UUID) ([]*models.VideoComparison, error) {
	query := `
		SELECT id, user_id, original_analysis_id, modified_analysis_id,
			   comparison_type, status, comparison_data, llm_assessment,
			   quality_score, created_at, updated_at, error_msg
		FROM video_comparisons 
		WHERE (original_analysis_id = $1 AND modified_analysis_id = $2)
		   OR (original_analysis_id = $2 AND modified_analysis_id = $1)
		ORDER BY created_at DESC`

	var comparisons []*models.VideoComparison
	err := r.db.SelectContext(ctx, &comparisons, query, originalID, modifiedID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comparison history: %w", err)
	}

	return comparisons, nil
}
