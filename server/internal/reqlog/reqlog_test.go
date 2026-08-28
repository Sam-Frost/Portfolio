package reqlog

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestLogger swaps slog.Default for one writing JSON into buf, wrapped in
// the same contextHandler production uses, and returns a restore func.
func newTestLogger(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	prev := slog.Default()
	slog.SetDefault(slog.New(&contextHandler{
		Handler: slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
	}))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return buf
}

func lastLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var m map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &m); err != nil {
		t.Fatalf("parse log line %q: %v", lines[len(lines)-1], err)
	}
	return m
}

func TestMiddlewareAssignsAndEchoesRequestID(t *testing.T) {
	buf := newTestLogger(t)

	var seen string
	h := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/todos", nil))

	if seen == "" {
		t.Fatal("handler saw no request id in context")
	}
	if got := rec.Header().Get(HeaderRequestID); got != seen {
		t.Fatalf("response header %q, want %q", got, seen)
	}
	entry := lastLine(t, buf)
	if entry["msg"] != "request completed" || entry["request_id"] != seen {
		t.Fatalf("completion line = %v", entry)
	}
	if entry["status"].(float64) != http.StatusOK {
		t.Fatalf("status = %v", entry["status"])
	}
}

func TestMiddlewareHonoursInboundRequestID(t *testing.T) {
	newTestLogger(t)

	var seen string
	h := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestID(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(HeaderRequestID, "trace-abc-123")
	h.ServeHTTP(httptest.NewRecorder(), req)

	if seen != "trace-abc-123" {
		t.Fatalf("request id = %q, want inbound value", seen)
	}
}

func TestMiddlewareRecoversPanicAsLogged500(t *testing.T) {
	buf := newTestLogger(t)

	h := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/todos", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(buf.String(), "request panic") {
		t.Fatalf("panic not logged: %s", buf.String())
	}
	if lastLine(t, buf)["level"] != "ERROR" {
		t.Fatalf("completion line not ERROR: %s", buf.String())
	}
}

func TestMiddlewareLogsRecordedError(t *testing.T) {
	buf := newTestLogger(t)

	h := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// mimic httpx.WriteError handing the hidden cause to the wrapper
		w.(interface{ RecordError(error) }).RecordError(context.DeadlineExceeded)
		w.WriteHeader(http.StatusInternalServerError)
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/cms/publish", nil))

	entry := lastLine(t, buf)
	if entry["level"] != "ERROR" || entry["msg"] != "request failed" {
		t.Fatalf("entry = %v", entry)
	}
	if entry["error"] != context.DeadlineExceeded.Error() {
		t.Fatalf("error attr = %v", entry["error"])
	}
}

func TestFromContextFallsBackOutsideRequest(t *testing.T) {
	if FromContext(context.Background()) != slog.Default() {
		t.Fatal("FromContext should fall back to slog.Default() with no request logger")
	}
}
