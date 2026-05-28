package auth

import (
	"net/http"

	"github.com/apudiu/quranprism/api/internal/transport/http/httperr"
)

// ErrInvalidCredentials is returned for any login failure mode where we
// must not leak whether the email exists.
var ErrInvalidCredentials = httperr.New(http.StatusUnauthorized, "invalid_credentials", "invalid email or password")

// ErrEmailNotVerified — login attempted before the user has clicked the
// verification link. Front-end can render a "resend" button.
var ErrEmailNotVerified = httperr.New(http.StatusForbidden, "email_not_verified", "verify your email before signing in")

// ErrAccountDisabled — admin disabled the account.
var ErrAccountDisabled = httperr.New(http.StatusForbidden, "account_disabled", "this account has been disabled")

// ErrAccountLocked — ACC-6: too many failed attempts in the window.
var ErrAccountLocked = httperr.New(http.StatusLocked, "account_locked", "too many failed attempts — try again later")

// ErrInvalidToken covers expired/used/unknown email-verification + reset
// tokens. We intentionally collapse the reasons so attackers don't get a
// hint about which state the token is in.
var ErrInvalidToken = httperr.New(http.StatusBadRequest, "invalid_token", "token is invalid or has expired")

// ErrMissingRefreshCookie — refresh / logout without the cookie set.
var ErrMissingRefreshCookie = httperr.New(http.StatusUnauthorized, "missing_refresh_cookie", "refresh cookie missing")

// ErrPasswordMismatch — change-password current_password didn't match.
var ErrPasswordMismatch = httperr.New(http.StatusUnauthorized, "password_mismatch", "current password is incorrect")
