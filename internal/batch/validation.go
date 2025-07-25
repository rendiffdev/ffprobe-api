package batch

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rendiffdev/ffprobe-api/internal/ffmpeg"
)

// ValidateBatchRequest validates a batch analysis request
func ValidateBatchRequest(request *BatchRequest) error {
	if request == nil {
		return fmt.Errorf("batch request cannot be nil")
	}

	// Validate files list
	if len(request.Files) == 0 {
		return fmt.Errorf("batch request must contain at least one file")
	}

	if len(request.Files) > 1000 {
		return fmt.Errorf("batch size too large: %d files (max 1000)", len(request.Files))
	}

	// Validate each file
	for i, file := range request.Files {
		if err := validateBatchFile(&file, i); err != nil {
			return fmt.Errorf("invalid file at index %d: %w", i, err)
		}
	}

	// Validate global options if provided
	if request.Options != nil {
		if err := validateFFprobeOptions(request.Options); err != nil {
			return fmt.Errorf("invalid global options: %w", err)
		}
	}

	// Validate priority
	if request.Priority != "" {
		validPriorities := []string{"low", "normal", "high", "urgent"}
		validPriority := false
		for _, valid := range validPriorities {
			if request.Priority == valid {
				validPriority = true
				break
			}
		}
		if !validPriority {
			return fmt.Errorf("invalid priority: %s", request.Priority)
		}
	}

	// Validate timeout
	if request.Timeout > 0 {
		maxTimeout := 24 * time.Hour
		if request.Timeout > maxTimeout {
			return fmt.Errorf("timeout too large: %v (max %v)", request.Timeout, maxTimeout)
		}
	}

	// Validate concurrency
	if request.Concurrency > 50 {
		return fmt.Errorf("concurrency too high: %d (max 50)", request.Concurrency)
	}

	return nil
}

// validateBatchFile validates a single file in a batch
func validateBatchFile(file *BatchFile, index int) error {
	if file == nil {
		return fmt.Errorf("file cannot be nil")
	}

	// Validate ID
	if strings.TrimSpace(file.ID) == "" {
		return fmt.Errorf("file ID cannot be empty")
	}

	if len(file.ID) > 256 {
		return fmt.Errorf("file ID too long (max 256 characters)")
	}

	// Check for dangerous characters in ID
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">", "\\", "/"}
	for _, char := range dangerousChars {
		if strings.Contains(file.ID, char) {
			return fmt.Errorf("file ID contains dangerous character: %s", char)
		}
	}

	// Validate file path
	if strings.TrimSpace(file.Path) == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	if len(file.Path) > 2000 {
		return fmt.Errorf("file path too long (max 2000 characters)")
	}

	// Check for dangerous characters in path
	for _, char := range dangerousChars {
		if strings.Contains(file.Path, char) {
			return fmt.Errorf("file path contains dangerous character: %s", char)
		}
	}

	// Check for path traversal
	if strings.Contains(file.Path, "..") {
		return fmt.Errorf("file path contains path traversal")
	}

	// Validate source type
	if file.SourceType != "" {
		validSourceTypes := []string{"local", "url", "s3", "gcs", "azure", "upload", "stream"}
		validSource := false
		for _, valid := range validSourceTypes {
			if file.SourceType == valid {
				validSource = true
				break
			}
		}
		if !validSource {
			return fmt.Errorf("invalid source type: %s", file.SourceType)
		}
	}

	// Validate file options if provided
	if file.Options != nil {
		if err := validateFFprobeOptions(file.Options); err != nil {
			return fmt.Errorf("invalid file options: %w", err)
		}
	}

	// Validate metadata
	if file.Metadata != nil {
		for key, value := range file.Metadata {
			if err := validateMetadataKey(key); err != nil {
				return fmt.Errorf("invalid metadata key: %w", err)
			}
			if err := validateMetadataValue(value); err != nil {
				return fmt.Errorf("invalid metadata value for key '%s': %w", key, err)
			}
		}
	}

	return nil
}

// validateFFprobeOptions validates FFprobe options
func validateFFprobeOptions(options *ffmpeg.FFprobeOptions) error {
	if options == nil {
		return nil
	}

	// Validate custom args
	if len(options.Args) > 100 {
		return fmt.Errorf("too many custom args: %d (max 100)", len(options.Args))
	}

	for i, arg := range options.Args {
		if strings.TrimSpace(arg) == "" {
			return fmt.Errorf("empty argument at index %d", i)
		}

		// Check for dangerous arguments
		dangerousArgs := []string{"-f", "null", "/dev/null", "|", "&", ";", "`", "$"}
		for _, dangerous := range dangerousArgs {
			if strings.Contains(arg, dangerous) {
				return fmt.Errorf("dangerous argument detected: %s", arg)
			}
		}
	}

	return nil
}

// ValidateBatchStatus validates batch status updates
func ValidateBatchStatus(batchID uuid.UUID, status string) error {
	if batchID == uuid.Nil {
		return fmt.Errorf("batch ID cannot be nil")
	}

	validStatuses := []string{"pending", "queued", "processing", "completed", "failed", "cancelled", "timeout"}
	validStatus := false
	for _, valid := range validStatuses {
		if status == valid {
			validStatus = true
			break
		}
	}

	if !validStatus {
		return fmt.Errorf("invalid batch status: %s", status)
	}

	return nil
}

// ValidateBatchResults validates batch processing results
func ValidateBatchResults(results []BatchResult) error {
	if results == nil {
		return fmt.Errorf("results cannot be nil")
	}

	for i, result := range results {
		if err := validateBatchResult(&result, i); err != nil {
			return fmt.Errorf("invalid result at index %d: %w", i, err)
		}
	}

	return nil
}

// validateBatchResult validates a single batch result
func validateBatchResult(result *BatchResult, index int) error {
	if result == nil {
		return fmt.Errorf("result cannot be nil")
	}

	// Validate file ID
	if strings.TrimSpace(result.FileID) == "" {
		return fmt.Errorf("file ID cannot be empty")
	}

	// Validate status
	validStatuses := []string{"pending", "processing", "completed", "failed", "skipped", "cancelled"}
	validStatus := false
	for _, valid := range validStatuses {
		if result.Status == valid {
			validStatus = true
			break
		}
	}

	if !validStatus {
		return fmt.Errorf("invalid result status: %s", result.Status)
	}

	// Validate processing time
	if result.ProcessingTime < 0 {
		return fmt.Errorf("processing time cannot be negative")
	}

	// Validate error message length
	if len(result.Error) > 2000 {
		return fmt.Errorf("error message too long (max 2000 characters)")
	}

	return nil
}

// validateMetadataKey validates metadata key format
func validateMetadataKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("metadata key cannot be empty")
	}

	if len(key) > 128 {
		return fmt.Errorf("metadata key too long (max 128 characters)")
	}

	// Check for valid characters (alphanumeric, dash, underscore)
	for _, char := range key {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || char == '-' || char == '_') {
			return fmt.Errorf("metadata key contains invalid character: %c", char)
		}
	}

	return nil
}

// validateMetadataValue validates metadata value
func validateMetadataValue(value string) error {
	if len(value) > 1024 {
		return fmt.Errorf("metadata value too long (max 1024 characters)")
	}

	// Check for control characters (except tab, newline, carriage return)
	for _, char := range value {
		if char < 32 && char != 9 && char != 10 && char != 13 {
			return fmt.Errorf("metadata value contains control character")
		}
	}

	return nil
}

// SanitizeBatchFileID sanitizes a batch file ID
func SanitizeBatchFileID(id string) string {
	// Remove dangerous characters
	unsafeChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">", "\\", "/", " "}
	for _, char := range unsafeChars {
		id = strings.ReplaceAll(id, char, "_")
	}

	// Remove path components
	id = filepath.Base(id)

	// Limit length
	if len(id) > 256 {
		id = id[:256]
	}

	// Ensure not empty
	if id == "" || id == "." {
		id = "batch_file"
	}

	return id
}

// ValidateBatchConfiguration validates batch processing configuration
func ValidateBatchConfiguration(config *BatchConfig) error {
	if config == nil {
		return fmt.Errorf("batch configuration cannot be nil")
	}

	// Validate max batch size
	if config.MaxBatchSize <= 0 {
		return fmt.Errorf("max batch size must be positive")
	}

	if config.MaxBatchSize > 10000 {
		return fmt.Errorf("max batch size too large: %d (max 10000)", config.MaxBatchSize)
	}

	// Validate max concurrency
	if config.MaxConcurrency <= 0 {
		return fmt.Errorf("max concurrency must be positive")
	}

	if config.MaxConcurrency > 100 {
		return fmt.Errorf("max concurrency too high: %d (max 100)", config.MaxConcurrency)
	}

	// Validate timeout
	if config.DefaultTimeout <= 0 {
		return fmt.Errorf("default timeout must be positive")
	}

	maxTimeout := 48 * time.Hour
	if config.DefaultTimeout > maxTimeout {
		return fmt.Errorf("default timeout too large: %v (max %v)", config.DefaultTimeout, maxTimeout)
	}

	// Validate queue size
	if config.QueueSize < 0 {
		return fmt.Errorf("queue size cannot be negative")
	}

	return nil
}

// Batch processing types for validation
type BatchRequest struct {
	Files       []BatchFile           `json:"files"`
	Options     *ffmpeg.FFprobeOptions `json:"options,omitempty"`
	Priority    string                `json:"priority,omitempty"`
	Timeout     time.Duration         `json:"timeout,omitempty"`
	Concurrency int                   `json:"concurrency,omitempty"`
	Async       bool                  `json:"async,omitempty"`
}

type BatchFile struct {
	ID         string                    `json:"id"`
	Path       string                    `json:"path"`
	SourceType string                    `json:"source_type,omitempty"`
	Options    *ffmpeg.FFprobeOptions     `json:"options,omitempty"`
	Metadata   map[string]string         `json:"metadata,omitempty"`
}

type BatchResult struct {
	FileID         string        `json:"file_id"`
	AnalysisID     uuid.UUID     `json:"analysis_id,omitempty"`
	Status         string        `json:"status"`
	Error          string        `json:"error,omitempty"`
	ProcessingTime time.Duration `json:"processing_time"`
	StartedAt      time.Time     `json:"started_at"`
	CompletedAt    *time.Time    `json:"completed_at,omitempty"`
}

type BatchConfig struct {
	MaxBatchSize   int           `json:"max_batch_size"`
	MaxConcurrency int           `json:"max_concurrency"`
	DefaultTimeout time.Duration `json:"default_timeout"`
	QueueSize      int           `json:"queue_size"`
}