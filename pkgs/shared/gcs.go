package shared

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"cloud.google.com/go/storage"
)

var (
	storageClient *storage.Client
	clientOnce    sync.Once
	clientErr     error
)

func getStorageClient(ctx context.Context) (*storage.Client, error) {
	clientOnce.Do(func() {
		storageClient, clientErr = storage.NewClient(ctx)
	})
	if clientErr != nil {
		return nil, fmt.Errorf("storage.NewClient: %w", clientErr)
	}
	return storageClient, nil
}

func UploadToGCS(ctx context.Context, GCSBucketName string, objectName string, data []byte) error {
	client, err := getStorageClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get GCS client: %w", err)
	}

	uploadCtx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	obj := client.Bucket(GCSBucketName).Object(objectName)
	wc := obj.NewWriter(uploadCtx)
	if _, err := wc.Write(data); err != nil {
		wc.Close()
		return fmt.Errorf("failed to write to GCS: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}

	return nil
}

func GetFileFromGCS(ctx context.Context, GCSBucketName, objectName string) ([]byte, error) {
	client, err := getStorageClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GCS client: %w", err)
	}

	readCtx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	obj := client.Bucket(GCSBucketName).Object(objectName)
	rc, err := obj.NewReader(readCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader for object %s: %w", objectName, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read object content for %s: %w", objectName, err)
	}

	return data, nil
}
