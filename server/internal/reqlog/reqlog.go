// Package reqlog is the shared "request logging" concern referred to in
// CLAUDE.md — one place, not copy-pasted per feature. It gives every HTTP
// request a trace ID, emits one structured log line per request (method,
// path, status, duration, bytes, and — for failures — the server-side error
// that the client response hides), turns a panic into a logged 500 instead
// of a dropped connection, and hands handlers/services a request-scoped
// slog.Logger whose lines carry the same request_id.
//
// Environment:
//
//	LOG_LEVEL   debug | info | warn | error   (default info)
//	LOG_FORMAT  text | json                    (default text; use json in prod)
package reqlog

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Sam-Frost/portfolio/internal/id"
)

// HeaderRequestID is the request/response header carrying the trace ID. An
// inbound value (from a trusted proxy or a client stitching its own logs) is
// honoured; otherwise one is generated.
const HeaderRequestID = "X-Request-Id"

type ctxKey int

const (
	loggerKey ctxKey = iota
	requestIDKey
)

// Init installs a process-wide slog default logger that stamps request_id
// onto every record made with a request context (see FromContext). Call it
// once, early in main. Returns the logger for convenience.
func Init() *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}
	var base slog.Handler
	if strings.EqualFold(os.Getenv("LOG_FORMAT"), "json") {
		base = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		base = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(&contextHandler{Handler: base})
	slog.SetDefault(logger)
	return logger
}

// contextHandler copies the request_id stashed in the context by Middleware
// onto every record, so a plain slog.InfoContext(ctx, ...) from deep inside a
// service is automatically correlated to its request without threading a
// logger through every call.
type contextHandler struct{ slog.Handler }

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if rid, ok := ctx.Value(requestIDKey).(string); ok && rid != "" {
		r.AddAttrs(slog.String("request_id", rid))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithGroup(name)}
}

// RequestID returns the trace ID assigned to ctx's request, or "" outside a
// request.
func RequestID(ctx context.Context) string {
	rid, _ := ctx.Value(requestIDKey).(string)
	return rid
}

// FromContext returns the request-scoped logger (already tagged with method
// and path; request_id is added automatically on every context-aware call).
// Use the ...Context methods so the correlation lands:
//
//	reqlog.FromContext(ctx).InfoContext(ctx, "published", "version", v)
//
// Falls back to slog.Default() outside a request.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// Middleware is the outermost HTTP wrapper: it owns the trace ID, the
// per-request log line, and panic recovery for everything inside it
// (CORS, auth, and every feature handler).
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rid := r.Header.Get(HeaderRequestID)
		if rid == "" {
			rid = id.New()
		}

		logger := slog.Default().With(
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		ctx := context.WithValue(r.Context(), requestIDKey, rid)
		ctx = context.WithValue(ctx, loggerKey, logger)
		r = r.WithContext(ctx)

		w.Header().Set(HeaderRequestID, rid)
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		defer func() {
			if rec := recover(); rec != nil {
				logger.LogAttrs(ctx, slog.LevelError, "request panic",
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
				)
				if !sw.wroteHeader {
					sw.WriteHeader(http.StatusInternalServerError)
				}
			}

			attrs := []slog.Attr{
				slog.Int("status", sw.status),
				slog.Int("bytes", sw.bytes),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote_addr", clientIP(r)),
			}
			if sw.err != nil {
				attrs = append(attrs, slog.String("error", sw.err.Error()))
			}

			msg, level := "request completed", slog.LevelInfo
			switch {
			case sw.status >= 500:
				msg, level = "request failed", slog.LevelError
			case sw.status >= 400:
				msg, level = "request rejected", slog.LevelWarn
			}
			logger.LogAttrs(ctx, level, msg, attrs...)
		}()

		next.ServeHTTP(sw, r)
	})
}

// statusWriter records what the response ended up being so the completion
// line can report it. It stays transparent to http.ResponseController
// (Flush, deadlines, …) via Unwrap, which the local document blob store's
// http.ServeContent path relies on.
type statusWriter struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
	err         error
}

func (w *statusWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.status = code
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// RecordError is the optional hook httpx.WriteError calls so the completion
// line carries the real server-side error even though the client response
// deliberately hides it. No-op when the response isn't wrapped (e.g. tests).
func (w *statusWriter) RecordError(err error) { w.err = err }

// Unwrap exposes the underlying writer to http.ResponseController.
func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		first, _, _ := strings.Cut(fwd, ",")
		return strings.TrimSpace(first)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
