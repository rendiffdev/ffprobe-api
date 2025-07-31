package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// LLMService handles AI analysis requests
type LLMService struct {
	logger     zerolog.Logger
	ollamaURL  string
	modelName  string
	httpClient *http.Client
}

// AnalysisRequest represents a request for AI analysis
type AnalysisRequest struct {
	AnalysisData map[string]interface{} `json:"analysis_data" binding:"required"`
	Prompt       string                 `json:"prompt,omitempty"`
}

// AnalysisResponse represents the response from AI analysis
type AnalysisResponse struct {
	Success    bool          `json:"success"`
	Report     string        `json:"report,omitempty"`
	Error      string        `json:"error,omitempty"`
	ProcessingTime time.Duration `json:"processing_time"`
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func main() {
	// Setup logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	
	service := &LLMService{
		logger:    logger,
		ollamaURL: getEnv("OLLAMA_URL", "http://ollama:11434"),
		modelName: getEnv("OLLAMA_MODEL", "phi3:mini"),
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", service.healthCheck)

	// Analysis endpoint
	router.POST("/analyze", service.generateAnalysis)

	// Start server
	port := getEnv("PORT", "8082")
	logger.Info().Str("port", port).Msg("Starting LLM service")
	
	if err := router.Run(":" + port); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}
}

func (s *LLMService) healthCheck(c *gin.Context) {
	// Check Ollama connectivity
	ollamaHealthy := s.checkOllamaHealth(c.Request.Context())
	
	status := "healthy"
	if !ollamaHealthy {
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        status,
		"service":       "llm-service",
		"timestamp":     time.Now().UTC(),
		"ollama_healthy": ollamaHealthy,
		"model":         s.modelName,
	})
}

func (s *LLMService) generateAnalysis(c *gin.Context) {
	start := time.Now()
	
	var req AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, AnalysisResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
			ProcessingTime: time.Since(start),
		})
		return
	}

	// Generate professional analysis prompt
	prompt := s.buildAnalysisPrompt(req.AnalysisData, req.Prompt)

	// Call Ollama for analysis
	report, err := s.callOllama(c.Request.Context(), prompt)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate analysis")
		c.JSON(http.StatusInternalServerError, AnalysisResponse{
			Success: false,
			Error:   "Failed to generate analysis: " + err.Error(),
			ProcessingTime: time.Since(start),
		})
		return
	}

	s.logger.Info().
		Dur("processing_time", time.Since(start)).
		Int("report_length", len(report)).
		Msg("AI analysis completed successfully")

	c.JSON(http.StatusOK, AnalysisResponse{
		Success: true,
		Report:  report,
		ProcessingTime: time.Since(start),
	})
}

func (s *LLMService) buildAnalysisPrompt(analysisData map[string]interface{}, customPrompt string) string {
	if customPrompt != "" {
		return customPrompt
	}

	// Use the professional video engineer prompt
	basePrompt := `You are a senior video engineer with 15+ years of experience in broadcast, streaming, and post-production. Analyze this media file comprehensively and provide a detailed professional assessment.

## Analysis Framework:
1. **Technical Specifications**: Format, codecs, bitrates, resolution, frame rate
2. **Quality Assessment**: Video/audio quality, compression efficiency, artifacts
3. **Compliance & Standards**: Broadcasting standards, streaming compatibility
4. **Performance Metrics**: File size efficiency, bandwidth requirements
5. **Workflow Integration**: Post-production compatibility, transcoding needs
6. **Issue Identification**: Technical problems, quality concerns, compatibility issues
7. **Optimization Recommendations**: Encoding improvements, workflow suggestions
8. **Professional Summary**: Executive summary with key findings and recommendations

Provide specific, actionable insights based on the technical data. Focus on practical implications for production workflows.`

	// Convert analysis data to string
	dataJSON, _ := json.MarshalIndent(analysisData, "", "  ")
	
	return fmt.Sprintf("%s\n\n## Media Analysis Data:\n```json\n%s\n```", basePrompt, string(dataJSON))
}

func (s *LLMService) callOllama(ctx context.Context, prompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  s.modelName,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"num_predict": 1500,
			"num_ctx":     2048,
			"temperature": 0.3,
			"num_thread":  4,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.ollamaURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama returned status %d", resp.StatusCode)
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return ollamaResp.Response, nil
}

func (s *LLMService) checkOllamaHealth(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", s.ollamaURL+"/api/version", nil)
	if err != nil {
		return false
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}