package storage

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateStorageConfig validates storage configuration
func ValidateStorageConfig(cfg Config) error {
	if cfg.Provider == "" {
		return fmt.Errorf("storage provider cannot be empty")
	}

	validProviders := []string{"s3", "aws", "gcs", "google", "azure", "azblob", "local", "filesystem"}
	validProvider := false
	for _, valid := range validProviders {
		if strings.ToLower(cfg.Provider) == valid {
			validProvider = true
			break
		}
	}

	if !validProvider {
		return fmt.Errorf("unsupported storage provider: %s", cfg.Provider)
	}

	// Provider-specific validation
	switch strings.ToLower(cfg.Provider) {
	case "s3", "aws":
		return validateS3Config(cfg)
	case "gcs", "google":
		return validateGCSConfig(cfg)
	case "azure", "azblob":
		return validateAzureConfig(cfg)
	case "local", "filesystem":
		return validateLocalConfig(cfg)
	}

	return nil
}

// validateS3Config validates S3-specific configuration
func validateS3Config(cfg Config) error {
	if cfg.Bucket == "" {
		return fmt.Errorf("S3 bucket name cannot be empty")
	}

	if cfg.AccessKey == "" {
		return fmt.Errorf("S3 access key cannot be empty")
	}

	if cfg.SecretKey == "" {
		return fmt.Errorf("S3 secret key cannot be empty")
	}

	if cfg.Region == "" {
		return fmt.Errorf("S3 region cannot be empty")
	}

	// Validate bucket name format
	if err := validateBucketName(cfg.Bucket); err != nil {
		return fmt.Errorf("invalid S3 bucket name: %w", err)
	}

	// Validate region format
	if err := validateAWSRegion(cfg.Region); err != nil {
		return fmt.Errorf("invalid AWS region: %w", err)
	}

	return nil
}

// validateGCSConfig validates Google Cloud Storage configuration
func validateGCSConfig(cfg Config) error {
	if cfg.Bucket == "" {
		return fmt.Errorf("GCS bucket name cannot be empty")
	}

	// GCS bucket name validation
	if err := validateGCSBucketName(cfg.Bucket); err != nil {
		return fmt.Errorf("invalid GCS bucket name: %w", err)
	}

	return nil
}

// validateAzureConfig validates Azure Blob Storage configuration
func validateAzureConfig(cfg Config) error {
	if cfg.Bucket == "" {
		return fmt.Errorf("Azure container name cannot be empty")
	}

	if cfg.AccessKey == "" {
		return fmt.Errorf("Azure account name cannot be empty")
	}

	if cfg.SecretKey == "" {
		return fmt.Errorf("Azure account key cannot be empty")
	}

	// Validate container name
	if err := validateAzureContainerName(cfg.Bucket); err != nil {
		return fmt.Errorf("invalid Azure container name: %w", err)
	}

	return nil
}

// validateLocalConfig validates local storage configuration
func validateLocalConfig(cfg Config) error {
	if cfg.Bucket == "" {
		return fmt.Errorf("local storage path cannot be empty")
	}

	// Check for dangerous path characters
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">"}
	for _, char := range dangerousChars {
		if strings.Contains(cfg.Bucket, char) {
			return fmt.Errorf("storage path contains dangerous character: %s", char)
		}
	}

	// Check for path traversal
	if strings.Contains(cfg.Bucket, "..") {
		return fmt.Errorf("storage path contains path traversal")
	}

	return nil
}

// ValidateStorageKey validates a storage key
func ValidateStorageKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("storage key cannot be empty")
	}

	if len(key) > 1024 {
		return fmt.Errorf("storage key too long (max 1024 characters)")
	}

	// Check for dangerous characters
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">", "\\"}
	for _, char := range dangerousChars {
		if strings.Contains(key, char) {
			return fmt.Errorf("storage key contains dangerous character: %s", char)
		}
	}

	// Check for path traversal
	if strings.Contains(key, "..") {
		return fmt.Errorf("storage key contains path traversal")
	}

	// Ensure key doesn't start with /
	if strings.HasPrefix(key, "/") {
		return fmt.Errorf("storage key cannot start with /")
	}

	return nil
}

// ValidateUploadOptions validates upload options
func ValidateUploadOptions(opts *UploadOptions) error {
	if opts == nil {
		return nil
	}

	// Validate content type
	if opts.ContentType != "" {
		if err := validateContentType(opts.ContentType); err != nil {
			return fmt.Errorf("invalid content type: %w", err)
		}
	}

	// Validate cache control
	if opts.CacheControl != "" {
		if len(opts.CacheControl) > 1000 {
			return fmt.Errorf("cache control header too long (max 1000 characters)")
		}
	}

	// Validate metadata
	if opts.Metadata != nil {
		for key, value := range opts.Metadata {
			if err := validateMetadataKey(key); err != nil {
				return fmt.Errorf("invalid metadata key '%s': %w", key, err)
			}
			if err := validateMetadataValue(value); err != nil {
				return fmt.Errorf("invalid metadata value for key '%s': %w", key, err)
			}
		}
	}

	return nil
}

// validateBucketName validates S3 bucket name
func validateBucketName(bucket string) error {
	if len(bucket) < 3 || len(bucket) > 63 {
		return fmt.Errorf("bucket name must be between 3 and 63 characters")
	}

	// Bucket name pattern (simplified)
	pattern := `^[a-z0-9][a-z0-9\-]*[a-z0-9]$`
	matched, err := regexp.MatchString(pattern, bucket)
	if err != nil {
		return fmt.Errorf("error validating bucket name: %w", err)
	}

	if !matched {
		return fmt.Errorf("bucket name contains invalid characters")
	}

	// Check for consecutive dots or dashes
	if strings.Contains(bucket, "..") || strings.Contains(bucket, "--") {
		return fmt.Errorf("bucket name cannot contain consecutive dots or dashes")
	}

	return nil
}

// validateGCSBucketName validates Google Cloud Storage bucket name
func validateGCSBucketName(bucket string) error {
	if len(bucket) < 3 || len(bucket) > 63 {
		return fmt.Errorf("GCS bucket name must be between 3 and 63 characters")
	}

	// GCS bucket name pattern
	pattern := `^[a-z0-9][a-z0-9\-_]*[a-z0-9]$`
	matched, err := regexp.MatchString(pattern, bucket)
	if err != nil {
		return fmt.Errorf("error validating GCS bucket name: %w", err)
	}

	if !matched {
		return fmt.Errorf("GCS bucket name contains invalid characters")
	}

	return nil
}

// validateAzureContainerName validates Azure container name
func validateAzureContainerName(container string) error {
	if len(container) < 3 || len(container) > 63 {
		return fmt.Errorf("Azure container name must be between 3 and 63 characters")
	}

	// Azure container name pattern
	pattern := `^[a-z0-9][a-z0-9\-]*[a-z0-9]$`
	matched, err := regexp.MatchString(pattern, container)
	if err != nil {
		return fmt.Errorf("error validating Azure container name: %w", err)
	}

	if !matched {
		return fmt.Errorf("Azure container name contains invalid characters")
	}

	// Cannot contain consecutive dashes
	if strings.Contains(container, "--") {
		return fmt.Errorf("Azure container name cannot contain consecutive dashes")
	}

	return nil
}

// validateAWSRegion validates AWS region format
func validateAWSRegion(region string) error {
	// AWS region pattern
	pattern := `^[a-z]{2}-[a-z]+-\d+$`
	matched, err := regexp.MatchString(pattern, region)
	if err != nil {
		return fmt.Errorf("error validating AWS region: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid AWS region format")
	}

	return nil
}

// validateContentType validates MIME content type
func validateContentType(contentType string) error {
	if len(contentType) > 255 {
		return fmt.Errorf("content type too long (max 255 characters)")
	}

	// Basic MIME type pattern
	pattern := `^[a-zA-Z][a-zA-Z0-9\-\+]*\/[a-zA-Z0-9\-\+\.]*$`
	matched, err := regexp.MatchString(pattern, contentType)
	if err != nil {
		return fmt.Errorf("error validating content type: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid content type format")
	}

	return nil
}

// validateMetadataKey validates metadata key
func validateMetadataKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("metadata key cannot be empty")
	}

	if len(key) > 256 {
		return fmt.Errorf("metadata key too long (max 256 characters)")
	}

	// Metadata key pattern (alphanumeric, dashes, underscores)
	pattern := `^[a-zA-Z0-9\-_]+$`
	matched, err := regexp.MatchString(pattern, key)
	if err != nil {
		return fmt.Errorf("error validating metadata key: %w", err)
	}

	if !matched {
		return fmt.Errorf("metadata key contains invalid characters")
	}

	return nil
}

// validateMetadataValue validates metadata value
func validateMetadataValue(value string) error {
	if len(value) > 2048 {
		return fmt.Errorf("metadata value too long (max 2048 characters)")
	}

	// Check for control characters
	for _, char := range value {
		if char < 32 && char != 9 && char != 10 && char != 13 {
			return fmt.Errorf("metadata value contains control characters")
		}
	}

	return nil
}

// SanitizeStorageKey sanitizes a storage key for safe usage
func SanitizeStorageKey(key string) string {
	// Remove dangerous characters
	unsafeChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">", "\\", " "}
	for _, char := range unsafeChars {
		key = strings.ReplaceAll(key, char, "_")
	}

	// Remove path traversal
	key = strings.ReplaceAll(key, "..", "_")

	// Remove leading slashes
	key = strings.TrimPrefix(key, "/")

	// Clean up the path
	key = filepath.Clean(key)

	// Limit length
	if len(key) > 1024 {
		ext := filepath.Ext(key)
		base := key[:1024-len(ext)]
		key = base + ext
	}

	// Ensure not empty
	if key == "" || key == "." {
		key = "unnamed_file"
	}

	return key
}

// ValidateFileSize validates file size limits
func ValidateFileSize(size int64, maxSize int64) error {
	if size < 0 {
		return fmt.Errorf("file size cannot be negative")
	}

	if maxSize > 0 && size > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed: %d > %d", size, maxSize)
	}

	return nil
}