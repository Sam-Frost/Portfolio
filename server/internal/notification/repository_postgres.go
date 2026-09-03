package notification

import (
	"context"
	"database/sql"
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

func (r *PostgresRepository) Subscribe(ctx context.Context, sub PushSubscription) error {
	const q = `
		INSERT INTO push_subscriptions (id, endpoint, p256dh, auth, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (endpoint) DO UPDATE SET
			p256dh = EXCLUDED.p256dh,
			auth = EXCLUDED.auth,
			user_agent = EXCLUDED.user_agent`
	if _, err := r.db.ExecContext(ctx, q, id.New(), sub.Endpoint, sub.P256dh, sub.Auth, sub.UserAgent, time.Now().UTC()); err != nil {
		return apperr.Internal("failed to save push subscription")
	}
	return nil
}

func (r *PostgresRepository) Resync(ctx context.Context, oldEndpoint string, sub PushSubscription) error {
	if oldEndpoint != "" && oldEndpoint != sub.Endpoint {
		const q = `UPDATE push_subscriptions
			SET endpoint = $1, p256dh = $2, auth = $3, user_agent = $4
			WHERE endpoint = $5`
		res, err := r.db.ExecContext(ctx, q, sub.Endpoint, sub.P256dh, sub.Auth, sub.UserAgent, oldEndpoint)
		if err != nil {
			return apperr.Internal("failed to resync push subscription")
		}
		if n, _ := res.RowsAffected(); n > 0 {
			return nil
		}
	}
	// No prior row to move (or same endpoint) — fall back to an upsert so the
	// device is registered either way.
	return r.Subscribe(ctx, sub)
}

func (r *PostgresRepository) Unsubscribe(ctx context.Context, endpoint string) error {
	const q = `DELETE FROM push_subscriptions WHERE endpoint = $1`
	if _, err := r.db.ExecContext(ctx, q, endpoint); err != nil {
		return apperr.Internal("failed to remove push subscription")
	}
	return nil
}

func (r *PostgresRepository) ListSubscriptions(ctx context.Context) ([]PushSubscription, error) {
	const q = `SELECT id, endpoint, p256dh, auth, user_agent, created_at FROM push_subscriptions ORDER BY created_at`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, apperr.Internal("failed to list push subscriptions")
	}
	defer rows.Close()

	subs := make([]PushSubscription, 0)
	for rows.Next() {
		var s PushSubscription
		if err := rows.Scan(&s.ID, &s.Endpoint, &s.P256dh, &s.Auth, &s.UserAgent, &s.CreatedAt); err != nil {
			return nil, apperr.Internal("failed to scan push subscription")
		}
		subs = append(subs, s)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list push subscriptions")
	}
	return subs, nil
}

func (r *PostgresRepository) LogExists(ctx context.Context, kind, istDate string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM notification_log WHERE kind = $1 AND ist_date = $2)`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, kind, istDate).Scan(&exists); err != nil {
		return false, apperr.Internal("failed to check notification log")
	}
	return exists, nil
}

func (r *PostgresRepository) InsertLog(ctx context.Context, kind, istDate string) error {
	const q = `
		INSERT INTO notification_log (id, kind, ist_date, sent_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (kind, ist_date) DO NOTHING`
	if _, err := r.db.ExecContext(ctx, q, id.New(), kind, istDate, time.Now().UTC()); err != nil {
		return apperr.Internal("failed to record notification log")
	}
	return nil
}
