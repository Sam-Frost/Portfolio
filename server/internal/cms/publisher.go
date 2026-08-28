package cms

import (
	"context"
	"log"
)

// Publisher ships a serialized content.json to wherever the public site
// reads it from. The real implementation (publisher_s3.go) puts the object
// in the site's S3 origin and invalidates CloudFront; noopPublisher is used
// when no S3 target is configured (local dev), so the rest of the CMS still
// works and Publish returns a clear "not configured" error.
type Publisher interface {
	// Enabled reports whether Publish will actually do anything. The service
	// checks this before Publish so an unconfigured server returns a 409
	// instead of silently "succeeding".
	Enabled() bool
	// Publish writes content (the full content.json bytes) as the live
	// document, tagging the archived copy with version.
	Publish(ctx context.Context, version int, content []byte) error
}

// noopPublisher logs and does nothing. Enabled() is false.
type noopPublisher struct{}

// NewNoopPublisher returns a Publisher that does nothing — used when
// CMS_S3_BUCKET is unset.
func NewNoopPublisher() Publisher { return noopPublisher{} }

func (noopPublisher) Enabled() bool { return false }

func (noopPublisher) Publish(_ context.Context, version int, content []byte) error {
	log.Printf("cms: noop publisher — would have shipped content.json v%d (%d bytes)", version, len(content))
	return nil
}
