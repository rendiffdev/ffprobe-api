package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rendiffdev/ffprobe-api/internal/circuitbreaker"
	"github.com/rendiffdev/ffprobe-api/internal/config"
	"github.com/rendiffdev/ffprobe-api/internal/models"
	"github.com/rs/zerolog"
)

// LLMService handles LLM operations for GenAI features
type LLMService struct {
	config                   *config.Config
	logger                   zerolog.Logger
	httpClient               *http.Client
	ollamaCircuitBreaker     *circuitbreaker.CircuitBreaker
	openrouterCircuitBreaker *circuitbreaker.CircuitBreaker
}

// NewLLMService creates a new LLM service with production-ready timeouts and circuit breakers
func NewLLMService(cfg *config.Config, logger zerolog.Logger) *LLMService {
	// CRITICAL FIX: Increased timeout for production LLM workloads
	// LLM calls typically take 30-120 seconds for complex analysis
	timeout := 120 * time.Second
	if cfg.LLMTimeout > 0 {
		timeout = time.Duration(cfg.LLMTimeout) * time.Second
	}

	logger.Info().
		Dur("timeout", timeout).
		Msg("Initializing LLM service with production timeout and circuit breakers")

	// Create circuit breaker for Ollama service
	ollamaTimeout := time.Duration(cfg.CircuitBreakerTimeout) * time.Second
	ollamaInterval := time.Duration(cfg.CircuitBreakerInterval) * time.Second

	ollamaCircuitBreaker := circuitbreaker.NewCircuitBreaker(circuitbreaker.Settings{
		Name:        "ollama-llm",
		MaxRequests: 3,              // Allow 3 requests in half-open state
		Interval:    ollamaInterval, // Configurable interval
		Timeout:     ollamaTimeout,  // Configurable timeout
		ReadyToTrip: func(counts circuitbreaker.Counts) bool {
			// Trip after 3 consecutive failures or 50% failure rate with at least 5 requests
			return counts.ConsecutiveFailures >= 3 ||
				(counts.Requests >= 5 && float64(counts.TotalFailures)/float64(counts.Requests) >= 0.5)
		},
		OnStateChange: func(name string, from circuitbreaker.State, to circuitbreaker.State) {
			logger.Warn().
				Str("service", name).
				Str("from_state", from.String()).
				Str("to_state", to.String()).
				Msg("Circuit breaker state changed")
		},
	})

	// Create circuit breaker for OpenRouter service
	openrouterTimeout := time.Duration(cfg.CircuitBreakerTimeout*2) * time.Second   // Double timeout for external API
	openrouterInterval := time.Duration(cfg.CircuitBreakerInterval*2) * time.Second // Double interval for external API

	openrouterCircuitBreaker := circuitbreaker.NewCircuitBreaker(circuitbreaker.Settings{
		Name:        "openrouter-llm",
		MaxRequests: 2,                  // More conservative for external API
		Interval:    openrouterInterval, // Configurable longer interval for external service
		Timeout:     openrouterTimeout,  // Configurable longer timeout before retry
		ReadyToTrip: func(counts circuitbreaker.Counts) bool {
			// Trip after 2 consecutive failures for external API
			return counts.ConsecutiveFailures >= 2 ||
				(counts.Requests >= 3 && float64(counts.TotalFailures)/float64(counts.Requests) >= 0.6)
		},
		OnStateChange: func(name string, from circuitbreaker.State, to circuitbreaker.State) {
			logger.Warn().
				Str("service", name).
				Str("from_state", from.String()).
				Str("to_state", to.String()).
				Msg("Circuit breaker state changed")
		},
	})

	return &LLMService{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		ollamaCircuitBreaker:     ollamaCircuitBreaker,
		openrouterCircuitBreaker: openrouterCircuitBreaker,
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

// generateWithLocalLLM attempts to use local LLM via Ollama with circuit breaker protection
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

	// Use circuit breaker to protect against cascading failures
	result, err := s.ollamaCircuitBreaker.Execute(func() (interface{}, error) {
		// Try primary model first (Gemma 3 270M - optimized for speed)
		response, err := s.generateWithOllamaModel(ctx, s.config.OllamaModel, prompt, map[string]interface{}{
			"temperature":    0.7,
			"top_p":          0.9,
			"top_k":          40,
			"repeat_penalty": 1.1,
			"num_predict":    1500, // Good for structured reports
			"num_ctx":        8192, // Gemma3 supports 8K context
			"num_batch":      16,   // Larger batch for faster processing
			"num_thread":     4,    // CPU threads to use
		})

		if err == nil {
			s.logger.Info().
				Str("model", s.config.OllamaModel).
				Str("circuit_breaker_state", s.ollamaCircuitBreaker.State().String()).
				Msg("Successfully generated with primary model")
			return response, nil
		}

		// If primary model fails, try fallback model (Phi-3 Mini - better reasoning)
		s.logger.Warn().
			Err(err).
			Str("model", s.config.OllamaModel).
			Msg("Primary model failed, trying fallback")

		if s.config.OllamaFallbackModel != "" {
			response, err = s.generateWithOllamaModel(ctx, s.config.OllamaFallbackModel, prompt, map[string]interface{}{
				"temperature":    0.7,
				"top_p":          0.9,
				"top_k":          40,
				"repeat_penalty": 1.1,
				"num_predict":    2000, // More tokens for complex analysis
				"num_ctx":        4096, // Phi3 mini context window
				"num_batch":      8,    // Standard batch size
				"num_thread":     4,    // CPU threads to use
			})

			if err == nil {
				s.logger.Info().
					Str("model", s.config.OllamaFallbackModel).
					Str("circuit_breaker_state", s.ollamaCircuitBreaker.State().String()).
					Msg("Successfully generated with fallback model")
				return response, nil
			}
		}

		return "", fmt.Errorf("both primary and fallback models failed: %w", err)
	})

	if err != nil {
		s.logger.Error().
			Err(err).
			Str("circuit_breaker_state", s.ollamaCircuitBreaker.State().String()).
			Interface("circuit_breaker_counts", s.ollamaCircuitBreaker.Counts()).
			Msg("Ollama LLM request failed through circuit breaker")
		return "", err
	}

	return result.(string), nil
}

// generateWithOllamaModel generates response using specific Ollama model
func (s *LLMService) generateWithOllamaModel(ctx context.Context, model string, prompt string, options map[string]interface{}) (string, error) {
	// Prepare Ollama request
	requestBody := map[string]interface{}{
		"model":   model,
		"prompt":  prompt,
		"stream":  false,
		"options": options,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Ollama request: %w", err)
	}

	// Create request with timeout (shorter for Gemma3, longer for Phi3)
	timeout := 60 * time.Second
	if model == s.config.OllamaFallbackModel {
		timeout = 120 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
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

	return strings.TrimSpace(response.Response), nil
}

// generateWithOpenRouter uses OpenRouter API as fallback with circuit breaker protection
func (s *LLMService) generateWithOpenRouter(ctx context.Context, prompt string) (string, error) {
	if s.config.OpenRouterAPIKey == "" {
		return "", fmt.Errorf("OpenRouter API key not configured")
	}

	// Use circuit breaker to protect against external API failures
	result, err := s.openrouterCircuitBreaker.Execute(func() (interface{}, error) {
		// Prepare request
		requestBody := map[string]interface{}{
			"model": "anthropic/claude-3-haiku",
			"messages": []map[string]string{
				{
					"role":    "user",
					"content": prompt,
				},
			},
			"max_tokens":  1000,
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
	})

	if err != nil {
		s.logger.Error().
			Err(err).
			Str("circuit_breaker_state", s.openrouterCircuitBreaker.State().String()).
			Interface("circuit_breaker_counts", s.openrouterCircuitBreaker.Counts()).
			Msg("OpenRouter LLM request failed through circuit breaker")
		return "", err
	}

	s.logger.Info().
		Str("circuit_breaker_state", s.openrouterCircuitBreaker.State().String()).
		Msg("Successfully generated with OpenRouter")

	return result.(string), nil
}

// buildAnalysisPrompt creates a prompt for general media analysis
func (s *LLMService) buildAnalysisPrompt(analysis *models.Analysis) string {
	var prompt strings.Builder

	prompt.WriteString("You are a senior video engineer and media processing expert working in a studio-quality post-production environment.\n\n")
	prompt.WriteString("Analyze the following FFprobe JSON output and provide a highly detailed and professional summary report.\n\n")
	prompt.WriteString("Break your response into the following structured sections:\n\n")
	prompt.WriteString("---\n\n")
	prompt.WriteString("🎬 1. **Basic Media Overview**\n")
	prompt.WriteString("- File name or container format (e.g., MP4, MKV, MOV)\n")
	prompt.WriteString("- Duration (in HH:MM:SS)\n")
	prompt.WriteString("- Overall file size (if present)\n")
	prompt.WriteString("- Bitrate (overall, and per stream if applicable)\n")
	prompt.WriteString("- Number and types of streams (video, audio, subtitle, data)\n\n")
	prompt.WriteString("---\n\n")
	prompt.WriteString("📺 2. **Video Stream(s) Details**\n")
	prompt.WriteString("- Codec name and profile (e.g., H.264 High@L4.1)\n")
	prompt.WriteString("- Resolution (width × height)\n")
	prompt.WriteString("- Aspect ratio (including SAR/DAR if available)\n")
	prompt.WriteString("- Frame rate (FPS, mention if CFR or VFR)\n")
	prompt.WriteString("- Bitrate of video stream\n")
	prompt.WriteString("- Scan type (progressive/interlaced)\n")
	prompt.WriteString("- GOP structure (if detectable)\n")
	prompt.WriteString("- Bit depth and color information (color primaries, transfer characteristics, matrix)\n")
	prompt.WriteString("- Hardware compatibility (e.g., playback issues on low-end devices, 4K support)\n\n")
	prompt.WriteString("---\n\n")
	prompt.WriteString("🎧 3. **Audio Stream(s) Details**\n")
	prompt.WriteString("- Codec name (e.g., AAC, AC3, Opus, DTS)\n")
	prompt.WriteString("- Sample rate and bit depth\n")
	prompt.WriteString("- Number of channels (e.g., mono, stereo, 5.1)\n")
	prompt.WriteString("- Language (if tagged)\n")
	prompt.WriteString("- Bitrate\n")
	prompt.WriteString("- Compression profile\n")
	prompt.WriteString("- Delivery suitability (e.g., for OTT, broadcast, DCP)\n\n")
	prompt.WriteString("---\n\n")
	prompt.WriteString("🔠 4. **Subtitle & Metadata Streams**\n")
	prompt.WriteString("- Subtitle format (e.g., SRT, MOV_TEXT)\n")
	prompt.WriteString("- Language(s)\n")
	prompt.WriteString("- Closed captions vs open subtitles\n")
	prompt.WriteString("- Any embedded timecode tracks or metadata (e.g., encoder, creation date)\n\n")
	prompt.WriteString("---\n\n")
	prompt.WriteString("🚧 5. **Technical Analysis and Issues**\n")
	prompt.WriteString("- Any inconsistencies in codec/container (e.g., H.264 in MKV)\n")
	prompt.WriteString("- Missing audio or video streams\n")
	prompt.WriteString("- Extremely high or low bitrates\n")
	prompt.WriteString("- VFR detection or improper timebase\n")
	prompt.WriteString("- Color profile mismatch\n")
	prompt.WriteString("- Channel layout mismatches or channel mapping issues\n")
	prompt.WriteString("- Codec settings not suitable for delivery or archiving\n\n")
	prompt.WriteString("---\n\n")
	prompt.WriteString("⚙️ 6. **Recommended FFmpeg Fixes or Optimizations**\n")
	prompt.WriteString("- Suggest FFmpeg commands to fix detected problems or re-encode for:\n")
	prompt.WriteString("  - Web/OTT delivery\n")
	prompt.WriteString("  - Archival/preservation\n")
	prompt.WriteString("  - YouTube/Vimeo upload\n")
	prompt.WriteString("  - Standard broadcast or DCP delivery\n\n")
	prompt.WriteString("Explain each command and why it helps.\n\n")
	prompt.WriteString("---\n\n")
	prompt.WriteString("🧑‍💻 7. **Summary for Non-Technical Users**\n")
	prompt.WriteString("- Translate the findings into simple language\n")
	prompt.WriteString("- Mention if the file is:\n")
	prompt.WriteString("  - Good for editing/post-production\n")
	prompt.WriteString("  - Suitable for YouTube or social media\n")
	prompt.WriteString("  - Playable on common devices\n")
	prompt.WriteString("  - Problematic in any way\n")
	prompt.WriteString("- Mention the quality of the audio and video in human terms (e.g., \"Stereo audio at medium quality, suitable for online use.\")\n\n")
	prompt.WriteString("---\n\n")
	prompt.WriteString("📦 8. **Delivery Readiness Tags**\n")
	prompt.WriteString("- ✅ Ready for upload\n")
	prompt.WriteString("- ⚠️ Needs optimization\n")
	prompt.WriteString("- ❌ Not recommended for delivery\n")
	prompt.WriteString("- Add 1-line justification per tag\n\n")
	prompt.WriteString("---\n\n")
	prompt.WriteString("JSON will be provided next. Parse all values and reason holistically. Be precise, professional, and use terms common in studios, broadcasting, and OTT.\n\n")

	prompt.WriteString(fmt.Sprintf("File: %s\n", analysis.FileName))
	prompt.WriteString(fmt.Sprintf("Size: %d bytes\n", analysis.FileSize))
	prompt.WriteString(fmt.Sprintf("Source: %s\n\n", analysis.SourceType))

	// Add ffprobe data if available
	if len(analysis.FFprobeData.Format) > 0 || len(analysis.FFprobeData.Streams) > 0 {
		prompt.WriteString("Technical Data:\n")
		jsonData, _ := json.MarshalIndent(analysis.FFprobeData, "", "  ")
		prompt.Write(jsonData)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("Be comprehensive and professional. Use industry-standard terminology.")

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
	if len(analysis.FFprobeData.Format) > 0 || len(analysis.FFprobeData.Streams) > 0 {
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
func (s *LLMService) GenerateQualityInsights(ctx context.Context, analysis *models.Analysis, metrics []models.QualityMetrics) (string, error) {
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

// GenerateResponse generates a response for a custom prompt (used by comparison service)
func (s *LLMService) GenerateResponse(ctx context.Context, prompt string) (string, error) {
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
func (s *LLMService) buildQualityInsightsPrompt(analysis *models.Analysis, metrics []models.QualityMetrics) string {
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

	// Check if configured models are available
	if s.config.OllamaModel != "" {
		status.ConfiguredModel = s.config.OllamaModel
		for _, model := range status.Models {
			if model == s.config.OllamaModel {
				status.ModelAvailable = true
				break
			}
		}
	}

	// Check fallback model availability
	if s.config.OllamaFallbackModel != "" {
		status.FallbackModel = s.config.OllamaFallbackModel
		for _, model := range status.Models {
			if model == s.config.OllamaFallbackModel {
				status.FallbackModelAvailable = true
				break
			}
		}
	}

	status.Healthy = len(status.Models) > 0
	return status, nil
}

// GetCircuitBreakerStatus returns the status of all circuit breakers
func (s *LLMService) GetCircuitBreakerStatus() map[string]interface{} {
	return map[string]interface{}{
		"ollama": map[string]interface{}{
			"name":   s.ollamaCircuitBreaker.Name(),
			"state":  s.ollamaCircuitBreaker.State().String(),
			"counts": s.ollamaCircuitBreaker.Counts(),
		},
		"openrouter": map[string]interface{}{
			"name":   s.openrouterCircuitBreaker.Name(),
			"state":  s.openrouterCircuitBreaker.State().String(),
			"counts": s.openrouterCircuitBreaker.Counts(),
		},
	}
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
	URL                    string    `json:"url"`
	Healthy                bool      `json:"healthy"`
	Models                 []string  `json:"models"`
	ConfiguredModel        string    `json:"configured_model"`
	ModelAvailable         bool      `json:"model_available"`
	FallbackModel          string    `json:"fallback_model,omitempty"`
	FallbackModelAvailable bool      `json:"fallback_model_available,omitempty"`
	Error                  string    `json:"error,omitempty"`
	Timestamp              time.Time `json:"timestamp"`
}
