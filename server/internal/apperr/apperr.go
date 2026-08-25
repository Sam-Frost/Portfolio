// Package apperr defines domain-level error kinds shared across features.
// It knows nothing about HTTP — internal/httpx is what maps a Kind to a
// status code — so the same errors stay meaningful if a feature is ever
// driven by something other than the REST handlers (a CLI, a gRPC layer,
// tests calling the service directly).
package apperr

type Kind int

const (
	KindInternal Kind = iota
	KindInvalidInput
	KindNotFound
	KindUnauthorized
)

type Error struct {
	Kind    Kind
	Message string
}

func (e *Error) Error() string { return e.Message }

func InvalidInput(msg string) *Error { return &Error{Kind: KindInvalidInput, Message: msg} }
func NotFound(msg string) *Error     { return &Error{Kind: KindNotFound, Message: msg} }
func Internal(msg string) *Error     { return &Error{Kind: KindInternal, Message: msg} }
func Unauthorized(msg string) *Error { return &Error{Kind: KindUnauthorized, Message: msg} }
