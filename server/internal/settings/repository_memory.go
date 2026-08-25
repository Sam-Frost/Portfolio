package settings

import (
	"context"
	"sync"
)

type MemoryRepository struct {
	mu       sync.Mutex
	settings Settings
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

func (r *MemoryRepository) Get(_ context.Context) (Settings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.settings, nil
}

func (r *MemoryRepository) Update(_ context.Context, input UpdateInput) (Settings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if input.DailyWorkTracker != nil {
		r.settings.DailyWorkTracker.TotalWorkHoursRequired = input.DailyWorkTracker.TotalWorkHoursRequired
	}

	return r.settings, nil
}
