package diary

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type MemoryRepository struct {
	mu      sync.Mutex
	entries map[string]Entry // keyed by EntryDate
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{entries: make(map[string]Entry)}
}

func (r *MemoryRepository) GetByDate(_ context.Context, date string) (Entry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	e, ok := r.entries[date]
	if !ok {
		return Entry{}, apperr.NotFound("diary entry not found")
	}
	return e, nil
}

// Upsert creates the entry for date if none exists yet, otherwise updates
// its content in place, preserving ID and CreatedAt.
func (r *MemoryRepository) Upsert(_ context.Context, date string, content string) (Entry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	e, ok := r.entries[date]
	if !ok {
		e = Entry{ID: id.New(), EntryDate: date, CreatedAt: now}
	}
	e.Content = content
	e.UpdatedAt = now
	r.entries[date] = e
	return e, nil
}

func (r *MemoryRepository) ListDates(_ context.Context, from, to string) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	dates := make([]string, 0)
	for d := range r.entries {
		if d >= from && d <= to {
			dates = append(dates, d)
		}
	}
	sort.Strings(dates)
	return dates, nil
}
