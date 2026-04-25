package storage

import (
	"context"
	"fmt"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/pkg/config"
)

// NewFromConfig constructs a domain.Storage implementation from config.
// Reads cfg.StorageBackend:
//
//   - "s3"              → S3Storage (AWS, R2, MinIO via aws-sdk-go-v2)
//   - "local" (default) → LocalStorage at cfg.StorageLocalPath
//
// When StorageBackend="s3" but StorageS3Bucket is empty, the factory
// returns an error rather than silently falling back — fail-loud is
// preferred to writing files to disk when the operator asked for S3.
func NewFromConfig(ctx context.Context, cfg *config.Config) (domain.Storage, error) {
	backend := cfg.StorageBackend
	if backend == "" {
		backend = "local"
	}
	switch backend {
	case "s3":
		if cfg.StorageS3Bucket == "" {
			return nil, fmt.Errorf("storage: STORAGE_BACKEND=s3 but STORAGE_S3_BUCKET is empty: %w", domain.ErrInvalidInput)
		}
		return NewS3Storage(ctx, S3Config{
			Bucket:          cfg.StorageS3Bucket,
			Region:          cfg.StorageS3Region,
			AccessKeyID:     cfg.StorageS3AccessKeyID,
			SecretAccessKey: cfg.StorageS3SecretAccessKey,
			Endpoint:        cfg.StorageS3Endpoint,
			UsePathStyle:    cfg.StorageS3UsePathStyle,
		})
	case "local":
		root := cfg.StorageLocalPath
		if root == "" {
			root = cfg.UploadDir
		}
		if root == "" {
			root = "./uploads"
		}
		return NewLocalStorage(root, cfg.StorageURLPrefix), nil
	default:
		return nil, fmt.Errorf("storage: unknown STORAGE_BACKEND %q (use \"local\" or \"s3\")", backend)
	}
}
