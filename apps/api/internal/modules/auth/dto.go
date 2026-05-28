package auth

import "time"

// SignupRequest — POST /v1/auth/signup body. ToSAccepted enforces PRV-18.
type SignupRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	ToSAccepted bool   `json:"tos_accepted"`
}

// LoginRequest — POST /v1/auth/login body.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// VerifyEmailRequest — POST /v1/auth/verify-email body.
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// ResendVerificationRequest — POST /v1/auth/resend-verification body.
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// ForgotPasswordRequest — POST /v1/auth/forgot-password body.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest — POST /v1/auth/reset-password body.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ChangePasswordRequest — POST /v1/me/change-password body.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// LoginResponse — body returned by login + refresh. The refresh token
// itself rides as an HttpOnly cookie, not in this body.
type LoginResponse struct {
	AccessToken     string    `json:"access_token"`
	AccessExpiresAt time.Time `json:"access_expires_at"`
}
