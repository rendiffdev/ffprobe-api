package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalProvider{
		basePath: basePath,
		baseURL:  cfg.BaseURL,
	}, nil
}

func (l *LocalProvider) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	filePath := filepath.Join(l.basePath, key)

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
	filePath := filepath.Join(l.basePath, key)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func (l *LocalProvider) Delete(ctx context.Context, key string) error {
	filePath := filepath.Join(l.basePath, key)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (l *LocalProvider) Exists(ctx context.Context, key string) (bool, error) {
	filePath := filepath.Join(l.basePath, key)
	_, err := os.Stat(filePath)
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
		return fmt.Sprintf("%s/%s", l.baseURL, key), nil
	}
	return fmt.Sprintf("file://%s", filepath.Join(l.basePath, key)), nil
}

func (l *LocalProvider) GetSignedURL(ctx context.Context, key string, expiration int64) (string, error) {
	return l.GetURL(ctx, key)
}
