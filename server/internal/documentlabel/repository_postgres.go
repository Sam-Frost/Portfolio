package documentlabel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

func (r *PostgresRepository) Create(ctx context.Context, l Label) (Label, error) {
	l.ID = id.New()

	const q = `INSERT INTO document_labels (id, name, color) VALUES ($1, $2, $3)`
	if _, err := r.db.ExecContext(ctx, q, l.ID, l.Name, l.Color); err != nil {
		if isUniqueViolation(err) {
			return Label{}, apperr.InvalidInput("a label with this name already exists")
		}
		return Label{}, apperr.Internal("failed to create label")
	}
	return l, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]Label, error) {
	const q = `SELECT id, name, color FROM document_labels ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, apperr.Internal("failed to list labels")
	}
	defer rows.Close()

	labels := make([]Label, 0)
	for rows.Next() {
		var l Label
		if err := rows.Scan(&l.ID, &l.Name, &l.Color); err != nil {
			return nil, apperr.Internal("failed to scan label")
		}
		labels = append(labels, l)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list labels")
	}

	return labels, nil
}

func (r *PostgresRepository) Update(ctx context.Context, labelID string, input UpdateInput) (Label, error) {
	sets := make([]string, 0, 2)
	args := make([]any, 0, 3)
	argN := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argN))
		args = append(args, strings.TrimSpace(*input.Name))
		argN++
	}
	if input.Color != nil {
		sets = append(sets, fmt.Sprintf("color = $%d", argN))
		args = append(args, *input.Color)
		argN++
	}

	if len(sets) == 0 {
		return r.get(ctx, labelID)
	}

	args = append(args, labelID)
	q := fmt.Sprintf(
		"UPDATE document_labels SET %s WHERE id = $%d RETURNING id, name, color",
		strings.Join(sets, ", "), argN,
	)

	var l Label
	err := r.db.QueryRowContext(ctx, q, args...).Scan(&l.ID, &l.Name, &l.Color)
	if errors.Is(err, sql.ErrNoRows) {
		return Label{}, apperr.NotFound("label not found")
	}
	if isUniqueViolation(err) {
		return Label{}, apperr.InvalidInput("a label with this name already exists")
	}
	if err != nil {
		return Label{}, apperr.Internal("failed to update label")
	}
	return l, nil
}

func (r *PostgresRepository) get(ctx context.Context, labelID string) (Label, error) {
	const q = `SELECT id, name, color FROM document_labels WHERE id = $1`

	var l Label
	err := r.db.QueryRowContext(ctx, q, labelID).Scan(&l.ID, &l.Name, &l.Color)
	if errors.Is(err, sql.ErrNoRows) {
		return Label{}, apperr.NotFound("label not found")
	}
	if err != nil {
		return Label{}, apperr.Internal("failed to get label")
	}
	return l, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, labelID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM document_labels WHERE id = $1`, labelID)
	if err != nil {
		return apperr.Internal("failed to delete label")
	}

	n, err := res.RowsAffected()
	if err != nil {
		return apperr.Internal("failed to delete label")
	}
	if n == 0 {
		return apperr.NotFound("label not found")
	}
	return nil
}

// isUniqueViolation reports whether err is a Postgres unique-constraint
// violation (SQLSTATE 23505) — document_labels.name is unique, so this is
// how a duplicate name surfaces as a 400 instead of a raw 500.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
