package reminder

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/Sam-Frost/portfolio/internal/id"
)

// todoState is the slice of a todo the memory repo needs to mimic the
// Postgres DueBefore join.
type todoState struct {
	Name string
	Done bool
}

type MemoryRepository struct {
	mu        sync.Mutex
	reminders map[string]Reminder
	// todos is set by tests so DueBefore can filter on done-ness and attach
	// a name, matching the real JOIN.
	todos map[string]todoState
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{reminders: map[string]Reminder{}, todos: map[string]todoState{}}
}

// SetTodo registers a todo's state for the DueBefore join (test helper).
func (r *MemoryRepository) SetTodo(id, name string, done bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.todos[id] = todoState{Name: name, Done: done}
}

func (r *MemoryRepository) Create(_ context.Context, rem Reminder) (Reminder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rem.ID = id.New()
	rem.CreatedAt = time.Now().UTC()
	r.reminders[rem.ID] = rem
	return rem, nil
}

func (r *MemoryRepository) ListByTodo(_ context.Context, todoID string) ([]Reminder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Reminder, 0)
	for _, rem := range r.reminders {
		if rem.TodoID == todoID {
			out = append(out, rem)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].FireAt.Before(out[j].FireAt) })
	return out, nil
}

func (r *MemoryRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.reminders, id)
	return nil
}

func (r *MemoryRepository) DueBefore(_ context.Context, t time.Time) ([]Due, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Due, 0)
	for _, rem := range r.reminders {
		td, ok := r.todos[rem.TodoID]
		if !ok || td.Done || rem.FireAt.After(t) {
			continue
		}
		out = append(out, Due{Reminder: rem, TodoName: td.Name})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].FireAt.Before(out[j].FireAt) })
	return out, nil
}

func (r *MemoryRepository) Reschedule(_ context.Context, id string, next time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rem, ok := r.reminders[id]; ok {
		rem.FireAt = next
		r.reminders[id] = rem
	}
	return nil
}
