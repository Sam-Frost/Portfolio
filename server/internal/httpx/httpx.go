// Package httpx holds HTTP response helpers shared by every feature handler,
// so each new feature (settings, credentials, ...) writes JSON responses and
// maps errors to status codes the same way without copy-pasting the
// boilerplate into its own handler.go.
package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

// errorRecorder is implemented by the request-scoped ResponseWriter wrapper
// (internal/reqlog). WriteError hands it the real error so the per-request
// log line can carry the server-side cause that the client response hides —
// without every handler having to thread a logger or a request in.
type errorRecorder interface{ RecordError(error) }

// WriteError maps err to an HTTP status and writes {"error": "..."}.
// Any *apperr.Error maps by its Kind; anything else (an error a service
// forgot to wrap, a genuinely unexpected failure) is a 500 with a generic
// message so internals never leak to the client.
func WriteError(w http.ResponseWriter, err error) {
	if rec, ok := w.(errorRecorder); ok {
		rec.RecordError(err)
	}

	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		WriteJSON(w, statusForKind(appErr.Kind), map[string]string{"error": appErr.Message})
		return
	}
	WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

func statusForKind(k apperr.Kind) int {
	switch k {
	case apperr.KindInvalidInput:
		return http.StatusBadRequest
	case apperr.KindNotFound:
		return http.StatusNotFound
	case apperr.KindUnauthorized:
		return http.StatusUnauthorized
	case apperr.KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// DecodeJSON decodes the request body into dst, returning an
// apperr.InvalidInput on malformed JSON so handlers can pass it straight to
// WriteError.
func DecodeJSON(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return apperr.InvalidInput("malformed request body")
	}
	return nil
}
