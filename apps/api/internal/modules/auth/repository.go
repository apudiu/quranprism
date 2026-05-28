package auth

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
)

// Repository owns the token + session + login_attempts tables. The user
// table is accessed via user.Service.
type Repository struct {
	q *sqlcdb.Queries
}

// NewRepository wires the auth Repository.
func NewRepository(q *sqlcdb.Queries) *Repository { return &Repository{q: q} }

// --- email verification tokens -------------------------------------

// CreateVerificationToken persists a fresh verification token for the
// user. expiresAt is absolute. Returns nothing — caller has the plaintext.
func (r *Repository) CreateVerificationToken(ctx context.Context, userID uuid.UUID, hash string, expiresAt time.Time) error {
	_, err := r.q.CreateEmailVerificationToken(ctx, sqlcdb.CreateEmailVerificationTokenParams{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return fmt.Errorf("auth: create verify token: %w", err)
	}
	return nil
}

// ConsumeVerificationToken validates the presented plaintext, marks the
// row used, and returns the underlying user ID. Returns ErrInvalidToken
// for expired/used/missing rows so the caller doesn't have to enumerate.
func (r *Repository) ConsumeVerificationToken(ctx context.Context, plaintext string) (uuid.UUID, error) {
	row, err := r.q.GetEmailVerificationToken(ctx, hashToken(plaintext))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrInvalidToken
		}
		return uuid.Nil, fmt.Errorf("auth: lookup verify token: %w", err)
	}
	if row.UsedAt.Valid || row.ExpiresAt.Before(time.Now()) {
		return uuid.Nil, ErrInvalidToken
	}
	if err := r.q.ConsumeEmailVerificationToken(ctx, row.ID); err != nil {
		return uuid.Nil, fmt.Errorf("auth: consume verify token: %w", err)
	}
	return row.UserID, nil
}

// --- password reset tokens -----------------------------------------

// CreateResetToken persists a fresh password-reset token.
func (r *Repository) CreateResetToken(ctx context.Context, userID uuid.UUID, hash string, expiresAt time.Time) error {
	_, err := r.q.CreatePasswordResetToken(ctx, sqlcdb.CreatePasswordResetTokenParams{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return fmt.Errorf("auth: create reset token: %w", err)
	}
	return nil
}

// ConsumeResetToken validates + marks used + returns the user ID.
func (r *Repository) ConsumeResetToken(ctx context.Context, plaintext string) (uuid.UUID, error) {
	row, err := r.q.GetPasswordResetToken(ctx, hashToken(plaintext))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrInvalidToken
		}
		return uuid.Nil, fmt.Errorf("auth: lookup reset token: %w", err)
	}
	if row.UsedAt.Valid || row.ExpiresAt.Before(time.Now()) {
		return uuid.Nil, ErrInvalidToken
	}
	if err := r.q.ConsumePasswordResetToken(ctx, row.ID); err != nil {
		return uuid.Nil, fmt.Errorf("auth: consume reset token: %w", err)
	}
	return row.UserID, nil
}

// --- refresh sessions ----------------------------------------------

// CreateRefreshSession records a new session row. The plaintext is
// embedded in the cookie sent to the user; only the hash lands here.
//
// userAgent / ip are best-effort: nil values are fine and stored as NULL.
func (r *Repository) CreateRefreshSession(ctx context.Context, userID uuid.UUID, hash string, userAgent string, ip *netip.Addr, expiresAt time.Time) (uuid.UUID, error) {
	var uaPtr *string
	if userAgent != "" {
		uaPtr = &userAgent
	}
	row, err := r.q.CreateRefreshSession(ctx, sqlcdb.CreateRefreshSessionParams{
		UserID:    userID,
		TokenHash: hash,
		UserAgent: uaPtr,
		Ip:        ip,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("auth: create refresh session: %w", err)
	}
	return row.ID, nil
}

// LookupRefreshSession returns the active session matching the hashed
// presented token. Returns ErrInvalidCredentials for missing/expired/
// revoked rows so refresh failure stays opaque to clients.
func (r *Repository) LookupRefreshSession(ctx context.Context, plaintext string) (sqlcdb.RefreshSession, error) {
	row, err := r.q.GetRefreshSessionByTokenHash(ctx, hashToken(plaintext))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlcdb.RefreshSession{}, ErrInvalidCredentials
		}
		return sqlcdb.RefreshSession{}, fmt.Errorf("auth: lookup refresh session: %w", err)
	}
	return row, nil
}

// TouchRefreshSession bumps last_used_at on a successful refresh.
func (r *Repository) TouchRefreshSession(ctx context.Context, id uuid.UUID) error {
	if err := r.q.TouchRefreshSession(ctx, id); err != nil {
		return fmt.Errorf("auth: touch refresh session: %w", err)
	}
	return nil
}

// RevokeRefreshSession marks a single session revoked (logout / rotation).
func (r *Repository) RevokeRefreshSession(ctx context.Context, id uuid.UUID) error {
	if err := r.q.RevokeRefreshSession(ctx, id); err != nil {
		return fmt.Errorf("auth: revoke refresh session: %w", err)
	}
	return nil
}

// RevokeAllRefreshSessions kicks the user off every device. Used by
// password reset.
func (r *Repository) RevokeAllRefreshSessions(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.RevokeAllRefreshSessionsForUser(ctx, userID); err != nil {
		return fmt.Errorf("auth: revoke all refresh sessions: %w", err)
	}
	return nil
}

// RevokeOtherRefreshSessions revokes every session except the one
// presenting the change-password request, so the active client keeps
// rolling its tokens without re-login.
func (r *Repository) RevokeOtherRefreshSessions(ctx context.Context, userID, keepID uuid.UUID) error {
	if err := r.q.RevokeOtherRefreshSessionsForUser(ctx, sqlcdb.RevokeOtherRefreshSessionsForUserParams{
		UserID: userID,
		ID:     keepID,
	}); err != nil {
		return fmt.Errorf("auth: revoke other refresh sessions: %w", err)
	}
	return nil
}

// --- login attempts ------------------------------------------------

// RecordLoginAttempt appends a row to login_attempts. ip is optional.
func (r *Repository) RecordLoginAttempt(ctx context.Context, email string, ip *netip.Addr, succeeded bool) error {
	if err := r.q.RecordLoginAttempt(ctx, sqlcdb.RecordLoginAttemptParams{
		Email:     email,
		Ip:        ip,
		Succeeded: succeeded,
	}); err != nil {
		return fmt.Errorf("auth: record login attempt: %w", err)
	}
	return nil
}

// CountRecentFailedLogins drives the ACC-6 lockout — returns the count
// of failed attempts on this email since the cutoff.
func (r *Repository) CountRecentFailedLogins(ctx context.Context, email string, since time.Time) (int, error) {
	n, err := r.q.CountRecentFailedLogins(ctx, sqlcdb.CountRecentFailedLoginsParams{
		Email:     email,
		CreatedAt: since,
	})
	if err != nil {
		return 0, fmt.Errorf("auth: count failed logins: %w", err)
	}
	return int(n), nil
}
