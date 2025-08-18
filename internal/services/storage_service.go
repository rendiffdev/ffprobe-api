package services

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/rendiffdev/ffprobe-api/internal/storage"
	"github.com/rs/zerolog"
)

type StorageService struct {
	provider storage.Provider
	logger   zerolog.Logger
}

func NewStorageService(provider storage.Provider, logger zerolog.Logger) *StorageService {
	return &StorageService{
		provider: provider,
		logger:   logger.With().Str("service", "storage").Logger(),
	}
}

func (s *StorageService) UploadFile(ctx context.Context, key string, reader io.Reader, size int64) error {
	s.logger.Info().
		Str("key", key).
		Int64("size", size).
		Msg("Uploading file to storage")

	if err := s.provider.Upload(ctx, key, reader, size); err != nil {
		s.logger.Error().
			Err(err).
			Str("key", key).
			Msg("Failed to upload file")
		return fmt.Errorf("failed to upload file: %w", err)
	}

	s.logger.Info().
		Str("key", key).
		Msg("File uploaded successfully")
	return nil
}

func (s *StorageService) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	s.logger.Info().
		Str("key", key).
		Msg("Downloading file from storage")

	reader, err := s.provider.Download(ctx, key)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("key", key).
			Msg("Failed to download file")
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return reader, nil
}

func (s *StorageService) DeleteFile(ctx context.Context, key string) error {
	s.logger.Info().
		Str("key", key).
		Msg("Deleting file from storage")

	if err := s.provider.Delete(ctx, key); err != nil {
		s.logger.Error().
			Err(err).
			Str("key", key).
			Msg("Failed to delete file")
		return fmt.Errorf("failed to delete file: %w", err)
	}

	s.logger.Info().
		Str("key", key).
		Msg("File deleted successfully")
	return nil
}

func (s *StorageService) FileExists(ctx context.Context, key string) (bool, error) {
	exists, err := s.provider.Exists(ctx, key)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("key", key).
			Msg("Failed to check if file exists")
		return false, fmt.Errorf("failed to check if file exists: %w", err)
	}

	return exists, nil
}

func (s *StorageService) GetFileURL(ctx context.Context, key string) (string, error) {
	url, err := s.provider.GetURL(ctx, key)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("key", key).
			Msg("Failed to get file URL")
		return "", fmt.Errorf("failed to get file URL: %w", err)
	}

	return url, nil
}

func (s *StorageService) GetSignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	s.logger.Info().
		Str("key", key).
		Dur("expiration", expiration).
		Msg("Generating signed URL")

	url, err := s.provider.GetSignedURL(ctx, key, int64(expiration.Seconds()))
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("key", key).
			Msg("Failed to generate signed URL")
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}