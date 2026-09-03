package settings

import (
	"context"
	"database/sql"
	"time"

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
	const q = `SELECT total_work_hours_required, time_left_goal_date, time_left_format,
		notification_recipient_email, notification_morning_time,
		notification_email_enabled, notification_push_enabled
		FROM settings WHERE id = $1`

	var s Settings
	var goalDate sql.NullTime
	var recipientEmail sql.NullString
	err := r.db.QueryRowContext(ctx, q, singletonID).Scan(
		&s.DailyWorkTracker.TotalWorkHoursRequired,
		&goalDate,
		&s.TimeLeftClock.Format,
		&recipientEmail,
		&s.Notifications.MorningTime,
		&s.Notifications.EmailEnabled,
		&s.Notifications.PushEnabled,
	)
	if err != nil {
		return Settings{}, apperr.Internal("failed to get settings")
	}
	if goalDate.Valid {
		formatted := goalDate.Time.Format(time.RFC3339)
		s.TimeLeftClock.GoalDate = &formatted
	}
	if recipientEmail.Valid {
		s.Notifications.RecipientEmail = &recipientEmail.String
	}
	return s, nil
}

func (r *PostgresRepository) Update(ctx context.Context, input UpdateInput) (Settings, error) {
	if input.DailyWorkTracker == nil && input.TimeLeftClock == nil && input.Notifications == nil {
		return r.Get(ctx)
	}

	if input.DailyWorkTracker != nil {
		const q = `UPDATE settings SET total_work_hours_required = $1 WHERE id = $2`
		if _, err := r.db.ExecContext(ctx, q, input.DailyWorkTracker.TotalWorkHoursRequired, singletonID); err != nil {
			return Settings{}, apperr.Internal("failed to update settings")
		}
	}

	if input.TimeLeftClock != nil {
		var goalDate any
		if input.TimeLeftClock.GoalDate != nil {
			t, err := time.Parse(time.RFC3339, *input.TimeLeftClock.GoalDate)
			if err != nil {
				return Settings{}, apperr.Internal("failed to update settings")
			}
			goalDate = t
		}

		const q = `UPDATE settings SET time_left_goal_date = $1, time_left_format = $2 WHERE id = $3`
		if _, err := r.db.ExecContext(ctx, q, goalDate, input.TimeLeftClock.Format, singletonID); err != nil {
			return Settings{}, apperr.Internal("failed to update settings")
		}
	}

	if input.Notifications != nil {
		var recipientEmail any
		if input.Notifications.RecipientEmail != nil && *input.Notifications.RecipientEmail != "" {
			recipientEmail = *input.Notifications.RecipientEmail
		}

		const q = `UPDATE settings SET
			notification_recipient_email = $1, notification_morning_time = $2,
			notification_email_enabled = $3, notification_push_enabled = $4
			WHERE id = $5`
		if _, err := r.db.ExecContext(ctx, q, recipientEmail, input.Notifications.MorningTime,
			input.Notifications.EmailEnabled, input.Notifications.PushEnabled, singletonID); err != nil {
			return Settings{}, apperr.Internal("failed to update settings")
		}
	}

	return r.Get(ctx)
}
