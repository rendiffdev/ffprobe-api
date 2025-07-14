package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/ffprobe-api/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalProvider(t *testing.T) {
	cfg := storage.Config{
		Provider: "local",
		Bucket:   "./test_storage",
	}

	provider, err := storage.NewProvider(cfg)
	require.NoError(t, err)
	assert.NotNil(t, provider)

	ctx := context.Background()
	testKey := "test/file.txt"
	testContent := "Hello, World!"

	t.Run("Upload file", func(t *testing.T) {
		reader := strings.NewReader(testContent)
		err := provider.Upload(ctx, testKey, reader, int64(len(testContent)))
		assert.NoError(t, err)
	})

	t.Run("File exists", func(t *testing.T) {
		exists, err := provider.Exists(ctx, testKey)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Download file", func(t *testing.T) {
		reader, err := provider.Download(ctx, testKey)
		assert.NoError(t, err)
		defer reader.Close()

		content := make([]byte, len(testContent))
		n, err := reader.Read(content)
		assert.NoError(t, err)
		assert.Equal(t, len(testContent), n)
		assert.Equal(t, testContent, string(content))
	})

	t.Run("Get URL", func(t *testing.T) {
		url, err := provider.GetURL(ctx, testKey)
		assert.NoError(t, err)
		assert.Contains(t, url, testKey)
	})

	t.Run("Get signed URL", func(t *testing.T) {
		url, err := provider.GetSignedURL(ctx, testKey, 3600)
		assert.NoError(t, err)
		assert.Contains(t, url, testKey)
	})

	t.Run("Delete file", func(t *testing.T) {
		err := provider.Delete(ctx, testKey)
		assert.NoError(t, err)

		exists, err := provider.Exists(ctx, testKey)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestStorageFactory(t *testing.T) {
	testCases := []struct {
		name        string
		provider    string
		expectError bool
	}{
		{"Local provider", "local", false},
		{"Filesystem provider", "filesystem", false},
		{"S3 provider", "s3", false},
		{"AWS provider", "aws", false},
		{"GCS provider", "gcs", false},
		{"Google provider", "google", false},
		{"Azure provider", "azure", false},
		{"AzBlob provider", "azblob", false},
		{"Unknown provider", "unknown", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := storage.Config{
				Provider: tc.provider,
				Bucket:   "test-bucket",
			}

			provider, err := storage.NewProvider(cfg)
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				if tc.provider == "local" || tc.provider == "filesystem" {
					assert.NoError(t, err)
					assert.NotNil(t, provider)
				} else {
					assert.NotNil(t, provider)
				}
			}
		})
	}
}