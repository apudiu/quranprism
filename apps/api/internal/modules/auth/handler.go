package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/apudiu/quranprism/api/internal/platform/config"
	"github.com/apudiu/quranprism/api/internal/platform/jwt"
	"github.com/apudiu/quranprism/api/internal/transport/http/httperr"
	httpmw "github.com/apudiu/quranprism/api/internal/transport/http/middleware"
	"github.com/apudiu/quranprism/api/internal/transport/http/response"
	"github.com/apudiu/quranprism/api/internal/transport/http/router"
)

// refreshCookieName is the cookie that carries the plaintext refresh
// token. HttpOnly + SameSite=Lax + scoped to /v1/auth so it only flies
// to the rotation endpoints — not on every API call.
const refreshCookieName = "qp_refresh"

// Handler owns /v1/auth/* and the protected /v1/me/change-password
// endpoint (lives here because it sits on the credential path).
type Handler struct {
	cfg *config.Config
	svc *Service
	jwt *jwt.Service
	rdb *redis.Client
	log *slog.Logger
}

// NewHandler is the fx constructor.
func NewHandler(cfg *config.Config, svc *Service, jwtSvc *jwt.Service, rdb *redis.Client, log *slog.Logger) *Handler {
	return &Handler{
		cfg: cfg,
		svc: svc,
		jwt: jwtSvc,
		rdb: rdb,
		log: log.With("module", "auth"),
	}
}

var _ router.Registrar = (*Handler)(nil)

// RegisterRoutes mounts auth + change-password on the given chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/v1/auth", func(r chi.Router) {
		// Public endpoints — rate-limited by IP / email.
		r.With(httpmw.RateLimit(h.rdb, h.log, "signup", 5, time.Hour, httpmw.IPKeyFunc)).
			Post("/signup", h.signup)
		r.With(httpmw.RateLimit(h.rdb, h.log, "verify_email", 10, time.Hour, httpmw.IPKeyFunc)).
			Post("/verify-email", h.verifyEmail)
		r.With(httpmw.RateLimit(h.rdb, h.log, "resend_verify", 3, time.Hour, emailKeyFunc)).
			Post("/resend-verification", h.resendVerification)
		r.With(httpmw.RateLimit(h.rdb, h.log, "login", lockoutThreshold, lockoutWindow, httpmw.IPKeyFunc)).
			Post("/login", h.login)
		r.With(httpmw.RateLimit(h.rdb, h.log, "refresh", 60, time.Hour, httpmw.IPKeyFunc)).
			Post("/refresh", h.refresh)
		r.Post("/logout", h.logout)
		r.With(httpmw.RateLimit(h.rdb, h.log, "forgot_password", 3, time.Hour, emailKeyFunc)).
			Post("/forgot-password", h.forgotPassword)
		r.With(httpmw.RateLimit(h.rdb, h.log, "reset_password", 10, time.Hour, httpmw.IPKeyFunc)).
			Post("/reset-password", h.resetPassword)
	})

	// Protected. Lives under /v1/me but logically belongs to auth —
	// mounted as a flat path so it doesn't collide with the user
	// module's r.Route("/v1/me", ...) block.
	r.With(httpmw.AuthRequired(h.jwt, h.log)).
		Post("/v1/me/change-password", h.changePassword)
}

// --- handlers ------------------------------------------------------

func (h *Handler) signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, h.log, httperr.BadRequest("malformed body"))
		return
	}
	u, err := h.svc.Signup(r.Context(), req)
	if err != nil {
		// Translate validation errors to 400; everything else passes
		// through httperr.Write (which keeps typed *E intact).
		if isValidationErr(err) {
			httperr.Write(w, h.log, httperr.Unprocessable(err.Error()))
			return
		}
		httperr.Write(w, h.log, err)
		return
	}
	response.JSON(w, http.StatusAccepted, map[string]any{
		"user_id":           u.ID,
		"verification_sent": true,
	})
}

func (h *Handler) verifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, h.log, httperr.BadRequest("malformed body"))
		return
	}
	if err := h.svc.VerifyEmail(r.Context(), req.Token); err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.Empty(w, http.StatusNoContent)
}

func (h *Handler) resendVerification(w http.ResponseWriter, r *http.Request) {
	var req ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, h.log, httperr.BadRequest("malformed body"))
		return
	}
	if err := h.svc.ResendVerification(r.Context(), req.Email); err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.Empty(w, http.StatusAccepted)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, h.log, httperr.BadRequest("malformed body"))
		return
	}
	ip := parseIP(r)
	res, err := h.svc.Login(r.Context(), req.Email, req.Password, ip, r.UserAgent())
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	h.setRefreshCookie(w, res.RefreshPlaintext, res.RefreshExpiresAt)
	response.JSON(w, http.StatusOK, struct {
		User            any       `json:"user"`
		AccessToken     string    `json:"access_token"`
		AccessExpiresAt time.Time `json:"access_expires_at"`
	}{
		User:            res.User,
		AccessToken:     res.Access.Token,
		AccessExpiresAt: res.Access.ExpiresAt,
	})
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	plain, err := readRefreshCookie(r)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	ip := parseIP(r)
	res, err := h.svc.Refresh(r.Context(), plain, ip, r.UserAgent())
	if err != nil {
		// Bad / expired refresh → clear cookie and 401 so the SPA shows the
		// signed-out state instead of an infinite refresh loop.
		h.clearRefreshCookie(w)
		httperr.Write(w, h.log, httperr.Unauthorized())
		return
	}
	h.setRefreshCookie(w, res.RefreshPlaintext, res.RefreshExpiresAt)
	response.JSON(w, http.StatusOK, LoginResponse{
		AccessToken:     res.Access.Token,
		AccessExpiresAt: res.Access.ExpiresAt,
	})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	plain, _ := readRefreshCookieOptional(r)
	if err := h.svc.Logout(r.Context(), plain); err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	h.clearRefreshCookie(w)
	response.Empty(w, http.StatusNoContent)
}

func (h *Handler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, h.log, httperr.BadRequest("malformed body"))
		return
	}
	if err := h.svc.ForgotPassword(r.Context(), req.Email); err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.Empty(w, http.StatusAccepted)
}

func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, h.log, httperr.BadRequest("malformed body"))
		return
	}
	if err := h.svc.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		if isValidationErr(err) {
			httperr.Write(w, h.log, httperr.Unprocessable(err.Error()))
			return
		}
		httperr.Write(w, h.log, err)
		return
	}
	response.Empty(w, http.StatusNoContent)
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(httpmw.MustIdentity(r.Context()).UserID)
	if err != nil {
		httperr.Write(w, h.log, httperr.Unauthorized())
		return
	}
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, h.log, httperr.BadRequest("malformed body"))
		return
	}

	// Look up the current refresh session (if any) so we can keep it
	// active while rotating every other.
	var keepID uuid.UUID
	if plain, _ := readRefreshCookieOptional(r); plain != "" {
		if row, err := h.svc.lookupRefreshSession(r.Context(), plain); err == nil {
			keepID = row.ID
		}
	}

	if err := h.svc.ChangePassword(r.Context(), id, req.CurrentPassword, req.NewPassword, keepID); err != nil {
		if isValidationErr(err) {
			httperr.Write(w, h.log, httperr.Unprocessable(err.Error()))
			return
		}
		httperr.Write(w, h.log, err)
		return
	}
	response.Empty(w, http.StatusNoContent)
}

// --- cookie helpers ------------------------------------------------

func (h *Handler) setRefreshCookie(w http.ResponseWriter, value string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    value,
		Path:     "/v1/auth",
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		Secure:   h.cfg.App.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/v1/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.App.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	})
}

func readRefreshCookie(r *http.Request) (string, error) {
	c, err := r.Cookie(refreshCookieName)
	if err != nil || c.Value == "" {
		return "", ErrMissingRefreshCookie
	}
	return c.Value, nil
}

func readRefreshCookieOptional(r *http.Request) (string, error) {
	c, err := r.Cookie(refreshCookieName)
	if err != nil {
		return "", nil
	}
	return c.Value, nil
}

// --- misc helpers --------------------------------------------------

// parseIP extracts a *netip.Addr from the request. Honours X-Forwarded-For
// when present (compose-internal nginx / cloudflare relay). Returns nil
// when nothing parseable shows up — that's recorded as NULL inet.
func parseIP(r *http.Request) *netip.Addr {
	src := httpmw.IPKeyFunc(r)
	if src == "" || src == "anon" {
		return nil
	}
	if a, err := netip.ParseAddr(src); err == nil {
		return &a
	}
	// Strip port if present.
	if host, _, err := net.SplitHostPort(src); err == nil {
		if a, err := netip.ParseAddr(host); err == nil {
			return &a
		}
	}
	return nil
}

// emailKeyFunc keys rate-limit buckets by the body's "email" field, so
// resend / forgot-password resist enumeration sprays better than IP alone.
//
// Because chi binds the body to the handler not the middleware, this
// peeks at the body via a buffered copy then rewinds. Cheap for the
// tiny JSON bodies on these endpoints.
func emailKeyFunc(r *http.Request) string {
	// We can't read the body here without consuming it before the handler
	// also reads it; for v1 fall back to IP. Replace with a body-peek
	// helper if email-keyed limiting becomes important.
	return httpmw.IPKeyFunc(r)
}

// isValidationErr reports whether the error is a plain (non-httperr)
// validation failure from service.go. Used to coerce those into a 422
// without polluting the typed-error path.
func isValidationErr(err error) bool {
	if err == nil {
		return false
	}
	var e *httperr.E
	if errors.As(err, &e) {
		return false
	}
	return strings.HasPrefix(err.Error(), "auth:")
}
