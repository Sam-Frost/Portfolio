package document

import "context"

// BlobStore is where document bytes actually live. The browser talks to it
// directly via the URLs PresignPut / PresignGet return, so file data never
// transits this server. blobstore_s3.go is the production implementation;
// blobstore_local.go is a local-disk stand-in for dev and the no-AWS
// deployment.
type BlobStore interface {
	// PresignPut returns a URL the browser can PUT the object bytes to,
	// valid for a few minutes. contentType is the type the PUT must send.
	PresignPut(ctx context.Context, key, contentType string) (url string, err error)
	// PresignGet returns a URL the browser can GET the object from as an
	// attachment named filename, valid for a few minutes.
	PresignGet(ctx context.Context, key, filename string) (url string, err error)
	// Stat reports the stored object's size. ok is false (nil error) when
	// the object isn't there yet — used to confirm an upload finished.
	Stat(ctx context.Context, key string) (size int64, ok bool, err error)
	// Delete removes objects. A missing key is not an error.
	Delete(ctx context.Context, keys ...string) error
}
