package notepad

import (
	"context"
	"sort"
	"sync"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type MemoryRepository struct {
	mu    sync.Mutex
	notes map[string]Note
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{notes: make(map[string]Note)}
}

func (r *MemoryRepository) Create(_ context.Context, n Note) (Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n.ID = id.New()
	r.notes[n.ID] = n
	return n, nil
}

func (r *MemoryRepository) List(_ context.Context) ([]NoteSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	summaries := make([]NoteSummary, 0, len(r.notes))
	for _, n := range r.notes {
		summaries = append(summaries, NoteSummary{ID: n.ID, Title: n.Title, CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt})
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].CreatedAt.After(summaries[j].CreatedAt) })
	return summaries, nil
}

func (r *MemoryRepository) Get(_ context.Context, noteID string) (Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, ok := r.notes[noteID]
	if !ok {
		return Note{}, apperr.NotFound("note not found")
	}
	return n, nil
}

func (r *MemoryRepository) Update(_ context.Context, noteID string, input UpdateInput) (Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, ok := r.notes[noteID]
	if !ok {
		return Note{}, apperr.NotFound("note not found")
	}

	if input.Title != nil {
		n.Title = *input.Title
	}
	if input.ContentHTML != nil {
		n.ContentHTML = *input.ContentHTML
	}
	n.UpdatedAt = updatedAtNow()

	r.notes[noteID] = n
	return n, nil
}

func (r *MemoryRepository) Delete(_ context.Context, noteID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.notes[noteID]; !ok {
		return apperr.NotFound("note not found")
	}
	delete(r.notes, noteID)
	return nil
}
