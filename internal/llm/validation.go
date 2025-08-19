package llm

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ValidateLLMConfig validates LLM service configuration
func ValidateLLMConfig(config *LLMConfig) error {
	if config == nil {
		return fmt.Errorf("LLM configuration cannot be nil")
	}

	// Validate provider
	if config.Provider == "" {
		return fmt.Errorf("LLM provider cannot be empty")
	}

	validProviders := []string{"openrouter", "openai", "anthropic", "ollama", "local"}
	validProvider := false
	for _, valid := range validProviders {
		if config.Provider == valid {
			validProvider = true
			break
		}
	}

	if !validProvider {
		return fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}

	// Provider-specific validation
	switch config.Provider {
	case "openrouter":
		return validateOpenRouterConfig(config)
	case "openai":
		return validateOpenAIConfig(config)
	case "anthropic":
		return validateAnthropicConfig(config)
	case "ollama":
		return validateOllamaConfig(config)
	case "local":
		return validateLocalConfig(config)
	}

	return nil
}

// validateOpenRouterConfig validates OpenRouter-specific configuration
func validateOpenRouterConfig(config *LLMConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("OpenRouter API key cannot be empty")
	}

	if len(config.APIKey) < 10 {
		return fmt.Errorf("OpenRouter API key appears invalid (too short)")
	}

	// Validate model name
	if config.Model == "" {
		config.Model = "anthropic/claude-3-haiku" // Default model
	}

	validModels := []string{
		"anthropic/claude-3-haiku",
		"anthropic/claude-3-sonnet",
		"anthropic/claude-3-opus",
		"openai/gpt-3.5-turbo",
		"openai/gpt-4",
		"openai/gpt-4-turbo",
		"meta-llama/llama-2-70b-chat",
		"google/gemini-pro",
	}

	validModel := false
	for _, valid := range validModels {
		if config.Model == valid {
			validModel = true
			break
		}
	}

	if !validModel {
		return fmt.Errorf("unsupported OpenRouter model: %s", config.Model)
	}

	return nil
}

// validateOpenAIConfig validates OpenAI-specific configuration
func validateOpenAIConfig(config *LLMConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("OpenAI API key cannot be empty")
	}

	// Check API key format (starts with sk-)
	if !strings.HasPrefix(config.APIKey, "sk-") {
		return fmt.Errorf("OpenAI API key should start with 'sk-'")
	}

	if len(config.APIKey) < 20 {
		return fmt.Errorf("OpenAI API key appears invalid (too short)")
	}

	// Validate model
	if config.Model == "" {
		config.Model = "gpt-3.5-turbo" // Default model
	}

	validModels := []string{
		"gpt-3.5-turbo",
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4o",
		"gpt-4o-mini",
	}

	validModel := false
	for _, valid := range validModels {
		if config.Model == valid {
			validModel = true
			break
		}
	}

	if !validModel {
		return fmt.Errorf("unsupported OpenAI model: %s", config.Model)
	}

	return nil
}

// validateAnthropicConfig validates Anthropic-specific configuration
func validateAnthropicConfig(config *LLMConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("Anthropic API key cannot be empty")
	}

	// Check API key format (starts with sk-ant-)
	if !strings.HasPrefix(config.APIKey, "sk-ant-") {
		return fmt.Errorf("Anthropic API key should start with 'sk-ant-'")
	}

	if len(config.APIKey) < 30 {
		return fmt.Errorf("Anthropic API key appears invalid (too short)")
	}

	// Validate model
	if config.Model == "" {
		config.Model = "claude-3-haiku-20240307" // Default model
	}

	validModels := []string{
		"claude-3-haiku-20240307",
		"claude-3-sonnet-20240229",
		"claude-3-opus-20240229",
		"claude-2.1",
		"claude-2.0",
	}

	validModel := false
	for _, valid := range validModels {
		if config.Model == valid {
			validModel = true
			break
		}
	}

	if !validModel {
		return fmt.Errorf("unsupported Anthropic model: %s", config.Model)
	}

	return nil
}

// validateOllamaConfig validates Ollama-specific configuration
func validateOllamaConfig(config *LLMConfig) error {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:11434" // Default Ollama URL
	}

	// Validate URL format
	if _, err := url.Parse(config.BaseURL); err != nil {
		return fmt.Errorf("invalid Ollama base URL: %w", err)
	}

	// Validate model name
	if config.Model == "" {
		return fmt.Errorf("Ollama model name cannot be empty")
	}

	// Basic model name validation (alphanumeric, dash, underscore, colon)
	modelPattern := `^[a-zA-Z0-9\-_:]+$`
	matched, err := regexp.MatchString(modelPattern, config.Model)
	if err != nil {
		return fmt.Errorf("error validating Ollama model name: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid Ollama model name format: %s", config.Model)
	}

	return nil
}

// validateLocalConfig validates local LLM configuration
func validateLocalConfig(config *LLMConfig) error {
	if config.ModelPath == "" {
		return fmt.Errorf("local model path cannot be empty")
	}

	// Check for dangerous path characters
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">"}
	for _, char := range dangerousChars {
		if strings.Contains(config.ModelPath, char) {
			return fmt.Errorf("model path contains dangerous character: %s", char)
		}
	}

	// Check for path traversal
	if strings.Contains(config.ModelPath, "..") {
		return fmt.Errorf("model path contains path traversal")
	}

	return nil
}

// ValidateLLMRequest validates an LLM generation request
func ValidateLLMRequest(request *LLMRequest) error {
	if request == nil {
		return fmt.Errorf("LLM request cannot be nil")
	}

	// Validate prompt
	if strings.TrimSpace(request.Prompt) == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	if len(request.Prompt) > 100000 {
		return fmt.Errorf("prompt too long (max 100,000 characters)")
	}

	// Check for potentially dangerous content
	if err := validatePromptSafety(request.Prompt); err != nil {
		return fmt.Errorf("unsafe prompt content: %w", err)
	}

	// Validate parameters
	if request.MaxTokens < 0 {
		return fmt.Errorf("max tokens cannot be negative")
	}

	if request.MaxTokens > 32000 {
		return fmt.Errorf("max tokens too large (max 32,000)")
	}

	if request.Temperature < 0 || request.Temperature > 2.0 {
		return fmt.Errorf("temperature must be between 0 and 2.0")
	}

	if request.TopP < 0 || request.TopP > 1.0 {
		return fmt.Errorf("top_p must be between 0 and 1.0")
	}

	// Validate timeout
	if request.Timeout > 0 {
		maxTimeout := 10 * time.Minute
		if request.Timeout > maxTimeout {
			return fmt.Errorf("timeout too large: %v (max %v)", request.Timeout, maxTimeout)
		}
	}

	return nil
}

// validatePromptSafety checks for potentially unsafe prompt content
func validatePromptSafety(prompt string) error {
	// Check for injection attempts
	dangerousPatterns := []string{
		"ignore previous instructions",
		"forget your role",
		"system:",
		"<script>",
		"javascript:",
		"eval(",
		"exec(",
		"__import__",
		"subprocess",
		"os.system",
	}

	lowerPrompt := strings.ToLower(prompt)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerPrompt, pattern) {
			return fmt.Errorf("prompt contains potentially dangerous pattern: %s", pattern)
		}
	}

	// Check for excessive repetition (potential DoS)
	if detectExcessiveRepetition(prompt) {
		return fmt.Errorf("prompt contains excessive repetition")
	}

	return nil
}

// detectExcessiveRepetition checks for patterns that might cause issues
func detectExcessiveRepetition(prompt string) bool {
	// Simple check: if any 10-character substring appears more than 50 times
	substrings := make(map[string]int)

	if len(prompt) < 10 {
		return false
	}

	for i := 0; i <= len(prompt)-10; i++ {
		substr := prompt[i : i+10]
		substrings[substr]++
		if substrings[substr] > 50 {
			return true
		}
	}

	return false
}

// ValidateLLMResponse validates an LLM response
func ValidateLLMResponse(response *LLMResponse) error {
	if response == nil {
		return fmt.Errorf("LLM response cannot be nil")
	}

	// Validate content
	if len(response.Content) > 500000 {
		return fmt.Errorf("response content too large (max 500,000 characters)")
	}

	// Validate usage metrics
	if response.Usage != nil {
		if response.Usage.PromptTokens < 0 {
			return fmt.Errorf("prompt tokens cannot be negative")
		}
		if response.Usage.CompletionTokens < 0 {
			return fmt.Errorf("completion tokens cannot be negative")
		}
		if response.Usage.TotalTokens < 0 {
			return fmt.Errorf("total tokens cannot be negative")
		}
	}

	// Validate processing time
	if response.ProcessingTime < 0 {
		return fmt.Errorf("processing time cannot be negative")
	}

	return nil
}

// SanitizePrompt sanitizes a prompt for safe usage
func SanitizePrompt(prompt string) string {
	// Remove control characters except newlines and tabs
	sanitized := ""
	for _, char := range prompt {
		if char >= 32 || char == 9 || char == 10 || char == 13 {
			sanitized += string(char)
		}
	}

	// Limit length
	if len(sanitized) > 100000 {
		sanitized = sanitized[:100000]
	}

	// Trim excessive whitespace
	lines := strings.Split(sanitized, "\n")
	cleanLines := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" || len(cleanLines) == 0 || cleanLines[len(cleanLines)-1] != "" {
			cleanLines = append(cleanLines, trimmed)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// LLM configuration and request types
type LLMConfig struct {
	Provider    string        `json:"provider"`
	APIKey      string        `json:"api_key,omitempty"`
	BaseURL     string        `json:"base_url,omitempty"`
	Model       string        `json:"model"`
	ModelPath   string        `json:"model_path,omitempty"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
	Timeout     time.Duration `json:"timeout"`
}

type LLMRequest struct {
	Prompt      string        `json:"prompt"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	Stop        []string      `json:"stop,omitempty"`
	Timeout     time.Duration `json:"timeout,omitempty"`
}

type LLMResponse struct {
	Content        string        `json:"content"`
	FinishReason   string        `json:"finish_reason,omitempty"`
	Usage          *LLMUsage     `json:"usage,omitempty"`
	ProcessingTime time.Duration `json:"processing_time"`
	Model          string        `json:"model,omitempty"`
}

type LLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
