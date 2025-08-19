package storage

import (
	"context"
	"io"
)

type Provider interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	GetURL(ctx context.Context, key string) (string, error)
	GetSignedURL(ctx context.Context, key string, expiration int64) (string, error)
}

type UploadOptions struct {
	ContentType     string
	CacheControl    string
	ContentEncoding string
	Metadata        map[string]string
}

type DownloadOptions struct {
	Range string
}

type Config struct {
	Provider  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	Endpoint  string
	UseSSL    bool
	BaseURL   string
}
