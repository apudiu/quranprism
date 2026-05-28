package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/apudiu/quranprism/api/internal/platform/jwt"
	"github.com/apudiu/quranprism/api/internal/transport/http/httperr"
)

// Identity is the authenticated principal extracted from the access
// token. Handlers read it via IdentityFromContext (cheap O(1) lookup,
// no DB hit). Domain modules that need the full user record load it
// from the user service themselves.
type Identity struct {
	UserID string // user UUID, stringified
}

type identityKey struct{}

// IdentityFromContext returns the principal attached by AuthRequired.
// Handlers should treat the second-return false as a programming
// mistake — AuthRequired wouldn't have called next without it.
func IdentityFromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey{}).(Identity)
	return id, ok
}

// MustIdentity panics when the context has no Identity. Use inside
// handlers behind AuthRequired so the missing-identity case is a clear
// programmer error rather than a 401 leaking from a misconfigured route.
func MustIdentity(ctx context.Context) Identity {
	id, ok := IdentityFromContext(ctx)
	if !ok {
		panic("middleware: missing identity in context — route not behind AuthRequired?")
	}
	return id
}

// AuthRequired validates the `Authorization: Bearer <jwt>` header and
// attaches the resulting Identity to the request context. Rejects with
// 401 on any failure (missing header, wrong scheme, bad token).
func AuthRequired(jwtSvc *jwt.Service, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(authz, prefix) {
				httperr.Write(w, log, httperr.Unauthorized())
				return
			}
			token := strings.TrimPrefix(authz, prefix)
			sub, err := jwtSvc.Verify(token)
			if err != nil {
				httperr.Write(w, log, httperr.Unauthorized())
				return
			}
			ctx := context.WithValue(r.Context(), identityKey{}, Identity{UserID: sub})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// EmailVerifier is the small contract EmailVerifiedRequired needs from
// the user module. Defined here so middleware doesn't import modules
// directly and keeps the transport → domain dependency arrow one-way.
type EmailVerifier interface {
	IsEmailVerified(ctx context.Context, userID string) (bool, error)
}

// EmailVerifiedRequired rejects with 403 when the authenticated user has
// not verified their email yet. Must be chained AFTER AuthRequired.
func EmailVerifiedRequired(v EmailVerifier, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := MustIdentity(r.Context())
			ok, err := v.IsEmailVerified(r.Context(), id.UserID)
			if err != nil {
				httperr.Write(w, log, err)
				return
			}
			if !ok {
				e := httperr.New(http.StatusForbidden, "email_not_verified", "email address not verified")
				httperr.Write(w, log, e)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
