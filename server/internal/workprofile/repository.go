package workprofile

import "context"

// Repository is the persistence boundary for work tabs and their tasks.
// Shaped for the Postgres implementation; repository_memory.go is the test
// stand-in.
type Repository interface {
	CreateTab(ctx context.Context, name string) (Tab, error)
	ListTabs(ctx context.Context) ([]Tab, error)
	UpdateTab(ctx context.Context, id string, input UpdateTabInput) (Tab, error)
	DeleteTab(ctx context.Context, id string) error

	CreateTask(ctx context.Context, tabID string, t Task) (Task, error)
	ListTasksByTab(ctx context.Context, tabID string) ([]Task, error)
	UpdateTask(ctx context.Context, id string, input UpdateTaskInput) (Task, error)
	DeleteTask(ctx context.Context, id string) error

	// ListOpenTasksWithTab returns every not-done task across all tabs, each
	// carrying its tab's name — the input to the overview.
	ListOpenTasksWithTab(ctx context.Context) ([]TaskWithTab, error)
}
