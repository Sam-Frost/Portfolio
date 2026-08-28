package db

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestCollapse(t *testing.T) {
	got := collapse("SELECT *\n\tFROM todos\n\tWHERE id = $1")
	if got != "SELECT * FROM todos WHERE id = $1" {
		t.Fatalf("collapse = %q", got)
	}
	long := collapse(strings.Repeat("x", 400))
	if len([]rune(long)) != 301 || !strings.HasSuffix(long, "…") {
		t.Fatalf("collapse did not truncate: len=%d", len([]rune(long)))
	}
}

func TestSlowQueryThreshold(t *testing.T) {
	t.Setenv("DB_SLOW_QUERY_MS", "")
	if slowQueryThreshold() != 200*time.Millisecond {
		t.Fatalf("default = %v", slowQueryThreshold())
	}
	t.Setenv("DB_SLOW_QUERY_MS", "50")
	if slowQueryThreshold() != 50*time.Millisecond {
		t.Fatalf("override = %v", slowQueryThreshold())
	}
	t.Setenv("DB_SLOW_QUERY_MS", "garbage")
	if slowQueryThreshold() != 200*time.Millisecond {
		t.Fatalf("bad value should fall back: %v", slowQueryThreshold())
	}
}

func TestQueryTracerLogsError(t *testing.T) {
	buf := &bytes.Buffer{}
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	defer slog.SetDefault(prev)

	tr := queryTracer{slow: 200 * time.Millisecond}
	ctx := tr.TraceQueryStart(context.Background(), nil, pgx.TraceQueryStartData{SQL: "SELECT 1"})
	tr.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{
		CommandTag: pgconn.CommandTag{},
		Err:        errors.New("connection reset"),
	})

	out := buf.String()
	if !strings.Contains(out, `"level":"ERROR"`) || !strings.Contains(out, "db query failed") || !strings.Contains(out, "connection reset") {
		t.Fatalf("error not logged as expected: %s", out)
	}
}
