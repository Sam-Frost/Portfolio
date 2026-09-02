// Package blobstore is where large binary objects (currently diary video
// logs) live. The browser talks to it directly via the URLs the Presign*
// methods return, so object bytes never transit this server.
//
// It supports both a single-shot PUT (small objects) and a multipart
// upload (large objects, uploaded part-by-part while they're still being
// produced — see the diary video recorder, which uploads parts as it
// records so a browser crash loses at most the last few seconds).
//
// s3.go is the production implementation; local.go is a local-disk
// stand-in for dev and the no-AWS deployment, presenting the same
// browser-direct upload/download flow via HMAC-signed URLs that point back
// at this server.
//
// This mirrors internal/document's own (older, single-shot only) blob
// store; that one should eventually be collapsed into this package.
package blobstore

import (
	"context"
	"net/http"
)

// Part is one piece of a multipart upload: the 1-based part number and the
// ETag the store returned when that part's bytes were PUT. The caller
// collects these as it uploads and hands the full set back to
// CompleteMultipart.
type Part struct {
	Number int    `json:"number"`
	ETag   string `json:"etag"`
}

// Store is the persistence boundary for blob bytes. Keys are opaque
// slash-separated paths (e.g. "<id>/clip.mp4"); an implementation may add
// its own prefix.
type Store interface {
	// PresignPut returns a URL the browser can PUT the whole object to,
	// valid for a few minutes. contentType is the type the PUT must send.
	PresignPut(ctx context.Context, key, contentType string) (url string, err error)
	// PresignGet returns a URL the browser can GET the object from, valid
	// for ttlSeconds. When inline is true the object is served for in-page
	// playback (Content-Disposition: inline); otherwise as an attachment
	// named filename.
	PresignGet(ctx context.Context, key, filename string, inline bool, ttlSeconds int) (url string, err error)
	// Stat reports the stored object's size. ok is false (nil error) when
	// the object isn't there yet — used to confirm an upload finished.
	Stat(ctx context.Context, key string) (size int64, ok bool, err error)
	// Delete removes objects. A missing key is not an error.
	Delete(ctx context.Context, keys ...string) error

	// CreateMultipart begins a multipart upload for key and returns its
	// upload ID, which every subsequent part / complete / abort call for
	// this key must carry.
	CreateMultipart(ctx context.Context, key, contentType string) (uploadID string, err error)
	// PresignUploadPart returns a URL the browser can PUT one part's bytes
	// to. partNumber is 1-based and ascending; every part except the last
	// must be at least 5 MiB (an S3 rule the local store doesn't enforce).
	// The PUT response's ETag header is what goes into the matching Part.
	PresignUploadPart(ctx context.Context, key, uploadID string, partNumber int) (url string, err error)
	// CompleteMultipart assembles the uploaded parts into the final object.
	// parts need not be sorted; the store orders them by Number.
	CompleteMultipart(ctx context.Context, key, uploadID string, parts []Part) error
	// AbortMultipart discards an unfinished multipart upload and any parts
	// already uploaded. A missing/unknown upload is not an error.
	AbortMultipart(ctx context.Context, key, uploadID string) error
}

// LocalServer is implemented by a Store that needs this server to serve the
// raw upload/download bytes itself (the local-disk store). The S3 store
// hands out URLs pointing straight at the bucket and implements nothing
// here. The caller registers ServeBlob on RoutePrefix and exempts that
// prefix from bearer auth — ServeBlob verifies its own URL signature.
type LocalServer interface {
	RoutePrefix() string
	ServeBlob(w http.ResponseWriter, r *http.Request)
}
