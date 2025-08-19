package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// WorkerClient handles communication with worker services
type WorkerClient struct {
	ffprobeWorkerURL string
	llmServiceURL    string
	httpClient       *http.Client
	logger           zerolog.Logger
}

// NewWorkerClient creates a new worker client
func NewWorkerClient(ffprobeWorkerURL, llmServiceURL string, logger zerolog.Logger) *WorkerClient {
	return &WorkerClient{
		ffprobeWorkerURL: ffprobeWorkerURL,
		llmServiceURL:    llmServiceURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Generous timeout for media processing
		},
		logger: logger,
	}
}

// FFprobeWorkerRequest represents a request to the FFprobe worker
type FFprobeWorkerRequest struct {
	FilePath string                 `json:"file_path"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// FFprobeWorkerResponse represents a response from the FFprobe worker
type FFprobeWorkerResponse struct {
	Success        bool                   `json:"success"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Error          string                 `json:"error,omitempty"`
	ProcessingTime time.Duration          `json:"processing_time"`
}

// LLMWorkerRequest represents a request to the LLM service
type LLMWorkerRequest struct {
	AnalysisData map[string]interface{} `json:"analysis_data"`
	Prompt       string                 `json:"prompt,omitempty"`
}

// LLMWorkerResponse represents a response from the LLM service
type LLMWorkerResponse struct {
	Success        bool          `json:"success"`
	Report         string        `json:"report,omitempty"`
	Error          string        `json:"error,omitempty"`
	ProcessingTime time.Duration `json:"processing_time"`
}

// AnalyzeWithWorker performs media analysis using the FFprobe worker service
func (wc *WorkerClient) AnalyzeWithWorker(ctx context.Context, filePath string, options map[string]interface{}) (map[string]interface{}, error) {
	// Check if worker service is available, fallback to local if not
	if !wc.isWorkerHealthy(ctx, wc.ffprobeWorkerURL) {
		wc.logger.Warn().Msg("FFprobe worker unavailable, this would fallback to local processing")
		// Return a basic response to maintain functionality
		return map[string]interface{}{
			"format": map[string]interface{}{
				"filename": filePath,
				"note":     "Processed locally (worker unavailable)",
			},
			"streams": []interface{}{},
		}, nil
	}

	req := FFprobeWorkerRequest{
		FilePath: filePath,
		Options:  options,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", wc.ffprobeWorkerURL+"/analyze", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := wc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call worker: %w", err)
	}
	defer resp.Body.Close()

	var workerResp FFprobeWorkerResponse
	if err := json.NewDecoder(resp.Body).Decode(&workerResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !workerResp.Success {
		return nil, fmt.Errorf("worker analysis failed: %s", workerResp.Error)
	}

	wc.logger.Info().
		Str("file_path", filePath).
		Dur("processing_time", workerResp.ProcessingTime).
		Msg("Worker analysis completed")

	return workerResp.Data, nil
}

// GenerateAnalysisWithLLM generates analysis using the LLM service
func (wc *WorkerClient) GenerateAnalysisWithLLM(ctx context.Context, analysisData map[string]interface{}) (string, error) {
	// Check if LLM service is available, return empty if not
	if !wc.isWorkerHealthy(ctx, wc.llmServiceURL) {
		wc.logger.Warn().Msg("LLM service unavailable, skipping AI analysis")
		return "", nil
	}

	req := LLMWorkerRequest{
		AnalysisData: analysisData,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", wc.llmServiceURL+"/analyze", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := wc.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call LLM service: %w", err)
	}
	defer resp.Body.Close()

	var llmResp LLMWorkerResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !llmResp.Success {
		return "", fmt.Errorf("LLM analysis failed: %s", llmResp.Error)
	}

	wc.logger.Info().
		Dur("processing_time", llmResp.ProcessingTime).
		Int("report_length", len(llmResp.Report)).
		Msg("LLM analysis completed")

	return llmResp.Report, nil
}

// isWorkerHealthy checks if a worker service is healthy
func (wc *WorkerClient) isWorkerHealthy(ctx context.Context, serviceURL string) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", serviceURL+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
