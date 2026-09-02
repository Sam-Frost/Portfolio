package diary

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/blobstore"
	"github.com/Sam-Frost/portfolio/internal/id"
)

// VideoService owns the diary video log: recording clips are uploaded
// straight from the browser to the blob store as a multipart upload (parts
// PUT while the recording is still going), and this service brokers the
// presigned URLs and confirms the result.
//
// It reuses the text diary's edit-lock rule (IsLocked): a clip can only be
// added to, or removed from, a day whose 24-hour grace window is still
// open. Playback is never locked.
type VideoService struct {
	repo VideoRepository
	blob blobstore.Store
}

func NewVideoService(repo VideoRepository, blob blobstore.Store) *VideoService {
	return &VideoService{repo: repo, blob: blob}
}

// contentTypeExt maps the two MIME types MediaRecorder produces to a file
// extension for the storage key (so the object streams back with a sane
// type and the local store can sniff it).
func contentTypeExt(contentType string) (ext string, ok bool) {
	switch {
	case strings.HasPrefix(contentType, "video/mp4"):
		return "mp4", true
	case strings.HasPrefix(contentType, "video/webm"):
		return "webm", true
	default:
		return "", false
	}
}

func (s *VideoService) ListByDate(ctx context.Context, date string) ([]Video, error) {
	if err := validateDate("date", date); err != nil {
		return nil, err
	}
	return s.repo.ListByDate(ctx, date)
}

func (s *VideoService) CountsByDateRange(ctx context.Context, from, to string) (map[string]int, error) {
	if err := validateDate("from", from); err != nil {
		return nil, err
	}
	if err := validateDate("to", to); err != nil {
		return nil, err
	}
	if to < from {
		return nil, apperr.InvalidInput("to must not be before from")
	}
	return s.repo.CountsByDateRange(ctx, from, to)
}

// CreateUpload opens a multipart upload for a new clip on date and returns
// the pending row plus the upload id. Mirrors document.Service.CreateDocument:
// the row is written first, then the store call, with the row rolled back
// if the store call fails so a retry starts clean.
func (s *VideoService) CreateUpload(ctx context.Context, date string, in CreateVideoInput) (CreatedVideo, error) {
	if err := validateDate("date", date); err != nil {
		return CreatedVideo{}, err
	}
	if IsLocked(date, time.Now().UTC()) {
		return CreatedVideo{}, apperr.Conflict("this day's edit window has closed")
	}
	ext, ok := contentTypeExt(in.ContentType)
	if !ok {
		return CreatedVideo{}, apperr.InvalidInput("contentType must be video/mp4 or video/webm")
	}

	var title *string
	if in.Title != nil {
		if t := strings.TrimSpace(*in.Title); t != "" {
			title = &t
		}
	}

	key := id.New() + "/clip." + ext
	created, err := s.repo.Create(ctx, Video{
		EntryDate:   date,
		Title:       title,
		ContentType: in.ContentType,
		Status:      VideoStatusPending,
		S3Key:       key,
	})
	if err != nil {
		return CreatedVideo{}, err
	}

	uploadID, err := s.blob.CreateMultipart(ctx, key, in.ContentType)
	if err != nil {
		if _, _, _, delErr := s.repo.Delete(ctx, created.ID); delErr != nil {
			log.Printf("diary: failed to roll back pending video row %s: %v", created.ID, delErr)
		}
		return CreatedVideo{}, apperr.Internal("failed to start upload")
	}
	if err := s.repo.SetUploadID(ctx, created.ID, uploadID); err != nil {
		_ = s.blob.AbortMultipart(ctx, key, uploadID)
		if _, _, _, delErr := s.repo.Delete(ctx, created.ID); delErr != nil {
			log.Printf("diary: failed to roll back pending video row %s: %v", created.ID, delErr)
		}
		return CreatedVideo{}, apperr.Internal("failed to start upload")
	}
	created.UploadID = uploadID

	return CreatedVideo{Video: created, UploadID: uploadID}, nil
}

// PartURL hands the browser a presigned URL to PUT one part's bytes to.
func (s *VideoService) PartURL(ctx context.Context, videoID string, partNumber int) (string, error) {
	if partNumber < 1 {
		return "", apperr.InvalidInput("partNumber must be at least 1")
	}
	v, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return "", err
	}
	if v.Status != VideoStatusPending {
		return "", apperr.Conflict("this upload is already complete")
	}
	if v.UploadID == "" {
		return "", apperr.Internal("upload was not initialized")
	}
	url, err := s.blob.PresignUploadPart(ctx, v.S3Key, v.UploadID, partNumber)
	if err != nil {
		return "", apperr.Internal("failed to prepare upload part")
	}
	return url, nil
}

// Complete assembles the uploaded parts, confirms the object landed, and
// flips the row to ready.
func (s *VideoService) Complete(ctx context.Context, videoID string, in CompleteVideoInput) (Video, error) {
	if len(in.Parts) == 0 {
		return Video{}, apperr.InvalidInput("parts must not be empty")
	}
	if in.DurationSeconds != nil && *in.DurationSeconds < 0 {
		return Video{}, apperr.InvalidInput("durationSeconds must not be negative")
	}

	v, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return Video{}, err
	}
	if v.Status == VideoStatusReady {
		return v, nil
	}
	if v.UploadID == "" {
		return Video{}, apperr.Internal("upload was not initialized")
	}

	if err := s.blob.CompleteMultipart(ctx, v.S3Key, v.UploadID, in.Parts); err != nil {
		return Video{}, apperr.Internal("failed to finalize upload")
	}

	size, ok, err := s.blob.Stat(ctx, v.S3Key)
	if err != nil {
		return Video{}, apperr.Internal("failed to verify upload")
	}
	if !ok {
		return Video{}, apperr.InvalidInput("upload not found — the recording was not received")
	}

	return s.repo.MarkReady(ctx, videoID, size, in.DurationSeconds)
}

// PlaybackURL returns a short-lived URL the browser can stream the clip
// from (S3 serves Range requests, so seeking works).
func (s *VideoService) PlaybackURL(ctx context.Context, videoID string) (string, error) {
	v, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return "", err
	}
	if v.Status != VideoStatusReady {
		return "", apperr.Conflict("this recording is still uploading")
	}
	name := "diary-" + v.EntryDate + "." + mustExt(v.ContentType)
	url, err := s.blob.PresignGet(ctx, v.S3Key, name, true, playbackURLTTLSeconds)
	if err != nil {
		return "", apperr.Internal("failed to prepare playback")
	}
	return url, nil
}

// Delete removes a clip. Allowed only while the day is still editable; an
// unfinished upload is aborted, a finished object is deleted.
func (s *VideoService) Delete(ctx context.Context, videoID string) error {
	v, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return err
	}
	if IsLocked(v.EntryDate, time.Now().UTC()) {
		return apperr.Conflict("this day's edit window has closed")
	}

	key, uploadID, status, err := s.repo.Delete(ctx, videoID)
	if err != nil {
		return err
	}
	// The row is gone; a stray blob is harmless, so don't fail the request.
	if status == VideoStatusPending && uploadID != "" {
		if abErr := s.blob.AbortMultipart(ctx, key, uploadID); abErr != nil {
			log.Printf("diary: failed to abort upload for video %s: %v", videoID, abErr)
		}
		return nil
	}
	if delErr := s.blob.Delete(ctx, key); delErr != nil {
		log.Printf("diary: failed to delete blob %s: %v", key, delErr)
	}
	return nil
}

func mustExt(contentType string) string {
	if ext, ok := contentTypeExt(contentType); ok {
		return ext
	}
	return "mp4"
}
