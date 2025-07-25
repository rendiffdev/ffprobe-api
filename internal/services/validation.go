package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rendiffdev/ffprobe-api/internal/models"
)

// ValidateAnalysisRequest validates an analysis creation request
func ValidateAnalysisRequest(request *models.CreateAnalysisRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate file name
	if strings.TrimSpace(request.FileName) == "" {
		return fmt.Errorf("file name cannot be empty")
	}

	if len(request.FileName) > 500 {
		return fmt.Errorf("file name too long (max 500 characters)")
	}

	// Validate file path
	if strings.TrimSpace(request.FilePath) == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	if len(request.FilePath) > 2000 {
		return fmt.Errorf("file path too long (max 2000 characters)")
	}

	// Check for dangerous characters in path
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">"}
	for _, char := range dangerousChars {
		if strings.Contains(request.FilePath, char) {
			return fmt.Errorf("file path contains dangerous character: %s", char)
		}
	}

	// Check for path traversal
	if strings.Contains(request.FilePath, "..") {
		return fmt.Errorf("file path contains path traversal")
	}

	// Validate source type
	validSourceTypes := []string{"local", "url", "s3", "gcs", "azure", "upload"}
	validSource := false
	for _, valid := range validSourceTypes {
		if request.SourceType == valid {
			validSource = true
			break
		}
	}

	if !validSource {
		return fmt.Errorf("invalid source type: %s", request.SourceType)
	}

	// Validate file size
	if request.FileSize < 0 {
		return fmt.Errorf("file size cannot be negative")
	}

	maxFileSize := int64(100 * 1024 * 1024 * 1024) // 100GB
	if request.FileSize > maxFileSize {
		return fmt.Errorf("file size too large: %d bytes (max %d)", request.FileSize, maxFileSize)
	}

	// For local files, validate existence and accessibility
	if request.SourceType == "local" && !strings.Contains(request.FilePath, "://") {
		if err := validateLocalFile(request.FilePath); err != nil {
			return fmt.Errorf("local file validation failed: %w", err)
		}
	}

	// Validate content hash format if provided
	if request.ContentHash != "" {
		if len(request.ContentHash) != 64 { // SHA-256 hash length
			return fmt.Errorf("invalid content hash format (expected SHA-256)")
		}

		// Check if hash contains only hex characters
		for _, char := range request.ContentHash {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return fmt.Errorf("content hash contains invalid characters")
			}
		}
	}

	return nil
}

// validateLocalFile validates a local file path
func validateLocalFile(filePath string) error {
	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Check if it's a file (not directory)
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	validExts := []string{
		".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v",
		".mpg", ".mpeg", ".3gp", ".3g2", ".mxf", ".ts", ".vob", ".ogv",
		".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a", ".opus",
		".m3u8", ".mpd", // Streaming formats
	}

	validExt := false
	for _, validExtension := range validExts {
		if ext == validExtension {
			validExt = true
			break
		}
	}

	if !validExt && ext != "" {
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	return nil
}

// ValidateAnalysisUpdate validates an analysis update request
func ValidateAnalysisUpdate(analysis *models.Analysis) error {
	if analysis == nil {
		return fmt.Errorf("analysis cannot be nil")
	}

	// Validate status
	validStatuses := []models.AnalysisStatus{
		models.StatusPending,
		models.StatusProcessing,
		models.StatusCompleted,
		models.StatusFailed,
	}

	validStatus := false
	for _, valid := range validStatuses {
		if analysis.Status == valid {
			validStatus = true
			break
		}
	}

	if !validStatus {
		return fmt.Errorf("invalid analysis status: %s", analysis.Status)
	}

	// Validate FFprobe data if present
	if analysis.FFprobeData != nil {
		if err := validateFFprobeData(analysis.FFprobeData); err != nil {
			return fmt.Errorf("invalid FFprobe data: %w", err)
		}
	}

	// Validate error message length
	if analysis.ErrorMsg != nil && len(*analysis.ErrorMsg) > 2000 {
		return fmt.Errorf("error message too long (max 2000 characters)")
	}

	return nil
}

// validateFFprobeData validates FFprobe output data
func validateFFprobeData(data interface{}) error {
	// Check if data can be marshaled to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("FFprobe data is not valid JSON: %w", err)
	}

	// Check data size (max 100MB JSON)
	maxSize := 100 * 1024 * 1024
	if len(jsonData) > maxSize {
		return fmt.Errorf("FFprobe data too large: %d bytes (max %d)", len(jsonData), maxSize)
	}

	// Try to unmarshal back to validate structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		return fmt.Errorf("FFprobe data has invalid structure: %w", err)
	}

	return nil
}

// ValidateAnalysisID validates an analysis ID format
func ValidateAnalysisID(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("analysis ID cannot be empty")
	}

	// Check UUID format (36 characters with dashes)
	if len(id) != 36 {
		return fmt.Errorf("invalid analysis ID format")
	}

	// Check for valid UUID pattern
	uuidPattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	matched, err := regexp.MatchString(uuidPattern, id)
	if err != nil {
		return fmt.Errorf("error validating analysis ID: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid analysis ID format")
	}

	return nil
}

// SanitizeFileName sanitizes a file name for safe storage
func SanitizeFileName(filename string) string {
	// Remove path components
	filename = filepath.Base(filename)

	// Replace unsafe characters
	unsafeChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t"}
	for _, char := range unsafeChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	// Remove leading/trailing spaces and dots
	filename = strings.Trim(filename, " .")

	// Limit length
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		base := filename[:255-len(ext)]
		filename = base + ext
	}

	// Ensure not empty
	if filename == "" {
		filename = "unnamed_file"
	}

	return filename
}

// ValidateUserPermissions validates user permissions for analysis operations
func ValidateUserPermissions(userID string, operation string) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	validOperations := []string{"create", "read", "update", "delete", "analyze"}
	validOperation := false
	for _, valid := range validOperations {
		if operation == valid {
			validOperation = true
			break
		}
	}

	if !validOperation {
		return fmt.Errorf("invalid operation: %s", operation)
	}

	// Additional permission checks can be added here
	// For example, checking user roles, quotas, etc.

	return nil
}