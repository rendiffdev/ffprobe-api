package validator

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// FilePathValidator validates file paths for security
type FilePathValidator struct {
	allowedExtensions []string
	maxPathLength     int
	blockPatterns     []*regexp.Regexp
}

// NewFilePathValidator creates a new file path validator
func NewFilePathValidator() *FilePathValidator {
	return &FilePathValidator{
		allowedExtensions: []string{
			".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm",
			".m4v", ".mpg", ".mpeg", ".3gp", ".3g2", ".mxf", ".ts",
			".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a",
			".opus", ".m3u8", ".mpd",
		},
		maxPathLength: 4096,
		blockPatterns: []*regexp.Regexp{
			regexp.MustCompile(`\.\.`),     // Directory traversal
			regexp.MustCompile(`^/etc/`),   // System files
			regexp.MustCompile(`^/proc/`),  // Process info
			regexp.MustCompile(`^/sys/`),   // System info
			regexp.MustCompile(`^/dev/`),   // Device files
			regexp.MustCompile(`\x00`),     // Null bytes
			regexp.MustCompile(`[<>"|*?]`), // Invalid characters
			regexp.MustCompile(`^\s*$`),    // Empty paths
		},
	}
}

// ValidateFilePath validates a file path
func (v *FilePathValidator) ValidateFilePath(path string) error {
	// Check empty
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Check length
	if len(path) > v.maxPathLength {
		return fmt.Errorf("file path too long: %d > %d", len(path), v.maxPathLength)
	}

	// Clean path
	cleanPath := filepath.Clean(path)

	// Check for blocked patterns
	for _, pattern := range v.blockPatterns {
		if pattern.MatchString(cleanPath) {
			return fmt.Errorf("invalid file path: contains blocked pattern")
		}
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(cleanPath))
	validExt := false
	for _, allowed := range v.allowedExtensions {
		if ext == allowed {
			validExt = true
			break
		}
	}

	if !validExt && ext != "" {
		return fmt.Errorf("unsupported file extension: %s", ext)
	}

	return nil
}

// ValidateURL validates a URL for security
func ValidateURL(urlStr string) error {
	// Check empty
	if strings.TrimSpace(urlStr) == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check scheme
	validSchemes := []string{"http", "https", "rtmp", "rtsp", "s3", "gs", "file"}
	schemeValid := false
	for _, scheme := range validSchemes {
		if parsedURL.Scheme == scheme {
			schemeValid = true
			break
		}
	}

	if !schemeValid {
		return fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	// Block localhost and private IPs for security
	host := strings.ToLower(parsedURL.Hostname())
	blockedHosts := []string{"localhost", "127.0.0.1", "0.0.0.0", "::1"}
	for _, blocked := range blockedHosts {
		if host == blocked {
			return fmt.Errorf("blocked host: %s", host)
		}
	}

	// Check for private IP ranges
	if isPrivateIP(host) {
		return fmt.Errorf("private IP addresses not allowed: %s", host)
	}

	return nil
}

// isPrivateIP checks if a host is a private IP
func isPrivateIP(host string) bool {
	privatePatterns := []string{
		`^10\.`,                         // 10.0.0.0/8
		`^172\.(1[6-9]|2[0-9]|3[01])\.`, // 172.16.0.0/12
		`^192\.168\.`,                   // 192.168.0.0/16
		`^169\.254\.`,                   // 169.254.0.0/16 (link-local)
		`^fc00:`,                        // IPv6 private
		`^fe80:`,                        // IPv6 link-local
	}

	for _, pattern := range privatePatterns {
		if matched, _ := regexp.MatchString(pattern, host); matched {
			return true
		}
	}

	return false
}

// SanitizeFilename sanitizes a filename for safe storage
func SanitizeFilename(filename string) string {
	// Remove path components
	filename = filepath.Base(filename)

	// Replace unsafe characters
	unsafe := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	filename = unsafe.ReplaceAllString(filename, "_")

	// Limit length
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		base := filename[:255-len(ext)]
		filename = base + ext
	}

	// Ensure not empty
	if filename == "" || filename == "." {
		filename = "unnamed_file"
	}

	return filename
}

// ValidateFileSize validates file size limits
func ValidateFileSize(size int64, maxSize int64) error {
	if size <= 0 {
		return fmt.Errorf("invalid file size: %d", size)
	}

	if size > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", size, maxSize)
	}

	return nil
}
