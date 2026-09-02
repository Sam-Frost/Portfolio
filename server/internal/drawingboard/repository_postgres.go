package drawingboard

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

func (r *PostgresRepository) Create(ctx context.Context, b Board) (Board, error) {
	b.ID = id.New()

	const q = `INSERT INTO drawing_boards (id, name, scene_data, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	if _, err := r.db.ExecContext(ctx, q, b.ID, b.Name, []byte(b.SceneData), b.CreatedAt, b.UpdatedAt); err != nil {
		return Board{}, apperr.Internal("failed to create board")
	}
	return b, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]BoardSummary, error) {
	const q = `SELECT id, name, created_at, updated_at
		FROM drawing_boards
		WHERE deleted_at IS NULL
		ORDER BY updated_at DESC`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, apperr.Internal("failed to list boards")
	}
	defer rows.Close()

	summaries := make([]BoardSummary, 0)
	for rows.Next() {
		var s BoardSummary
		if err := rows.Scan(&s.ID, &s.Name, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, apperr.Internal("failed to scan board")
		}
		summaries = append(summaries, s)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list boards")
	}

	return summaries, nil
}

func (r *PostgresRepository) Get(ctx context.Context, boardID string) (Board, error) {
	const q = `SELECT id, name, scene_data, created_at, updated_at
		FROM drawing_boards WHERE id = $1 AND deleted_at IS NULL`

	var b Board
	var sceneData []byte
	err := r.db.QueryRowContext(ctx, q, boardID).Scan(&b.ID, &b.Name, &sceneData, &b.CreatedAt, &b.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Board{}, apperr.NotFound("board not found")
	}
	if err != nil {
		return Board{}, apperr.Internal("failed to get board")
	}
	b.SceneData = sceneData
	return b, nil
}

func (r *PostgresRepository) Update(ctx context.Context, boardID string, input UpdateInput) (Board, error) {
	sets := make([]string, 0, 3)
	args := make([]any, 0, 4)
	argN := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argN))
		args = append(args, strings.TrimSpace(*input.Name))
		argN++
	}
	if input.SceneData != nil {
		sets = append(sets, fmt.Sprintf("scene_data = $%d", argN))
		args = append(args, []byte(input.SceneData))
		argN++
	}

	if len(sets) == 0 {
		return r.Get(ctx, boardID)
	}

	sets = append(sets, fmt.Sprintf("updated_at = $%d", argN))
	args = append(args, updatedAtNow())
	argN++

	args = append(args, boardID)
	q := fmt.Sprintf(
		"UPDATE drawing_boards SET %s WHERE id = $%d AND deleted_at IS NULL "+
			"RETURNING id, name, scene_data, created_at, updated_at",
		strings.Join(sets, ", "), argN,
	)

	var b Board
	var sceneData []byte
	err := r.db.QueryRowContext(ctx, q, args...).Scan(&b.ID, &b.Name, &sceneData, &b.CreatedAt, &b.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Board{}, apperr.NotFound("board not found")
	}
	if err != nil {
		return Board{}, apperr.Internal("failed to update board")
	}
	b.SceneData = sceneData
	return b, nil
}

// Delete is a soft delete: it stamps deleted_at rather than removing the
// row, so the board drops out of List/Get/Update but its data is retained.
func (r *PostgresRepository) Delete(ctx context.Context, boardID string) error {
	res, err := r.db.ExecContext(
		ctx, `UPDATE drawing_boards SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, updatedAtNow(), boardID,
	)
	if err != nil {
		return apperr.Internal("failed to delete board")
	}

	n, err := res.RowsAffected()
	if err != nil {
		return apperr.Internal("failed to delete board")
	}
	if n == 0 {
		return apperr.NotFound("board not found")
	}
	return nil
}
