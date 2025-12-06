package repositories

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// JSONB represents a PostgreSQL JSONB column type
type JSONB map[string]interface{}

// Value implements the driver Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}

	return json.Unmarshal(bytes, j)
}

// VMAFModel represents a custom VMAF model in the database
type VMAFModel struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	UserID      uuid.UUID  `db:"user_id" json:"user_id"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	Version     string     `db:"version" json:"version"`
	FilePath    string     `db:"file_path" json:"file_path"`
	FileSize    int64      `db:"file_size" json:"file_size"`
	FileHash    string     `db:"file_hash" json:"file_hash"`
	IsPublic    bool       `db:"is_public" json:"is_public"`
	IsDefault   bool       `db:"is_default" json:"is_default"`
	Metadata    JSONB      `db:"metadata" json:"metadata"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// VMAFModelRepository handles database operations for VMAF models
type VMAFModelRepository struct {
	db *sqlx.DB
}

// NewVMAFModelRepository creates a new VMAF model repository
func NewVMAFModelRepository(db *sqlx.DB) *VMAFModelRepository {
	return &VMAFModelRepository{db: db}
}

// Create creates a new VMAF model
func (r *VMAFModelRepository) Create(ctx context.Context, model *VMAFModel) error {
	query := `
		INSERT INTO vmaf_models (
			id, user_id, name, description, version, 
			file_path, file_size, file_hash, is_public, 
			is_default, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, 
			$6, $7, $8, $9, 
			$10, $11, $12, $13
		)`

	now := time.Now()
	model.ID = uuid.New()
	model.CreatedAt = now
	model.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		model.ID, model.UserID, model.Name, model.Description, model.Version,
		model.FilePath, model.FileSize, model.FileHash, model.IsPublic,
		model.IsDefault, model.Metadata, model.CreatedAt, model.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create VMAF model: %w", err)
	}

	return nil
}

// GetByID retrieves a VMAF model by ID
func (r *VMAFModelRepository) GetByID(ctx context.Context, id uuid.UUID) (*VMAFModel, error) {
	var model VMAFModel
	query := `
		SELECT id, user_id, name, description, version, 
			   file_path, file_size, file_hash, is_public, 
			   is_default, metadata, created_at, updated_at, deleted_at
		FROM vmaf_models 
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("VMAF model not found")
		}
		return nil, fmt.Errorf("failed to get VMAF model: %w", err)
	}

	return &model, nil
}

// GetByName retrieves a VMAF model by name
func (r *VMAFModelRepository) GetByName(ctx context.Context, name string) (*VMAFModel, error) {
	var model VMAFModel
	query := `
		SELECT id, user_id, name, description, version, 
			   file_path, file_size, file_hash, is_public, 
			   is_default, metadata, created_at, updated_at, deleted_at
		FROM vmaf_models 
		WHERE name = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1`

	err := r.db.GetContext(ctx, &model, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("VMAF model not found")
		}
		return nil, fmt.Errorf("failed to get VMAF model: %w", err)
	}

	return &model, nil
}

// List retrieves all VMAF models accessible to a user
func (r *VMAFModelRepository) List(ctx context.Context, userID uuid.UUID, includePublic bool) ([]*VMAFModel, error) {
	var models []*VMAFModel
	query := `
		SELECT id, user_id, name, description, version, 
			   file_path, file_size, file_hash, is_public, 
			   is_default, metadata, created_at, updated_at
		FROM vmaf_models 
		WHERE deleted_at IS NULL AND (user_id = $1`

	args := []interface{}{userID}

	if includePublic {
		query += ` OR is_public = true`
	}

	query += `) ORDER BY is_default DESC, created_at DESC`

	err := r.db.SelectContext(ctx, &models, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list VMAF models: %w", err)
	}

	return models, nil
}

// ListPublic retrieves all public VMAF models
func (r *VMAFModelRepository) ListPublic(ctx context.Context) ([]*VMAFModel, error) {
	var models []*VMAFModel
	query := `
		SELECT id, user_id, name, description, version, 
			   file_path, file_size, file_hash, is_public, 
			   is_default, metadata, created_at, updated_at
		FROM vmaf_models 
		WHERE deleted_at IS NULL AND is_public = true
		ORDER BY is_default DESC, created_at DESC`

	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list public VMAF models: %w", err)
	}

	return models, nil
}

// Update updates a VMAF model
func (r *VMAFModelRepository) Update(ctx context.Context, model *VMAFModel) error {
	query := `
		UPDATE vmaf_models 
		SET name = $2, description = $3, version = $4, 
			is_public = $5, is_default = $6, metadata = $7, 
			updated_at = $8
		WHERE id = $1 AND deleted_at IS NULL`

	model.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		model.ID, model.Name, model.Description, model.Version,
		model.IsPublic, model.IsDefault, model.Metadata, model.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update VMAF model: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("VMAF model not found")
	}

	return nil
}

// Delete soft deletes a VMAF model
func (r *VMAFModelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE vmaf_models 
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, id, now)
	if err != nil {
		return fmt.Errorf("failed to delete VMAF model: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("VMAF model not found")
	}

	return nil
}

// SetDefault sets a model as the default VMAF model
func (r *VMAFModelRepository) SetDefault(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Remove default from all models
	_, err = tx.ExecContext(ctx, `UPDATE vmaf_models SET is_default = false WHERE is_default = true`)
	if err != nil {
		return fmt.Errorf("failed to unset default models: %w", err)
	}

	// Set the new default
	result, err := tx.ExecContext(ctx,
		`UPDATE vmaf_models SET is_default = true, updated_at = $2 WHERE id = $1 AND deleted_at IS NULL`,
		id, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to set default model: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("VMAF model not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDefault retrieves the default VMAF model
func (r *VMAFModelRepository) GetDefault(ctx context.Context) (*VMAFModel, error) {
	var model VMAFModel
	query := `
		SELECT id, user_id, name, description, version, 
			   file_path, file_size, file_hash, is_public, 
			   is_default, metadata, created_at, updated_at
		FROM vmaf_models 
		WHERE is_default = true AND deleted_at IS NULL
		LIMIT 1`

	err := r.db.GetContext(ctx, &model, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no default VMAF model found")
		}
		return nil, fmt.Errorf("failed to get default VMAF model: %w", err)
	}

	return &model, nil
}

// ExistsByName checks if a model with the given name exists
func (r *VMAFModelRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM vmaf_models WHERE name = $1 AND deleted_at IS NULL)`

	err := r.db.GetContext(ctx, &exists, query, name)
	if err != nil {
		return false, fmt.Errorf("failed to check VMAF model existence: %w", err)
	}

	return exists, nil
}

// GetByUser retrieves all models for a specific user
func (r *VMAFModelRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]*VMAFModel, error) {
	var models []*VMAFModel
	query := `
		SELECT id, user_id, name, description, version, 
			   file_path, file_size, file_hash, is_public, 
			   is_default, metadata, created_at, updated_at
		FROM vmaf_models 
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &models, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user's VMAF models: %w", err)
	}

	return models, nil
}
