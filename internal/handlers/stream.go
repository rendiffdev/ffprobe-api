package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/ffmpeg"
	"github.com/rendiffdev/ffprobe-api/internal/models"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// StreamHandler handles real-time streaming analysis
type StreamHandler struct {
	analysisService *services.AnalysisService
	upgrader        websocket.Upgrader
	logger          zerolog.Logger
}

// NewStreamHandler creates a new stream handler
func NewStreamHandler(analysisService *services.AnalysisService, logger zerolog.Logger) *StreamHandler {
	return &StreamHandler{
		analysisService: analysisService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from any origin
				// In production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		logger: logger,
	}
}

// StreamMessage represents a WebSocket message
type StreamMessage struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// StreamAnalysisRequest represents a streaming analysis request
type StreamAnalysisRequest struct {
	URL         string                 `json:"url" binding:"required"`
	Options     *ffmpeg.FFprobeOptions `json:"options,omitempty"`
	Interval    int                    `json:"interval,omitempty"` // Update interval in seconds
}

// StreamAnalysisUpdate represents a streaming analysis update
type StreamAnalysisUpdate struct {
	Timestamp   time.Time              `json:"timestamp"`
	AnalysisID  uuid.UUID              `json:"analysis_id"`
	Status      string                 `json:"status"`
	Progress    float64                `json:"progress,omitempty"`
	Result      *ffmpeg.FFprobeResult  `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// ProgressUpdate represents a progress update
type ProgressUpdate struct {
	AnalysisID uuid.UUID `json:"analysis_id"`
	Progress   float64   `json:"progress"`
	Message    string    `json:"message,omitempty"`
}

// StreamAnalysis handles WebSocket connection for real-time analysis
// @Summary Stream analysis updates
// @Description Connect via WebSocket to receive real-time analysis updates
// @Tags stream
// @Router /api/v1/stream/analysis [get]
func (h *StreamHandler) StreamAnalysis(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}
	defer conn.Close()

	clientID := uuid.New().String()
	h.logger.Info().Str("client_id", clientID).Msg("WebSocket client connected")

	// Set up ping/pong to keep connection alive
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Message handling loop
	go func() {
		for {
			select {
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	// Read messages from client
	for {
		var msg StreamMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error().Err(err).Str("client_id", clientID).Msg("WebSocket read error")
			}
			break
		}

		h.logger.Debug().
			Str("client_id", clientID).
			Str("type", msg.Type).
			Msg("Received WebSocket message")

		switch msg.Type {
		case "analyze":
			h.handleStreamAnalyze(conn, msg)
		case "subscribe":
			h.handleSubscribe(conn, msg)
		case "unsubscribe":
			h.handleUnsubscribe(conn, msg)
		case "ping":
			h.sendMessage(conn, StreamMessage{Type: "pong", ID: msg.ID})
		default:
			h.sendError(conn, "Unknown message type: "+msg.Type)
		}
	}

	h.logger.Info().Str("client_id", clientID).Msg("WebSocket client disconnected")
}

// StreamProgress streams analysis progress updates
// @Summary Stream progress updates
// @Description Get real-time progress updates for an analysis via Server-Sent Events
// @Tags stream
// @Produce text/event-stream
// @Param id path string true "Analysis ID"
// @Router /api/v1/stream/progress/{id} [get]
func (h *StreamHandler) StreamProgress(c *gin.Context) {
	analysisIDStr := c.Param("id")
	analysisID, err := uuid.Parse(analysisIDStr)
	if err != nil {
		h.logger.Error().Err(err).Str("id", analysisIDStr).Msg("Invalid analysis ID")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid analysis ID",
			Details: err.Error(),
		})
		return
	}

	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Create a channel for progress updates
	progressChan := make(chan ProgressUpdate, 10)
	defer close(progressChan)

	// Create context for cancellation
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Simulate progress updates (in production, this would come from the analysis service)
	go func() {
		for i := 0; i <= 100; i += 10 {
			select {
			case <-ctx.Done():
				return
			case progressChan <- ProgressUpdate{
				AnalysisID: analysisID,
				Progress:   float64(i) / 100,
				Message:    "Processing...",
			}:
				time.Sleep(1 * time.Second)
			}
		}
	}()

	// Stream progress updates
	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case progress, ok := <-progressChan:
			if !ok {
				return false
			}

			data, _ := json.Marshal(progress)
			c.SSEvent("progress", string(data))
			c.Writer.Flush()

			if progress.Progress >= 1.0 {
				// Send completion event
				c.SSEvent("complete", string(data))
				c.Writer.Flush()
				return false
			}
			return true
		case <-time.After(30 * time.Second):
			// Send heartbeat
			c.SSEvent("heartbeat", "{}")
			c.Writer.Flush()
			return true
		}
	})
}

// LiveStreamAnalysis analyzes a live stream
// @Summary Analyze live stream
// @Description Analyze a live streaming URL (RTMP, RTSP, HLS)
// @Tags stream
// @Accept json
// @Produce json
// @Param request body StreamAnalysisRequest true "Stream analysis request"
// @Success 200 {object} ProbeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/stream/live [post]
func (h *StreamHandler) LiveStreamAnalysis(c *gin.Context) {
	var req StreamAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid stream analysis request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate stream URL
	if !isStreamURL(req.URL) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid stream URL",
			Details: "URL must be a valid streaming protocol (rtmp://, rtsp://, http://*.m3u8)",
		})
		return
	}

	// Create analysis for stream
	analysisReq := &models.CreateAnalysisRequest{
		FileName:   req.URL,
		FilePath:   req.URL,
		SourceType: "stream",
	}

	analysis, err := h.analysisService.CreateAnalysis(c.Request.Context(), analysisReq)
	if err != nil {
		h.logger.Error().Err(err).Str("url", req.URL).Msg("Failed to create stream analysis")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create analysis",
			Details: err.Error(),
		})
		return
	}

	// Set specific options for streaming
	if req.Options == nil {
		req.Options = ffmpeg.NewOptionsBuilder().
			Input(req.URL).
			BasicInfo().
			TimeoutSeconds(30). // 30 second timeout for streams
			Build()
	} else {
		req.Options.Input = req.URL
		if req.Options.Timeout == 0 {
			req.Options.Timeout = 30 * time.Second
		}
	}

	// Process stream analysis
	if err := h.analysisService.ProcessAnalysis(c.Request.Context(), analysis.ID, req.Options); err != nil {
		h.logger.Error().Err(err).Str("analysis_id", analysis.ID.String()).Msg("Stream analysis failed")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Stream analysis failed",
			Details: err.Error(),
		})
		return
	}

	// Get updated analysis
	updatedAnalysis, err := h.analysisService.GetAnalysis(c.Request.Context(), analysis.ID)
	if err != nil {
		h.logger.Error().Err(err).Str("analysis_id", analysis.ID.String()).Msg("Failed to get stream analysis result")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get analysis result",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ProbeResponse{
		AnalysisID: analysis.ID,
		Status:     "completed",
		Analysis:   updatedAnalysis,
	})
}

// Helper methods

func (h *StreamHandler) handleStreamAnalyze(conn *websocket.Conn, msg StreamMessage) {
	var req StreamAnalysisRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.sendError(conn, "Invalid analysis request: "+err.Error())
		return
	}

	// Create analysis
	analysisReq := &models.CreateAnalysisRequest{
		FileName:   req.URL,
		FilePath:   req.URL,
		SourceType: "stream",
	}

	ctx := context.Background()
	analysis, err := h.analysisService.CreateAnalysis(ctx, analysisReq)
	if err != nil {
		h.sendError(conn, "Failed to create analysis: "+err.Error())
		return
	}

	// Send initial response
	h.sendUpdate(conn, StreamAnalysisUpdate{
		Timestamp:  time.Now(),
		AnalysisID: analysis.ID,
		Status:     "started",
	})

	// Process analysis with progress callback
	progressCallback := func(progress float64) {
		h.sendUpdate(conn, StreamAnalysisUpdate{
			Timestamp:  time.Now(),
			AnalysisID: analysis.ID,
			Status:     "processing",
			Progress:   progress,
		})
	}

	// Process in goroutine
	go func() {
		if err := h.analysisService.ProcessWithProgress(ctx, analysis.ID, req.Options, progressCallback); err != nil {
			h.sendUpdate(conn, StreamAnalysisUpdate{
				Timestamp:  time.Now(),
				AnalysisID: analysis.ID,
				Status:     "failed",
				Error:      err.Error(),
			})
			return
		}

		// Get final result
		if result, err := h.analysisService.GetAnalysis(ctx, analysis.ID); err == nil {
			h.sendUpdate(conn, StreamAnalysisUpdate{
				Timestamp:  time.Now(),
				AnalysisID: analysis.ID,
				Status:     "completed",
				Progress:   1.0,
			})
		}
	}()
}

func (h *StreamHandler) handleSubscribe(conn *websocket.Conn, msg StreamMessage) {
	// In a production system, this would subscribe to analysis updates
	h.sendMessage(conn, StreamMessage{
		Type: "subscribed",
		ID:   msg.ID,
	})
}

func (h *StreamHandler) handleUnsubscribe(conn *websocket.Conn, msg StreamMessage) {
	// In a production system, this would unsubscribe from analysis updates
	h.sendMessage(conn, StreamMessage{
		Type: "unsubscribed",
		ID:   msg.ID,
	})
}

func (h *StreamHandler) sendMessage(conn *websocket.Conn, msg StreamMessage) {
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.WriteJSON(msg); err != nil {
		h.logger.Error().Err(err).Msg("Failed to send WebSocket message")
	}
}

func (h *StreamHandler) sendError(conn *websocket.Conn, errMsg string) {
	h.sendMessage(conn, StreamMessage{
		Type:  "error",
		Error: errMsg,
	})
}

func (h *StreamHandler) sendUpdate(conn *websocket.Conn, update StreamAnalysisUpdate) {
	data, _ := json.Marshal(update)
	h.sendMessage(conn, StreamMessage{
		Type: "update",
		Data: data,
	})
}

func isStreamURL(url string) bool {
	// Check for common streaming protocols
	streamPrefixes := []string{
		"rtmp://",
		"rtmps://",
		"rtsp://",
		"rtsps://",
		"http://",
		"https://",
	}

	for _, prefix := range streamPrefixes {
		if strings.HasPrefix(strings.ToLower(url), prefix) {
			// Additional check for HLS
			if strings.HasPrefix(url, "http") && strings.HasSuffix(strings.ToLower(url), ".m3u8") {
				return true
			}
			// For RTMP/RTSP, prefix is enough
			if strings.HasPrefix(strings.ToLower(url), "rtmp") || strings.HasPrefix(strings.ToLower(url), "rtsp") {
				return true
			}
		}
	}

	return false
}