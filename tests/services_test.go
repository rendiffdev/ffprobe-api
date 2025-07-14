package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ffprobe-api/internal/services"
	"github.com/ffprobe-api/internal/storage"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageService(t *testing.T) {
	logger := zerolog.Nop()

	cfg := storage.Config{
		Provider: "local",
		Bucket:   "./test_storage_service",
	}

	provider, err := storage.NewProvider(cfg)
	require.NoError(t, err)

	storageService := services.NewStorageService(provider, logger)

	ctx := context.Background()
	testKey := "test/service-file.txt"
	testContent := "Hello from service test!"

	t.Run("Upload file via service", func(t *testing.T) {
		reader := strings.NewReader(testContent)
		err := storageService.UploadFile(ctx, testKey, reader, int64(len(testContent)))
		assert.NoError(t, err)
	})

	t.Run("Check file exists via service", func(t *testing.T) {
		exists, err := storageService.FileExists(ctx, testKey)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Download file via service", func(t *testing.T) {
		reader, err := storageService.DownloadFile(ctx, testKey)
		assert.NoError(t, err)
		defer reader.Close()

		content := make([]byte, len(testContent))
		n, err := reader.Read(content)
		assert.NoError(t, err)
		assert.Equal(t, len(testContent), n)
		assert.Equal(t, testContent, string(content))
	})

	t.Run("Get file URL via service", func(t *testing.T) {
		url, err := storageService.GetFileURL(ctx, testKey)
		assert.NoError(t, err)
		assert.Contains(t, url, testKey)
	})

	t.Run("Get signed URL via service", func(t *testing.T) {
		expiration := time.Hour
		url, err := storageService.GetSignedURL(ctx, testKey, expiration)
		assert.NoError(t, err)
		assert.Contains(t, url, testKey)
	})

	t.Run("Delete file via service", func(t *testing.T) {
		err := storageService.DeleteFile(ctx, testKey)
		assert.NoError(t, err)

		exists, err := storageService.FileExists(ctx, testKey)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestStorageServiceErrorHandling(t *testing.T) {
	logger := zerolog.Nop()

	cfg := storage.Config{
		Provider: "local",
		Bucket:   "./test_storage_service_error",
	}

	provider, err := storage.NewProvider(cfg)
	require.NoError(t, err)

	storageService := services.NewStorageService(provider, logger)

	ctx := context.Background()
	nonExistentKey := "non-existent/file.txt"

	t.Run("Download non-existent file", func(t *testing.T) {
		_, err := storageService.DownloadFile(ctx, nonExistentKey)
		assert.Error(t, err)
	})

	t.Run("Check non-existent file", func(t *testing.T) {
		exists, err := storageService.FileExists(ctx, nonExistentKey)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Delete non-existent file", func(t *testing.T) {
		err := storageService.DeleteFile(ctx, nonExistentKey)
		assert.Error(t, err)
	})

	t.Run("Get URL for non-existent file", func(t *testing.T) {
		url, err := storageService.GetFileURL(ctx, nonExistentKey)
		assert.NoError(t, err)
		assert.Contains(t, url, nonExistentKey)
	})
}