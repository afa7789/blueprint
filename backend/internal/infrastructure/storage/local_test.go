package storage_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
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

// TestLocalStorage_SymlinkEscapeRejected proves the EvalSymlinks-based
// containment check rejects keys whose resolved path lands outside root,
// even when the textual key passes validateKey. We plant a symlink under
// root pointing at /tmp and try to write through it.
func TestLocalStorage_SymlinkEscapeRejected(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks under Windows require elevated privileges")
	}
	root := t.TempDir()
	outside := t.TempDir() // a directory outside root

	// Plant: <root>/escape -> <outside>
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	s := storage.NewLocalStorage(root, "/static")
	ctx := context.Background()

	// All three operations must reject the escape — Upload requires the
	// containment check before CreateTemp, and Download/Exists/Delete
	// require it before touching the resolved file.
	if _, err := s.Upload(ctx, "escape/foo.txt", strings.NewReader("x"), ""); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("upload escape: expected ErrInvalidInput, got %v", err)
	}
	if _, err := s.Download(ctx, "escape/foo.txt"); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("download escape: expected ErrInvalidInput, got %v", err)
	}
	if _, err := s.Exists(ctx, "escape/foo.txt"); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("exists escape: expected ErrInvalidInput, got %v", err)
	}
	if err := s.Delete(ctx, "escape/foo.txt"); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("delete escape: expected ErrInvalidInput, got %v", err)
	}

	// Sanity: nothing was written to the outside dir.
	entries, _ := os.ReadDir(outside)
	if len(entries) != 0 {
		t.Fatalf("outside dir should be empty, got %d entries", len(entries))
	}
}

// TestLocalStorage_FileMode0644 verifies the post-rename file mode is
// 0o644 — CreateTemp uses 0o600 and the chmod step is what makes uploads
// readable by a separate static-server user.
func TestLocalStorage_FileMode0644(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes do not apply on Windows")
	}
	dir := t.TempDir()
	s := storage.NewLocalStorage(dir, "/static")
	if _, err := s.Upload(context.Background(), "f.txt", strings.NewReader("x"), ""); err != nil {
		t.Fatalf("upload: %v", err)
	}
	st, err := os.Stat(filepath.Join(dir, "f.txt"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if got := st.Mode().Perm(); got != 0o644 {
		t.Fatalf("file mode = %o, want 0644", got)
	}
}
