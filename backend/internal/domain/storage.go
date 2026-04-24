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
type Storage interface {
	Upload(ctx context.Context, key string, r io.Reader, contentType string) (url string, err error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Exists(ctx context.Context, key string) (bool, error)
	SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}
