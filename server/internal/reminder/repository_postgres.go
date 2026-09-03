package reminder

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, rem Reminder) (Reminder, error) {
	rem.ID = id.New()
	rem.CreatedAt = time.Now().UTC()

	const q = `INSERT INTO reminders (id, todo_id, kind, fire_at, interval_seconds, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := r.db.ExecContext(ctx, q, rem.ID, rem.TodoID, rem.Kind, rem.FireAt, rem.IntervalSeconds, rem.CreatedAt); err != nil {
		if isForeignKeyViolation(err) {
			return Reminder{}, apperr.NotFound("todo not found")
		}
		return Reminder{}, apperr.Internal("failed to create reminder")
	}
	return rem, nil
}

func (r *PostgresRepository) ListByTodo(ctx context.Context, todoID string) ([]Reminder, error) {
	const q = `SELECT id, todo_id, kind, fire_at, interval_seconds, created_at
		FROM reminders WHERE todo_id = $1 ORDER BY fire_at`
	rows, err := r.db.QueryContext(ctx, q, todoID)
	if err != nil {
		return nil, apperr.Internal("failed to list reminders")
	}
	defer rows.Close()

	out := make([]Reminder, 0)
	for rows.Next() {
		var rem Reminder
		if err := rows.Scan(&rem.ID, &rem.TodoID, &rem.Kind, &rem.FireAt, &rem.IntervalSeconds, &rem.CreatedAt); err != nil {
			return nil, apperr.Internal("failed to scan reminder")
		}
		out = append(out, rem)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list reminders")
	}
	return out, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM reminders WHERE id = $1`, id); err != nil {
		return apperr.Internal("failed to delete reminder")
	}
	return nil
}

func (r *PostgresRepository) DueBefore(ctx context.Context, t time.Time) ([]Due, error) {
	const q = `SELECT r.id, r.todo_id, r.kind, r.fire_at, r.interval_seconds, r.created_at, td.name
		FROM reminders r
		JOIN todos td ON td.id = r.todo_id
		WHERE td.done = FALSE AND r.fire_at <= $1
		ORDER BY r.fire_at`
	rows, err := r.db.QueryContext(ctx, q, t)
	if err != nil {
		return nil, apperr.Internal("failed to query due reminders")
	}
	defer rows.Close()

	out := make([]Due, 0)
	for rows.Next() {
		var d Due
		if err := rows.Scan(&d.ID, &d.TodoID, &d.Kind, &d.FireAt, &d.IntervalSeconds, &d.CreatedAt, &d.TodoName); err != nil {
			return nil, apperr.Internal("failed to scan due reminder")
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to query due reminders")
	}
	return out, nil
}

func (r *PostgresRepository) Reschedule(ctx context.Context, id string, next time.Time) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE reminders SET fire_at = $1 WHERE id = $2`, next, id); err != nil {
		return apperr.Internal("failed to reschedule reminder")
	}
	return nil
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
