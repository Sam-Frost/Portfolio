package notepad

import "context"

// Repository is the persistence boundary for notes, shaped for the eventual
// Postgres implementation (see repository_postgres.go); repository_memory.go
// is the in-memory stand-in used by tests. Get exists separately from List
// because List deliberately omits ContentHTML.
type Repository interface {
	Create(ctx context.Context, n Note) (Note, error)
	List(ctx context.Context, filter ListFilter) ([]NoteSummary, error)
	// Scratch returns the singleton free-form scratch note ("Random
	// Notepad"), creating it on first access. Unlike Create it takes no
	// title, and the note it returns is never included in List — it's a
	// quick-jot buffer kept separate from the titled, organized notes.
	// Callers edit it through the ordinary Update path once they have its ID.
	Scratch(ctx context.Context) (Note, error)
	Get(ctx context.Context, id string) (Note, error)
	Update(ctx context.Context, id string, input UpdateInput) (Note, error)
	Delete(ctx context.Context, id string) error
}
