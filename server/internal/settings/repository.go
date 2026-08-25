package settings

import "context"

// Repository is the persistence boundary for the single settings record,
// shaped for the eventual Postgres implementation (see
// repository_postgres.go); repository_memory.go is the in-memory stand-in
// used by tests.
type Repository interface {
	Get(ctx context.Context) (Settings, error)
	Update(ctx context.Context, input UpdateInput) (Settings, error)
}
