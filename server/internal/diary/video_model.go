package diary

import (
	"time"

	"github.com/Sam-Frost/portfolio/internal/blobstore"
)

// Video status values, mirroring internal/document's.
const (
	VideoStatusPending = "pending" // row created, multipart upload open, bytes not all in
	VideoStatusReady   = "ready"   // upload completed and confirmed present in the store
)

// playbackURLTTLSeconds is how long a playback URL stays valid. Video
// sessions run much longer than a document download, and an S3 presigned
// URL that expires mid-playback breaks Range requests, so this is
// deliberately generous (2h).
const playbackURLTTLSeconds = 2 * 60 * 60

// Video is one recorded diary clip for a calendar day. A day can have
// several (they're their own rows, independent of the day's written Entry).
// S3Key / UploadID are storage internals and never sent to clients.
type Video struct {
	ID              string     `json:"id"`
	EntryDate       string     `json:"entryDate"`
	Title           *string    `json:"title"`
	ContentType     string     `json:"contentType"`
	SizeBytes       int64      `json:"sizeBytes"`
	DurationSeconds *int       `json:"durationSeconds"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"createdAt"`
	UploadedAt      *time.Time `json:"uploadedAt"`

	S3Key    string `json:"-"`
	UploadID string `json:"-"`
}

// CreateVideoInput opens a new upload for a day. ContentType is the MIME
// type the browser's MediaRecorder settled on (video/mp4 or video/webm).
type CreateVideoInput struct {
	ContentType string  `json:"contentType"`
	Title       *string `json:"title"`
}

// CreatedVideo is the create response: the pending row plus the multipart
// upload id the browser threads through every part / complete call.
type CreatedVideo struct {
	Video    Video  `json:"video"`
	UploadID string `json:"uploadId"`
}

// CompleteVideoInput finalizes an upload: the ETags the browser collected
// from each uploaded part, plus the clip's measured duration.
type CompleteVideoInput struct {
	Parts           []blobstore.Part `json:"parts"`
	DurationSeconds *int             `json:"durationSeconds"`
}
