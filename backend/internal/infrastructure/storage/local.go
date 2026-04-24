// Package storage provides concrete domain.Storage implementations for
// persisting uploaded files. LocalStorage writes to a configurable
// filesystem root; S3Storage writes to an S3-compatible bucket with
// presigned GET URLs.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/afa/blueprint/backend/internal/domain"
)

// LocalStorage is a domain.Storage implementation that persists objects
// under a filesystem root. Writes are atomic (write-to-tmp + rename)
// and directory-traversal attempts are rejected. The URL returned from
// Upload is a relative path "/static/{key}" — the server mounts a static
// file handler at that mount point.
type LocalStorage struct {
	root      string
	urlPrefix string
}

// NewLocalStorage constructs a LocalStorage rooted at the given path.
// The directory is created (mkdir -p) on first use. urlPrefix is the
// URL path prefix returned by Upload/SignedURL (e.g. "/static").
func NewLocalStorage(root, urlPrefix string) *LocalStorage {
	if urlPrefix == "" {
		urlPrefix = "/static"
	}
	return &LocalStorage{root: root, urlPrefix: strings.TrimRight(urlPrefix, "/")}
}

// Root returns the filesystem root this LocalStorage writes to.
func (l *LocalStorage) Root() string { return l.root }

// validateKey rejects empty keys, absolute paths, and traversal attempts.
func validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("storage: empty key: %w", domain.ErrInvalidInput)
	}
	cleaned := filepath.ToSlash(filepath.Clean(key))
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") ||
		strings.Contains(cleaned, "/../") || strings.HasSuffix(cleaned, "/..") {
		return fmt.Errorf("storage: traversal key %q: %w", key, domain.ErrInvalidInput)
	}
	for _, seg := range strings.Split(strings.ReplaceAll(key, "\\", "/"), "/") {
		if seg == ".." {
			return fmt.Errorf("storage: traversal key %q: %w", key, domain.ErrInvalidInput)
		}
	}
	if filepath.IsAbs(key) {
		return fmt.Errorf("storage: absolute key %q: %w", key, domain.ErrInvalidInput)
	}
	return nil
}

func (l *LocalStorage) resolve(key string) (string, error) {
	if err := validateKey(key); err != nil {
		return "", err
	}
	return filepath.Join(l.root, filepath.FromSlash(key)), nil
}

func (l *LocalStorage) publicURL(key string) string {
	return l.urlPrefix + "/" + strings.TrimPrefix(filepath.ToSlash(key), "/")
}

// Upload writes the body to {root}/{key} via a write-to-tmp + rename
// atomic swap. Parent directories are created on demand.
func (l *LocalStorage) Upload(ctx context.Context, key string, r io.Reader, _ string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	dst, err := l.resolve(key)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", fmt.Errorf("storage: mkdir: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".upload-*.tmp")
	if err != nil {
		return "", fmt.Errorf("storage: create tmp: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }
	if _, err := io.Copy(tmp, r); err != nil {
		_ = tmp.Close()
		cleanup()
		return "", fmt.Errorf("storage: copy body: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanup()
		return "", fmt.Errorf("storage: fsync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return "", fmt.Errorf("storage: close tmp: %w", err)
	}
	if err := os.Rename(tmpName, dst); err != nil {
		cleanup()
		return "", fmt.Errorf("storage: rename: %w", err)
	}
	return l.publicURL(key), nil
}

// Download opens {root}/{key} for reading. Returns domain.ErrNotFound
// if the object is absent.
func (l *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	src, err := l.resolve(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("storage: open: %w", err)
	}
	return f, nil
}

// Exists reports whether {root}/{key} is a regular file.
func (l *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	p, err := l.resolve(key)
	if err != nil {
		return false, err
	}
	st, err := os.Stat(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("storage: stat: %w", err)
	}
	return st.Mode().IsRegular(), nil
}

// SignedURL returns the same public URL as Upload. Local storage has
// no cryptographic signing — the TTL is advisory and ignored.
func (l *LocalStorage) SignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	if err := validateKey(key); err != nil {
		return "", err
	}
	return l.publicURL(key), nil
}

// Delete removes {root}/{key}. Missing keys are not an error.
func (l *LocalStorage) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	p, err := l.resolve(key)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("storage: delete: %w", err)
	}
	return nil
}

var _ domain.Storage = (*LocalStorage)(nil)
