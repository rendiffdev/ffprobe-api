package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/ffmpeg"
	"github.com/rendiffdev/ffprobe-api/internal/models"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// UploadHandler handles file upload operations
type UploadHandler struct {
	analysisService *services.AnalysisService
	uploadDir       string
	maxFileSize     int64
	allowedFormats  []string
	logger          zerolog.Logger
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(analysisService *services.AnalysisService, uploadDir string, logger zerolog.Logger) *UploadHandler {
	return &UploadHandler{
		analysisService: analysisService,
		uploadDir:       uploadDir,
		maxFileSize:     50 * 1024 * 1024 * 1024, // 50GB default
		allowedFormats: []string{
			".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm",
			".m4v", ".mpg", ".mpeg", ".3gp", ".ts", ".mts", ".m2ts",
			".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a",
			".opus", ".mxf", ".dv", ".f4v", ".vob", ".ogv", ".m3u8",
		},
		logger: logger,
	}
}

// SetMaxFileSize sets the maximum allowed file size
func (h *UploadHandler) SetMaxFileSize(size int64) {
	h.maxFileSize = size
}

// SetAllowedFormats sets the allowed file formats
func (h *UploadHandler) SetAllowedFormats(formats []string) {
	h.allowedFormats = formats
}

// UploadRequest represents a file upload request
type UploadRequest struct {
	Async         bool                   `form:"async"`
	AutoAnalyze   bool                   `form:"auto_analyze"`
	DeleteOnComplete bool                `form:"delete_on_complete"`
	Options       *ffmpeg.FFprobeOptions `form:"options"`
}

// UploadResponse represents the upload response
type UploadResponse struct {
	ID           uuid.UUID `json:"id"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	UploadPath   string    `json:"upload_path"`
	ContentHash  string    `json:"content_hash,omitempty"`
	AnalysisID   uuid.UUID `json:"analysis_id,omitempty"`
	Status       string    `json:"status"`
	Message      string    `json:"message,omitempty"`
}

// ChunkedUploadRequest represents a chunked upload request
type ChunkedUploadRequest struct {
	UploadID    string `form:"upload_id" binding:"required"`
	ChunkNumber int    `form:"chunk_number" binding:"required"`
	TotalChunks int    `form:"total_chunks" binding:"required"`
	FileName    string `form:"file_name"`
}

// ChunkedUploadResponse represents a chunked upload response
type ChunkedUploadResponse struct {
	UploadID     string `json:"upload_id"`
	ChunkNumber  int    `json:"chunk_number"`
	TotalChunks  int    `json:"total_chunks"`
	BytesReceived int64  `json:"bytes_received"`
	Complete     bool   `json:"complete"`
	FilePath     string `json:"file_path,omitempty"`
}

// UploadFile handles single file upload
// @Summary Upload a media file
// @Description Upload a media file for analysis
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Media file to upload"
// @Param async formData bool false "Process asynchronously"
// @Param auto_analyze formData bool false "Automatically analyze after upload"
// @Param delete_on_complete formData bool false "Delete file after analysis"
// @Success 200 {object} UploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 413 {object} ErrorResponse "File too large"
// @Failure 415 {object} ErrorResponse "Unsupported media type"
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/upload [post]
func (h *UploadHandler) UploadFile(c *gin.Context) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB buffer
		h.logger.Error().Err(err).Msg("Failed to parse multipart form")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Failed to parse upload form",
			Details: err.Error(),
		})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get file from form")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "No file provided",
			Details: err.Error(),
		})
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > h.maxFileSize {
		h.logger.Error().
			Int64("size", header.Size).
			Int64("max_size", h.maxFileSize).
			Msg("File too large")
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{
			Error:   "File too large",
			Details: fmt.Sprintf("File size %d exceeds maximum allowed size %d", header.Size, h.maxFileSize),
		})
		return
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !h.isAllowedFormat(ext) {
		h.logger.Error().Str("extension", ext).Msg("Unsupported file format")
		c.JSON(http.StatusUnsupportedMediaType, ErrorResponse{
			Error:   "Unsupported file format",
			Details: fmt.Sprintf("File extension %s is not supported", ext),
		})
		return
	}

	// Parse request options
	var req UploadRequest
	req.Async = c.PostForm("async") == "true"
	req.AutoAnalyze = c.PostForm("auto_analyze") == "true"
	req.DeleteOnComplete = c.PostForm("delete_on_complete") == "true"

	// Generate unique filename
	uploadID := uuid.New()
	filename := fmt.Sprintf("%s_%s", uploadID.String(), header.Filename)
	uploadPath := filepath.Join(h.uploadDir, filename)

	// Ensure upload directory exists
	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		h.logger.Error().Err(err).Str("dir", h.uploadDir).Msg("Failed to create upload directory")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create upload directory",
			Details: err.Error(),
		})
		return
	}

	// Create destination file
	dst, err := os.Create(uploadPath)
	if err != nil {
		h.logger.Error().Err(err).Str("path", uploadPath).Msg("Failed to create destination file")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create destination file",
			Details: err.Error(),
		})
		return
	}
	defer dst.Close()

	// Copy file content
	written, err := io.Copy(dst, file)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to save uploaded file")
		os.Remove(uploadPath) // Clean up on error
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to save uploaded file",
			Details: err.Error(),
		})
		return
	}

	h.logger.Info().
		Str("upload_id", uploadID.String()).
		Str("filename", header.Filename).
		Int64("size", written).
		Msg("File uploaded successfully")

	response := UploadResponse{
		ID:         uploadID,
		FileName:   header.Filename,
		FileSize:   written,
		UploadPath: uploadPath,
		Status:     "uploaded",
	}

	// Auto-analyze if requested
	if req.AutoAnalyze {
		analysisReq := &models.CreateAnalysisRequest{
			FileName:   header.Filename,
			FilePath:   uploadPath,
			FileSize:   written,
			SourceType: "upload",
		}

		analysis, err := h.analysisService.CreateAnalysis(c.Request.Context(), analysisReq)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to create analysis for uploaded file")
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "File uploaded but analysis failed",
				Details: err.Error(),
			})
			return
		}

		response.AnalysisID = analysis.ID

		if req.Async {
			// Start async processing
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
				defer cancel()

				if err := h.analysisService.ProcessAnalysis(ctx, analysis.ID, req.Options); err != nil {
					h.logger.Error().Err(err).Str("analysis_id", analysis.ID.String()).Msg("Async analysis failed")
				}

				// Delete file if requested
				if req.DeleteOnComplete {
					if err := os.Remove(uploadPath); err != nil {
						h.logger.Error().Err(err).Str("path", uploadPath).Msg("Failed to delete file after analysis")
					}
				}
			}()

			response.Status = "processing"
			response.Message = "File uploaded and analysis started"
		} else {
			// Synchronous processing
			if err := h.analysisService.ProcessAnalysis(c.Request.Context(), analysis.ID, req.Options); err != nil {
				h.logger.Error().Err(err).Msg("Analysis failed")
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Analysis failed",
					Details: err.Error(),
				})
				return
			}

			// Delete file if requested
			if req.DeleteOnComplete {
				if err := os.Remove(uploadPath); err != nil {
					h.logger.Error().Err(err).Str("path", uploadPath).Msg("Failed to delete file after analysis")
				}
			}

			response.Status = "completed"
			response.Message = "File uploaded and analyzed successfully"
		}
	}

	c.JSON(http.StatusOK, response)
}

// UploadChunk handles chunked file upload
// @Summary Upload a file chunk
// @Description Upload a chunk of a large file
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param chunk formData file true "File chunk"
// @Param upload_id formData string true "Upload session ID"
// @Param chunk_number formData int true "Chunk number"
// @Param total_chunks formData int true "Total number of chunks"
// @Param file_name formData string false "Original filename (required for first chunk)"
// @Success 200 {object} ChunkedUploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/upload/chunk [post]
func (h *UploadHandler) UploadChunk(c *gin.Context) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse multipart form")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Failed to parse upload form",
			Details: err.Error(),
		})
		return
	}

	// Parse request
	var req ChunkedUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid chunk upload request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	// Get chunk file
	chunk, _, err := c.Request.FormFile("chunk")
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get chunk from form")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "No chunk provided",
			Details: err.Error(),
		})
		return
	}
	defer chunk.Close()

	// Create chunk directory
	chunkDir := filepath.Join(h.uploadDir, "chunks", req.UploadID)
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		h.logger.Error().Err(err).Str("dir", chunkDir).Msg("Failed to create chunk directory")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create chunk directory",
			Details: err.Error(),
		})
		return
	}

	// Save chunk
	chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d", req.ChunkNumber))
	dst, err := os.Create(chunkPath)
	if err != nil {
		h.logger.Error().Err(err).Str("path", chunkPath).Msg("Failed to create chunk file")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create chunk file",
			Details: err.Error(),
		})
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, chunk)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to save chunk")
		os.Remove(chunkPath)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to save chunk",
			Details: err.Error(),
		})
		return
	}

	h.logger.Info().
		Str("upload_id", req.UploadID).
		Int("chunk", req.ChunkNumber).
		Int("total", req.TotalChunks).
		Int64("size", written).
		Msg("Chunk uploaded")

	response := ChunkedUploadResponse{
		UploadID:     req.UploadID,
		ChunkNumber:  req.ChunkNumber,
		TotalChunks:  req.TotalChunks,
		BytesReceived: written,
		Complete:     false,
	}

	// Check if all chunks are uploaded
	if req.ChunkNumber == req.TotalChunks {
		// Save filename for final assembly
		if req.FileName != "" {
			metaPath := filepath.Join(chunkDir, "metadata.txt")
			os.WriteFile(metaPath, []byte(req.FileName), 0644)
		}

		// Check if all chunks exist
		allChunksExist := true
		for i := 1; i <= req.TotalChunks; i++ {
			chunkFile := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d", i))
			if _, err := os.Stat(chunkFile); err != nil {
				allChunksExist = false
				break
			}
		}

		if allChunksExist {
			// Assemble file
			finalPath, err := h.assembleChunks(req.UploadID, req.TotalChunks)
			if err != nil {
				h.logger.Error().Err(err).Msg("Failed to assemble chunks")
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to assemble file",
					Details: err.Error(),
				})
				return
			}

			response.Complete = true
			response.FilePath = finalPath

			// Clean up chunks
			os.RemoveAll(chunkDir)
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetUploadStatus gets the status of an upload
// @Summary Get upload status
// @Description Get the current status of a chunked upload
// @Tags upload
// @Produce json
// @Param id path string true "Upload ID"
// @Success 200 {object} ChunkedUploadResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/upload/status/{id} [get]
func (h *UploadHandler) GetUploadStatus(c *gin.Context) {
	uploadID := c.Param("id")
	chunkDir := filepath.Join(h.uploadDir, "chunks", uploadID)

	// Check if upload exists
	if _, err := os.Stat(chunkDir); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Upload not found",
		})
		return
	}

	// Count chunks
	chunks, err := os.ReadDir(chunkDir)
	if err != nil {
		h.logger.Error().Err(err).Str("dir", chunkDir).Msg("Failed to read chunk directory")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to read upload status",
			Details: err.Error(),
		})
		return
	}

	chunkCount := 0
	totalSize := int64(0)
	for _, chunk := range chunks {
		if strings.HasPrefix(chunk.Name(), "chunk_") {
			chunkCount++
			if info, err := chunk.Info(); err == nil {
				totalSize += info.Size()
			}
		}
	}

	c.JSON(http.StatusOK, ChunkedUploadResponse{
		UploadID:     uploadID,
		ChunkNumber:  chunkCount,
		BytesReceived: totalSize,
		Complete:     false,
	})
}

// Helper methods

func (h *UploadHandler) isAllowedFormat(ext string) bool {
	for _, allowed := range h.allowedFormats {
		if ext == allowed {
			return true
		}
	}
	return false
}

func (h *UploadHandler) assembleChunks(uploadID string, totalChunks int) (string, error) {
	chunkDir := filepath.Join(h.uploadDir, "chunks", uploadID)
	
	// Read metadata
	metaPath := filepath.Join(chunkDir, "metadata.txt")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return "", fmt.Errorf("failed to read metadata: %w", err)
	}

	fileName := string(metaData)
	if fileName == "" {
		fileName = fmt.Sprintf("%s_assembled.bin", uploadID)
	}

	finalPath := filepath.Join(h.uploadDir, fmt.Sprintf("%s_%s", uploadID, fileName))
	
	// Create final file
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return "", fmt.Errorf("failed to create final file: %w", err)
	}
	defer finalFile.Close()

	// Assemble chunks in order
	for i := 1; i <= totalChunks; i++ {
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d", i))
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			return "", fmt.Errorf("failed to read chunk %d: %w", i, err)
		}

		if _, err := finalFile.Write(chunkData); err != nil {
			return "", fmt.Errorf("failed to write chunk %d: %w", i, err)
		}
	}

	return finalPath, nil
}