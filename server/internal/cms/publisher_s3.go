package cms

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// s3API / cloudfrontAPI are the slices of the AWS SDK the publisher uses,
// pulled out as interfaces so tests can substitute fakes.
type s3API interface {
	PutObject(ctx context.Context, in *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type cloudfrontAPI interface {
	CreateInvalidation(ctx context.Context, in *cloudfront.CreateInvalidationInput, optFns ...func(*cloudfront.Options)) (*cloudfront.CreateInvalidationOutput, error)
}

// S3Config configures the live publisher. Bucket is required; everything
// else has a sensible default or is optional.
type S3Config struct {
	Bucket         string
	Prefix         string // key prefix inside the bucket, e.g. "sat0ru". "" = bucket root.
	ContentKey     string // basename of the live document, default "content.json".
	DistributionID string // CloudFront distribution to invalidate; "" skips invalidation.
	Region         string
}

type s3Publisher struct {
	s3         s3API
	cf         cloudfrontAPI
	bucket     string
	contentKey string // full key of the live document
	historyDir string // full key prefix for archived versions
	distID     string
}

// NewS3Publisher builds a Publisher backed by S3 + CloudFront, loading AWS
// credentials from the default chain (env vars, shared config, or an EC2
// instance role).
func NewS3Publisher(ctx context.Context, cfg S3Config) (Publisher, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("cms: S3Config.Bucket is required")
	}

	opts := []func(*awsconfig.LoadOptions) error{}
	if cfg.Region != "" {
		opts = append(opts, awsconfig.WithRegion(cfg.Region))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("cms: load aws config: %w", err)
	}

	return newS3PublisherWithClients(s3.NewFromConfig(awsCfg), cloudfront.NewFromConfig(awsCfg), cfg), nil
}

func newS3PublisherWithClients(s3c s3API, cfc cloudfrontAPI, cfg S3Config) *s3Publisher {
	contentBase := cfg.ContentKey
	if contentBase == "" {
		contentBase = "content.json"
	}
	return &s3Publisher{
		s3:         s3c,
		cf:         cfc,
		bucket:     cfg.Bucket,
		contentKey: path.Join(cfg.Prefix, contentBase),
		historyDir: path.Join(cfg.Prefix, "content", "history"),
		distID:     cfg.DistributionID,
	}
}

func (p *s3Publisher) Enabled() bool { return true }

func (p *s3Publisher) Publish(ctx context.Context, version int, content []byte) error {
	// Archive the versioned copy first: if this succeeds but the live PUT
	// fails, we still have the artifact; the reverse would leave a live doc
	// with no history entry.
	historyKey := path.Join(p.historyDir, fmt.Sprintf("v%d-%s.json", version, time.Now().UTC().Format("20060102T150405Z")))
	if err := p.put(ctx, historyKey, content, "private, max-age=31536000, immutable"); err != nil {
		return fmt.Errorf("archive %s: %w", historyKey, err)
	}

	if err := p.put(ctx, p.contentKey, content, "no-cache, must-revalidate"); err != nil {
		return fmt.Errorf("put %s: %w", p.contentKey, err)
	}

	if p.distID == "" {
		return nil
	}
	if _, err := p.cf.CreateInvalidation(ctx, &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(p.distID),
		InvalidationBatch: &cftypes.InvalidationBatch{
			CallerReference: aws.String("cms-" + strconv.Itoa(version) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)),
			Paths: &cftypes.Paths{
				Quantity: aws.Int32(1),
				Items:    []string{"/" + path.Base(p.contentKey)},
			},
		},
	}); err != nil {
		return fmt.Errorf("cloudfront invalidation: %w", err)
	}
	return nil
}

func (p *s3Publisher) put(ctx context.Context, key string, body []byte, cacheControl string) error {
	_, err := p.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(p.bucket),
		Key:          aws.String(key),
		Body:         bytes.NewReader(body),
		ContentType:  aws.String("application/json"),
		CacheControl: aws.String(cacheControl),
	})
	return err
}
