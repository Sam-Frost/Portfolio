package worksession

import (
	"context"
	"sync"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

// MemoryRepository is the in-memory stand-in for PostgresRepository, used
// in tests.
type MemoryRepository struct {
	mu       sync.Mutex
	sessions map[string]WorkSession
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{sessions: make(map[string]WorkSession)}
}

func (r *MemoryRepository) Create(_ context.Context, session WorkSession) (WorkSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	session.ID = id.New()
	r.sessions[session.ID] = session
	return session, nil
}

func (r *MemoryRepository) GetRunning(_ context.Context) (WorkSession, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, s := range r.sessions {
		if s.Status == StatusRunning {
			return s, true, nil
		}
	}
	return WorkSession{}, false, nil
}

func (r *MemoryRepository) Finish(_ context.Context, sessionID string, input FinishInput) (WorkSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.sessions[sessionID]
	if !ok || s.Status != StatusRunning {
		return WorkSession{}, apperr.InvalidInput("session is not running")
	}

	s.Status = input.Status
	s.Note = input.Note
	endedAt := input.EndedAt
	s.EndedAt = &endedAt
	actual := input.ActualMinutes
	s.ActualMinutes = &actual

	r.sessions[sessionID] = s
	return s, nil
}

func (r *MemoryRepository) ListRange(_ context.Context, from, to time.Time) ([]WorkSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]WorkSession, 0)
	for _, s := range r.sessions {
		if !s.StartedAt.Before(to) {
			continue
		}
		if s.EndedAt != nil && s.EndedAt.Before(from) {
			continue
		}
		result = append(result, s)
	}
	return result, nil
}
