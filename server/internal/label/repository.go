package label

import "context"

// Repository is the persistence boundary for labels, shaped for the
// eventual Postgres implementation (see repository_postgres.go);
// repository_memory.go is the in-memory stand-in used by tests.
type Repository interface {
	Create(ctx context.Context, l Label) (Label, error)
	List(ctx context.Context) ([]Label, error)
	Update(ctx context.Context, id string, input UpdateInput) (Label, error)
	Delete(ctx context.Context, id string) error
}
