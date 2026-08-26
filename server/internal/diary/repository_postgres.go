package diary

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

func (r *PostgresRepository) GetByDate(ctx context.Context, date string) (Entry, error) {
	const q = `SELECT id, entry_date, content, created_at, updated_at FROM diary_entries WHERE entry_date = $1`

	var e Entry
	var d time.Time
	err := r.db.QueryRowContext(ctx, q, date).Scan(&e.ID, &d, &e.Content, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Entry{}, apperr.NotFound("diary entry not found")
	}
	if err != nil {
		return Entry{}, apperr.Internal("failed to get diary entry")
	}
	e.EntryDate = d.Format(EntryDateLayout)
	return e, nil
}

// Upsert relies on entry_date's UNIQUE constraint (see migration
// 0010_add_diary.sql) to do "one entry per day, edited over time" as a
// single round trip instead of a read-then-write.
func (r *PostgresRepository) Upsert(ctx context.Context, date string, content string) (Entry, error) {
	now := time.Now().UTC()
	const q = `
		INSERT INTO diary_entries (id, entry_date, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		ON CONFLICT (entry_date) DO UPDATE SET content = EXCLUDED.content, updated_at = EXCLUDED.updated_at
		RETURNING id, entry_date, content, created_at, updated_at
	`

	var e Entry
	var d time.Time
	err := r.db.QueryRowContext(ctx, q, id.New(), date, content, now).Scan(&e.ID, &d, &e.Content, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return Entry{}, apperr.Internal("failed to save diary entry")
	}
	e.EntryDate = d.Format(EntryDateLayout)
	return e, nil
}

func (r *PostgresRepository) ListDates(ctx context.Context, from, to string) ([]string, error) {
	const q = `SELECT entry_date FROM diary_entries WHERE entry_date BETWEEN $1 AND $2 ORDER BY entry_date`

	rows, err := r.db.QueryContext(ctx, q, from, to)
	if err != nil {
		return nil, apperr.Internal("failed to list diary entries")
	}
	defer rows.Close()

	dates := make([]string, 0)
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return nil, apperr.Internal("failed to scan diary entry")
		}
		dates = append(dates, d.Format(EntryDateLayout))
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list diary entries")
	}

	return dates, nil
}
