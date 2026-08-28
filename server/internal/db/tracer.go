package db

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// slowQueryThreshold reads DB_SLOW_QUERY_MS (default 200ms). A value of 0
// disables the "slow db query" WARN line.
func slowQueryThreshold() time.Duration {
	ms := 200
	if v := os.Getenv("DB_SLOW_QUERY_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			ms = n
		}
	}
	return time.Duration(ms) * time.Millisecond
}

// queryTracer is a pgx.QueryTracer that emits one structured log line per
// SQL statement. It logs through slog.Default(), so the request_id that
// internal/reqlog stashes on the context (and injects via its slog handler)
// rides along automatically — a query is traceable back to the HTTP request
// that issued it without this package importing reqlog.
//
// Levels keep normal operation quiet:
//   - success            -> DEBUG (silent unless LOG_LEVEL=debug)
//   - slower than slowMS  -> WARN
//   - error               -> ERROR
//
// Statement arguments are never logged — only their count — so credentials,
// note bodies, diary text etc. don't leak into logs.
type queryTracer struct {
	slow time.Duration
}

type traceKey struct{}

type traceCtx struct {
	start time.Time
	sql   string
}

func (t queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, traceKey{}, &traceCtx{start: time.Now(), sql: data.SQL})
}

func (t queryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	tc, ok := ctx.Value(traceKey{}).(*traceCtx)
	if !ok {
		return
	}
	elapsed := time.Since(tc.start)

	attrs := []slog.Attr{
		slog.String("sql", collapse(tc.sql)),
		slog.Duration("duration", elapsed),
	}

	level := slog.LevelDebug
	msg := "db query"
	switch {
	case data.Err != nil:
		level, msg = slog.LevelError, "db query failed"
		attrs = append(attrs, slog.String("error", data.Err.Error()))
	case t.slow > 0 && elapsed >= t.slow:
		level, msg = slog.LevelWarn, "slow db query"
		attrs = append(attrs, slog.Int64("rows", data.CommandTag.RowsAffected()))
	default:
		attrs = append(attrs, slog.Int64("rows", data.CommandTag.RowsAffected()))
	}

	slog.Default().LogAttrs(ctx, level, msg, attrs...)
}

// collapse squeezes a multi-line SQL statement onto one line and trims it so
// a log line stays readable.
func collapse(sql string) string {
	const max = 300
	s := strings.Join(strings.Fields(sql), " ")
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}
