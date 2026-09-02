package drawingboard

import (
	"context"
	"sort"
	"sync"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type MemoryRepository struct {
	mu      sync.Mutex
	boards  map[string]Board
	deleted map[string]bool
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{boards: make(map[string]Board), deleted: make(map[string]bool)}
}

func (r *MemoryRepository) Create(_ context.Context, b Board) (Board, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b.ID = id.New()
	r.boards[b.ID] = b
	return b, nil
}

func (r *MemoryRepository) List(_ context.Context) ([]BoardSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	summaries := make([]BoardSummary, 0, len(r.boards))
	for _, b := range r.boards {
		if r.deleted[b.ID] {
			continue
		}
		summaries = append(summaries, BoardSummary{
			ID: b.ID, Name: b.Name, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt,
		})
	}
	// Most recently edited board first.
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].UpdatedAt.After(summaries[j].UpdatedAt)
	})
	return summaries, nil
}

func (r *MemoryRepository) Get(_ context.Context, boardID string) (Board, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.boards[boardID]
	if !ok || r.deleted[boardID] {
		return Board{}, apperr.NotFound("board not found")
	}
	return b, nil
}

func (r *MemoryRepository) Update(_ context.Context, boardID string, input UpdateInput) (Board, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.boards[boardID]
	if !ok || r.deleted[boardID] {
		return Board{}, apperr.NotFound("board not found")
	}

	if input.Name != nil {
		b.Name = *input.Name
	}
	if input.SceneData != nil {
		b.SceneData = input.SceneData
	}
	b.UpdatedAt = updatedAtNow()

	r.boards[boardID] = b
	return b, nil
}

// Delete is a soft delete: the board is flagged rather than removed, so it
// drops out of List/Get/Update but its data isn't destroyed.
func (r *MemoryRepository) Delete(_ context.Context, boardID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.boards[boardID]; !ok || r.deleted[boardID] {
		return apperr.NotFound("board not found")
	}
	r.deleted[boardID] = true
	return nil
}
