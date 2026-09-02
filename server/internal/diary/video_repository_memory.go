package diary

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type MemoryVideoRepository struct {
	mu     sync.Mutex
	videos map[string]Video // keyed by ID
}

func NewMemoryVideoRepository() *MemoryVideoRepository {
	return &MemoryVideoRepository{videos: make(map[string]Video)}
}

func (r *MemoryVideoRepository) Create(_ context.Context, v Video) (Video, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	v.ID = id.New()
	v.CreatedAt = time.Now().UTC()
	r.videos[v.ID] = v
	return v, nil
}

func (r *MemoryVideoRepository) GetByID(_ context.Context, videoID string) (Video, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	v, ok := r.videos[videoID]
	if !ok {
		return Video{}, apperr.NotFound("diary video not found")
	}
	return v, nil
}

func (r *MemoryVideoRepository) ListByDate(_ context.Context, date string) ([]Video, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]Video, 0)
	for _, v := range r.videos {
		if v.EntryDate == date {
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *MemoryVideoRepository) CountsByDateRange(_ context.Context, from, to string) (map[string]int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	counts := make(map[string]int)
	for _, v := range r.videos {
		if v.Status == VideoStatusReady && v.EntryDate >= from && v.EntryDate <= to {
			counts[v.EntryDate]++
		}
	}
	return counts, nil
}

func (r *MemoryVideoRepository) SetUploadID(_ context.Context, videoID, uploadID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	v, ok := r.videos[videoID]
	if !ok {
		return apperr.NotFound("diary video not found")
	}
	v.UploadID = uploadID
	r.videos[videoID] = v
	return nil
}

func (r *MemoryVideoRepository) MarkReady(_ context.Context, videoID string, size int64, durationSeconds *int) (Video, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	v, ok := r.videos[videoID]
	if !ok {
		return Video{}, apperr.NotFound("diary video not found")
	}
	now := time.Now().UTC()
	v.Status = VideoStatusReady
	v.SizeBytes = size
	v.DurationSeconds = durationSeconds
	v.UploadedAt = &now
	r.videos[videoID] = v
	return v, nil
}

func (r *MemoryVideoRepository) Delete(_ context.Context, videoID string) (s3Key, uploadID, status string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	v, ok := r.videos[videoID]
	if !ok {
		return "", "", "", apperr.NotFound("diary video not found")
	}
	delete(r.videos, videoID)
	return v.S3Key, v.UploadID, v.Status, nil
}
