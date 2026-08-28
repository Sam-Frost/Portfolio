// Package db owns the Postgres connection and a minimal, dependency-free
// migration runner. There's exactly one table today (todos), so this is
// deliberately not a full migration framework — just enough to apply the
// .sql files under migrations/ once, in order, tracked in a
// schema_migrations table.
package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Connect opens the Postgres pool with a query tracer attached, so every SQL
// statement is logged (errors and slow queries always; all queries when
// LOG_LEVEL=debug) and correlated to its HTTP request via the request_id on
// the context. DB_SLOW_QUERY_MS tunes the slow-query threshold (default 200,
// 0 disables the WARN line).
func Connect(ctx context.Context, databaseURL string) (*sql.DB, error) {
	connConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	connConfig.Tracer = queryTracer{slow: slowQueryThreshold()}

	sqlDB := sql.OpenDB(stdlib.GetConnector(*connConfig))

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return sqlDB, nil
}

// Migrate applies every migration under migrations/ that isn't already
// recorded in schema_migrations, in filename order, each in its own
// transaction.
func Migrate(ctx context.Context, sqlDB *sql.DB) error {
	if _, err := sqlDB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		if err := applyMigration(ctx, sqlDB, name); err != nil {
			return err
		}
	}

	return nil
}

func applyMigration(ctx context.Context, sqlDB *sql.DB, name string) error {
	var applied bool
	err := sqlDB.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, name,
	).Scan(&applied)
	if err != nil {
		return fmt.Errorf("check migration %s: %w", name, err)
	}
	if applied {
		return nil
	}

	contents, err := migrationsFS.ReadFile("migrations/" + name)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", name, err)
	}

	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
		return fmt.Errorf("apply migration %s: %w", name, err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
		return fmt.Errorf("record migration %s: %w", name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}

	return nil
}
