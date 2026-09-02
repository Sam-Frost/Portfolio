package blobstore

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const putURLTTL = 15 * time.Minute

// S3Config configures the production store. Bucket is required.
type S3Config struct {
	Bucket string
	Prefix string // key prefix inside the bucket, e.g. "diary-videos". "" = bucket root.
	Region string
}

type s3Store struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
	prefix  string
}

// NewS3 builds an S3-backed Store, loading AWS credentials from the default
// chain (env vars, shared config, or an instance role) — same as
// cms.NewS3Publisher and document.NewS3BlobStore.
func NewS3(ctx context.Context, cfg S3Config) (Store, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("blobstore: S3Config.Bucket is required")
	}

	opts := []func(*awsconfig.LoadOptions) error{}
	if cfg.Region != "" {
		opts = append(opts, awsconfig.WithRegion(cfg.Region))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("blobstore: load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	return &s3Store{
		client:  client,
		presign: s3.NewPresignClient(client),
		bucket:  cfg.Bucket,
		prefix:  cfg.Prefix,
	}, nil
}

func (s *s3Store) fullKey(key string) string { return path.Join(s.prefix, key) }

func (s *s3Store) PresignPut(ctx context.Context, key, contentType string) (string, error) {
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

func (s *s3Store) PresignGet(ctx context.Context, key, filename string, inline bool, ttlSeconds int) (string, error) {
	disposition := fmt.Sprintf("attachment; filename=%q", filename)
	if inline {
		disposition = "inline"
	}
	req, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s.bucket),
		Key:                        aws.String(s.fullKey(key)),
		ResponseContentDisposition: aws.String(disposition),
	}, s3.WithPresignExpires(time.Duration(ttlSeconds)*time.Second))
	if err != nil {
		return "", fmt.Errorf("presign get: %w", err)
	}
	return req.URL, nil
}

func (s *s3Store) Stat(ctx context.Context, key string) (int64, bool, error) {
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

func (s *s3Store) Delete(ctx context.Context, keys ...string) error {
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

func (s *s3Store) CreateMultipart(ctx context.Context, key, contentType string) (string, error) {
	out, err := s.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s.fullKey(key)),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("create multipart upload: %w", err)
	}
	if out.UploadId == nil {
		return "", fmt.Errorf("create multipart upload: no upload id returned")
	}
	return *out.UploadId, nil
}

func (s *s3Store) PresignUploadPart(ctx context.Context, key, uploadID string, partNumber int) (string, error) {
	req, err := s.presign.PresignUploadPart(ctx, &s3.UploadPartInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(s.fullKey(key)),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(int32(partNumber)),
	}, s3.WithPresignExpires(putURLTTL))
	if err != nil {
		return "", fmt.Errorf("presign upload part: %w", err)
	}
	return req.URL, nil
}

func (s *s3Store) CompleteMultipart(ctx context.Context, key, uploadID string, parts []Part) error {
	ordered := append([]Part(nil), parts...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Number < ordered[j].Number })

	completed := make([]s3types.CompletedPart, 0, len(ordered))
	for _, p := range ordered {
		completed = append(completed, s3types.CompletedPart{
			ETag:       aws.String(p.ETag),
			PartNumber: aws.Int32(int32(p.Number)),
		})
	}

	_, err := s.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(s.bucket),
		Key:             aws.String(s.fullKey(key)),
		UploadId:        aws.String(uploadID),
		MultipartUpload: &s3types.CompletedMultipartUpload{Parts: completed},
	})
	if err != nil {
		return fmt.Errorf("complete multipart upload: %w", err)
	}
	return nil
}

func (s *s3Store) AbortMultipart(ctx context.Context, key, uploadID string) error {
	_, err := s.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(s.fullKey(key)),
		UploadId: aws.String(uploadID),
	})
	if err != nil {
		var noUpload *s3types.NoSuchUpload
		if errors.As(err, &noUpload) {
			return nil
		}
		return fmt.Errorf("abort multipart upload: %w", err)
	}
	return nil
}
