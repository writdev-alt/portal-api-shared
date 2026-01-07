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
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("GCS_BUCKET_NAME environment variable is required")
	}

	basePath := os.Getenv("GCS_BASE_PATH")
	if basePath == "" {
		basePath = "webhooks"
	}

	ctx := context.Background()

	// Initialize GCS client
	// If GCS_CREDENTIALS_FILE is set, use it; otherwise use default credentials
	var client *storage.Client
	var err error

	credentialsFile := os.Getenv("GCS_CREDENTIALS_FILE")
	if credentialsFile != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	} else {
		// Use default credentials (e.g., from service account or gcloud auth)
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

// SaveWebhookJSON saves webhook JSON payload to GCS
// Returns the GCS object path (gs://bucket/path) and public URL if available
func (g *GCSClient) SaveWebhookJSON(ctx context.Context, provider, transactionType, trxID string, payload interface{}) (string, string, error) {
	// Generate file path: webhooks/{provider}/{transactionType}/{year}/{month}/{day}/{trxID}-{timestamp}.json
	now := time.Now()
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	timestamp := now.Format("20060102-150405")

	// Generate filename
	var filename string
	if trxID != "" {
		filename = fmt.Sprintf("%s-%s.json", trxID, timestamp)
	} else {
		filename = fmt.Sprintf("%s.json", timestamp)
	}

	objectPath := filepath.Join(g.basePath, provider, transactionType, datePath, filename)

	// Convert payload to JSON
	jsonData, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Upload to GCS
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)

	// Set object metadata
	writer := obj.NewWriter(ctx)
	writer.ContentType = "application/json"
	writer.Metadata = map[string]string{
		"provider":         provider,
		"transaction_type": transactionType,
		"trx_id":           trxID,
		"uploaded_at":      now.Format(time.RFC3339),
	}

	// Write JSON data
	if _, err := writer.Write(jsonData); err != nil {
		writer.Close()
		return "", "", fmt.Errorf("failed to write to GCS: %w", err)
	}

	// Close writer to finalize upload
	if err := writer.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close GCS writer: %w", err)
	}

	// Generate GCS path and public URL
	gcsPath := fmt.Sprintf("gs://%s/%s", g.bucketName, objectPath)

	// Generate public URL (if bucket is public)
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucketName, objectPath)

	return gcsPath, publicURL, nil
}

// SaveWebhookJSONFromBytes saves webhook JSON from raw bytes to GCS
func (g *GCSClient) SaveWebhookJSONFromBytes(ctx context.Context, provider, transactionType, trxID string, jsonBytes []byte) (string, string, error) {
	// Generate file path
	now := time.Now()
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	timestamp := now.Format("20060102-150405")

	var filename string
	if trxID != "" {
		filename = fmt.Sprintf("%s-%s.json", trxID, timestamp)
	} else {
		filename = fmt.Sprintf("%s.json", timestamp)
	}

	objectPath := filepath.Join(g.basePath, provider, transactionType, datePath, filename)

	// Upload to GCS
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)

	writer := obj.NewWriter(ctx)
	writer.ContentType = "application/json"
	writer.Metadata = map[string]string{
		"provider":         provider,
		"transaction_type": transactionType,
		"trx_id":           trxID,
		"uploaded_at":      now.Format(time.RFC3339),
	}

	if _, err := writer.Write(jsonBytes); err != nil {
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

// ReadWebhookJSON reads webhook JSON from GCS
func (g *GCSClient) ReadWebhookJSON(ctx context.Context, objectPath string) ([]byte, error) {
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

// DeleteWebhookJSON deletes webhook JSON from GCS
func (g *GCSClient) DeleteWebhookJSON(ctx context.Context, objectPath string) error {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)

	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete from GCS: %w", err)
	}

	return nil
}

// Close closes the GCS client
func (g *GCSClient) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}
