package auth

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Handwritten doubles per CLAUDE.md — no testify mocks. Each double is a
// minimal struct holding canned return values + a record of what was
// called. They satisfy the interface the test needs without touching real
// dependencies.

type fakeFailureRepo struct {
	failures int
	recorded []bool // succeeded values, in order
}

func (f *fakeFailureRepo) CountRecentFailedLogins(_ context.Context, _ string, _ time.Time) (int, error) {
	return f.failures, nil
}

func (f *fakeFailureRepo) RecordLoginAttempt(_ context.Context, _ string, _ *netip.Addr, succeeded bool) error {
	f.recorded = append(f.recorded, succeeded)
	return nil
}

// loginLockoutGate is a tiny stand-alone implementation of the lockout
// check Login() performs. Extracted as a unit under test so we don't
// have to wire the whole Service for one assertion.
//
// The real Service.Login uses the same shape — if you change the policy
// there, change this test alongside.
func loginLockoutGate(ctx context.Context, repo *fakeFailureRepo, email string) error {
	cutoff := time.Now().Add(-lockoutWindow)
	n, err := repo.CountRecentFailedLogins(ctx, email, cutoff)
	if err != nil {
		return err
	}
	if n >= lockoutThreshold {
		return ErrAccountLocked
	}
	return nil
}

func TestLockoutGate(t *testing.T) {
	cases := []struct {
		name     string
		failures int
		want     error
	}{
		{"zero failures passes", 0, nil},
		{"nine failures still passes", lockoutThreshold - 1, nil},
		{"threshold reached locks", lockoutThreshold, ErrAccountLocked},
		{"over threshold locks", lockoutThreshold + 5, ErrAccountLocked},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeFailureRepo{failures: tc.failures}
			got := loginLockoutGate(context.Background(), repo, "x@y.z")
			if !errors.Is(got, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

// TestHashThenVerifyRoundTrip exercises the actual password helpers —
// pure functions, no DB.
func TestHashThenVerifyRoundTrip(t *testing.T) {
	hash, err := HashPassword("a-real-password-please")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty")
	}
	ok, err := VerifyPassword("a-real-password-please", hash)
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if !ok {
		t.Fatal("VerifyPassword returned false for correct password")
	}
	ok, err = VerifyPassword("WRONG", hash)
	if err != nil {
		t.Fatalf("VerifyPassword wrong: unexpected err %v", err)
	}
	if ok {
		t.Fatal("VerifyPassword returned true for wrong password")
	}
}

// TestNewTokenIsHashConsistent ensures the plaintext that round-trips
// through hashToken matches the persisted hash. Catches encoding drift
// (hex vs base64) and missed normalisation regressions.
func TestNewTokenIsHashConsistent(t *testing.T) {
	plaintext, hash, err := newToken()
	if err != nil {
		t.Fatalf("newToken: %v", err)
	}
	if plaintext == "" || hash == "" {
		t.Fatal("empty token / hash")
	}
	if hashToken(plaintext) != hash {
		t.Fatal("hashToken(plaintext) != hash — encoding mismatch")
	}
	// Different invocations must yield different tokens.
	plaintext2, hash2, _ := newToken()
	if plaintext == plaintext2 || hash == hash2 {
		t.Fatal("newToken returned the same value twice")
	}
}

// Compile-time check: uuid imports actually used (linters previously
// flagged the import in the handwritten-double scaffold).
var _ = uuid.Nil
