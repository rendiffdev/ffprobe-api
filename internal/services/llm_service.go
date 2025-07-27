package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/config"
	"github.com/rendiffdev/ffprobe-api/internal/models"
)

// LLMService handles LLM operations for GenAI features
type LLMService struct {
	config    *config.Config
	logger    zerolog.Logger
	httpClient *http.Client
}

// NewLLMService creates a new LLM service
func NewLLMService(cfg *config.Config, logger zerolog.Logger) *LLMService {
	return &LLMService{
		config:    cfg,
		logger:    logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateAnalysis generates human-readable analysis from ffprobe data
func (s *LLMService) GenerateAnalysis(ctx context.Context, analysis *models.Analysis) (string, error) {
	// Create prompt for media analysis
	prompt := s.buildAnalysisPrompt(analysis)
	
	// Try local LLM first (if available), then fallback to OpenRouter
	response, err := s.generateWithLocalLLM(ctx, prompt)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Local LLM failed, falling back to OpenRouter")
		response, err = s.generateWithOpenRouter(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("both local and remote LLM failed: %w", err)
		}
	}
	
	return response, nil
}

// AnswerQuestion answers a question about media file using LLM
func (s *LLMService) AnswerQuestion(ctx context.Context, analysis *models.Analysis, question string) (string, error) {
	// Create prompt for Q&A
	prompt := s.buildQAPrompt(analysis, question)
	
	// Try local LLM first, then fallback to OpenRouter
	response, err := s.generateWithLocalLLM(ctx, prompt)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Local LLM failed, falling back to OpenRouter")
		response, err = s.generateWithOpenRouter(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("both local and remote LLM failed: %w", err)
		}
	}
	
	return response, nil
}

// generateWithLocalLLM attempts to use local LLM via Ollama
func (s *LLMService) generateWithLocalLLM(ctx context.Context, prompt string) (string, error) {
	// Check if local LLM is enabled
	if !s.config.EnableLocalLLM {
		return "", fmt.Errorf("local LLM disabled")
	}

	if s.config.OllamaURL == "" {
		return "", fmt.Errorf("Ollama URL not configured")
	}

	if s.config.OllamaModel == "" {
		return "", fmt.Errorf("Ollama model not configured")
	}

	// Prepare Ollama request
	requestBody := map[string]interface{}{
		"model":  s.config.OllamaModel,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature":    0.7,
			"top_p":          0.9,
			"top_k":          40,
			"repeat_penalty": 1.1,
			"num_predict":    1000,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Ollama request: %w", err)
	}

	// Create request with timeout
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second) // Increase timeout for local LLM
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.OllamaURL+"/api/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create Ollama request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request to Ollama
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send Ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}

	// Parse Ollama response
	var response struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
		Error    string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	if response.Error != "" {
		return "", fmt.Errorf("Ollama API error: %s", response.Error)
	}

	if !response.Done {
		return "", fmt.Errorf("Ollama response incomplete")
	}

	if response.Response == "" {
		return "", fmt.Errorf("empty response from Ollama")
	}

	s.logger.Info().Str("model", s.config.OllamaModel).Msg("Successfully generated response with local LLM")
	return strings.TrimSpace(response.Response), nil
}

// generateWithOpenRouter uses OpenRouter API as fallback
func (s *LLMService) generateWithOpenRouter(ctx context.Context, prompt string) (string, error) {
	if s.config.OpenRouterAPIKey == "" {
		return "", fmt.Errorf("OpenRouter API key not configured")
	}
	
	// Prepare request
	requestBody := map[string]interface{}{
		"model": "anthropic/claude-3-haiku",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens": 1000,
		"temperature": 0.7,
	}
	
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.OpenRouterAPIKey)
	req.Header.Set("HTTP-Referer", "https://ffprobe-api.local")
	req.Header.Set("X-Title", "FFprobe API")
	
	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenRouter API returned status %d", resp.StatusCode)
	}
	
	// Parse response
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	if response.Error.Message != "" {
		return "", fmt.Errorf("OpenRouter API error: %s", response.Error.Message)
	}
	
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenRouter API")
	}
	
	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}

// buildAnalysisPrompt creates a prompt for general media analysis
func (s *LLMService) buildAnalysisPrompt(analysis *models.Analysis) string {
	var prompt strings.Builder
	
	prompt.WriteString("You are an expert media analyst. Analyze the following media file information and provide a comprehensive, human-readable summary.\n\n")
	prompt.WriteString("Focus on:\n")
	prompt.WriteString("- Key characteristics and quality\n")
	prompt.WriteString("- Technical specifications\n")
	prompt.WriteString("- Compatibility and usage recommendations\n")
	prompt.WriteString("- Any potential issues or noteworthy aspects\n\n")
	
	prompt.WriteString(fmt.Sprintf("File: %s\n", analysis.FileName))
	prompt.WriteString(fmt.Sprintf("Size: %d bytes\n", analysis.FileSize))
	prompt.WriteString(fmt.Sprintf("Source: %s\n\n", analysis.SourceType))
	
	// Add ffprobe data if available
	if analysis.FFprobeData != nil {
		prompt.WriteString("Technical Data:\n")
		jsonData, _ := json.MarshalIndent(analysis.FFprobeData, "", "  ")
		prompt.Write(jsonData)
		prompt.WriteString("\n\n")
	}
	
	prompt.WriteString("Please provide a clear, informative analysis in plain English, suitable for both technical and non-technical users.")
	
	return prompt.String()
}

// buildQAPrompt creates a prompt for Q&A about media file
func (s *LLMService) buildQAPrompt(analysis *models.Analysis, question string) string {
	var prompt strings.Builder
	
	prompt.WriteString("You are an expert media analyst. Answer the following question about this media file.\n\n")
	
	prompt.WriteString(fmt.Sprintf("File: %s\n", analysis.FileName))
	prompt.WriteString(fmt.Sprintf("Size: %d bytes\n", analysis.FileSize))
	prompt.WriteString(fmt.Sprintf("Source: %s\n\n", analysis.SourceType))
	
	// Add ffprobe data if available
	if analysis.FFprobeData != nil {
		prompt.WriteString("Technical Data:\n")
		jsonData, _ := json.MarshalIndent(analysis.FFprobeData, "", "  ")
		prompt.Write(jsonData)
		prompt.WriteString("\n\n")
	}
	
	prompt.WriteString(fmt.Sprintf("Question: %s\n\n", question))
	prompt.WriteString("Please provide a helpful, accurate answer based on the technical data above.")
	
	return prompt.String()
}

// GenerateQualityInsights generates insights about video quality metrics
func (s *LLMService) GenerateQualityInsights(ctx context.Context, analysis *models.Analysis, metrics []*models.QualityMetric) (string, error) {
	prompt := s.buildQualityInsightsPrompt(analysis, metrics)
	
	// Try local LLM first, then fallback to OpenRouter
	response, err := s.generateWithLocalLLM(ctx, prompt)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Local LLM failed, falling back to OpenRouter")
		response, err = s.generateWithOpenRouter(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("both local and remote LLM failed: %w", err)
		}
	}
	
	return response, nil
}

// buildQualityInsightsPrompt creates a prompt for quality metrics analysis
func (s *LLMService) buildQualityInsightsPrompt(analysis *models.Analysis, metrics []*models.QualityMetric) string {
	var prompt strings.Builder
	
	prompt.WriteString("You are an expert video quality analyst. Analyze the following quality metrics and provide insights.\n\n")
	prompt.WriteString("Focus on:\n")
	prompt.WriteString("- Overall quality assessment\n")
	prompt.WriteString("- Comparison to industry standards\n")
	prompt.WriteString("- Recommendations for improvement\n")
	prompt.WriteString("- Suitability for different use cases\n\n")
	
	prompt.WriteString(fmt.Sprintf("File: %s\n\n", analysis.FileName))
	
	prompt.WriteString("Quality Metrics:\n")
	for _, metric := range metrics {
		prompt.WriteString(fmt.Sprintf("- %s: Overall=%.2f, Min=%.2f, Max=%.2f, Mean=%.2f\n", 
			metric.MetricType, 
			metric.OverallScore, 
			metric.MinScore, 
			metric.MaxScore, 
			metric.MeanScore))
	}
	
	prompt.WriteString("\nPlease provide practical insights and recommendations based on these quality metrics.")
	
	return prompt.String()
}

// CheckOllamaHealth checks if Ollama service is healthy and models are available
func (s *LLMService) CheckOllamaHealth(ctx context.Context) (*OllamaHealthStatus, error) {
	if s.config.OllamaURL == "" {
		return nil, fmt.Errorf("Ollama URL not configured")
	}

	status := &OllamaHealthStatus{
		URL:       s.config.OllamaURL,
		Healthy:   false,
		Models:    []string{},
		Error:     "",
		Timestamp: time.Now(),
	}

	// Check if Ollama is responding
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", s.config.OllamaURL+"/api/version", nil)
	if err != nil {
		status.Error = fmt.Sprintf("Failed to create version request: %v", err)
		return status, nil
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		status.Error = fmt.Sprintf("Failed to connect to Ollama: %v", err)
		return status, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status.Error = fmt.Sprintf("Ollama health check failed with status %d", resp.StatusCode)
		return status, nil
	}

	// Get list of available models
	req, err = http.NewRequestWithContext(ctx, "GET", s.config.OllamaURL+"/api/tags", nil)
	if err != nil {
		status.Error = fmt.Sprintf("Failed to create models request: %v", err)
		return status, nil
	}

	resp, err = s.httpClient.Do(req)
	if err != nil {
		status.Error = fmt.Sprintf("Failed to get models from Ollama: %v", err)
		return status, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var modelsResp struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err == nil {
			for _, model := range modelsResp.Models {
				status.Models = append(status.Models, model.Name)
			}
		}
	}

	// Check if configured model is available
	if s.config.OllamaModel != "" {
		status.ConfiguredModel = s.config.OllamaModel
		for _, model := range status.Models {
			if model == s.config.OllamaModel {
				status.ModelAvailable = true
				break
			}
		}
	}

	status.Healthy = len(status.Models) > 0
	return status, nil
}

// PullModel downloads a model to Ollama
func (s *LLMService) PullModel(ctx context.Context, modelName string) error {
	if s.config.OllamaURL == "" {
		return fmt.Errorf("Ollama URL not configured")
	}

	requestBody := map[string]interface{}{
		"name": modelName,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal pull request: %w", err)
	}

	// Use longer timeout for model pulling
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.OllamaURL+"/api/pull", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send pull request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("model pull failed with status %d", resp.StatusCode)
	}

	s.logger.Info().Str("model", modelName).Msg("Successfully pulled model to Ollama")
	return nil
}

// OllamaHealthStatus represents the health status of Ollama service
type OllamaHealthStatus struct {
	URL             string    `json:"url"`
	Healthy         bool      `json:"healthy"`
	Models          []string  `json:"models"`
	ConfiguredModel string    `json:"configured_model"`
	ModelAvailable  bool      `json:"model_available"`
	Error           string    `json:"error,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
}