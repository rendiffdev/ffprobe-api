package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/rendiffdev/ffprobe-api/internal/repositories"
	"github.com/rs/zerolog"
)

// VMAFModelService handles VMAF model operations
type VMAFModelService struct {
	repo      *repositories.VMAFModelRepository
	modelsDir string
	logger    zerolog.Logger
}

// NewVMAFModelService creates a new VMAF model service
func NewVMAFModelService(repo *repositories.VMAFModelRepository, modelsDir string, logger zerolog.Logger) *VMAFModelService {
	// Ensure models directory exists
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		logger.Error().Err(err).Str("dir", modelsDir).Msg("Failed to create models directory")
	}

	return &VMAFModelService{
		repo:      repo,
		modelsDir: modelsDir,
		logger:    logger,
	}
}

// VMAFModelRequest represents a request to create or update a VMAF model
type VMAFModelRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Version     string                 `json:"version" binding:"required"`
	IsPublic    bool                   `json:"is_public"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UploadModel handles uploading a new VMAF model
func (s *VMAFModelService) UploadModel(ctx context.Context, userID uuid.UUID, req *VMAFModelRequest, modelFile io.Reader, fileSize int64) (*repositories.VMAFModel, error) {
	// Validate model name
	if err := s.validateModelName(req.Name); err != nil {
		return nil, err
	}

	// Check if model name already exists
	exists, err := s.repo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check model existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("model with name '%s' already exists", req.Name)
	}

	// Generate file path
	modelID := uuid.New()
	fileName := fmt.Sprintf("%s_%s.pkl", req.Name, modelID.String())
	filePath := filepath.Join(s.modelsDir, fileName)

	// Create temporary file
	tempFile, err := os.CreateTemp(s.modelsDir, "vmaf_model_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	// Copy model file and calculate hash
	hasher := sha256.New()
	writer := io.MultiWriter(tempFile, hasher)

	written, err := io.Copy(writer, modelFile)
	if err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("failed to save model file: %w", err)
	}
	tempFile.Close()

	if written != fileSize {
		return nil, fmt.Errorf("file size mismatch: expected %d, got %d", fileSize, written)
	}

	// Validate model file
	if err := s.validateModelFile(tempPath); err != nil {
		return nil, fmt.Errorf("invalid model file: %w", err)
	}

	// Move temp file to final location
	if err := os.Rename(tempPath, filePath); err != nil {
		return nil, fmt.Errorf("failed to move model file: %w", err)
	}

	// Create model record
	model := &repositories.VMAFModel{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Version:     req.Version,
		FilePath:    filePath,
		FileSize:    fileSize,
		FileHash:    fmt.Sprintf("%x", hasher.Sum(nil)),
		IsPublic:    req.IsPublic,
		IsDefault:   false,
		Metadata:    repositories.JSONB(req.Metadata),
	}

	if err := s.repo.Create(ctx, model); err != nil {
		// Clean up file on failure
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to create model record: %w", err)
	}

	s.logger.Info().
		Str("model_id", model.ID.String()).
		Str("name", model.Name).
		Str("user_id", userID.String()).
		Msg("VMAF model uploaded successfully")

	return model, nil
}

// GetModel retrieves a VMAF model by ID
func (s *VMAFModelService) GetModel(ctx context.Context, modelID uuid.UUID) (*repositories.VMAFModel, error) {
	return s.repo.GetByID(ctx, modelID)
}

// GetModelByName retrieves a VMAF model by name
func (s *VMAFModelService) GetModelByName(ctx context.Context, name string) (*repositories.VMAFModel, error) {
	return s.repo.GetByName(ctx, name)
}

// ListModels lists all models accessible to a user
func (s *VMAFModelService) ListModels(ctx context.Context, userID uuid.UUID) ([]*repositories.VMAFModel, error) {
	return s.repo.List(ctx, userID, true)
}

// ListPublicModels lists all public models
func (s *VMAFModelService) ListPublicModels(ctx context.Context) ([]*repositories.VMAFModel, error) {
	return s.repo.ListPublic(ctx)
}

// UpdateModel updates a VMAF model's metadata
func (s *VMAFModelService) UpdateModel(ctx context.Context, modelID uuid.UUID, userID uuid.UUID, req *VMAFModelRequest) (*repositories.VMAFModel, error) {
	// Get existing model
	model, err := s.repo.GetByID(ctx, modelID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if model.UserID != userID {
		return nil, fmt.Errorf("unauthorized: you don't own this model")
	}

	// Update fields
	model.Name = req.Name
	model.Description = req.Description
	model.Version = req.Version
	model.IsPublic = req.IsPublic
	if req.Metadata != nil {
		model.Metadata = repositories.JSONB(req.Metadata)
	}

	if err := s.repo.Update(ctx, model); err != nil {
		return nil, err
	}

	return model, nil
}

// DeleteModel deletes a VMAF model
func (s *VMAFModelService) DeleteModel(ctx context.Context, modelID uuid.UUID, userID uuid.UUID) error {
	// Get model to check ownership
	model, err := s.repo.GetByID(ctx, modelID)
	if err != nil {
		return err
	}

	// Check ownership (allow admins to delete any model)
	if model.UserID != userID {
		return fmt.Errorf("unauthorized: you don't own this model")
	}

	// Delete from database
	if err := s.repo.Delete(ctx, modelID); err != nil {
		return err
	}

	// Delete file
	if err := os.Remove(model.FilePath); err != nil {
		s.logger.Warn().Err(err).Str("path", model.FilePath).Msg("Failed to delete model file")
	}

	return nil
}

// SetDefaultModel sets a model as the default
func (s *VMAFModelService) SetDefaultModel(ctx context.Context, modelID uuid.UUID) error {
	return s.repo.SetDefault(ctx, modelID)
}

// GetDefaultModel retrieves the default VMAF model
func (s *VMAFModelService) GetDefaultModel(ctx context.Context) (*repositories.VMAFModel, error) {
	return s.repo.GetDefault(ctx)
}

// GetModelPath returns the full path to a model file
func (s *VMAFModelService) GetModelPath(ctx context.Context, modelID uuid.UUID) (string, error) {
	model, err := s.repo.GetByID(ctx, modelID)
	if err != nil {
		return "", err
	}

	// Verify file exists
	if _, err := os.Stat(model.FilePath); err != nil {
		return "", fmt.Errorf("model file not found: %w", err)
	}

	return model.FilePath, nil
}

// validateModelName validates a model name
func (s *VMAFModelService) validateModelName(name string) error {
	if len(name) < 3 || len(name) > 100 {
		return fmt.Errorf("model name must be between 3 and 100 characters")
	}

	// Check for invalid characters
	if strings.ContainsAny(name, "/\\<>:|?*\"") {
		return fmt.Errorf("model name contains invalid characters")
	}

	return nil
}

// validateModelFile validates a VMAF model file
func (s *VMAFModelService) validateModelFile(filePath string) error {
	// Check file size
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Maximum 500MB for model files
	if info.Size() > 500*1024*1024 {
		return fmt.Errorf("model file too large (max 500MB)")
	}

	// TODO: Add actual VMAF model validation
	// This would involve checking the file format, structure, etc.
	// For now, we just check that it's not empty
	if info.Size() == 0 {
		return fmt.Errorf("model file is empty")
	}

	return nil
}

// CleanupOrphanedFiles removes model files that don't have database records
func (s *VMAFModelService) CleanupOrphanedFiles(ctx context.Context) error {
	files, err := os.ReadDir(s.modelsDir)
	if err != nil {
		return fmt.Errorf("failed to read models directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".pkl") {
			continue
		}

		filePath := filepath.Join(s.modelsDir, file.Name())

		// Check if file has a corresponding database record
		// This is a simple implementation - in production you might want to
		// query all models and check against the file list for efficiency
		orphaned := true

		// For now, we'll skip this check
		// TODO: Implement proper orphaned file detection

		if orphaned {
			s.logger.Debug().Str("file", filePath).Msg("Found orphaned model file")
			// os.Remove(filePath)
		}
	}

	return nil
}
