package worksession

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, session WorkSession) (WorkSession, error) {
	session.ID = id.New()

	const q = `
		INSERT INTO work_sessions (id, planned_minutes, started_at, status)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := r.db.ExecContext(ctx, q, session.ID, session.PlannedMinutes, session.StartedAt, session.Status); err != nil {
		return WorkSession{}, apperr.Internal("failed to create work session")
	}
	return session, nil
}

func (r *PostgresRepository) GetRunning(ctx context.Context) (WorkSession, bool, error) {
	const q = `
		SELECT id, planned_minutes, started_at, ended_at, status, note, actual_minutes
		FROM work_sessions WHERE status = $1
		ORDER BY started_at DESC LIMIT 1
	`
	var s WorkSession
	err := r.db.QueryRowContext(ctx, q, StatusRunning).Scan(
		&s.ID, &s.PlannedMinutes, &s.StartedAt, &s.EndedAt, &s.Status, &s.Note, &s.ActualMinutes,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkSession{}, false, nil
	}
	if err != nil {
		return WorkSession{}, false, apperr.Internal("failed to load running work session")
	}
	return s, true, nil
}

func (r *PostgresRepository) Finish(ctx context.Context, sessionID string, input FinishInput) (WorkSession, error) {
	const q = `
		UPDATE work_sessions
		SET status = $1, note = $2, ended_at = $3, actual_minutes = $4
		WHERE id = $5 AND status = $6
		RETURNING id, planned_minutes, started_at, ended_at, status, note, actual_minutes
	`
	var s WorkSession
	err := r.db.QueryRowContext(ctx, q, input.Status, input.Note, input.EndedAt, input.ActualMinutes, sessionID, StatusRunning).Scan(
		&s.ID, &s.PlannedMinutes, &s.StartedAt, &s.EndedAt, &s.Status, &s.Note, &s.ActualMinutes,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkSession{}, apperr.InvalidInput("session is not running")
	}
	if err != nil {
		return WorkSession{}, apperr.Internal("failed to finish work session")
	}
	return s, nil
}

func (r *PostgresRepository) ListRange(ctx context.Context, from, to time.Time) ([]WorkSession, error) {
	const q = `
		SELECT id, planned_minutes, started_at, ended_at, status, note, actual_minutes
		FROM work_sessions
		WHERE started_at < $1 AND (ended_at IS NULL OR ended_at >= $2)
		ORDER BY started_at ASC
	`
	rows, err := r.db.QueryContext(ctx, q, to, from)
	if err != nil {
		return nil, apperr.Internal("failed to list work sessions")
	}
	defer rows.Close()

	sessions := make([]WorkSession, 0)
	for rows.Next() {
		var s WorkSession
		if err := rows.Scan(&s.ID, &s.PlannedMinutes, &s.StartedAt, &s.EndedAt, &s.Status, &s.Note, &s.ActualMinutes); err != nil {
			return nil, apperr.Internal("failed to scan work session")
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list work sessions")
	}

	return sessions, nil
}
