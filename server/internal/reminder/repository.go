package reminder

import (
	"context"
	"time"
)

// Repository is the persistence boundary for reminders. Shaped for the
// Postgres implementation (repository_postgres.go); repository_memory.go is
// the test stand-in.
type Repository interface {
	Create(ctx context.Context, r Reminder) (Reminder, error)
	// ListByTodo returns the reminders on a todo, soonest first.
	ListByTodo(ctx context.Context, todoID string) ([]Reminder, error)
	// Delete removes a reminder; a missing id is not an error.
	Delete(ctx context.Context, id string) error
	// DueBefore returns every reminder whose FireAt is at or before t and
	// whose todo is not done, newest-due first. This is the join that makes
	// a repeating reminder stop once its todo is completed.
	DueBefore(ctx context.Context, t time.Time) ([]Due, error)
	// Reschedule moves a repeating reminder's next fire to next.
	Reschedule(ctx context.Context, id string, next time.Time) error
}
