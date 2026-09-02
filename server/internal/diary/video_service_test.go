package diary

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/blobstore"
)

// fakeBlob is an in-memory blobstore.Store for video service tests. The
// multipart calls just record what happened; Stat reports a size once
// CompleteMultipart has run.
type fakeBlob struct {
	mu        sync.Mutex
	completed map[string]int64 // key -> size after CompleteMultipart
	aborted   []string
	deleted   []string
}

func newFakeBlob() *fakeBlob { return &fakeBlob{completed: map[string]int64{}} }

func (f *fakeBlob) PresignPut(context.Context, string, string) (string, error) { return "put://", nil }
func (f *fakeBlob) PresignGet(_ context.Context, key, _ string, _ bool, _ int) (string, error) {
	return "get://" + key, nil
}
func (f *fakeBlob) Stat(_ context.Context, key string) (int64, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	size, ok := f.completed[key]
	return size, ok, nil
}
func (f *fakeBlob) Delete(_ context.Context, keys ...string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleted = append(f.deleted, keys...)
	return nil
}
func (f *fakeBlob) CreateMultipart(context.Context, string, string) (string, error) {
	return "upload-1", nil
}
func (f *fakeBlob) PresignUploadPart(_ context.Context, key, _ string, part int) (string, error) {
	return "part://" + key, nil
}
func (f *fakeBlob) CompleteMultipart(_ context.Context, key, _ string, parts []blobstore.Part) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.completed[key] = int64(len(parts) * 1000)
	return nil
}
func (f *fakeBlob) AbortMultipart(_ context.Context, key, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.aborted = append(f.aborted, key)
	return nil
}

func newTestVideoService() (*VideoService, *MemoryVideoRepository, *fakeBlob) {
	repo := NewMemoryVideoRepository()
	blob := newFakeBlob()
	return NewVideoService(repo, blob), repo, blob
}

func today() string { return time.Now().In(IST).Format(EntryDateLayout) }

func TestVideoService_CreateUploadRejectsLockedDay(t *testing.T) {
	svc, _, _ := newTestVideoService()

	_, err := svc.CreateUpload(context.Background(), "2020-01-01", CreateVideoInput{ContentType: "video/mp4"})
	assertKind(t, err, apperr.KindConflict)
}

func TestVideoService_CreateUploadRejectsBadContentType(t *testing.T) {
	svc, _, _ := newTestVideoService()

	_, err := svc.CreateUpload(context.Background(), today(), CreateVideoInput{ContentType: "video/x-matroska"})
	assertKind(t, err, apperr.KindInvalidInput)
}

func TestVideoService_CreateUploadOpensPendingRow(t *testing.T) {
	svc, _, _ := newTestVideoService()

	created, err := svc.CreateUpload(context.Background(), today(), CreateVideoInput{ContentType: "video/mp4"})
	if err != nil {
		t.Fatalf("CreateUpload: %v", err)
	}
	if created.Video.Status != VideoStatusPending {
		t.Errorf("Status = %q, want pending", created.Video.Status)
	}
	if created.UploadID == "" {
		t.Errorf("UploadID is empty")
	}
	if !strings.HasSuffix(created.Video.S3Key, ".mp4") {
		t.Errorf("S3Key = %q, want a .mp4 suffix", created.Video.S3Key)
	}
}

func TestVideoService_CompleteFlipsToReadyWithSizeAndDuration(t *testing.T) {
	svc, _, _ := newTestVideoService()
	created, err := svc.CreateUpload(context.Background(), today(), CreateVideoInput{ContentType: "video/mp4"})
	if err != nil {
		t.Fatalf("CreateUpload: %v", err)
	}

	dur := 42
	done, err := svc.Complete(context.Background(), created.Video.ID, CompleteVideoInput{
		Parts:           []blobstore.Part{{Number: 1, ETag: "a"}, {Number: 2, ETag: "b"}},
		DurationSeconds: &dur,
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if done.Status != VideoStatusReady {
		t.Fatalf("Status = %q, want ready", done.Status)
	}
	if done.SizeBytes != 2000 {
		t.Errorf("SizeBytes = %d, want 2000", done.SizeBytes)
	}
	if done.DurationSeconds == nil || *done.DurationSeconds != 42 {
		t.Errorf("DurationSeconds = %v, want 42", done.DurationSeconds)
	}
}

func TestVideoService_CompleteRejectsEmptyParts(t *testing.T) {
	svc, _, _ := newTestVideoService()
	created, _ := svc.CreateUpload(context.Background(), today(), CreateVideoInput{ContentType: "video/webm"})

	_, err := svc.Complete(context.Background(), created.Video.ID, CompleteVideoInput{})
	assertKind(t, err, apperr.KindInvalidInput)
}

func TestVideoService_PlaybackURLRequiresReady(t *testing.T) {
	svc, _, _ := newTestVideoService()
	created, _ := svc.CreateUpload(context.Background(), today(), CreateVideoInput{ContentType: "video/mp4"})

	_, err := svc.PlaybackURL(context.Background(), created.Video.ID)
	assertKind(t, err, apperr.KindConflict)

	if _, err := svc.Complete(context.Background(), created.Video.ID, CompleteVideoInput{Parts: []blobstore.Part{{Number: 1, ETag: "a"}}}); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	url, err := svc.PlaybackURL(context.Background(), created.Video.ID)
	if err != nil {
		t.Fatalf("PlaybackURL: %v", err)
	}
	if !strings.HasPrefix(url, "get://") {
		t.Errorf("url = %q", url)
	}
}

func TestVideoService_DeletePendingAbortsUpload(t *testing.T) {
	svc, _, blob := newTestVideoService()
	created, _ := svc.CreateUpload(context.Background(), today(), CreateVideoInput{ContentType: "video/mp4"})

	if err := svc.Delete(context.Background(), created.Video.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(blob.aborted) != 1 || blob.aborted[0] != created.Video.S3Key {
		t.Errorf("aborted = %v, want [%s]", blob.aborted, created.Video.S3Key)
	}
}

func TestVideoService_DeleteReadyDeletesBlob(t *testing.T) {
	svc, _, blob := newTestVideoService()
	created, _ := svc.CreateUpload(context.Background(), today(), CreateVideoInput{ContentType: "video/mp4"})
	if _, err := svc.Complete(context.Background(), created.Video.ID, CompleteVideoInput{Parts: []blobstore.Part{{Number: 1, ETag: "a"}}}); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	if err := svc.Delete(context.Background(), created.Video.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(blob.deleted) != 1 || blob.deleted[0] != created.Video.S3Key {
		t.Errorf("deleted = %v, want [%s]", blob.deleted, created.Video.S3Key)
	}
}

func TestVideoService_DeleteRejectsLockedDay(t *testing.T) {
	svc, repo, _ := newTestVideoService()
	// Seed a ready row on a long-past date directly through the repo.
	v, _ := repo.Create(context.Background(), Video{EntryDate: "2020-01-01", ContentType: "video/mp4", Status: VideoStatusReady, S3Key: "x/clip.mp4"})

	err := svc.Delete(context.Background(), v.ID)
	assertKind(t, err, apperr.KindConflict)
}

func TestVideoService_CountsByDateRangeCountsOnlyReady(t *testing.T) {
	svc, repo, _ := newTestVideoService()
	ctx := context.Background()
	repo.Create(ctx, Video{EntryDate: "2026-09-01", Status: VideoStatusReady, S3Key: "a"})
	repo.Create(ctx, Video{EntryDate: "2026-09-01", Status: VideoStatusReady, S3Key: "b"})
	repo.Create(ctx, Video{EntryDate: "2026-09-02", Status: VideoStatusPending, S3Key: "c"})

	counts, err := svc.CountsByDateRange(ctx, "2026-09-01", "2026-09-30")
	if err != nil {
		t.Fatalf("CountsByDateRange: %v", err)
	}
	if counts["2026-09-01"] != 2 {
		t.Errorf("2026-09-01 count = %d, want 2", counts["2026-09-01"])
	}
	if _, ok := counts["2026-09-02"]; ok {
		t.Errorf("pending-only day should not be counted")
	}
}
