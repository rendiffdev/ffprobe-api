package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GCSProvider struct {
	client *storage.Client
	bucket string
}

func NewGCSProvider(cfg Config) (*GCSProvider, error) {
	var client *storage.Client
	var err error

	if cfg.AccessKey != "" {
		client, err = storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(cfg.AccessKey)))
	} else {
		client, err = storage.NewClient(context.Background())
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSProvider{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (g *GCSProvider) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	obj := g.client.Bucket(g.bucket).Object(key)
	writer := obj.NewWriter(ctx)
	defer writer.Close()

	if _, err := io.Copy(writer, reader); err != nil {
		return fmt.Errorf("failed to upload to GCS: %w", err)
	}

	return nil
}

func (g *GCSProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	obj := g.client.Bucket(g.bucket).Object(key)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to download from GCS: %w", err)
	}
	return reader, nil
}

func (g *GCSProvider) Delete(ctx context.Context, key string) error {
	obj := g.client.Bucket(g.bucket).Object(key)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete from GCS: %w", err)
	}
	return nil
}

func (g *GCSProvider) Exists(ctx context.Context, key string) (bool, error) {
	obj := g.client.Bucket(g.bucket).Object(key)
	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if object exists in GCS: %w", err)
	}
	return true, nil
}

func (g *GCSProvider) GetURL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucket, key), nil
}

func (g *GCSProvider) GetSignedURL(ctx context.Context, key string, expiration int64) (string, error) {

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(time.Duration(expiration) * time.Second),
	}

	url, err := g.client.Bucket(g.bucket).SignedURL(key, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}
