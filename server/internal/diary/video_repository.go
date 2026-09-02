package diary

import "context"

// VideoRepository is the persistence boundary for diary video rows, shaped
// for the Postgres implementation (video_repository_postgres.go);
// video_repository_memory.go is the in-memory stand-in used by tests.
//
// As with the text Entry repository, lock enforcement is not the
// repository's job — the VideoService checks IsLocked before Create/Delete.
type VideoRepository interface {
	// Create inserts a pending row (v.Status is set by the caller) and
	// returns it with ID/CreatedAt populated.
	Create(ctx context.Context, v Video) (Video, error)
	GetByID(ctx context.Context, id string) (Video, error)
	// ListByDate returns the day's videos, oldest first. Pending rows are
	// included so a resumable upload is still visible.
	ListByDate(ctx context.Context, date string) ([]Video, error)
	// CountsByDateRange returns, for [from, to] inclusive, how many ready
	// videos each date has — just enough for the calendar to mark days.
	CountsByDateRange(ctx context.Context, from, to string) (map[string]int, error)
	// SetUploadID records the store's multipart upload id on a pending row.
	SetUploadID(ctx context.Context, id, uploadID string) error
	// MarkReady flips a pending row to ready with its confirmed size and
	// (optional) measured duration.
	MarkReady(ctx context.Context, id string, size int64, durationSeconds *int) (Video, error)
	// Delete removes the row and returns its storage key, multipart upload
	// id, and prior status so the service can clean up the blob (abort an
	// unfinished upload, or delete a finished object).
	Delete(ctx context.Context, id string) (s3Key, uploadID, status string, err error)
}
