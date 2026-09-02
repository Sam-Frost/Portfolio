package notepadlabel

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type MemoryRepository struct {
	mu     sync.Mutex
	labels map[string]Label
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{labels: make(map[string]Label)}
}

func (r *MemoryRepository) nameTaken(name, excludeID string) bool {
	for otherID, existing := range r.labels {
		if otherID != excludeID && strings.EqualFold(existing.Name, name) {
			return true
		}
	}
	return false
}

func (r *MemoryRepository) Create(_ context.Context, l Label) (Label, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.nameTaken(l.Name, "") {
		return Label{}, apperr.InvalidInput("a label with this name already exists")
	}

	l.ID = id.New()
	r.labels[l.ID] = l
	return l, nil
}

func (r *MemoryRepository) List(_ context.Context) ([]Label, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	labels := make([]Label, 0, len(r.labels))
	for _, l := range r.labels {
		labels = append(labels, l)
	}
	sort.Slice(labels, func(i, j int) bool { return labels[i].Name < labels[j].Name })
	return labels, nil
}

func (r *MemoryRepository) Update(_ context.Context, labelID string, input UpdateInput) (Label, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	l, ok := r.labels[labelID]
	if !ok {
		return Label{}, apperr.NotFound("label not found")
	}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if r.nameTaken(trimmed, labelID) {
			return Label{}, apperr.InvalidInput("a label with this name already exists")
		}
		l.Name = trimmed
	}
	if input.Color != nil {
		l.Color = *input.Color
	}

	r.labels[labelID] = l
	return l, nil
}

func (r *MemoryRepository) Delete(_ context.Context, labelID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.labels[labelID]; !ok {
		return apperr.NotFound("label not found")
	}
	delete(r.labels, labelID)
	return nil
}
