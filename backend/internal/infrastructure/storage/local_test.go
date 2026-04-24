package storage_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/internal/infrastructure/storage"
)

func TestLocalStorage_UploadDownload(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewLocalStorage(dir, "/static")
	ctx := context.Background()

	url, err := s.Upload(ctx, "covers/x.png", bytes.NewReader([]byte("hello")), "image/png")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if url != "/static/covers/x.png" {
		t.Fatalf("unexpected url: %s", url)
	}

	// Persisted on disk
	if _, err := os.Stat(filepath.Join(dir, "covers", "x.png")); err != nil {
		t.Fatalf("file not persisted: %v", err)
	}

	rc, err := s.Download(ctx, "covers/x.png")
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	defer func() { _ = rc.Close() }()
	b, _ := io.ReadAll(rc)
	if string(b) != "hello" {
		t.Fatalf("round-trip mismatch: %s", b)
	}
}

func TestLocalStorage_TraversalRejected(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewLocalStorage(dir, "/static")
	ctx := context.Background()

	cases := []string{"../etc/passwd", "a/../../b", "/abs/path", ""}
	for _, k := range cases {
		if _, err := s.Upload(ctx, k, strings.NewReader("x"), ""); err == nil {
			t.Fatalf("expected rejection for key %q", k)
		} else if !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for %q, got %v", k, err)
		}
	}
}

func TestLocalStorage_DownloadMissing(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewLocalStorage(dir, "/static")
	_, err := s.Download(context.Background(), "missing.txt")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestLocalStorage_SurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	s1 := storage.NewLocalStorage(dir, "/static")
	if _, err := s1.Upload(ctx, "a/b.txt", strings.NewReader("persist"), ""); err != nil {
		t.Fatalf("upload: %v", err)
	}

	// Simulate restart: brand-new storage instance, same root
	s2 := storage.NewLocalStorage(dir, "/static")
	ok, err := s2.Exists(ctx, "a/b.txt")
	if err != nil || !ok {
		t.Fatalf("persistence check: ok=%v err=%v", ok, err)
	}
	rc, err := s2.Download(ctx, "a/b.txt")
	if err != nil {
		t.Fatalf("download after restart: %v", err)
	}
	defer func() { _ = rc.Close() }()
	b, _ := io.ReadAll(rc)
	if string(b) != "persist" {
		t.Fatalf("content lost across restart: %s", b)
	}
}

func TestLocalStorage_DeleteIdempotent(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewLocalStorage(dir, "/static")
	ctx := context.Background()
	if _, err := s.Upload(ctx, "k.txt", strings.NewReader("x"), ""); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, "k.txt"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := s.Delete(ctx, "k.txt"); err != nil {
		t.Fatalf("second delete should be idempotent: %v", err)
	}
}
