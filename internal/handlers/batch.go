package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rendiffdev/rendiff-probe/internal/errors"
	"github.com/rendiffdev/rendiff-probe/internal/ffmpeg"
	"github.com/rendiffdev/rendiff-probe/internal/models"
	"github.com/rendiffdev/rendiff-probe/internal/services"
	"github.com/rendiffdev/rendiff-probe/internal/validator"
	"github.com/rs/zerolog"
)

// BatchHandler handles batch processing operations
type BatchHandler struct {
	analysisService *services.AnalysisService
	maxBatchSize    int
	logger          zerolog.Logger
}

// NewBatchHandler creates a new batch handler
func NewBatchHandler(analysisService *services.AnalysisService, logger zerolog.Logger) *BatchHandler {
	return &BatchHandler{
		analysisService: analysisService,
		maxBatchSize:    100, // Default max batch size
		logger:          logger,
	}
}

// BatchAnalysisRequest represents a batch analysis request
type BatchAnalysisRequest struct {
	Files   []BatchFileItem        `json:"files" binding:"required,min=1"`
	Options *ffmpeg.FFprobeOptions `json:"options,omitempty"`
	Async   bool                   `json:"async,omitempty"`
}

// BatchFileItem represents a single file in a batch
type BatchFileItem struct {
	ID         string                 `json:"id" binding:"required"`
	FilePath   string                 `json:"file_path" binding:"required"`
	SourceType string                 `json:"source_type,omitempty"`
	Options    *ffmpeg.FFprobeOptions `json:"options,omitempty"`
}

// BatchAnalysisResponse represents the batch analysis response
type BatchAnalysisResponse struct {
	BatchID   uuid.UUID         `json:"batch_id"`
	Status    string            `json:"status"`
	Total     int               `json:"total"`
	Completed int               `json:"completed"`
	Failed    int               `json:"failed"`
	Results   []BatchResultItem `json:"results,omitempty"`
}

// BatchResultItem represents a single result in a batch
type BatchResultItem struct {
	ID         string           `json:"id"`
	AnalysisID uuid.UUID        `json:"analysis_id,omitempty"`
	Status     string           `json:"status"`
	Error      string           `json:"error,omitempty"`
	Analysis   *models.Analysis `json:"analysis,omitempty"`
}

// BatchStatusResponse represents batch status information
type BatchStatusResponse struct {
	BatchID     uuid.UUID         `json:"batch_id"`
	Status      string            `json:"status"`
	Total       int               `json:"total"`
	Completed   int               `json:"completed"`
	Failed      int               `json:"failed"`
	InProgress  int               `json:"in_progress"`
	StartedAt   time.Time         `json:"started_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Results     []BatchResultItem `json:"results"`
}

// Batch tracking (in-memory for now, should be in database for production)
var (
	batchStore = make(map[uuid.UUID]*BatchStatusResponse)
	batchMutex sync.RWMutex
)

// CreateBatch creates a new batch analysis
// @Summary Create batch analysis
// @Description Analyze multiple files in a single batch operation
// @Tags batch
// @Accept json
// @Produce json
// @Param request body BatchAnalysisRequest true "Batch analysis request"
// @Success 200 {object} BatchAnalysisResponse
// @Success 202 {object} BatchAnalysisResponse "Accepted for async processing"
// @Failure 400 {object} ErrorResponse
// @Failure 413 {object} ErrorResponse "Batch too large"
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/batch/analyze [post]
func (h *BatchHandler) CreateBatch(c *gin.Context) {
	var req BatchAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid batch request")
		errors.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Check batch size
	if len(req.Files) > h.maxBatchSize {
		h.logger.Error().
			Int("size", len(req.Files)).
			Int("max", h.maxBatchSize).
			Msg("Batch too large")
		errors.RespondWithError(c, http.StatusRequestEntityTooLarge, errors.CodeBadRequest, "Batch too large", fmt.Sprintf("Maximum batch size is %d", h.maxBatchSize))
		return
	}

	// Validate each file in the batch
	fileValidator := validator.NewFilePathValidator()
	for i, file := range req.Files {
		// Validate file path or URL depending on source type
		if file.SourceType == "url" || file.SourceType == "" && (len(file.FilePath) > 4 && file.FilePath[:4] == "http") {
			if err := validator.ValidateURL(file.FilePath); err != nil {
				h.logger.Warn().Err(err).Str("url", file.FilePath).Int("index", i).Msg("Invalid URL in batch")
				errors.ValidationError(c, fmt.Sprintf("Invalid URL in file %d", i+1), err.Error())
				return
			}
		} else {
			if err := fileValidator.ValidateFilePath(file.FilePath); err != nil {
				h.logger.Warn().Err(err).Str("path", file.FilePath).Int("index", i).Msg("Invalid file path in batch")
				errors.ValidationError(c, fmt.Sprintf("Invalid file path in file %d", i+1), err.Error())
				return
			}
		}
	}

	// Create batch ID
	batchID := uuid.New()

	// Initialize batch status
	batchStatus := &BatchStatusResponse{
		BatchID:    batchID,
		Status:     "pending",
		Total:      len(req.Files),
		Completed:  0,
		Failed:     0,
		InProgress: 0,
		StartedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Results:    make([]BatchResultItem, 0, len(req.Files)),
	}

	// Store batch status
	batchMutex.Lock()
	batchStore[batchID] = batchStatus
	batchMutex.Unlock()

	h.logger.Info().
		Str("batch_id", batchID.String()).
		Int("total_files", len(req.Files)).
		Bool("async", req.Async).
		Msg("Batch analysis created")

	if req.Async {
		// Process asynchronously
		go h.processBatchAsync(batchID, req)

		c.JSON(http.StatusAccepted, BatchAnalysisResponse{
			BatchID:   batchID,
			Status:    "processing",
			Total:     len(req.Files),
			Completed: 0,
			Failed:    0,
		})
	} else {
		// Process synchronously
		results := h.processBatchSync(c.Request.Context(), batchID, req)

		c.JSON(http.StatusOK, BatchAnalysisResponse{
			BatchID:   batchID,
			Status:    "completed",
			Total:     len(req.Files),
			Completed: countByStatus(results, "completed"),
			Failed:    countByStatus(results, "failed"),
			Results:   results,
		})
	}
}

// GetBatchStatus gets the status of a batch
// @Summary Get batch status
// @Description Get the current status and results of a batch analysis
// @Tags batch
// @Produce json
// @Param id path string true "Batch ID"
// @Success 200 {object} BatchStatusResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/batch/status/{id} [get]
func (h *BatchHandler) GetBatchStatus(c *gin.Context) {
	batchIDStr := c.Param("id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		h.logger.Error().Err(err).Str("id", batchIDStr).Msg("Invalid batch ID")
		errors.ValidationError(c, "Invalid batch ID", err.Error())
		return
	}

	batchMutex.RLock()
	status, exists := batchStore[batchID]
	batchMutex.RUnlock()

	if !exists {
		errors.NotFound(c, "Batch not found", "")
		return
	}

	c.JSON(http.StatusOK, status)
}

// CancelBatch cancels a batch analysis
// @Summary Cancel batch analysis
// @Description Cancel an in-progress batch analysis
// @Tags batch
// @Param id path string true "Batch ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/batch/{id}/cancel [post]
func (h *BatchHandler) CancelBatch(c *gin.Context) {
	batchIDStr := c.Param("id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		h.logger.Error().Err(err).Str("id", batchIDStr).Msg("Invalid batch ID")
		errors.ValidationError(c, "Invalid batch ID", err.Error())
		return
	}

	batchMutex.Lock()
	status, exists := batchStore[batchID]
	if exists && status.Status == "processing" {
		status.Status = "cancelled"
		status.UpdatedAt = time.Now()
		now := time.Now()
		status.CompletedAt = &now
	}
	batchMutex.Unlock()

	if !exists {
		errors.NotFound(c, "Batch not found", "")
		return
	}

	h.logger.Info().Str("batch_id", batchID.String()).Msg("Batch cancelled")
	c.Status(http.StatusNoContent)
}

// ListBatches lists recent batch operations
// @Summary List batch operations
// @Description List recent batch analysis operations
// @Tags batch
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param status query string false "Filter by status"
// @Success 200 {object} ListBatchesResponse
// @Router /api/v1/batch [get]
func (h *BatchHandler) ListBatches(c *gin.Context) {
	statusFilter := c.Query("status")

	batchMutex.RLock()
	defer batchMutex.RUnlock()

	batches := make([]BatchStatusResponse, 0)
	for _, batch := range batchStore {
		if statusFilter == "" || batch.Status == statusFilter {
			batches = append(batches, *batch)
		}
	}

	c.JSON(http.StatusOK, ListBatchesResponse{
		Batches: batches,
		Count:   len(batches),
	})
}

// Helper methods

func (h *BatchHandler) processBatchSync(ctx context.Context, batchID uuid.UUID, req BatchAnalysisRequest) []BatchResultItem {
	results := make([]BatchResultItem, len(req.Files))

	for i, file := range req.Files {
		result := BatchResultItem{
			ID:     file.ID,
			Status: "processing",
		}

		// Create analysis
		analysisReq := &models.CreateAnalysisRequest{
			FileName:   file.FilePath,
			FilePath:   file.FilePath,
			SourceType: file.SourceType,
		}
		if analysisReq.SourceType == "" {
			analysisReq.SourceType = "batch"
		}

		analysis, err := h.analysisService.CreateAnalysis(ctx, analysisReq)
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			results[i] = result
			continue
		}

		result.AnalysisID = analysis.ID

		// Process analysis
		options := file.Options
		if options == nil {
			options = req.Options
		}

		if err := h.analysisService.ProcessAnalysis(ctx, analysis.ID, options); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
		} else {
			result.Status = "completed"
			// Get updated analysis
			if updatedAnalysis, err := h.analysisService.GetAnalysis(ctx, analysis.ID); err == nil {
				result.Analysis = updatedAnalysis
			}
		}

		results[i] = result
	}

	// Update batch status
	h.updateBatchStatus(batchID, results)

	return results
}

func (h *BatchHandler) processBatchAsync(batchID uuid.UUID, req BatchAnalysisRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	// Update status to processing
	batchMutex.Lock()
	if status, exists := batchStore[batchID]; exists {
		status.Status = "processing"
		status.UpdatedAt = time.Now()
	}
	batchMutex.Unlock()

	// Process files concurrently with limited concurrency
	sem := make(chan struct{}, 5) // Limit to 5 concurrent analyses
	var wg sync.WaitGroup
	results := make([]BatchResultItem, len(req.Files))

	for i, file := range req.Files {
		wg.Add(1)
		go func(index int, fileItem BatchFileItem) {
			defer wg.Done()

			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			result := BatchResultItem{
				ID:     fileItem.ID,
				Status: "processing",
			}

			// Check if batch was cancelled
			batchMutex.RLock()
			cancelled := false
			if status, exists := batchStore[batchID]; exists && status.Status == "cancelled" {
				cancelled = true
			}
			batchMutex.RUnlock()

			if cancelled {
				result.Status = "cancelled"
				results[index] = result
				return
			}

			// Create analysis
			analysisReq := &models.CreateAnalysisRequest{
				FileName:   fileItem.FilePath,
				FilePath:   fileItem.FilePath,
				SourceType: fileItem.SourceType,
			}
			if analysisReq.SourceType == "" {
				analysisReq.SourceType = "batch"
			}

			analysis, err := h.analysisService.CreateAnalysis(ctx, analysisReq)
			if err != nil {
				result.Status = "failed"
				result.Error = err.Error()
				results[index] = result
				h.updateBatchProgress(batchID, results)
				return
			}

			result.AnalysisID = analysis.ID

			// Process analysis
			options := fileItem.Options
			if options == nil {
				options = req.Options
			}

			if err := h.analysisService.ProcessAnalysis(ctx, analysis.ID, options); err != nil {
				result.Status = "failed"
				result.Error = err.Error()
			} else {
				result.Status = "completed"
			}

			results[index] = result
			h.updateBatchProgress(batchID, results)
		}(i, file)
	}

	wg.Wait()

	// Final update
	h.updateBatchStatus(batchID, results)
}

func (h *BatchHandler) updateBatchProgress(batchID uuid.UUID, results []BatchResultItem) {
	batchMutex.Lock()
	defer batchMutex.Unlock()

	if status, exists := batchStore[batchID]; exists {
		status.Results = results
		status.Completed = countByStatus(results, "completed")
		status.Failed = countByStatus(results, "failed")
		status.InProgress = countByStatus(results, "processing")
		status.UpdatedAt = time.Now()
	}
}

func (h *BatchHandler) updateBatchStatus(batchID uuid.UUID, results []BatchResultItem) {
	batchMutex.Lock()
	defer batchMutex.Unlock()

	if status, exists := batchStore[batchID]; exists {
		status.Results = results
		status.Completed = countByStatus(results, "completed")
		status.Failed = countByStatus(results, "failed")
		status.InProgress = 0
		status.Status = "completed"
		status.UpdatedAt = time.Now()
		now := time.Now()
		status.CompletedAt = &now
	}
}

func countByStatus(results []BatchResultItem, status string) int {
	count := 0
	for _, result := range results {
		if result.Status == status {
			count++
		}
	}
	return count
}

// Response types
type ListBatchesResponse struct {
	Batches []BatchStatusResponse `json:"batches"`
	Count   int                   `json:"count"`
}
