package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// tokenByteLen is the entropy of every token the auth module emits
// (verification, password reset, refresh). 32 bytes = 256 bits, base64-
// url encodes to 43 chars without padding.
const tokenByteLen = 32

// newToken generates a fresh random token and returns the plaintext
// (sent to the user via email or cookie) together with its SHA-256 hex
// digest (persisted in Postgres). Plaintext is never persisted.
func newToken() (plaintext, hash string, err error) {
	buf := make([]byte, tokenByteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("auth: rand: %w", err)
	}
	plaintext = hex.EncodeToString(buf)
	hash = hashToken(plaintext)
	return plaintext, hash, nil
}

// hashToken returns the canonical SHA-256 digest of plaintext. Used both
// when persisting on token issue and when looking up a presented token.
func hashToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}
