package worksession

import (
	"context"
	"time"
)

// FinishInput is what both Complete and Cancel boil down to at the
// persistence layer: a transition off "running" to a terminal status, with
// an end time, elapsed minutes, and an optional note. Shaped like
// todo.UpdateInput — fields a SQL "UPDATE ... SET ... WHERE id = $1 AND
// status = 'running'" can use directly — rather than a mutation closure.
type FinishInput struct {
	Status        Status
	Note          *string
	EndedAt       time.Time
	ActualMinutes int
}

// Repository is the persistence boundary for work sessions.
// repository_postgres.go is the real implementation, repository_memory.go
// the in-memory stand-in used by tests.
type Repository interface {
	Create(ctx context.Context, session WorkSession) (WorkSession, error)
	// GetRunning returns the single 'running' session, if any — there is
	// at most one at a time (enforced in Service.Start).
	GetRunning(ctx context.Context) (WorkSession, bool, error)
	// Finish transitions the running session with the given id to a
	// terminal status. It only applies while that session is still
	// 'running' (guards against a double-complete/cancel race), returning
	// apperr.InvalidInput if it isn't found in that state.
	Finish(ctx context.Context, id string, input FinishInput) (WorkSession, error)
	// ListRange returns every session that overlaps [from, to) — i.e.
	// started_at < to AND (ended_at IS NULL OR ended_at >= from). from/to
	// are UTC instants at IST-midnight boundaries (see ParseISTDayStart);
	// day-bucketing/splitting happens in the service, not here.
	ListRange(ctx context.Context, from, to time.Time) ([]WorkSession, error)
}
