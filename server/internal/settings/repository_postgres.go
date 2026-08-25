package settings

import (
	"context"
	"database/sql"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

// singletonID is the fixed row id for the one settings record; the
// migration seeds it so Get/Update never have to special-case a missing row.
const singletonID = "singleton"

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Get(ctx context.Context) (Settings, error) {
	const q = `SELECT total_work_hours_required FROM settings WHERE id = $1`

	var s Settings
	err := r.db.QueryRowContext(ctx, q, singletonID).Scan(&s.DailyWorkTracker.TotalWorkHoursRequired)
	if err != nil {
		return Settings{}, apperr.Internal("failed to get settings")
	}
	return s, nil
}

func (r *PostgresRepository) Update(ctx context.Context, input UpdateInput) (Settings, error) {
	if input.DailyWorkTracker == nil {
		return r.Get(ctx)
	}

	const q = `
		UPDATE settings
		SET total_work_hours_required = $1
		WHERE id = $2
		RETURNING total_work_hours_required
	`

	var s Settings
	err := r.db.QueryRowContext(ctx, q, input.DailyWorkTracker.TotalWorkHoursRequired, singletonID).
		Scan(&s.DailyWorkTracker.TotalWorkHoursRequired)
	if err != nil {
		return Settings{}, apperr.Internal("failed to update settings")
	}
	return s, nil
}
