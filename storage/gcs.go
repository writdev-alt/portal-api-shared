package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// GCSClient handles Google Cloud Storage operations
type GCSClient struct {
	client     *storage.Client
	bucketName string
	basePath   string
}

// NewGCSClient creates a new GCS client
func NewGCSClient() (*GCSClient, error) {
	return NewGCSClientWithBasePath("")
}

// NewGCSClientWithBasePath creates a new GCS client with custom base path
func NewGCSClientWithBasePath(basePath string) (*GCSClient, error) {
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("GCS_BUCKET_NAME environment variable is required")
	}

	if basePath == "" {
		basePath = os.Getenv("GCS_BASE_PATH")
		if basePath == "" {
			basePath = "webhooks"
		}
	}

	ctx := context.Background()

	// Initialize GCS client
	var client *storage.Client
	var err error

	credentialsFile := os.Getenv("GCS_CREDENTIALS_FILE")
	if credentialsFile != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	} else {
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSClient{
		client:     client,
		bucketName: bucketName,
		basePath:   basePath,
	}, nil
}

// GetBucketName returns the configured bucket name
func (g *GCSClient) GetBucketName() string {
	return g.bucketName
}

// GetBasePath returns the configured base path
func (g *GCSClient) GetBasePath() string {
	return g.basePath
}

// UploadFile uploads a file to GCS
// Returns the GCS object path (gs://bucket/path) and public URL
func (g *GCSClient) UploadFile(ctx context.Context, objectPath string, data []byte, contentType string) (string, string, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)

	writer := obj.NewWriter(ctx)
	if contentType != "" {
		writer.ContentType = contentType
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return "", "", fmt.Errorf("failed to write to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close GCS writer: %w", err)
	}

	gcsPath := fmt.Sprintf("gs://%s/%s", g.bucketName, objectPath)
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucketName, objectPath)

	return gcsPath, publicURL, nil
}

// UploadFileWithMetadata uploads a file with metadata to GCS
func (g *GCSClient) UploadFileWithMetadata(ctx context.Context, objectPath string, data []byte, contentType string, metadata map[string]string) (string, string, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)

	writer := obj.NewWriter(ctx)
	if contentType != "" {
		writer.ContentType = contentType
	}
	if metadata != nil {
		writer.Metadata = metadata
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return "", "", fmt.Errorf("failed to write to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close GCS writer: %w", err)
	}

	gcsPath := fmt.Sprintf("gs://%s/%s", g.bucketName, objectPath)
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucketName, objectPath)

	return gcsPath, publicURL, nil
}

// ReadFile reads a file from GCS
func (g *GCSClient) ReadFile(ctx context.Context, objectPath string) ([]byte, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS reader: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from GCS: %w", err)
	}

	return data, nil
}

// ReadFileAsReader returns a reader for a file from GCS
func (g *GCSClient) ReadFileAsReader(ctx context.Context, objectPath string) (io.ReadCloser, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS reader: %w", err)
	}

	return reader, nil
}

// DeleteFile deletes a file from GCS
func (g *GCSClient) DeleteFile(ctx context.Context, objectPath string) error {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)

	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete from GCS: %w", err)
	}

	return nil
}

// FileExists checks if a file exists in GCS
func (g *GCSClient) FileExists(ctx context.Context, objectPath string) (bool, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)

	_, err := obj.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// ListFiles lists files in GCS with the given prefix
func (g *GCSClient) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	bucket := g.client.Bucket(g.bucketName)
	query := &storage.Query{
		Prefix: prefix,
	}

	var objectNames []string
	it := bucket.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == storage.ErrObjectNotExist || err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}
		objectNames = append(objectNames, attrs.Name)
	}

	return objectNames, nil
}

// Close closes the GCS client
func (g *GCSClient) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

// --- Webhook-specific helpers ---

// SaveWebhookJSON saves webhook JSON payload to GCS
// Returns the GCS object path (gs://bucket/path) and public URL
func (g *GCSClient) SaveWebhookJSON(ctx context.Context, provider, transactionType, trxID string, payload interface{}) (string, string, error) {
	now := time.Now()
	objectPath := g.generateWebhookPath(provider, transactionType, trxID, now)

	// Convert payload to JSON
	jsonData, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	metadata := map[string]string{
		"provider":         provider,
		"transaction_type": transactionType,
		"trx_id":           trxID,
		"uploaded_at":      now.Format(time.RFC3339),
	}

	return g.UploadFileWithMetadata(ctx, objectPath, jsonData, "application/json", metadata)
}

// SaveWebhookJSONFromBytes saves webhook JSON from raw bytes to GCS
func (g *GCSClient) SaveWebhookJSONFromBytes(ctx context.Context, provider, transactionType, trxID string, jsonBytes []byte) (string, string, error) {
	now := time.Now()
	objectPath := g.generateWebhookPath(provider, transactionType, trxID, now)

	metadata := map[string]string{
		"provider":         provider,
		"transaction_type": transactionType,
		"trx_id":           trxID,
		"uploaded_at":      now.Format(time.RFC3339),
	}

	return g.UploadFileWithMetadata(ctx, objectPath, jsonBytes, "application/json", metadata)
}

// ReadWebhookJSON reads webhook JSON from GCS (alias for ReadFile)
func (g *GCSClient) ReadWebhookJSON(ctx context.Context, objectPath string) ([]byte, error) {
	return g.ReadFile(ctx, objectPath)
}

// DeleteWebhookJSON deletes webhook JSON from GCS (alias for DeleteFile)
func (g *GCSClient) DeleteWebhookJSON(ctx context.Context, objectPath string) error {
	return g.DeleteFile(ctx, objectPath)
}

// --- Avatar-specific helpers ---

// UploadAvatar uploads an avatar image to GCS
// Returns the GCS object path and public URL
func (g *GCSClient) UploadAvatar(ctx context.Context, entityType string, entityID uint64, fileData []byte, contentType, extension string) (string, string, error) {
	now := time.Now()
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	timestamp := now.Format("20060102-150405")

	filename := fmt.Sprintf("%d-%s%s", entityID, timestamp, extension)
	objectPath := filepath.Join("avatars", entityType, datePath, filename)

	return g.UploadFile(ctx, objectPath, fileData, contentType)
}

// --- Internal helpers ---

func (g *GCSClient) generateWebhookPath(provider, transactionType, trxID string, now time.Time) string {
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	timestamp := now.Format("20060102-150405")

	var filename string
	if trxID != "" {
		filename = fmt.Sprintf("%s-%s.json", trxID, timestamp)
	} else {
		filename = fmt.Sprintf("%s.json", timestamp)
	}

	return filepath.Join(g.basePath, provider, transactionType, datePath, filename)
}
