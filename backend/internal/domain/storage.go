package domain

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
)

// Storage persists and retrieves binary objects (uploaded files, receipts,
// covers, etc.). Production impls: infrastructure/storage/local and
// infrastructure/storage/s3.
//
// URL contract — read carefully before persisting Upload's return value:
//
//   - LocalStorage.Upload returns a stable, non-expiring relative path
//     (e.g. "/static/covers/xyz.png") served by the app's static handler.
//     Safe to persist in the database.
//
//   - S3Storage.Upload returns a presigned GET URL valid for 24h. Storing
//     it directly will produce broken links after expiry. Callers that
//     need durable links should persist the key and call SignedURL on
//     read, or proxy reads through the backend.
//
// SignedURL is always a time-limited URL. For LocalStorage it returns the
// same stable path as Upload (TTL is advisory and ignored). For S3Storage
// it returns a fresh presigned URL valid for ttl.
type Storage interface {
	Upload(ctx context.Context, key string, r io.Reader, contentType string) (url string, err error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Exists(ctx context.Context, key string) (bool, error)
	SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}
