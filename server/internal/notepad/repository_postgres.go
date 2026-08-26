package notepad

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, n Note) (Note, error) {
	n.ID = id.New()

	const q = `INSERT INTO notes (id, title, content_html, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	if _, err := r.db.ExecContext(ctx, q, n.ID, n.Title, n.ContentHTML, n.CreatedAt, n.UpdatedAt); err != nil {
		return Note{}, apperr.Internal("failed to create note")
	}
	return n, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]NoteSummary, error) {
	const q = `SELECT id, title, created_at, updated_at FROM notes WHERE deleted_at IS NULL ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, apperr.Internal("failed to list notes")
	}
	defer rows.Close()

	summaries := make([]NoteSummary, 0)
	for rows.Next() {
		var s NoteSummary
		if err := rows.Scan(&s.ID, &s.Title, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, apperr.Internal("failed to scan note")
		}
		summaries = append(summaries, s)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list notes")
	}

	return summaries, nil
}

func (r *PostgresRepository) Get(ctx context.Context, noteID string) (Note, error) {
	const q = `SELECT id, title, content_html, created_at, updated_at FROM notes WHERE id = $1 AND deleted_at IS NULL`

	var n Note
	err := r.db.QueryRowContext(ctx, q, noteID).Scan(&n.ID, &n.Title, &n.ContentHTML, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Note{}, apperr.NotFound("note not found")
	}
	if err != nil {
		return Note{}, apperr.Internal("failed to get note")
	}
	return n, nil
}

func (r *PostgresRepository) Update(ctx context.Context, noteID string, input UpdateInput) (Note, error) {
	sets := make([]string, 0, 3)
	args := make([]any, 0, 4)
	argN := 1

	if input.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", argN))
		args = append(args, strings.TrimSpace(*input.Title))
		argN++
	}
	if input.ContentHTML != nil {
		sets = append(sets, fmt.Sprintf("content_html = $%d", argN))
		args = append(args, *input.ContentHTML)
		argN++
	}

	if len(sets) == 0 {
		return r.Get(ctx, noteID)
	}

	sets = append(sets, fmt.Sprintf("updated_at = $%d", argN))
	args = append(args, updatedAtNow())
	argN++

	args = append(args, noteID)
	q := fmt.Sprintf(
		"UPDATE notes SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING id, title, content_html, created_at, updated_at",
		strings.Join(sets, ", "), argN,
	)

	var n Note
	err := r.db.QueryRowContext(ctx, q, args...).Scan(&n.ID, &n.Title, &n.ContentHTML, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Note{}, apperr.NotFound("note not found")
	}
	if err != nil {
		return Note{}, apperr.Internal("failed to update note")
	}
	return n, nil
}

// Delete is a soft delete: it stamps deleted_at rather than removing the
// row, so the note drops out of List/Get/Update but its data is retained.
func (r *PostgresRepository) Delete(ctx context.Context, noteID string) error {
	res, err := r.db.ExecContext(
		ctx, `UPDATE notes SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, updatedAtNow(), noteID,
	)
	if err != nil {
		return apperr.Internal("failed to delete note")
	}

	n, err := res.RowsAffected()
	if err != nil {
		return apperr.Internal("failed to delete note")
	}
	if n == 0 {
		return apperr.NotFound("note not found")
	}
	return nil
}
