package storage_test

import (
	"context"
	"testing"

	"github.com/afa/blueprint/backend/internal/infrastructure/storage"
)

// TestNewS3Storage_CustomEndpoint verifies that custom endpoints
// (R2, MinIO, B2, etc.) are accepted by the constructor and do not
// fail boot. Does not validate live connectivity — that requires
// real credentials and a reachable service.
func TestNewS3Storage_CustomEndpoint(t *testing.T) {
	ctx := context.Background()

	// MinIO-style: path-style endpoint local
	if _, err := storage.NewS3Storage(ctx, storage.S3Config{
		Bucket:          "test-bucket",
		Region:          "us-east-1",
		AccessKeyID:     "minio",
		SecretAccessKey: "minio123",
		Endpoint:        "http://localhost:9000",
		UsePathStyle:    true,
	}); err != nil {
		t.Fatalf("MinIO config: %v", err)
	}

	// Cloudflare R2: virtual-hosted endpoint
	if _, err := storage.NewS3Storage(ctx, storage.S3Config{
		Bucket:          "test-bucket",
		Region:          "auto",
		AccessKeyID:     "r2-key",
		SecretAccessKey: "r2-secret",
		Endpoint:        "https://abc123.r2.cloudflarestorage.com",
	}); err != nil {
		t.Fatalf("R2 config: %v", err)
	}

	// Backblaze B2
	if _, err := storage.NewS3Storage(ctx, storage.S3Config{
		Bucket:          "test-bucket",
		Region:          "us-west-002",
		AccessKeyID:     "b2-key",
		SecretAccessKey: "b2-secret",
		Endpoint:        "https://s3.us-west-002.backblazeb2.com",
	}); err != nil {
		t.Fatalf("B2 config: %v", err)
	}

	// DigitalOcean Spaces
	if _, err := storage.NewS3Storage(ctx, storage.S3Config{
		Bucket:          "test-bucket",
		Region:          "nyc3",
		AccessKeyID:     "do-key",
		SecretAccessKey: "do-secret",
		Endpoint:        "https://nyc3.digitaloceanspaces.com",
	}); err != nil {
		t.Fatalf("DO Spaces config: %v", err)
	}

	// AWS default: no endpoint
	if _, err := storage.NewS3Storage(ctx, storage.S3Config{
		Bucket:          "test-bucket",
		Region:          "us-east-1",
		AccessKeyID:     "aws-key",
		SecretAccessKey: "aws-secret",
	}); err != nil {
		t.Fatalf("AWS default config: %v", err)
	}
}

// TestNewS3Storage_MissingBucket guarantees fail-fast.
func TestNewS3Storage_MissingBucket(t *testing.T) {
	_, err := storage.NewS3Storage(context.Background(), storage.S3Config{
		Region:   "us-east-1",
		Endpoint: "http://localhost:9000",
	})
	if err == nil {
		t.Fatal("expected error when Bucket is empty")
	}
}
