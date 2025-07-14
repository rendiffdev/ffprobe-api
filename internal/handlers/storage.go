package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ffprobe-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type StorageHandler struct {
	storageService *services.StorageService
	logger         zerolog.Logger
}

func NewStorageHandler(storageService *services.StorageService, logger zerolog.Logger) *StorageHandler {
	return &StorageHandler{
		storageService: storageService,
		logger:         logger.With().Str("handler", "storage").Logger(),
	}
}

type UploadResponse struct {
	UploadID string `json:"upload_id"`
	Key      string `json:"key"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
}

type DownloadResponse struct {
	Key  string `json:"key"`
	URL  string `json:"url"`
	Size int64  `json:"size"`
}

type SignedURLRequest struct {
	Key        string `json:"key" binding:"required"`
	Expiration int64  `json:"expiration,omitempty"`
}

type SignedURLResponse struct {
	Key        string `json:"key"`
	URL        string `json:"url"`
	Expiration int64  `json:"expiration"`
}

func (h *StorageHandler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get file from form")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file from form"})
		return
	}
	defer file.Close()

	uploadID := uuid.New().String()
	ext := filepath.Ext(header.Filename)
	key := fmt.Sprintf("uploads/%s%s", uploadID, ext)

	if err := h.storageService.UploadFile(c.Request.Context(), key, file, header.Size); err != nil {
		h.logger.Error().Err(err).Str("key", key).Msg("Failed to upload file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	url, err := h.storageService.GetFileURL(c.Request.Context(), key)
	if err != nil {
		h.logger.Error().Err(err).Str("key", key).Msg("Failed to get file URL")
		url = ""
	}

	response := UploadResponse{
		UploadID: uploadID,
		Key:      key,
		URL:      url,
		Size:     header.Size,
	}

	c.JSON(http.StatusOK, response)
}

func (h *StorageHandler) DownloadFile(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required"})
		return
	}

	exists, err := h.storageService.FileExists(c.Request.Context(), key)
	if err != nil {
		h.logger.Error().Err(err).Str("key", key).Msg("Failed to check if file exists")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if file exists"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	reader, err := h.storageService.DownloadFile(c.Request.Context(), key)
	if err != nil {
		h.logger.Error().Err(err).Str("key", key).Msg("Failed to download file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file"})
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(key)))
	c.Header("Content-Type", "application/octet-stream")

	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

func (h *StorageHandler) DeleteFile(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required"})
		return
	}

	exists, err := h.storageService.FileExists(c.Request.Context(), key)
	if err != nil {
		h.logger.Error().Err(err).Str("key", key).Msg("Failed to check if file exists")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if file exists"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if err := h.storageService.DeleteFile(c.Request.Context(), key); err != nil {
		h.logger.Error().Err(err).Str("key", key).Msg("Failed to delete file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

func (h *StorageHandler) GetFileInfo(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required"})
		return
	}

	exists, err := h.storageService.FileExists(c.Request.Context(), key)
	if err != nil {
		h.logger.Error().Err(err).Str("key", key).Msg("Failed to check if file exists")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if file exists"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	url, err := h.storageService.GetFileURL(c.Request.Context(), key)
	if err != nil {
		h.logger.Error().Err(err).Str("key", key).Msg("Failed to get file URL")
		url = ""
	}

	response := DownloadResponse{
		Key: key,
		URL: url,
	}

	c.JSON(http.StatusOK, response)
}

func (h *StorageHandler) GetSignedURL(c *gin.Context) {
	var req SignedURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expiration := time.Hour
	if req.Expiration > 0 {
		expiration = time.Duration(req.Expiration) * time.Second
	}

	exists, err := h.storageService.FileExists(c.Request.Context(), req.Key)
	if err != nil {
		h.logger.Error().Err(err).Str("key", req.Key).Msg("Failed to check if file exists")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if file exists"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	url, err := h.storageService.GetSignedURL(c.Request.Context(), req.Key, expiration)
	if err != nil {
		h.logger.Error().Err(err).Str("key", req.Key).Msg("Failed to generate signed URL")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate signed URL"})
		return
	}

	response := SignedURLResponse{
		Key:        req.Key,
		URL:        url,
		Expiration: time.Now().Add(expiration).Unix(),
	}

	c.JSON(http.StatusOK, response)
}