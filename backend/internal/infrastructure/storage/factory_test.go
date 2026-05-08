package storage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/internal/infrastructure/storage"
	"github.com/afa/blueprint/backend/pkg/config"
)

func TestFactory_LocalDefault(t *testing.T) {
	cfg := &config.Config{StorageBackend: "local", StorageLocalPath: t.TempDir(), StorageURLPrefix: "/static"}
	s, err := storage.NewFromConfig(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("nil storage")
	}
}

func TestFactory_S3MissingBucketFails(t *testing.T) {
	cfg := &config.Config{StorageBackend: "s3"}
	_, err := storage.NewFromConfig(context.Background(), cfg)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestFactory_UnknownBackend(t *testing.T) {
	cfg := &config.Config{StorageBackend: "floppy"}
	_, err := storage.NewFromConfig(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for unknown backend")
	}
}

func TestFactory_NormalizesBackend(t *testing.T) {
	for _, bk := range []string{"LOCAL", " Local ", "local"} {
		cfg := &config.Config{StorageBackend: bk, UploadDir: t.TempDir()}
		if _, err := storage.NewFromConfig(context.Background(), cfg); err != nil {
			t.Fatalf("backend=%q: %v", bk, err)
		}
	}
}

func TestFactory_EmptyDefaultsToLocal(t *testing.T) {
	cfg := &config.Config{StorageBackend: "", UploadDir: t.TempDir()}
	s, err := storage.NewFromConfig(context.Background(), cfg)
	if err != nil || s == nil {
		t.Fatalf("expected local default, got err=%v", err)
	}
}
