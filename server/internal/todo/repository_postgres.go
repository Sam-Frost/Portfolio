package todo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

func (r *PostgresRepository) Create(ctx context.Context, t Todo) (Todo, error) {
	t.ID = id.New()
	t.DateAdded = time.Now().UTC()
	t.Done = false
	t.CompletedAt = nil

	const q = `INSERT INTO todos (id, name, description, date_added, target_date, done, completed_at, label_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err := r.db.ExecContext(ctx, q, t.ID, t.Name, t.Description, t.DateAdded, t.TargetDate, t.Done, t.CompletedAt, t.LabelID); err != nil {
		if isForeignKeyViolation(err) {
			return Todo{}, apperr.InvalidInput("label not found")
		}
		return Todo{}, apperr.Internal("failed to create todo")
	}
	return t, nil
}

func (r *PostgresRepository) List(ctx context.Context, sortField SortField, order SortOrder, labelID *string) ([]Todo, error) {
	orderSQL := "DESC"
	if order == SortAsc {
		orderSQL = "ASC"
	}

	// NULL target dates / completion timestamps always sort last, regardless
	// of direction, mirroring the in-memory repository's `less`.
	orderClause := " ORDER BY date_added " + orderSQL
	switch sortField {
	case SortByTargetDate:
		orderClause = " ORDER BY target_date IS NULL, target_date " + orderSQL
	case SortByCompletedAt:
		orderClause = " ORDER BY completed_at IS NULL, completed_at " + orderSQL
	}

	q := "SELECT id, name, description, date_added, target_date, done, completed_at, label_id FROM todos"
	args := make([]any, 0, 1)
	if labelID != nil {
		q += " WHERE label_id = $1"
		args = append(args, *labelID)
	}
	q += orderClause

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, apperr.Internal("failed to list todos")
	}
	defer rows.Close()

	todos := make([]Todo, 0)
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.DateAdded, &t.TargetDate, &t.Done, &t.CompletedAt, &t.LabelID); err != nil {
			return nil, apperr.Internal("failed to scan todo")
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list todos")
	}

	return todos, nil
}

func (r *PostgresRepository) Update(ctx context.Context, todoID string, input UpdateInput) (Todo, error) {
	sets := make([]string, 0, 5)
	args := make([]any, 0, 6)
	argN := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argN))
		args = append(args, strings.TrimSpace(*input.Name))
		argN++
	}
	if input.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argN))
		args = append(args, *input.Description)
		argN++
	}
	if input.TargetDate != nil {
		sets = append(sets, fmt.Sprintf("target_date = $%d", argN))
		if *input.TargetDate == "" {
			args = append(args, nil)
		} else {
			args = append(args, *input.TargetDate)
		}
		argN++
	}
	if input.Done != nil {
		sets = append(sets, fmt.Sprintf("done = $%d", argN))
		args = append(args, *input.Done)
		argN++

		// Keep completed_at in lockstep with done (the DB enforces this too,
		// via todos_done_completed_at_consistent): set it on completion,
		// clear it on undo so a later redo records a fresh timestamp.
		if *input.Done {
			sets = append(sets, fmt.Sprintf("completed_at = $%d", argN))
			args = append(args, time.Now().UTC())
			argN++
		} else {
			sets = append(sets, "completed_at = NULL")
		}
	}
	if input.LabelID != nil {
		sets = append(sets, fmt.Sprintf("label_id = $%d", argN))
		if *input.LabelID == "" {
			args = append(args, nil)
		} else {
			args = append(args, *input.LabelID)
		}
		argN++
	}

	if len(sets) == 0 {
		return r.get(ctx, todoID)
	}

	args = append(args, todoID)
	q := fmt.Sprintf(
		"UPDATE todos SET %s WHERE id = $%d RETURNING id, name, description, date_added, target_date, done, completed_at, label_id",
		strings.Join(sets, ", "), argN,
	)

	var t Todo
	err := r.db.QueryRowContext(ctx, q, args...).Scan(&t.ID, &t.Name, &t.Description, &t.DateAdded, &t.TargetDate, &t.Done, &t.CompletedAt, &t.LabelID)
	if errors.Is(err, sql.ErrNoRows) {
		return Todo{}, apperr.NotFound("todo not found")
	}
	if isForeignKeyViolation(err) {
		return Todo{}, apperr.InvalidInput("label not found")
	}
	if err != nil {
		return Todo{}, apperr.Internal("failed to update todo")
	}
	return t, nil
}

func (r *PostgresRepository) get(ctx context.Context, todoID string) (Todo, error) {
	const q = `SELECT id, name, description, date_added, target_date, done, completed_at, label_id FROM todos WHERE id = $1`

	var t Todo
	err := r.db.QueryRowContext(ctx, q, todoID).Scan(&t.ID, &t.Name, &t.Description, &t.DateAdded, &t.TargetDate, &t.Done, &t.CompletedAt, &t.LabelID)
	if errors.Is(err, sql.ErrNoRows) {
		return Todo{}, apperr.NotFound("todo not found")
	}
	if err != nil {
		return Todo{}, apperr.Internal("failed to get todo")
	}
	return t, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, todoID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM todos WHERE id = $1`, todoID)
	if err != nil {
		return apperr.Internal("failed to delete todo")
	}

	n, err := res.RowsAffected()
	if err != nil {
		return apperr.Internal("failed to delete todo")
	}
	if n == 0 {
		return apperr.NotFound("todo not found")
	}
	return nil
}

func (r *PostgresRepository) CountActive(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM todos WHERE done = false`).Scan(&count); err != nil {
		return 0, apperr.Internal("failed to count todos")
	}
	return count, nil
}

// isForeignKeyViolation reports whether err is a Postgres foreign-key
// violation (SQLSTATE 23503) — todos.label_id references labels.id, so this
// is how an unknown label surfaces as a 400 instead of a raw 500.
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
