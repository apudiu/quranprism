// Package auth implements the credential lifecycle:
//   - signup + email verification (ACC-1, ACC-2)
//   - login + refresh + logout (ACC-5)
//   - password reset + change-password
//   - lockout policy (ACC-6)
//
// User-table reads/writes are delegated to the user module. The auth
// module owns tokens, refresh sessions, and login_attempts.
package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost is the work factor used for password hashing. Per ACC-4 we
// stay at 12 minimum; bump it together with hardware capacity, never
// silently below.
const bcryptCost = 12

// HashPassword returns the bcrypt hash of plain.
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", fmt.Errorf("auth: empty password")
	}
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("auth: hash: %w", err)
	}
	return string(b), nil
}

// VerifyPassword reports whether plain matches the stored bcrypt hash.
// Returns (true, nil) on match, (false, nil) on mismatch, and a non-nil
// error only for unexpected failures (malformed hash, allocator OOM, ...).
func VerifyPassword(plain, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	return false, fmt.Errorf("auth: verify: %w", err)
}
