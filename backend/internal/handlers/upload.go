package handlers

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/google/uuid"
)

// UploadFormFile persists a multipart upload via the configured storage
// backend. The returned URL is suitable for direct use in APIs (e.g.
// /static/covers/<uuid>.png for local, or a presigned GET URL for S3).
func UploadFormFile(ctx context.Context, storage domain.Storage, file *multipart.FileHeader, prefix string) (string, error) {
	if storage == nil {
		return "", fmt.Errorf("upload: storage is nil")
	}
	ext := filepath.Ext(file.Filename)
	key := fmt.Sprintf("%s/%s%s", prefix, uuid.NewString(), ext)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = src.Close() }()

	contentType := file.Header.Get("Content-Type")
	return storage.Upload(ctx, key, src, contentType)
}
