package todo

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type MemoryRepository struct {
	mu    sync.Mutex
	todos map[string]Todo
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{todos: make(map[string]Todo)}
}

func (r *MemoryRepository) Create(_ context.Context, todo Todo) (Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	todo.ID = id.New()
	todo.DateAdded = time.Now().UTC()
	todo.Done = false
	r.todos[todo.ID] = todo
	return todo, nil
}

func (r *MemoryRepository) List(_ context.Context, sortField SortField, order SortOrder, labelID *string) ([]Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	todos := make([]Todo, 0, len(r.todos))
	for _, t := range r.todos {
		if labelID != nil && (t.LabelID == nil || *t.LabelID != *labelID) {
			continue
		}
		todos = append(todos, t)
	}

	sort.Slice(todos, func(i, j int) bool {
		return less(todos[i], todos[j], sortField, order)
	})

	return todos, nil
}

// less mirrors the frontend's compareTodos for the fields the server now
// sorts: null target dates always sort to the end, regardless of order.
func less(a, b Todo, field SortField, order SortOrder) bool {
	ascending := order != SortDesc

	if field == SortByTargetDate {
		if a.TargetDate == nil && b.TargetDate == nil {
			return false
		}
		if a.TargetDate == nil {
			return false
		}
		if b.TargetDate == nil {
			return true
		}
		if ascending {
			return *a.TargetDate < *b.TargetDate
		}
		return *a.TargetDate > *b.TargetDate
	}

	if ascending {
		return a.DateAdded.Before(b.DateAdded)
	}
	return a.DateAdded.After(b.DateAdded)
}

func (r *MemoryRepository) Update(_ context.Context, todoID string, input UpdateInput) (Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.todos[todoID]
	if !ok {
		return Todo{}, apperr.NotFound("todo not found")
	}

	if input.Name != nil {
		t.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		t.Description = input.Description
	}
	if input.TargetDate != nil {
		if *input.TargetDate == "" {
			t.TargetDate = nil
		} else {
			t.TargetDate = input.TargetDate
		}
	}
	if input.Done != nil {
		t.Done = *input.Done
	}
	if input.LabelID != nil {
		if *input.LabelID == "" {
			t.LabelID = nil
		} else {
			t.LabelID = input.LabelID
		}
	}

	r.todos[todoID] = t
	return t, nil
}

func (r *MemoryRepository) Delete(_ context.Context, todoID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.todos[todoID]; !ok {
		return apperr.NotFound("todo not found")
	}
	delete(r.todos, todoID)
	return nil
}

func (r *MemoryRepository) CountActive(_ context.Context) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for _, t := range r.todos {
		if !t.Done {
			count++
		}
	}
	return count, nil
}
