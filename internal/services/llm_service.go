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

// generateWithLocalLLM attempts to use local LLM (placeholder for ollama integration)
func (s *LLMService) generateWithLocalLLM(ctx context.Context, prompt string) (string, error) {
	// This would integrate with ollama or go-llama.cpp
	// For now, return an error to trigger fallback
	return "", fmt.Errorf("local LLM not available")
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