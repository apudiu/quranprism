// Package httperr maps domain errors onto HTTP status codes + stable
// machine-readable codes.
//
// Domain modules expose typed sentinel errors (e.g. user.ErrNotFound) and
// the handler layer wraps them through Write — keeping HTTP concerns out
// of the domain itself and letting the same business error map to a
// status here when the transport is different (gRPC, queue, etc).
package httperr

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/apudiu/quranprism/api/internal/transport/http/response"
)

// E represents a typed application error. Domain modules either return
// an E directly or return a sentinel that handlers wrap with one.
type E struct {
	Status  int            // HTTP status
	Code    string         // machine-readable code, snake_case
	Message string         // human-readable, safe for users
	Details map[string]any // optional structured details (e.g. validation errors)
	Cause   error          // wrapped for errors.Is/As; never serialised
}

func (e *E) Error() string {
	if e.Cause != nil {
		return e.Code + ": " + e.Cause.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *E) Unwrap() error { return e.Cause }

// New builds an E. Use the convenience constructors below when the
// status / code pair is reused.
func New(status int, code, message string) *E {
	return &E{Status: status, Code: code, Message: message}
}

// Wrap attaches a non-user-facing cause to an E for logging.
func (e *E) Wrap(cause error) *E { e.Cause = cause; return e }

// WithDetails attaches structured detail fields to an E.
func (e *E) WithDetails(d map[string]any) *E { e.Details = d; return e }

// Common 4xx/5xx envelopes. Keep this list lean — most modules will use
// these directly. Add new ones only when the same (status, code) pair
// recurs across modules.
func BadRequest(message string) *E   { return New(http.StatusBadRequest, "bad_request", message) }
func Unauthorized() *E               { return New(http.StatusUnauthorized, "unauthorized", "authentication required") }
func Forbidden() *E                  { return New(http.StatusForbidden, "forbidden", "not allowed") }
func NotFound(what string) *E        { return New(http.StatusNotFound, "not_found", what+" not found") }
func Conflict(message string) *E     { return New(http.StatusConflict, "conflict", message) }
func Unprocessable(message string) *E {
	return New(http.StatusUnprocessableEntity, "unprocessable", message)
}
func TooManyRequests() *E { return New(http.StatusTooManyRequests, "rate_limited", "rate limit exceeded") }
func Locked(message string) *E { return New(http.StatusLocked, "locked", message) }
func Internal() *E              { return New(http.StatusInternalServerError, "internal", "internal server error") }

// Write serialises any error onto the response. Typed *E errors keep
// their status / code; untyped errors are logged at error level and
// rendered as a generic 500 to avoid leaking implementation details.
func Write(w http.ResponseWriter, log *slog.Logger, err error) {
	var e *E
	if errors.As(err, &e) {
		response.Error(w, e.Status, e.Code, e.Message, e.Details)
		return
	}
	log.Error("unhandled error", "err", err)
	response.Error(w, http.StatusInternalServerError, "internal", "internal server error", nil)
}
