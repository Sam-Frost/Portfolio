package notepad

import "context"

// Repository is the persistence boundary for notes, shaped for the eventual
// Postgres implementation (see repository_postgres.go); repository_memory.go
// is the in-memory stand-in used by tests. Get exists separately from List
// because List deliberately omits ContentHTML.
type Repository interface {
	Create(ctx context.Context, n Note) (Note, error)
	List(ctx context.Context) ([]NoteSummary, error)
	Get(ctx context.Context, id string) (Note, error)
	Update(ctx context.Context, id string, input UpdateInput) (Note, error)
	Delete(ctx context.Context, id string) error
}
