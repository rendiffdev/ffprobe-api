package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rendiffdev/rendiff-probe/internal/models"
)

// Stub implementation to resolve compilation issues
// The actual quality service implementation would go here

var (
	// ErrNotImplemented is returned for unimplemented methods
	ErrNotImplemented = errors.New("feature not yet implemented")
)

// GetQualityMetrics retrieves quality metrics for an analysis (stub)
func (s *AnalysisService) GetQualityMetricsStub(ctx context.Context, analysisID uuid.UUID) ([]models.QualityMetrics, error) {
	return nil, ErrNotImplemented
}

// GenerateQualityInsights generates insights about video quality metrics (stub)
func (s *LLMService) GenerateQualityInsightsStub(ctx context.Context, analysis *models.Analysis, metrics []models.QualityMetrics) (string, error) {
	return "", ErrNotImplemented
}

