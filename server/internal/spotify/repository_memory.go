package spotify

import (
	"context"
	"sync"
)

// MemoryRepository is the in-memory stand-in for PostgresRepository, used
// in tests. There is exactly one token set (the whole domain area shares a
// single Spotify connection), so this is just a guarded optional value.
type MemoryRepository struct {
	mu     sync.Mutex
	tokens *TokenSet
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

func (r *MemoryRepository) Get(_ context.Context) (TokenSet, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tokens == nil {
		return TokenSet{}, false, nil
	}
	return *r.tokens, true, nil
}

func (r *MemoryRepository) Save(_ context.Context, tokens TokenSet) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tokens = &tokens
	return nil
}
