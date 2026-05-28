// Package middleware holds the cross-cutting chi middleware the api uses.
//
// Anything that decorates every request (request logging, CORS, recovery,
// rate limiting, auth) lives here. Domain-specific middleware stays in
// the module that owns it.
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Logger emits one slog line per request after the response is written.
// Status, byte count, and elapsed time come from chi's WrapResponseWriter
// so we don't have to wrap it ourselves.
func Logger(base *slog.Logger) func(http.Handler) http.Handler {
	log := base.With("subsystem", "http")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"elapsed_ms", time.Since(start).Milliseconds(),
				"request_id", chimiddleware.GetReqID(r.Context()),
				"remote", r.RemoteAddr,
			)
		})
	}
}
