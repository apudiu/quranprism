// Package jwt issues and validates the application's HS256-signed access
// tokens.
//
// JWT_SECRETS is a comma-separated rotating list: position 0 is the
// **current** signing secret; every position is accepted on verify. To
// rotate without invalidating live tokens: prepend the new secret, keep
// the old one in place until every issued token has expired
// (`AccessTTL`), then drop it.
package jwt

import (
	"errors"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"

	"github.com/apudiu/quranprism/api/internal/platform/config"
)

// Issued holds the SignedString and its absolute expiry — handy because
// HTTP responses commonly need both ("token plus expires_at").
type Issued struct {
	Token     string
	ExpiresAt time.Time
}

// ErrInvalidToken is returned by Verify for any malformed/expired/wrong-sig
// token. Wrapped so callers can errors.Is it without depending on jwt v5.
var ErrInvalidToken = errors.New("jwt: invalid token")

// Service signs and verifies access tokens.
type Service struct {
	secrets   [][]byte
	accessTTL time.Duration
}

// New parses the configured rotating-secret list and returns a Service.
// Returns an error if no usable secret is configured (config.Load already
// enforces this, but defence in depth).
func New(cfg *config.Config) (*Service, error) {
	secrets := make([][]byte, 0, len(cfg.JWT.Secrets))
	for _, s := range cfg.JWT.Secrets {
		if s != "" {
			secrets = append(secrets, []byte(s))
		}
	}
	if len(secrets) == 0 {
		return nil, fmt.Errorf("jwt: no signing secrets configured")
	}
	return &Service{secrets: secrets, accessTTL: cfg.JWT.AccessTTL}, nil
}

// Sign issues a new access token for the given subject (the user's UUID
// stringified). Returns the signed JWT and its absolute expiry.
func (s *Service) Sign(subject string) (Issued, error) {
	now := time.Now().UTC()
	exp := now.Add(s.accessTTL)

	claims := gojwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  gojwt.NewNumericDate(now),
		ExpiresAt: gojwt.NewNumericDate(exp),
	}
	tok := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(s.secrets[0])
	if err != nil {
		return Issued{}, fmt.Errorf("jwt: sign: %w", err)
	}
	return Issued{Token: signed, ExpiresAt: exp}, nil
}

// Verify parses a token and returns its subject, trying each configured
// secret in order. Returns ErrInvalidToken (wrapping the underlying jwt
// error) when none accept the signature.
func (s *Service) Verify(token string) (string, error) {
	var last error
	for _, secret := range s.secrets {
		secret := secret // capture
		c := &gojwt.RegisteredClaims{}
		_, err := gojwt.ParseWithClaims(token, c, func(t *gojwt.Token) (any, error) {
			if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method %q", t.Header["alg"])
			}
			return secret, nil
		})
		if err == nil {
			return c.Subject, nil
		}
		last = err
	}
	return "", fmt.Errorf("%w: %v", ErrInvalidToken, last)
}
