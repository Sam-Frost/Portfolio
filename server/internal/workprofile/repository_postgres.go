package workprofile

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

// --- tabs ---

func (r *PostgresRepository) CreateTab(ctx context.Context, name string) (Tab, error) {
	tab := Tab{ID: id.New(), Name: name, CreatedAt: time.Now().UTC()}

	// New tab goes to the end of the bar.
	const q = `INSERT INTO work_tabs (id, name, position, created_at)
		VALUES ($1, $2, COALESCE((SELECT MAX(position) + 1 FROM work_tabs), 0), $3)
		RETURNING position`
	if err := r.db.QueryRowContext(ctx, q, tab.ID, tab.Name, tab.CreatedAt).Scan(&tab.Position); err != nil {
		return Tab{}, apperr.Internal("failed to create tab")
	}
	return tab, nil
}

func (r *PostgresRepository) ListTabs(ctx context.Context) ([]Tab, error) {
	const q = `SELECT id, name, position, created_at FROM work_tabs ORDER BY position, created_at`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, apperr.Internal("failed to list tabs")
	}
	defer rows.Close()

	out := make([]Tab, 0)
	for rows.Next() {
		var t Tab
		if err := rows.Scan(&t.ID, &t.Name, &t.Position, &t.CreatedAt); err != nil {
			return nil, apperr.Internal("failed to scan tab")
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list tabs")
	}
	return out, nil
}

func (r *PostgresRepository) UpdateTab(ctx context.Context, tabID string, input UpdateTabInput) (Tab, error) {
	sets := make([]string, 0, 2)
	args := make([]any, 0, 3)
	argN := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argN))
		args = append(args, strings.TrimSpace(*input.Name))
		argN++
	}
	if input.Position != nil {
		sets = append(sets, fmt.Sprintf("position = $%d", argN))
		args = append(args, *input.Position)
		argN++
	}
	if len(sets) == 0 {
		return r.getTab(ctx, tabID)
	}

	args = append(args, tabID)
	q := fmt.Sprintf(
		"UPDATE work_tabs SET %s WHERE id = $%d RETURNING id, name, position, created_at",
		strings.Join(sets, ", "), argN,
	)
	var t Tab
	err := r.db.QueryRowContext(ctx, q, args...).Scan(&t.ID, &t.Name, &t.Position, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Tab{}, apperr.NotFound("tab not found")
	}
	if err != nil {
		return Tab{}, apperr.Internal("failed to update tab")
	}
	return t, nil
}

func (r *PostgresRepository) DeleteTab(ctx context.Context, tabID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM work_tabs WHERE id = $1`, tabID)
	if err != nil {
		return apperr.Internal("failed to delete tab")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperr.NotFound("tab not found")
	}
	return nil
}

func (r *PostgresRepository) getTab(ctx context.Context, tabID string) (Tab, error) {
	const q = `SELECT id, name, position, created_at FROM work_tabs WHERE id = $1`
	var t Tab
	err := r.db.QueryRowContext(ctx, q, tabID).Scan(&t.ID, &t.Name, &t.Position, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Tab{}, apperr.NotFound("tab not found")
	}
	if err != nil {
		return Tab{}, apperr.Internal("failed to get tab")
	}
	return t, nil
}

// --- tasks ---

const taskCols = `id, tab_id, name, description, target_date, done, completed_at, jira_acknowledged, created_at`

func scanTask(s interface {
	Scan(dest ...any) error
}) (Task, error) {
	var t Task
	err := s.Scan(&t.ID, &t.TabID, &t.Name, &t.Description, &t.TargetDate, &t.Done, &t.CompletedAt, &t.JiraAcknowledged, &t.CreatedAt)
	return t, err
}

func (r *PostgresRepository) CreateTask(ctx context.Context, tabID string, t Task) (Task, error) {
	t.ID = id.New()
	t.TabID = tabID
	t.CreatedAt = time.Now().UTC()
	t.Done = false
	t.CompletedAt = nil
	t.JiraAcknowledged = false

	const q = `INSERT INTO work_tasks (` + taskCols + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	if _, err := r.db.ExecContext(ctx, q, t.ID, t.TabID, t.Name, t.Description, t.TargetDate, t.Done, t.CompletedAt, t.JiraAcknowledged, t.CreatedAt); err != nil {
		if isForeignKeyViolation(err) {
			return Task{}, apperr.NotFound("tab not found")
		}
		return Task{}, apperr.Internal("failed to create task")
	}
	return t, nil
}

func (r *PostgresRepository) ListTasksByTab(ctx context.Context, tabID string) ([]Task, error) {
	const q = `SELECT ` + taskCols + ` FROM work_tasks WHERE tab_id = $1
		ORDER BY done, target_date IS NULL, target_date, created_at`
	rows, err := r.db.QueryContext(ctx, q, tabID)
	if err != nil {
		return nil, apperr.Internal("failed to list tasks")
	}
	defer rows.Close()

	out := make([]Task, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, apperr.Internal("failed to scan task")
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list tasks")
	}
	return out, nil
}

func (r *PostgresRepository) UpdateTask(ctx context.Context, taskID string, input UpdateTaskInput) (Task, error) {
	sets := make([]string, 0, 6)
	args := make([]any, 0, 7)
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

		if *input.Done {
			sets = append(sets, fmt.Sprintf("completed_at = $%d", argN))
			args = append(args, time.Now().UTC())
			argN++
			// Completing implies the Jira acknowledgement (the service has
			// already rejected the request otherwise).
			sets = append(sets, "jira_acknowledged = TRUE")
		} else {
			// Undo: clear the timestamp and the ack so re-completing
			// prompts again.
			sets = append(sets, "completed_at = NULL", "jira_acknowledged = FALSE")
		}
	}

	if len(sets) == 0 {
		return r.getTask(ctx, taskID)
	}

	args = append(args, taskID)
	q := fmt.Sprintf("UPDATE work_tasks SET %s WHERE id = $%d RETURNING "+taskCols, strings.Join(sets, ", "), argN)
	t, err := scanTask(r.db.QueryRowContext(ctx, q, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, apperr.NotFound("task not found")
	}
	if err != nil {
		return Task{}, apperr.Internal("failed to update task")
	}
	return t, nil
}

func (r *PostgresRepository) DeleteTask(ctx context.Context, taskID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM work_tasks WHERE id = $1`, taskID)
	if err != nil {
		return apperr.Internal("failed to delete task")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperr.NotFound("task not found")
	}
	return nil
}

func (r *PostgresRepository) getTask(ctx context.Context, taskID string) (Task, error) {
	t, err := scanTask(r.db.QueryRowContext(ctx, `SELECT `+taskCols+` FROM work_tasks WHERE id = $1`, taskID))
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, apperr.NotFound("task not found")
	}
	if err != nil {
		return Task{}, apperr.Internal("failed to get task")
	}
	return t, nil
}

func (r *PostgresRepository) ListOpenTasksWithTab(ctx context.Context) ([]TaskWithTab, error) {
	q := `SELECT ` + prefixed("wt", taskCols) + `, tb.name
		FROM work_tasks wt JOIN work_tabs tb ON tb.id = wt.tab_id
		WHERE wt.done = FALSE
		ORDER BY wt.target_date IS NULL, wt.target_date, wt.created_at`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, apperr.Internal("failed to list open tasks")
	}
	defer rows.Close()

	out := make([]TaskWithTab, 0)
	for rows.Next() {
		var tw TaskWithTab
		if err := rows.Scan(
			&tw.ID, &tw.TabID, &tw.Name, &tw.Description, &tw.TargetDate,
			&tw.Done, &tw.CompletedAt, &tw.JiraAcknowledged, &tw.CreatedAt, &tw.TabName,
		); err != nil {
			return nil, apperr.Internal("failed to scan open task")
		}
		out = append(out, tw)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list open tasks")
	}
	return out, nil
}

// prefixed turns "a, b, c" into "p.a, p.b, p.c" for a join select list.
func prefixed(p, cols string) string {
	parts := strings.Split(cols, ", ")
	for i, c := range parts {
		parts[i] = p + "." + c
	}
	return strings.Join(parts, ", ")
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
