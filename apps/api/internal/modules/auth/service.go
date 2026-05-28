package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"strings"
	"time"

	"github.com/google/uuid"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
	"github.com/apudiu/quranprism/api/internal/modules/acl"
	"github.com/apudiu/quranprism/api/internal/modules/user"
	"github.com/apudiu/quranprism/api/internal/platform/config"
	"github.com/apudiu/quranprism/api/internal/platform/jwt"
	"github.com/apudiu/quranprism/api/internal/platform/mailer"
)

// Token lifetimes specific to the auth module. Refresh sessions use
// cfg.JWT.RefreshTTL; verify + reset are fixed to sensible defaults that
// can be overridden later if needed.
const (
	verificationTokenTTL = 24 * time.Hour
	resetTokenTTL        = 1 * time.Hour

	// Lockout policy — ACC-6. 10 failed attempts inside lockoutWindow
	// locks the email.
	lockoutThreshold = 10
	lockoutWindow    = 15 * time.Minute
)

// Service orchestrates the auth flows. Stateless; safe to share across
// goroutines.
type Service struct {
	cfg    *config.Config
	repo   *Repository
	users  *user.Service
	aclSvc *acl.Service
	jwt    *jwt.Service
	mail   mailer.Mailer
	log    *slog.Logger
}

// NewService is the fx constructor.
func NewService(
	cfg *config.Config,
	repo *Repository,
	users *user.Service,
	aclSvc *acl.Service,
	jwtSvc *jwt.Service,
	mail mailer.Mailer,
	log *slog.Logger,
) *Service {
	return &Service{
		cfg:    cfg,
		repo:   repo,
		users:  users,
		aclSvc: aclSvc,
		jwt:    jwtSvc,
		mail:   mail,
		log:    log.With("module", "auth"),
	}
}

// --- signup + verification -----------------------------------------

// Signup creates the user, joins them to the Default user group,
// generates a verification token, and emails the verification link.
// Returns the created user — caller responds 202.
func (s *Service) Signup(ctx context.Context, req SignupRequest) (*user.User, error) {
	if !req.ToSAccepted {
		return nil, fmt.Errorf("auth: tos must be accepted")
	}
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("auth: name required")
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	u, err := s.users.Create(ctx, normaliseEmail(req.Email), strings.TrimSpace(req.Name), hash)
	if err != nil {
		return nil, err
	}

	// New users land with zero groups. Default-tier features (catalog
	// view, playlists, bookmarks) are reachable by any authenticated
	// user since their routes don't carry RequirePermission. Admin
	// powers come from explicit group membership (see prd/acl.md).

	if err := s.sendVerificationEmail(ctx, u); err != nil {
		// Don't fail the signup — verification can be re-requested. Log
		// loudly so an operator notices a misconfigured mailer.
		s.log.Warn("verification email failed; user can re-request", "user_id", u.ID, "err", err)
	}

	return u, nil
}

// VerifyEmail consumes the token and marks the user verified.
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	userID, err := s.repo.ConsumeVerificationToken(ctx, token)
	if err != nil {
		return err
	}
	if err := s.users.MarkEmailVerified(ctx, userID); err != nil {
		return err
	}
	return nil
}

// ResendVerification re-issues the verification token. Look-up is by
// email so the user can ask for it before they've signed in.
//
// Returns nil whether or not the email exists — no enumeration.
func (s *Service) ResendVerification(ctx context.Context, email string) error {
	row, err := s.users.GetCredentialsByEmail(ctx, normaliseEmail(email))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil
		}
		return err
	}
	if row.EmailVerifiedAt.Valid {
		return nil // already verified — no-op
	}
	u, err := s.users.GetByID(ctx, row.ID)
	if err != nil {
		return err
	}
	return s.sendVerificationEmail(ctx, u)
}

func (s *Service) sendVerificationEmail(ctx context.Context, u *user.User) error {
	plaintext, hash, err := newToken()
	if err != nil {
		return err
	}
	if err := s.repo.CreateVerificationToken(ctx, u.ID, hash, time.Now().Add(verificationTokenTTL)); err != nil {
		return err
	}
	link := fmt.Sprintf("%s/verify-email?token=%s", s.cfg.App.BaseURL, plaintext)
	body := fmt.Sprintf(
		"Hi %s,\n\nClick the link below to verify your email:\n\n%s\n\nThis link expires in 24 hours.\n\n— QuranPrism",
		u.Name, link,
	)
	return s.mail.Send(ctx, mailer.Mail{
		To:      u.Email,
		Subject: "Verify your QuranPrism email",
		Text:    body,
	})
}

// --- login / refresh / logout --------------------------------------

// LoginResult is the bundle the handler needs to populate the response
// and the refresh cookie.
type LoginResult struct {
	User              *user.User
	Access            jwt.Issued
	RefreshPlaintext  string
	RefreshExpiresAt  time.Time
	RefreshSessionID  uuid.UUID
}

// Login verifies the credentials, applies the lockout policy, and issues
// the access JWT + a new refresh session.
//
// ip / userAgent are best-effort identification for the refresh_sessions
// row; passing nil / "" is fine.
func (s *Service) Login(ctx context.Context, email, password string, ip *netip.Addr, userAgent string) (*LoginResult, error) {
	email = normaliseEmail(email)

	// ACC-6 lockout check before we touch bcrypt: if you're locked we
	// don't even leak whether the password would have worked.
	failures, err := s.repo.CountRecentFailedLogins(ctx, email, time.Now().Add(-lockoutWindow))
	if err != nil {
		return nil, err
	}
	if failures >= lockoutThreshold {
		return nil, ErrAccountLocked
	}

	row, err := s.users.GetCredentialsByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			// Record the failure under the supplied email so the lockout
			// applies even when the email doesn't exist (slows enumeration).
			_ = s.repo.RecordLoginAttempt(ctx, email, ip, false)
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if row.IsDisabled {
		return nil, ErrAccountDisabled
	}

	ok, err := VerifyPassword(password, row.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !ok {
		_ = s.repo.RecordLoginAttempt(ctx, email, ip, false)
		return nil, ErrInvalidCredentials
	}

	if !row.EmailVerifiedAt.Valid {
		// Don't issue tokens for an unverified account.
		_ = s.repo.RecordLoginAttempt(ctx, email, ip, false)
		return nil, ErrEmailNotVerified
	}

	if err := s.repo.RecordLoginAttempt(ctx, email, ip, true); err != nil {
		s.log.Warn("record login attempt failed", "err", err)
	}

	// Per PRV-4: a successful login inside the 30-day grace cancels the
	// pending deletion.
	if row.DeletionRequestedAt.Valid {
		_ = s.users.CancelDeletionGrace(ctx, row.ID)
	}

	issued, err := s.jwt.Sign(row.ID.String())
	if err != nil {
		return nil, err
	}

	plaintext, hash, err := newToken()
	if err != nil {
		return nil, err
	}
	refreshExp := time.Now().Add(s.cfg.JWT.RefreshTTL)
	sessionID, err := s.repo.CreateRefreshSession(ctx, row.ID, hash, userAgent, ip, refreshExp)
	if err != nil {
		return nil, err
	}

	u, err := s.users.GetByID(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:             u,
		Access:           issued,
		RefreshPlaintext: plaintext,
		RefreshExpiresAt: refreshExp,
		RefreshSessionID: sessionID,
	}, nil
}

// RefreshResult is the bundle the refresh handler returns.
type RefreshResult struct {
	Access            jwt.Issued
	RefreshPlaintext  string
	RefreshExpiresAt  time.Time
	RefreshSessionID  uuid.UUID
}

// Refresh rotates the supplied refresh token: old session is revoked, a
// new one is created, and a new access JWT is signed.
func (s *Service) Refresh(ctx context.Context, plaintext string, ip *netip.Addr, userAgent string) (*RefreshResult, error) {
	if plaintext == "" {
		return nil, ErrMissingRefreshCookie
	}
	row, err := s.repo.LookupRefreshSession(ctx, plaintext)
	if err != nil {
		return nil, err
	}

	// Rotation: revoke old, mint new. Touch is implied by the revoke.
	if err := s.repo.RevokeRefreshSession(ctx, row.ID); err != nil {
		return nil, err
	}

	newPlain, newHash, err := newToken()
	if err != nil {
		return nil, err
	}
	exp := time.Now().Add(s.cfg.JWT.RefreshTTL)
	newID, err := s.repo.CreateRefreshSession(ctx, row.UserID, newHash, userAgent, ip, exp)
	if err != nil {
		return nil, err
	}

	issued, err := s.jwt.Sign(row.UserID.String())
	if err != nil {
		return nil, err
	}

	return &RefreshResult{
		Access:           issued,
		RefreshPlaintext: newPlain,
		RefreshExpiresAt: exp,
		RefreshSessionID: newID,
	}, nil
}

// Logout revokes the refresh session matching the supplied plaintext.
// No-op for an unknown / already-revoked token — logout is idempotent.
func (s *Service) Logout(ctx context.Context, plaintext string) error {
	if plaintext == "" {
		return nil
	}
	row, err := s.repo.LookupRefreshSession(ctx, plaintext)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return nil
		}
		return err
	}
	return s.repo.RevokeRefreshSession(ctx, row.ID)
}

// --- forgot / reset / change password ------------------------------

// ForgotPassword issues a reset token if the email exists, emails the
// link, and always returns nil so the response is identical whether or
// not the account exists.
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	row, err := s.users.GetCredentialsByEmail(ctx, normaliseEmail(email))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil
		}
		return err
	}
	plaintext, hash, err := newToken()
	if err != nil {
		return err
	}
	if err := s.repo.CreateResetToken(ctx, row.ID, hash, time.Now().Add(resetTokenTTL)); err != nil {
		return err
	}
	link := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.App.BaseURL, plaintext)
	body := fmt.Sprintf(
		"Hi,\n\nA password reset was requested for your account. Click the link below to choose a new password:\n\n%s\n\nIf you didn't request this, ignore this email.\n\nThis link expires in 1 hour.\n\n— QuranPrism",
		link,
	)
	return s.mail.Send(ctx, mailer.Mail{
		To:      row.Email,
		Subject: "Reset your QuranPrism password",
		Text:    body,
	})
}

// ResetPassword consumes a reset token, updates the password, and
// revokes every refresh session for the user (re-login required
// everywhere).
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	userID, err := s.repo.ConsumeResetToken(ctx, token)
	if err != nil {
		return err
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, userID, hash); err != nil {
		return err
	}
	return s.repo.RevokeAllRefreshSessions(ctx, userID)
}

// ChangePassword verifies the current password, updates to the new one,
// and revokes every OTHER refresh session so the active device stays
// logged in. The current session is preserved by passing currentSessionID;
// pass uuid.Nil to revoke every session.
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, current, newPassword string, currentSessionID uuid.UUID) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	row, err := s.users.GetCredentialsByID(ctx, userID)
	if err != nil {
		return err
	}
	ok, err := VerifyPassword(current, row.PasswordHash)
	if err != nil {
		return err
	}
	if !ok {
		return ErrPasswordMismatch
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, userID, hash); err != nil {
		return err
	}
	if currentSessionID == uuid.Nil {
		return s.repo.RevokeAllRefreshSessions(ctx, userID)
	}
	return s.repo.RevokeOtherRefreshSessions(ctx, userID, currentSessionID)
}

// lookupRefreshSession exposes the repo lookup to the change-password
// handler without making the repo field public.
func (s *Service) lookupRefreshSession(ctx context.Context, plaintext string) (sqlcdb.RefreshSession, error) {
	return s.repo.LookupRefreshSession(ctx, plaintext)
}

// --- validation helpers --------------------------------------------

func normaliseEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func validateEmail(s string) error {
	s = strings.TrimSpace(s)
	if s == "" || !strings.Contains(s, "@") || len(s) > 320 {
		return fmt.Errorf("auth: invalid email")
	}
	return nil
}

func validatePassword(s string) error {
	if len(s) < 12 {
		return fmt.Errorf("auth: password must be at least 12 characters")
	}
	if len(s) > 256 {
		return fmt.Errorf("auth: password too long")
	}
	return nil
}
