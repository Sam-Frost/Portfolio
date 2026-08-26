package spotify

import "context"

// Repository is the persistence boundary for the domain area's single
// Spotify connection — there's no per-user auth to key multiple
// connections by (see server/internal/auth), so this holds exactly one
// TokenSet. Get's ok return is false until HandleCallback has completed
// once; repository_postgres.go is the real implementation,
// repository_memory.go the in-memory stand-in used by tests.
type Repository interface {
	Get(ctx context.Context) (TokenSet, bool, error)
	Save(ctx context.Context, tokens TokenSet) error
}
