package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalProvider struct {
	basePath string
	baseURL  string
}

func NewLocalProvider(cfg Config) (*LocalProvider, error) {
	basePath := cfg.Bucket
	if basePath == "" {
		basePath = "./storage"
	}

	// Get absolute path for base directory
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base path: %w", err)
	}

	if err := os.MkdirAll(absBasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalProvider{
		basePath: absBasePath,
		baseURL:  cfg.BaseURL,
	}, nil
}

// securePath validates and returns a safe file path within the base directory.
// Returns an error if the path would escape the base directory.
func (l *LocalProvider) securePath(key string) (string, error) {
	// Clean the key to remove any .. or other traversal attempts
	cleanKey := filepath.Clean(key)

	// Join with base path
	filePath := filepath.Join(l.basePath, cleanKey)

	// Get absolute path to resolve any remaining traversal
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// Verify the resolved path is within the base directory
	// Use filepath.Clean on basePath to ensure consistent comparison
	if !strings.HasPrefix(absPath, l.basePath+string(filepath.Separator)) && absPath != l.basePath {
		return "", fmt.Errorf("path traversal detected: access denied")
	}

	// Check for symlinks that might escape the base directory
	// Only check if the path exists (for read operations)
	if _, err := os.Lstat(absPath); err == nil {
		realPath, err := filepath.EvalSymlinks(absPath)
		if err == nil {
			if !strings.HasPrefix(realPath, l.basePath+string(filepath.Separator)) && realPath != l.basePath {
				return "", fmt.Errorf("symlink escape detected: access denied")
			}
		}
	}

	return absPath, nil
}

func (l *LocalProvider) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	filePath, err := l.securePath(key)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (l *LocalProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	filePath, err := l.securePath(key)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func (l *LocalProvider) Delete(ctx context.Context, key string) error {
	filePath, err := l.securePath(key)
	if err != nil {
		return err
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (l *LocalProvider) Exists(ctx context.Context, key string) (bool, error) {
	filePath, err := l.securePath(key)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if file exists: %w", err)
	}
	return true, nil
}

func (l *LocalProvider) GetURL(ctx context.Context, key string) (string, error) {
	if l.baseURL != "" {
		// Sanitize key for URL to prevent injection
		cleanKey := filepath.Clean(key)
		return fmt.Sprintf("%s/%s", l.baseURL, cleanKey), nil
	}

	filePath, err := l.securePath(key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file://%s", filePath), nil
}

func (l *LocalProvider) GetSignedURL(ctx context.Context, key string, expiration int64) (string, error) {
	return l.GetURL(ctx, key)
}
