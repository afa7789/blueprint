package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"github.com/afa/blueprint/backend/internal/domain"
)

// defaultPresignTTL is the TTL applied when Upload returns a presigned
// GET URL.
const defaultPresignTTL = 24 * time.Hour

// S3Client is the minimal subset of the aws-sdk-go-v2 s3.Client surface
// that S3Storage consumes. Narrowed so tests can inject a mock.
type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

// S3Presigner presigns GET requests.
type S3Presigner interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*PresignedHTTPRequest, error)
}

// PresignedHTTPRequest mirrors v4.PresignedHTTPRequest's URL-only slice.
type PresignedHTTPRequest struct {
	URL string
}

// S3Storage is a domain.Storage backed by an S3-compatible bucket
// (AWS S3, Cloudflare R2, MinIO).
type S3Storage struct {
	bucket    string
	client    S3Client
	presigner S3Presigner
	ttl       time.Duration
}

// S3Config bundles the fields required to construct an S3Storage.
// Endpoint is optional (fill for R2, MinIO, B2). UsePathStyle is
// required for MinIO; leave false for AWS S3; R2 works with either.
type S3Config struct {
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
	UsePathStyle    bool
}

// NewS3Storage constructs an S3Storage using aws-sdk-go-v2.
func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("storage: S3 bucket required: %w", domain.ErrInvalidInput)
	}
	loadOpts := []func(*awsconfig.LoadOptions) error{}
	if cfg.Region != "" {
		loadOpts = append(loadOpts, awsconfig.WithRegion(cfg.Region))
	}
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("storage: load aws config: %w", err)
	}
	s3Opts := func(o *s3.Options) {
		if cfg.Endpoint != "" {
			ep := cfg.Endpoint
			o.BaseEndpoint = &ep
		}
		if cfg.UsePathStyle {
			o.UsePathStyle = true
		}
	}
	client := s3.NewFromConfig(awsCfg, s3Opts)
	presignClient := s3.NewPresignClient(client, func(po *s3.PresignOptions) {
		po.ClientOptions = append(po.ClientOptions, s3Opts)
	})
	return &S3Storage{
		bucket:    cfg.Bucket,
		client:    client,
		presigner: presignerAdapter{p: presignClient},
		ttl:       defaultPresignTTL,
	}, nil
}

// NewS3StorageWithClient is the test-friendly constructor.
func NewS3StorageWithClient(bucket string, client S3Client, presigner S3Presigner, ttl time.Duration) *S3Storage {
	if ttl <= 0 {
		ttl = defaultPresignTTL
	}
	return &S3Storage{bucket: bucket, client: client, presigner: presigner, ttl: ttl}
}

func (s *S3Storage) Bucket() string { return s.bucket }
func (s *S3Storage) Client() S3Client { return s.client }

// Upload PUTs body under key and returns a presigned GET URL valid
// for s.ttl (24h by default).
func (s *S3Storage) Upload(ctx context.Context, key string, r io.Reader, contentType string) (string, error) {
	if err := validateKey(key); err != nil {
		return "", err
	}
	body, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("storage: read body: %w", err)
	}
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytesReadSeeker(body),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	if _, err := s.client.PutObject(ctx, input); err != nil {
		return "", fmt.Errorf("storage: s3 put: %w", err)
	}
	return s.SignedURL(ctx, key, s.ttl)
}

// Download returns a ReadCloser streaming the object body.
func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := validateKey(key); err != nil {
		return nil, err
	}
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isS3NotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("storage: s3 get: %w", err)
	}
	return out.Body, nil
}

// Exists returns true when HEAD succeeds, false on NotFound.
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	if err := validateKey(key); err != nil {
		return false, err
	}
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isS3NotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("storage: s3 head: %w", err)
	}
	return true, nil
}

// SignedURL returns a presigned GET URL for key valid for ttl.
func (s *S3Storage) SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if err := validateKey(key); err != nil {
		return "", err
	}
	if ttl <= 0 {
		ttl = s.ttl
	}
	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(o *s3.PresignOptions) { o.Expires = ttl })
	if err != nil {
		return "", fmt.Errorf("storage: presign: %w", err)
	}
	return req.URL, nil
}

// Delete removes key from the bucket. S3 DeleteObject is idempotent.
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("storage: s3 delete: %w", err)
	}
	return nil
}

func isS3NotFound(err error) bool {
	var nsk *s3types.NoSuchKey
	if errors.As(err, &nsk) {
		return true
	}
	var nf *s3types.NotFound
	if errors.As(err, &nf) {
		return true
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchKey", "NotFound", "404":
			return true
		}
	}
	return false
}

type presignerAdapter struct {
	p *s3.PresignClient
}

func (a presignerAdapter) PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*PresignedHTTPRequest, error) {
	req, err := a.p.PresignGetObject(ctx, params, optFns...)
	if err != nil {
		return nil, err
	}
	return &PresignedHTTPRequest{URL: req.URL}, nil
}

func bytesReadSeeker(b []byte) *byteReader { return &byteReader{b: b} }

type byteReader struct {
	b   []byte
	pos int64
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= int64(len(r.b)) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.pos:])
	r.pos += int64(n)
	return n, nil
}

func (r *byteReader) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = r.pos + offset
	case io.SeekEnd:
		abs = int64(len(r.b)) + offset
	default:
		return 0, fmt.Errorf("byteReader: invalid whence %d", whence)
	}
	if abs < 0 {
		return 0, fmt.Errorf("byteReader: negative position")
	}
	r.pos = abs
	return abs, nil
}

var _ domain.Storage = (*S3Storage)(nil)
