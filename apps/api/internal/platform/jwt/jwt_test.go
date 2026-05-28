package jwt

import (
	"errors"
	"testing"
	"time"

	"github.com/apudiu/quranprism/api/internal/platform/config"
)

func mkService(t *testing.T, secrets ...string) *Service {
	t.Helper()
	svc, err := New(&config.Config{
		JWT: config.JWTConfig{
			Secrets:   secrets,
			AccessTTL: 15 * time.Minute,
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return svc
}

func TestSignVerifyRoundtrip(t *testing.T) {
	svc := mkService(t, "secret-A")
	issued, err := svc.Sign("user-1")
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	sub, err := svc.Verify(issued.Token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if sub != "user-1" {
		t.Fatalf("subject: got %q, want %q", sub, "user-1")
	}
	if issued.ExpiresAt.Before(time.Now()) {
		t.Fatalf("ExpiresAt in the past: %v", issued.ExpiresAt)
	}
}

func TestVerifyAcceptsPreviousSecret(t *testing.T) {
	// A token signed under the previous current-secret still verifies
	// after rotation: we prepend "secret-B" but keep "secret-A".
	oldSvc := mkService(t, "secret-A")
	issued, err := oldSvc.Sign("user-1")
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	rotated := mkService(t, "secret-B", "secret-A")
	sub, err := rotated.Verify(issued.Token)
	if err != nil {
		t.Fatalf("Verify after rotation: %v", err)
	}
	if sub != "user-1" {
		t.Fatalf("subject after rotation: got %q", sub)
	}
}

func TestVerifyRejectsUnknownSecret(t *testing.T) {
	signed := mkService(t, "secret-A")
	issued, _ := signed.Sign("user-1")

	stranger := mkService(t, "secret-C")
	if _, err := stranger.Verify(issued.Token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyRejectsExpired(t *testing.T) {
	svc, err := New(&config.Config{
		JWT: config.JWTConfig{
			Secrets:   []string{"secret-A"},
			AccessTTL: -time.Second, // already expired
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	issued, _ := svc.Sign("user-1")
	if _, err := svc.Verify(issued.Token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for expired, got %v", err)
	}
}

func TestVerifyRejectsAlgNone(t *testing.T) {
	svc := mkService(t, "secret-A")
	// jwt.io-style "alg":"none" token. Header: {"alg":"none","typ":"JWT"};
	// payload: {"sub":"user-1"}.
	const algNoneToken = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1c2VyLTEifQ."
	if _, err := svc.Verify(algNoneToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for alg=none, got %v", err)
	}
}
