package user

import (
	"net/http"

	"github.com/apudiu/quranprism/api/internal/transport/http/httperr"
)

// ErrNotFound is the canonical "no such user" surface. Wrapped via
// httperr at the handler layer to produce the 404 response.
var ErrNotFound = httperr.New(http.StatusNotFound, "user_not_found", "user not found")

// ErrEmailTaken — signup path collision.
var ErrEmailTaken = httperr.New(http.StatusConflict, "email_taken", "an account with that email already exists")

// ErrInvalidCredentials covers wrong-email-or-password from the auth
// module too; defined here so a misuse from elsewhere can wrap it.
var ErrInvalidCredentials = httperr.New(http.StatusUnauthorized, "invalid_credentials", "invalid email or password")

// ErrPasswordMismatch — change-password path: supplied current_password
// doesn't match.
var ErrPasswordMismatch = httperr.New(http.StatusUnauthorized, "password_mismatch", "current password is incorrect")
