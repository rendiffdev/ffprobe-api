package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

type AzureProvider struct {
	client    *azblob.Client
	container string
	account   string
}

func NewAzureProvider(cfg Config) (*AzureProvider, error) {
	credential, err := azblob.NewSharedKeyCredential(cfg.AccessKey, cfg.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credentials: %w", err)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", cfg.AccessKey)
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure blob client: %w", err)
	}

	return &AzureProvider{
		client:    client,
		container: cfg.Bucket,
		account:   cfg.AccessKey,
	}, nil
}

func (a *AzureProvider) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	_, err := a.client.UploadStream(ctx, a.container, key, reader, nil)
	if err != nil {
		return fmt.Errorf("failed to upload to Azure: %w", err)
	}
	return nil
}

func (a *AzureProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	response, err := a.client.DownloadStream(ctx, a.container, key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download from Azure: %w", err)
	}
	return response.Body, nil
}

func (a *AzureProvider) Delete(ctx context.Context, key string) error {
	_, err := a.client.DeleteBlob(ctx, a.container, key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete from Azure: %w", err)
	}
	return nil
}

func (a *AzureProvider) Exists(ctx context.Context, key string) (bool, error) {
	_, err := a.client.NewBlobClient(a.container, key).GetProperties(ctx, nil)
	if err != nil {
		var storageError *azblob.StorageError
		if errors.As(err, &storageError) && storageError.ErrorCode == azblob.StorageErrorCodeBlobNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if blob exists in Azure: %w", err)
	}
	return true, nil
}

func (a *AzureProvider) GetURL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", a.account, a.container, key), nil
}

func (a *AzureProvider) GetSignedURL(ctx context.Context, key string, expiration int64) (string, error) {
	credential, err := azblob.NewSharedKeyCredential(a.account, "")
	if err != nil {
		return "", fmt.Errorf("failed to create credential for signed URL: %w", err)
	}

	sasQueryParams, err := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		ExpiryTime:    time.Now().UTC().Add(time.Duration(expiration) * time.Second),
		Permissions:   to.Ptr(sas.BlobPermissions{Read: true}).String(),
		ContainerName: a.container,
		BlobName:      key,
	}.SignWithSharedKey(credential)
	if err != nil {
		return "", fmt.Errorf("failed to sign URL: %w", err)
	}

	sasURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s?%s", a.account, a.container, key, sasQueryParams.Encode())
	return sasURL, nil
}