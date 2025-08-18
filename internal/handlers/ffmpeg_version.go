package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/services"
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
func (h *FFmpegVersionHandler) GetCurrentVersion(w http.ResponseWriter, r *http.Request) {
	version, err := h.updater.GetCurrentVersion()
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get current FFmpeg version")
		http.Error(w, "Failed to get current version", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version": version,
		"status":  "current",
	})
}

// CheckForUpdates checks for available FFmpeg updates
func (h *FFmpegVersionHandler) CheckForUpdates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	updateInfo, err := h.updater.CheckForUpdates(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to check for updates")
		http.Error(w, "Failed to check for updates", http.StatusInternalServerError)
		return
	}

	// Add user action required flag for major updates
	response := map[string]interface{}{
		"current":             updateInfo.Current,
		"available":           updateInfo.Available,
		"update_available":    updateInfo.Available != nil && h.updater.CompareVersions(updateInfo.Available, updateInfo.Current) > 0,
		"is_major_upgrade":    updateInfo.IsMajor,
		"is_minor_upgrade":    updateInfo.IsMinor,
		"is_patch_upgrade":    updateInfo.IsPatch,
		"stability":           updateInfo.Stability,
		"recommendation":      updateInfo.Recommendation,
		"user_approval_required": updateInfo.IsMajor, // Require approval for major updates
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateFFmpeg handles FFmpeg update requests
func (h *FFmpegVersionHandler) UpdateFFmpeg(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Confirm bool   `json:"confirm"`
		Version string `json:"version,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if user confirmed the update
	if !req.Confirm {
		http.Error(w, "Update must be confirmed", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get update information
	updateInfo, err := h.updater.CheckForUpdates(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to check for updates")
		http.Error(w, "Failed to check for updates", http.StatusInternalServerError)
		return
	}

	// Check if update is available
	if h.updater.CompareVersions(updateInfo.Available, updateInfo.Current) <= 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "major_update_warning",
			"message": "This is a major version upgrade that may contain breaking changes",
			"current": updateInfo.Current,
			"new":     updateInfo.Available,
			"action_required": "Please review the changelog and test in a non-production environment first",
			"changelog_url": "https://github.com/FFmpeg/FFmpeg/blob/master/Changelog",
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
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case progress, ok := <-progressChan:
			if !ok {
				// Channel closed, check for error
				if err := <-errorChan; err != nil {
					h.logger.Error().Err(err).Msg("FFmpeg update failed")
					json.NewEncoder(w).Encode(map[string]string{
						"event": "error",
						"data":  err.Error(),
					})
				} else {
					json.NewEncoder(w).Encode(map[string]string{
						"event": "complete",
						"data":  "FFmpeg updated successfully",
					})
				}
				return
			}
			
			// Send progress update
			json.NewEncoder(w).Encode(map[string]interface{}{
				"event":   "progress",
				"percent": progress,
			})
			flusher.Flush()
			
		case <-r.Context().Done():
			// Client disconnected
			return
		}
	}
}

// RollbackFFmpeg rolls back to the previous FFmpeg version
func (h *FFmpegVersionHandler) RollbackFFmpeg(w http.ResponseWriter, r *http.Request) {
	// This would implement rollback functionality
	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "not_implemented",
		"message": "Rollback functionality will be available in a future release",
	})
}

// RegisterRoutes registers the FFmpeg version management routes
func (h *FFmpegVersionHandler) RegisterRoutes(router *mux.Router) {
	// Admin routes for FFmpeg version management
	adminRouter := router.PathPrefix("/api/v1/admin/ffmpeg").Subrouter()
	
	adminRouter.HandleFunc("/version", h.GetCurrentVersion).Methods("GET")
	adminRouter.HandleFunc("/check-updates", h.CheckForUpdates).Methods("GET")
	adminRouter.HandleFunc("/update", h.UpdateFFmpeg).Methods("POST")
	adminRouter.HandleFunc("/rollback", h.RollbackFFmpeg).Methods("POST")
}