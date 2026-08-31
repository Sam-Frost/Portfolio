package notepad

import (
	"context"
	"sort"
	"sync"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type MemoryRepository struct {
	mu        sync.Mutex
	notes     map[string]Note
	deleted   map[string]bool
	scratchID string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{notes: make(map[string]Note), deleted: make(map[string]bool)}
}

func (r *MemoryRepository) Create(_ context.Context, n Note) (Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n.ID = id.New()
	r.notes[n.ID] = n
	return n, nil
}

func (r *MemoryRepository) List(_ context.Context, filter ListFilter) ([]NoteSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	summaries := make([]NoteSummary, 0, len(r.notes))
	for _, n := range r.notes {
		if r.deleted[n.ID] || n.Scratch || n.Archived != filter.Archived {
			continue
		}
		summaries = append(summaries, NoteSummary{
			ID: n.ID, Title: n.Title, Pinned: n.Pinned, Archived: n.Archived, Locked: n.Locked,
			CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt,
		})
	}
	// Pinned notes float to the top of the working set; within each group,
	// newest first. The archive view has no pinned notes (archiving clears
	// the pin) so it's a plain newest-first list.
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Pinned != summaries[j].Pinned {
			return summaries[i].Pinned
		}
		return summaries[i].CreatedAt.After(summaries[j].CreatedAt)
	})
	return summaries, nil
}

func (r *MemoryRepository) Scratch(_ context.Context) (Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if n, ok := r.notes[r.scratchID]; ok && !r.deleted[n.ID] {
		return n, nil
	}

	now := updatedAtNow()
	n := Note{ID: id.New(), Scratch: true, CreatedAt: now, UpdatedAt: now}
	r.notes[n.ID] = n
	r.scratchID = n.ID
	return n, nil
}

func (r *MemoryRepository) Get(_ context.Context, noteID string) (Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, ok := r.notes[noteID]
	if !ok || r.deleted[noteID] {
		return Note{}, apperr.NotFound("note not found")
	}
	return n, nil
}

func (r *MemoryRepository) Update(_ context.Context, noteID string, input UpdateInput) (Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, ok := r.notes[noteID]
	if !ok || r.deleted[noteID] {
		return Note{}, apperr.NotFound("note not found")
	}

	if input.Title != nil {
		n.Title = *input.Title
	}
	if input.ContentHTML != nil {
		n.ContentHTML = *input.ContentHTML
	}
	if input.Pinned != nil {
		n.Pinned = *input.Pinned
	}
	if input.Locked != nil {
		n.Locked = *input.Locked
	}
	if input.Archived != nil {
		n.Archived = *input.Archived
		if n.Archived {
			// An archived note can't also be pinned to the working set.
			n.Pinned = false
		}
	}
	n.UpdatedAt = updatedAtNow()

	r.notes[noteID] = n
	return n, nil
}

// Delete is a soft delete: the note is flagged rather than removed, so it
// drops out of List/Get/Update but its data isn't destroyed.
func (r *MemoryRepository) Delete(_ context.Context, noteID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.notes[noteID]; !ok || r.deleted[noteID] {
		return apperr.NotFound("note not found")
	}
	r.deleted[noteID] = true
	return nil
}
