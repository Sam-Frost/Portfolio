package todo

import "context"

type SortField string

const (
	SortByDateAdded  SortField = "dateAdded"
	SortByTargetDate SortField = "targetDate"
)

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

// Repository is the persistence boundary for todos. Repository is an
// in-memory stand-in (repository_memory.go) for a future Postgres-backed
// one — service/handler only depend on this interface, so swapping the
// backing store later doesn't touch either of them. Update takes a full
// UpdateInput (not a mutation closure) so a SQL implementation can build a
// real "SET" clause instead of reading a row just to satisfy a callback.
type Repository interface {
	Create(ctx context.Context, todo Todo) (Todo, error)
	// List returns todos, optionally restricted to labelID when non-nil.
	List(ctx context.Context, sortField SortField, order SortOrder, labelID *string) ([]Todo, error)
	Update(ctx context.Context, id string, input UpdateInput) (Todo, error)
	Delete(ctx context.Context, id string) error
	CountActive(ctx context.Context) (int, error)
}
