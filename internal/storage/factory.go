package storage

import (
	"fmt"
	"strings"
)

func NewProvider(cfg Config) (Provider, error) {
	switch strings.ToLower(cfg.Provider) {
	case "s3", "aws":
		return NewS3Provider(cfg)
	case "gcs", "google":
		return NewGCSProvider(cfg)
	case "azure", "azblob":
		return NewAzureProvider(cfg)
	case "local", "filesystem":
		return NewLocalProvider(cfg)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", cfg.Provider)
	}
}
