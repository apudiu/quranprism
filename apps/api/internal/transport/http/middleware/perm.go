package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/apudiu/quranprism/api/internal/transport/http/httperr"
)

// PermissionLister is the contract RequirePermission needs from the
// acl module. Defined here so middleware doesn't import acl directly —
// keeps the transport → domain dependency arrow one-way (mirrors the
// EmailVerifier pattern above).
type PermissionLister interface {
	ListPermissionsForUser(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type permsKey struct{}

// permsFromContext returns a previously cached perm slice if a prior
// RequirePermission on the same request already fetched it. Returns
// nil + false if no cache yet.
func permsFromContext(ctx context.Context) ([]string, bool) {
	v, ok := ctx.Value(permsKey{}).([]string)
	return v, ok
}

// RequirePermission rejects with 403 / `insufficient_permissions` when
// the authenticated user does not hold the named permission. Must be
// chained AFTER AuthRequired — calls MustIdentity, which panics if
// Identity is missing (programming error, never a 401).
//
// The user's effective perm slice is fetched once per request and
// cached in context so chained RequirePermission calls on the same
// request don't re-query the DB.
func RequirePermission(p PermissionLister, log *slog.Logger, perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := MustIdentity(r.Context())

			perms, cached := permsFromContext(r.Context())
			if !cached {
				uid, err := uuid.Parse(id.UserID)
				if err != nil {
					// Identity carries an opaque user id; if it's not a UUID
					// the JWT was minted wrong. Treat as 401.
					httperr.Write(w, log, httperr.Unauthorized())
					return
				}
				perms, err = p.ListPermissionsForUser(r.Context(), uid)
				if err != nil {
					httperr.Write(w, log, err)
					return
				}
				ctx := context.WithValue(r.Context(), permsKey{}, perms)
				r = r.WithContext(ctx)
			}

			for _, have := range perms {
				if have == perm {
					next.ServeHTTP(w, r)
					return
				}
			}

			httperr.Write(w, log, httperr.New(
				http.StatusForbidden,
				"insufficient_permissions",
				"you do not have permission to perform this action",
			))
		})
	}
}
