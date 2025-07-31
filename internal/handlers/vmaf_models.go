package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// VMAFModelsHandler handles VMAF model management endpoints
type VMAFModelsHandler struct {
	modelService *services.VMAFModelService
	logger       zerolog.Logger
}

// NewVMAFModelsHandler creates a new VMAF models handler
func NewVMAFModelsHandler(modelService *services.VMAFModelService, logger zerolog.Logger) *VMAFModelsHandler {
	return &VMAFModelsHandler{
		modelService: modelService,
		logger:       logger,
	}
}

// UploadModel handles VMAF model upload
// @Summary Upload custom VMAF model
// @Description Upload a custom VMAF model file
// @Tags quality
// @Accept multipart/form-data
// @Produce json
// @Param model formData file true "VMAF model file"
// @Param name formData string true "Model name"
// @Param description formData string false "Model description"
// @Param version formData string true "Model version"
// @Param is_public formData bool false "Make model public"
// @Success 201 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/quality/models [post]
func (h *VMAFModelsHandler) UploadModel(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "User ID not found in context",
		})
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Details: err.Error(),
		})
		return
	}

	// Parse multipart form
	file, header, err := c.Request.FormFile("model")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Model file required",
			Details: err.Error(),
		})
		return
	}
	defer file.Close()

	// Get form values
	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Model name required",
		})
		return
	}

	version := c.PostForm("version")
	if version == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Model version required",
		})
		return
	}

	description := c.PostForm("description")
	isPublic := c.PostForm("is_public") == "true"

	// Create model request
	req := &services.VMAFModelRequest{
		Name:        name,
		Description: description,
		Version:     version,
		IsPublic:    isPublic,
	}

	// Upload model
	model, err := h.modelService.UploadModel(c.Request.Context(), userID, req, file, header.Size)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to upload VMAF model")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to upload model",
			Details: err.Error(),
		})
		return
	}

	h.logger.Info().
		Str("model_id", model.ID.String()).
		Str("name", model.Name).
		Str("user_id", userID.String()).
		Msg("VMAF model uploaded successfully")

	c.JSON(http.StatusCreated, gin.H{
		"message": "Model uploaded successfully",
		"model":   model,
	})
}

// ListModels lists available VMAF models
// @Summary List VMAF models
// @Description List all VMAF models accessible to the user
// @Tags quality
// @Accept json
// @Produce json
// @Param public_only query bool false "Show only public models"
// @Success 200 {object} gin.H
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/quality/models [get]
func (h *VMAFModelsHandler) ListModels(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "User ID not found in context",
		})
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Details: err.Error(),
		})
		return
	}

	publicOnly := c.Query("public_only") == "true"

	var models interface{}
	if publicOnly {
		models, err = h.modelService.ListPublicModels(c.Request.Context())
	} else {
		models, err = h.modelService.ListModels(c.Request.Context(), userID)
	}

	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list VMAF models")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to list models",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
	})
}

// GetModel retrieves a specific VMAF model
// @Summary Get VMAF model
// @Description Get details of a specific VMAF model
// @Tags quality
// @Accept json
// @Produce json
// @Param id path string true "Model ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/quality/models/{id} [get]
func (h *VMAFModelsHandler) GetModel(c *gin.Context) {
	modelIDStr := c.Param("id")
	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid model ID",
			Details: err.Error(),
		})
		return
	}

	model, err := h.modelService.GetModel(c.Request.Context(), modelID)
	if err != nil {
		h.logger.Error().Err(err).Str("model_id", modelID.String()).Msg("Failed to get VMAF model")
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Model not found",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"model": model,
	})
}

// UpdateModel updates a VMAF model's metadata
// @Summary Update VMAF model
// @Description Update metadata of a VMAF model
// @Tags quality
// @Accept json
// @Produce json
// @Param id path string true "Model ID"
// @Param request body services.VMAFModelRequest true "Update request"
// @Success 200 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/quality/models/{id} [put]
func (h *VMAFModelsHandler) UpdateModel(c *gin.Context) {
	modelIDStr := c.Param("id")
	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid model ID",
			Details: err.Error(),
		})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "User ID not found in context",
		})
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Details: err.Error(),
		})
		return
	}

	var req services.VMAFModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	model, err := h.modelService.UpdateModel(c.Request.Context(), modelID, userID, &req)
	if err != nil {
		h.logger.Error().Err(err).Str("model_id", modelID.String()).Msg("Failed to update VMAF model")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update model",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Model updated successfully",
		"model":   model,
	})
}

// DeleteModel deletes a VMAF model
// @Summary Delete VMAF model
// @Description Delete a custom VMAF model
// @Tags quality
// @Accept json
// @Produce json
// @Param id path string true "Model ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/quality/models/{id} [delete]
func (h *VMAFModelsHandler) DeleteModel(c *gin.Context) {
	modelIDStr := c.Param("id")
	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid model ID",
			Details: err.Error(),
		})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "User ID not found in context",
		})
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Details: err.Error(),
		})
		return
	}

	if err := h.modelService.DeleteModel(c.Request.Context(), modelID, userID); err != nil {
		h.logger.Error().Err(err).Str("model_id", modelID.String()).Msg("Failed to delete VMAF model")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete model",
			Details: err.Error(),
		})
		return
	}

	h.logger.Info().
		Str("model_id", modelID.String()).
		Str("user_id", userID.String()).
		Msg("VMAF model deleted successfully")

	c.Status(http.StatusNoContent)
}

// SetDefaultModel sets a model as the default
// @Summary Set default VMAF model
// @Description Set a VMAF model as the default for all analyses
// @Tags quality
// @Accept json
// @Produce json
// @Param id path string true "Model ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/quality/models/{id}/default [put]
func (h *VMAFModelsHandler) SetDefaultModel(c *gin.Context) {
	// This endpoint should be admin-only
	roles, exists := c.Get("roles")
	if !exists {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: "Insufficient permissions",
		})
		return
	}

	hasAdmin := false
	if roleList, ok := roles.([]string); ok {
		for _, role := range roleList {
			if role == "admin" {
				hasAdmin = true
				break
			}
		}
	}

	if !hasAdmin {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: "Admin role required",
		})
		return
	}

	modelIDStr := c.Param("id")
	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid model ID",
			Details: err.Error(),
		})
		return
	}

	if err := h.modelService.SetDefaultModel(c.Request.Context(), modelID); err != nil {
		h.logger.Error().Err(err).Str("model_id", modelID.String()).Msg("Failed to set default VMAF model")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to set default model",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Default model set successfully",
		"model_id": modelID,
	})
}

// DownloadModel downloads a VMAF model file
// @Summary Download VMAF model
// @Description Download a VMAF model file
// @Tags quality
// @Accept json
// @Produce octet-stream
// @Param id path string true "Model ID"
// @Success 200 {file} binary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/quality/models/{id}/download [get]
func (h *VMAFModelsHandler) DownloadModel(c *gin.Context) {
	modelIDStr := c.Param("id")
	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid model ID",
			Details: err.Error(),
		})
		return
	}

	model, err := h.modelService.GetModel(c.Request.Context(), modelID)
	if err != nil {
		h.logger.Error().Err(err).Str("model_id", modelID.String()).Msg("Failed to get VMAF model")
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Model not found",
			Details: err.Error(),
		})
		return
	}

	// Check if model is public or user owns it
	userIDStr, exists := c.Get("user_id")
	if exists {
		if userID, err := uuid.Parse(userIDStr.(string)); err == nil {
			if !model.IsPublic && model.UserID != userID {
				c.JSON(http.StatusForbidden, ErrorResponse{
					Error: "Access denied to private model",
				})
				return
			}
		}
	} else if !model.IsPublic {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: "Authentication required for private models",
		})
		return
	}

	// Get model file path
	modelPath, err := h.modelService.GetModelPath(c.Request.Context(), modelID)
	if err != nil {
		h.logger.Error().Err(err).Str("model_id", modelID.String()).Msg("Failed to get model path")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get model file",
			Details: err.Error(),
		})
		return
	}

	// Set headers for download
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pkl", model.Name))
	c.Header("Content-Description", "VMAF Model File")
	
	c.File(modelPath)
}