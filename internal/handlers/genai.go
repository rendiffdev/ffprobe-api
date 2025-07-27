package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// GenAIHandler handles GenAI/LLM-powered endpoints
type GenAIHandler struct {
	analysisService *services.AnalysisService
	llmService      *services.LLMService
	logger          zerolog.Logger
}

// NewGenAIHandler creates a new GenAI handler
func NewGenAIHandler(analysisService *services.AnalysisService, llmService *services.LLMService, logger zerolog.Logger) *GenAIHandler {
	return &GenAIHandler{
		analysisService: analysisService,
		llmService:      llmService,
		logger:          logger,
	}
}

// AskQuestionRequest represents a request to ask a question about media
type AskQuestionRequest struct {
	// Either provide analysis_id for existing analysis or source for new analysis
	AnalysisID string `json:"analysis_id,omitempty"`
	Source     string `json:"source,omitempty"` // URL, file path, or upload
	Question   string `json:"question" binding:"required"`
	// Optional parameters for new analysis
	FFprobeArgs []string `json:"ffprobe_args,omitempty"`
}

// AskQuestionResponse represents the response from asking a question
type AskQuestionResponse struct {
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	AnalysisID string `json:"analysis_id"`
	Source     string `json:"source"`
}

// GenerateAnalysisRequest represents a request to generate AI analysis
type GenerateAnalysisRequest struct {
	AnalysisID string `json:"analysis_id" binding:"required"`
}

// GenerateAnalysisResponse represents AI-generated analysis
type GenerateAnalysisResponse struct {
	AnalysisID  string `json:"analysis_id"`
	Analysis    string `json:"analysis"`
	GeneratedAt string `json:"generated_at"`
}

// AskQuestion handles the /ask endpoint for interactive Q&A about media files
// @Summary Ask question about media
// @Description Ask a question about a media file using AI/LLM
// @Tags genai
// @Accept json
// @Produce json
// @Param request body AskQuestionRequest true "Question request"
// @Success 200 {object} AskQuestionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ask [post]
func (h *GenAIHandler) AskQuestion(c *gin.Context) {
	var req AskQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate that either analysis_id or source is provided
	if req.AnalysisID == "" && req.Source == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Either analysis_id or source must be provided",
		})
		return
	}

	var analysisID string
	var analysis *services.AnalysisResult

	// Get or create analysis
	if req.AnalysisID != "" {
		// Use existing analysis
		analysisID = req.AnalysisID
		
		// Validate UUID format
		if _, err := uuid.Parse(analysisID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid analysis ID format",
			})
			return
		}

		// Get existing analysis
		result, err := h.analysisService.GetAnalysisResult(c.Request.Context(), analysisID)
		if err != nil {
			if err == services.ErrAnalysisNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Analysis not found",
				})
				return
			}
			
			h.logger.Error().Err(err).Msg("Failed to get analysis")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get analysis",
			})
			return
		}
		analysis = result
	} else {
		// Create new analysis
		userID, _ := c.Get("user_id")
		userIDStr, _ := userID.(string)

		// Generate new analysis ID
		analysisID = uuid.New().String()

		// Start analysis in the background
		go func() {
			ctx := c.Request.Context()
			opts := services.AnalysisOptions{
				Source:      req.Source,
				UserID:      userIDStr,
				FFprobeArgs: req.FFprobeArgs,
			}

			if err := h.analysisService.AnalyzeMedia(ctx, analysisID, opts); err != nil {
				h.logger.Error().Err(err).Str("analysis_id", analysisID).Msg("Analysis failed")
			}
		}()

		// For new analysis, we need to wait a bit or return a different response
		// For now, we'll return an error suggesting to use existing analysis
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Analysis started. Please wait for completion and use the analysis_id to ask questions.",
			"analysis_id": analysisID,
			"suggestion": "Use GET /api/v1/probe/status/" + analysisID + " to check status, then ask your question with the analysis_id.",
		})
		return
	}

	// Generate answer using LLM
	answer, err := h.llmService.AnswerQuestion(c.Request.Context(), analysis.Analysis, req.Question)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate answer")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate answer",
			"details": "LLM service unavailable. Please try again later.",
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, AskQuestionResponse{
		Question:   req.Question,
		Answer:     answer,
		AnalysisID: analysisID,
		Source:     analysis.Analysis.FilePath,
	})
}

// GenerateAnalysis generates AI-powered analysis for an existing analysis
// @Summary Generate AI analysis
// @Description Generate human-readable AI analysis for media file
// @Tags genai
// @Accept json
// @Produce json
// @Param request body GenerateAnalysisRequest true "Analysis generation request"
// @Success 200 {object} GenerateAnalysisResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/genai/analysis [post]
func (h *GenAIHandler) GenerateAnalysis(c *gin.Context) {
	var req GenerateAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(req.AnalysisID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid analysis ID format",
		})
		return
	}

	// Get analysis
	result, err := h.analysisService.GetAnalysisResult(c.Request.Context(), req.AnalysisID)
	if err != nil {
		if err == services.ErrAnalysisNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Analysis not found",
			})
			return
		}
		
		h.logger.Error().Err(err).Msg("Failed to get analysis")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get analysis",
		})
		return
	}

	// Generate AI analysis
	aiAnalysis, err := h.llmService.GenerateAnalysis(c.Request.Context(), result.Analysis)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate AI analysis")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate AI analysis",
			"details": "LLM service unavailable. Please try again later.",
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, GenerateAnalysisResponse{
		AnalysisID:  req.AnalysisID,
		Analysis:    aiAnalysis,
		GeneratedAt: "now",
	})
}

// GenerateQualityInsights generates AI insights about quality metrics
// @Summary Generate quality insights
// @Description Generate AI insights about video quality metrics
// @Tags genai
// @Accept json
// @Produce json
// @Param analysis_id path string true "Analysis ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/genai/quality-insights/{analysis_id} [get]
func (h *GenAIHandler) GenerateQualityInsights(c *gin.Context) {
	analysisID := c.Param("analysis_id")
	
	// Validate UUID format
	analysisUUID, err := uuid.Parse(analysisID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid analysis ID format",
		})
		return
	}

	// Get analysis
	result, err := h.analysisService.GetAnalysisResult(c.Request.Context(), analysisID)
	if err != nil {
		if err == services.ErrAnalysisNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Analysis not found",
			})
			return
		}
		
		h.logger.Error().Err(err).Msg("Failed to get analysis")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get analysis",
		})
		return
	}

	// Get quality metrics
	metrics, err := h.analysisService.GetQualityMetrics(c.Request.Context(), analysisUUID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get quality metrics")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get quality metrics",
		})
		return
	}

	if len(metrics) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No quality metrics found for this analysis",
		})
		return
	}

	// Generate quality insights
	insights, err := h.llmService.GenerateQualityInsights(c.Request.Context(), result.Analysis, metrics)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate quality insights")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate quality insights",
			"details": "LLM service unavailable. Please try again later.",
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"analysis_id": analysisID,
		"insights":    insights,
		"metrics_count": len(metrics),
		"generated_at": "now",
	})
}

// GetLLMHealth checks the health of LLM services
// @Summary Check LLM health
// @Description Check the health status of local and remote LLM services
// @Tags genai
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/genai/health [get]
func (h *GenAIHandler) GetLLMHealth(c *gin.Context) {
	ctx := c.Request.Context()
	health := map[string]interface{}{
		"timestamp": time.Now().UTC(),
		"services":  map[string]interface{}{},
	}

	// Check Ollama health
	ollamaHealth, err := h.llmService.CheckOllamaHealth(ctx)
	if err != nil {
		health["services"].(map[string]interface{})["ollama"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	} else {
		health["services"].(map[string]interface{})["ollama"] = ollamaHealth
	}

	// Overall status
	allHealthy := true
	if ollamaHealth == nil || !ollamaHealth.Healthy {
		allHealthy = false
	}

	health["overall_status"] = "healthy"
	if !allHealthy {
		health["overall_status"] = "degraded"
	}

	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, health)
}

// PullModel downloads a model to Ollama
// @Summary Pull LLM model
// @Description Download a model to the local Ollama service
// @Tags genai
// @Accept json
// @Produce json
// @Param request body map[string]string true "Model pull request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/genai/pull-model [post]
func (h *GenAIHandler) PullModel(c *gin.Context) {
	var req struct {
		Model string `json:"model" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate model name
	if len(req.Model) == 0 || len(req.Model) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Model name must be between 1 and 100 characters",
		})
		return
	}

	ctx := c.Request.Context()

	// Start model pull
	go func() {
		if err := h.llmService.PullModel(ctx, req.Model); err != nil {
			h.logger.Error().Err(err).Str("model", req.Model).Msg("Failed to pull model")
		} else {
			h.logger.Info().Str("model", req.Model).Msg("Successfully pulled model")
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Model pull started",
		"model":   req.Model,
		"status":  "pulling",
	})
}