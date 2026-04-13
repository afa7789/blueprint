package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/google/uuid"
)

func UploadFile(file *multipart.FileHeader, prefix string, cfg *config.Config) (string, error) {
	// For now, only local storage
	// TODO: Add S3 support when AWS SDK is integrated

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s/%s%s", prefix, uuid.NewString(), ext)
	destPath := filepath.Join(cfg.UploadDir, filename)

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return fmt.Sprintf("/static/%s", filename), nil
}
