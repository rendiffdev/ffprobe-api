package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rendiffdev/ffprobe-api/internal/services"
	"github.com/rs/zerolog"
)

// FFmpegVersionHandler handles FFmpeg version management endpoints
type FFmpegVersionHandler struct {
	updater *services.FFmpegUpdater
	logger  zerolog.Logger
}

// NewFFmpegVersionHandler creates a new FFmpeg version handler
func NewFFmpegVersionHandler(updater *services.FFmpegUpdater, logger zerolog.Logger) *FFmpegVersionHandler {
	return &FFmpegVersionHandler{
		updater: updater,
		logger:  logger,
	}
}

// GetCurrentVersion returns the current FFmpeg version
func (h *FFmpegVersionHandler) GetCurrentVersion(c *gin.Context) {
	version, err := h.updater.GetCurrentVersion()
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get current FFmpeg version")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"version": version,
		"status":  "current",
	})
}

// CheckForUpdates checks for available FFmpeg updates
func (h *FFmpegVersionHandler) CheckForUpdates(c *gin.Context) {
	ctx := c.Request.Context()

	updateInfo, err := h.updater.CheckForUpdates(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to check for updates")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for updates"})
		return
	}

	// Add user action required flag for major updates
	c.JSON(http.StatusOK, gin.H{
		"current":                updateInfo.Current,
		"available":              updateInfo.Available,
		"update_available":       updateInfo.Available != nil && h.updater.CompareVersions(updateInfo.Available, updateInfo.Current) > 0,
		"is_major_upgrade":       updateInfo.IsMajor,
		"is_minor_upgrade":       updateInfo.IsMinor,
		"is_patch_upgrade":       updateInfo.IsPatch,
		"stability":              updateInfo.Stability,
		"recommendation":         updateInfo.Recommendation,
		"user_approval_required": updateInfo.IsMajor, // Require approval for major updates
	})
}

// UpdateFFmpeg handles FFmpeg update requests
func (h *FFmpegVersionHandler) UpdateFFmpeg(c *gin.Context) {
	var req struct {
		Confirm bool   `json:"confirm"`
		Version string `json:"version,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if user confirmed the update
	if !req.Confirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Update must be confirmed"})
		return
	}

	ctx := c.Request.Context()

	// Get update information
	updateInfo, err := h.updater.CheckForUpdates(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to check for updates")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for updates"})
		return
	}

	// Check if update is available
	if h.updater.CompareVersions(updateInfo.Available, updateInfo.Current) <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":  "no_update",
			"message": "Already on the latest version",
		})
		return
	}

	// For major updates, require explicit confirmation
	if updateInfo.IsMajor {
		h.logger.Info().
			Interface("current", updateInfo.Current).
			Interface("new", updateInfo.Available).
			Msg("Major FFmpeg update requested")

		// Send warning about major update
		c.JSON(http.StatusOK, gin.H{
			"status":          "major_update_warning",
			"message":         "This is a major version upgrade that may contain breaking changes",
			"current":         updateInfo.Current,
			"new":             updateInfo.Available,
			"action_required": "Please review the changelog and test in a non-production environment first",
			"changelog_url":   "https://github.com/FFmpeg/FFmpeg/blob/master/Changelog",
		})
		return
	}

	// Perform the update
	progressChan := make(chan int, 100)
	errorChan := make(chan error, 1)

	go func() {
		err := h.updater.DownloadUpdate(ctx, updateInfo.Available, func(percent int) {
			progressChan <- percent
		})
		errorChan <- err
		close(progressChan)
	}()

	// Stream progress to client using Server-Sent Events
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	c.Stream(func(w io.Writer) bool {
		select {
		case progress, ok := <-progressChan:
			if !ok {
				// Channel closed, check for error
				if err := <-errorChan; err != nil {
					h.logger.Error().Err(err).Msg("FFmpeg update failed")
					c.SSEvent("error", gin.H{"data": err.Error()})
				} else {
					c.SSEvent("complete", gin.H{"data": "FFmpeg updated successfully"})
				}
				return false
			}
			c.SSEvent("progress", gin.H{"percent": progress})
			return true

		case <-c.Request.Context().Done():
			return false
		}
	})
}

// RollbackFFmpeg rolls back to the previous FFmpeg version
func (h *FFmpegVersionHandler) RollbackFFmpeg(c *gin.Context) {
	// This would implement rollback functionality
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"status":  "not_implemented",
		"message": "Rollback functionality will be available in a future release",
	})
}

// RegisterRoutes registers the FFmpeg version management routes with Gin
func (h *FFmpegVersionHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Admin routes for FFmpeg version management
	admin := router.Group("/admin/ffmpeg")
	{
		admin.GET("/version", h.GetCurrentVersion)
		admin.GET("/check-updates", h.CheckForUpdates)
		admin.POST("/update", h.UpdateFFmpeg)
		admin.POST("/rollback", h.RollbackFFmpeg)
	}
}
