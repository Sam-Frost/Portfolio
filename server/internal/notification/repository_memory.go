package notification

import (
	"context"
	"sync"
	"time"

	"github.com/Sam-Frost/portfolio/internal/id"
)

type MemoryRepository struct {
	mu   sync.Mutex
	subs map[string]PushSubscription // keyed by endpoint
	log  map[string]bool             // keyed by kind + "|" + istDate
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{subs: map[string]PushSubscription{}, log: map[string]bool{}}
}

func (r *MemoryRepository) Subscribe(_ context.Context, sub PushSubscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	sub.ID = id.New()
	sub.CreatedAt = time.Now().UTC()
	r.subs[sub.Endpoint] = sub
	return nil
}

func (r *MemoryRepository) Resync(ctx context.Context, oldEndpoint string, sub PushSubscription) error {
	r.mu.Lock()
	if oldEndpoint != "" && oldEndpoint != sub.Endpoint {
		if _, ok := r.subs[oldEndpoint]; ok {
			delete(r.subs, oldEndpoint)
			sub.ID = id.New()
			sub.CreatedAt = time.Now().UTC()
			r.subs[sub.Endpoint] = sub
			r.mu.Unlock()
			return nil
		}
	}
	r.mu.Unlock()
	return r.Subscribe(ctx, sub)
}

func (r *MemoryRepository) Unsubscribe(_ context.Context, endpoint string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.subs, endpoint)
	return nil
}

func (r *MemoryRepository) ListSubscriptions(_ context.Context) ([]PushSubscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]PushSubscription, 0, len(r.subs))
	for _, s := range r.subs {
		out = append(out, s)
	}
	return out, nil
}

func (r *MemoryRepository) LogExists(_ context.Context, kind, istDate string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.log[kind+"|"+istDate], nil
}

func (r *MemoryRepository) InsertLog(_ context.Context, kind, istDate string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.log[kind+"|"+istDate] = true
	return nil
}
