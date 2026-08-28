package document

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	putURLTTL = 15 * time.Minute
	getURLTTL = 5 * time.Minute
)

// S3Config configures the production blob store. Bucket is required.
type S3Config struct {
	Bucket string
	Prefix string // key prefix inside the bucket, e.g. "documents". "" = bucket root.
	Region string
}

type s3BlobStore struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
	prefix  string
}

// NewS3BlobStore builds an S3-backed BlobStore, loading AWS credentials
// from the default chain (env vars, shared config, or an instance role) —
// same as cms.NewS3Publisher.
func NewS3BlobStore(ctx context.Context, cfg S3Config) (BlobStore, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("document: S3Config.Bucket is required")
	}

	opts := []func(*awsconfig.LoadOptions) error{}
	if cfg.Region != "" {
		opts = append(opts, awsconfig.WithRegion(cfg.Region))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("document: load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	return &s3BlobStore{
		client:  client,
		presign: s3.NewPresignClient(client),
		bucket:  cfg.Bucket,
		prefix:  cfg.Prefix,
	}, nil
}

func (s *s3BlobStore) fullKey(key string) string {
	return path.Join(s.prefix, key)
}

func (s *s3BlobStore) PresignPut(ctx context.Context, key, contentType string) (string, error) {
	req, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s.fullKey(key)),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(putURLTTL))
	if err != nil {
		return "", fmt.Errorf("presign put: %w", err)
	}
	return req.URL, nil
}

func (s *s3BlobStore) PresignGet(ctx context.Context, key, filename string) (string, error) {
	req, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s.bucket),
		Key:                        aws.String(s.fullKey(key)),
		ResponseContentDisposition: aws.String(fmt.Sprintf("attachment; filename=%q", filename)),
	}, s3.WithPresignExpires(getURLTTL))
	if err != nil {
		return "", fmt.Errorf("presign get: %w", err)
	}
	return req.URL, nil
}

func (s *s3BlobStore) Stat(ctx context.Context, key string) (int64, bool, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.fullKey(key)),
	})
	if err != nil {
		var notFound *s3types.NotFound
		if errors.As(err, &notFound) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("head object: %w", err)
	}
	var size int64
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	return size, true, nil
}

func (s *s3BlobStore) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	objects := make([]s3types.ObjectIdentifier, 0, len(keys))
	for _, k := range keys {
		objects = append(objects, s3types.ObjectIdentifier{Key: aws.String(s.fullKey(k))})
	}

	_, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucket),
		Delete: &s3types.Delete{Objects: objects, Quiet: aws.Bool(true)},
	})
	if err != nil {
		return fmt.Errorf("delete objects: %w", err)
	}
	return nil
}
