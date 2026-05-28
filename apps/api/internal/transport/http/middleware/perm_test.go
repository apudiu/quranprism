package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// fakeLister records calls so the per-request cache assertion works.
type fakeLister struct {
	perms []string
	err   error
	calls int
}

func (f *fakeLister) ListPermissionsForUser(_ context.Context, _ uuid.UUID) ([]string, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.perms, nil
}

func nopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func withIdentity(r *http.Request, uid string) *http.Request {
	ctx := context.WithValue(r.Context(), identityKey{}, Identity{UserID: uid})
	return r.WithContext(ctx)
}

func TestRequirePermission_Allow(t *testing.T) {
	uid := uuid.New().String()
	fl := &fakeLister{perms: []string{"group:view", "group:create"}}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	h := RequirePermission(fl, nopLogger(), "group:view")(next)
	rr := httptest.NewRecorder()
	r := withIdentity(httptest.NewRequest("GET", "/", nil), uid)
	h.ServeHTTP(rr, r)

	if !called {
		t.Fatal("next handler should have been called")
	}
	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rr.Code)
	}
}

func TestRequirePermission_Deny(t *testing.T) {
	uid := uuid.New().String()
	fl := &fakeLister{perms: []string{"group:view"}}

	called := false
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { called = true })

	h := RequirePermission(fl, nopLogger(), "group:delete")(next)
	rr := httptest.NewRecorder()
	r := withIdentity(httptest.NewRequest("GET", "/", nil), uid)
	h.ServeHTTP(rr, r)

	if called {
		t.Fatal("next should not be called on deny")
	}
	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rr.Code)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("body parse: %v", err)
	}
	if body.Error.Code != "insufficient_permissions" {
		t.Errorf("code = %q, want insufficient_permissions", body.Error.Code)
	}
}

func TestRequirePermission_Cached(t *testing.T) {
	uid := uuid.New().String()
	fl := &fakeLister{perms: []string{"group:view", "group:create"}}

	// Chain two RequirePermission middlewares — the second should hit
	// the cache, not the lister.
	final := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	chain := RequirePermission(fl, nopLogger(), "group:view")(
		RequirePermission(fl, nopLogger(), "group:create")(final),
	)

	rr := httptest.NewRecorder()
	r := withIdentity(httptest.NewRequest("GET", "/", nil), uid)
	chain.ServeHTTP(rr, r)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rr.Code)
	}
	if fl.calls != 1 {
		t.Errorf("lister calls = %d, want 1 (second perm check should hit cache)", fl.calls)
	}
}

func TestRequirePermission_NoIdentity_Panics(t *testing.T) {
	fl := &fakeLister{perms: []string{}}
	h := RequirePermission(fl, nopLogger(), "group:view")(http.NotFoundHandler())

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when Identity missing from context")
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
}

func TestRequirePermission_ListerError_500(t *testing.T) {
	uid := uuid.New().String()
	fl := &fakeLister{err: errors.New("db down")}

	h := RequirePermission(fl, nopLogger(), "group:view")(http.NotFoundHandler())
	rr := httptest.NewRecorder()
	r := withIdentity(httptest.NewRequest("GET", "/", nil), uid)
	h.ServeHTTP(rr, r)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

func TestRequirePermission_BadUserID_401(t *testing.T) {
	fl := &fakeLister{perms: []string{}}
	h := RequirePermission(fl, nopLogger(), "group:view")(http.NotFoundHandler())
	rr := httptest.NewRecorder()
	r := withIdentity(httptest.NewRequest("GET", "/", nil), "not-a-uuid")
	h.ServeHTTP(rr, r)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}
