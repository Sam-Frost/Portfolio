package worksession

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

// goalSlice maps []Goal to/from the goals JSONB column, so the repo doesn't
// need a Postgres array dependency (pgx/v5 is used via the stdlib
// database/sql driver here). Mirrors cms.strSlice.
type goalSlice []Goal

func (g goalSlice) Value() (driver.Value, error) {
	if g == nil {
		return "[]", nil
	}
	b, err := json.Marshal([]Goal(g))
	return string(b), err
}

func (g *goalSlice) Scan(src any) error {
	if src == nil {
		*g = []Goal{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("worksession: cannot scan %T into goalSlice", src)
	}
	if len(b) == 0 {
		*g = []Goal{}
		return nil
	}
	return json.Unmarshal(b, (*[]Goal)(g))
}

const sessionCols = `id, planned_minutes, category, started_at, ended_at, status, goals, start_note, note, actual_minutes`

func scanSession(sc interface{ Scan(...any) error }) (WorkSession, error) {
	var s WorkSession
	var goals goalSlice
	if err := sc.Scan(
		&s.ID, &s.PlannedMinutes, &s.Category, &s.StartedAt, &s.EndedAt,
		&s.Status, &goals, &s.StartNote, &s.Note, &s.ActualMinutes,
	); err != nil {
		return WorkSession{}, err
	}
	s.Goals = goals
	return s, nil
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, session WorkSession) (WorkSession, error) {
	session.ID = id.New()

	const q = `
		INSERT INTO work_sessions (id, planned_minutes, category, started_at, status, goals, start_note)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if _, err := r.db.ExecContext(ctx, q,
		session.ID, session.PlannedMinutes, session.Category, session.StartedAt,
		session.Status, goalSlice(session.Goals), session.StartNote,
	); err != nil {
		return WorkSession{}, apperr.Internal("failed to create work session")
	}
	return session, nil
}

func (r *PostgresRepository) GetRunning(ctx context.Context) (WorkSession, bool, error) {
	const q = `
		SELECT ` + sessionCols + `
		FROM work_sessions WHERE status = $1
		ORDER BY started_at DESC LIMIT 1
	`
	s, err := scanSession(r.db.QueryRowContext(ctx, q, StatusRunning))
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
		SET status = $1, goals = $2, note = $3, ended_at = $4, actual_minutes = $5
		WHERE id = $6 AND status = $7
		RETURNING ` + sessionCols + `
	`
	s, err := scanSession(r.db.QueryRowContext(ctx, q,
		input.Status, goalSlice(input.Goals), input.Note, input.EndedAt, input.ActualMinutes,
		sessionID, StatusRunning,
	))
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
		SELECT ` + sessionCols + `
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
		s, err := scanSession(rows)
		if err != nil {
			return nil, apperr.Internal("failed to scan work session")
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list work sessions")
	}

	return sessions, nil
}
