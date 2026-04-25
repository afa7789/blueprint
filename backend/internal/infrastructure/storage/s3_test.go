package storage_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/internal/infrastructure/storage"
)

// fakeS3Client is a no-op S3Client used to satisfy NewS3StorageWithClient
// where the client is not exercised. Calling any method panics so misuse
// surfaces as a test failure.
type fakeS3Client struct{}

func (fakeS3Client) PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	panic("fakeS3Client.PutObject called")
}
func (fakeS3Client) GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	panic("fakeS3Client.GetObject called")
}
func (fakeS3Client) DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	panic("fakeS3Client.DeleteObject called")
}
func (fakeS3Client) HeadObject(context.Context, *s3.HeadObjectInput, ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	panic("fakeS3Client.HeadObject called")
}

type fakePresigner struct{}

func (fakePresigner) PresignGetObject(context.Context, *s3.GetObjectInput, ...func(*s3.PresignOptions)) (*storage.PresignedHTTPRequest, error) {
	panic("fakePresigner.PresignGetObject called")
}

// recordingClient buffers PutObject inputs for assertion.
type recordingClient struct {
	fakeS3Client
	puts []*s3.PutObjectInput
}

func (r *recordingClient) PutObject(_ context.Context, in *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	r.puts = append(r.puts, in)
	return &s3.PutObjectOutput{}, nil
}

func mustPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic containing %q, got none", want)
		}
		msg, _ := r.(string)
		if !strings.Contains(msg, want) {
			t.Fatalf("expected panic containing %q, got %v", want, r)
		}
	}()
	fn()
}

func TestNewS3StorageWithClient_PanicsOnEmptyBucket(t *testing.T) {
	mustPanic(t, "bucket is required", func() {
		storage.NewS3StorageWithClient("", fakeS3Client{}, fakePresigner{}, 0)
	})
}

func TestNewS3StorageWithClient_PanicsOnNilClient(t *testing.T) {
	mustPanic(t, "client is required", func() {
		storage.NewS3StorageWithClient("bucket", nil, fakePresigner{}, 0)
	})
}

func TestNewS3StorageWithClient_PanicsOnNilPresigner(t *testing.T) {
	mustPanic(t, "presigner is required", func() {
		storage.NewS3StorageWithClient("bucket", fakeS3Client{}, nil, 0)
	})
}

// TestS3_Upload_RejectsInvalidKeyBeforeReadingBody guards against a
// regression where a traversal key would still cause io.ReadAll on the
// caller-supplied body. Using a recording client confirms PutObject is
// never invoked when the key is invalid.
func TestS3_Upload_RejectsInvalidKeyBeforeReadingBody(t *testing.T) {
	rc := &recordingClient{}
	s := storage.NewS3StorageWithClient("bucket", rc, fakePresigner{}, 0)

	_, err := s.Upload(context.Background(), "../escape", bytes.NewReader([]byte("x")), "")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if len(rc.puts) != 0 {
		t.Fatalf("PutObject should not have been called for invalid key")
	}
}
