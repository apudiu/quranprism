package user

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/apudiu/quranprism/api/internal/platform/jwt"
	"github.com/apudiu/quranprism/api/internal/transport/http/httperr"
	httpmw "github.com/apudiu/quranprism/api/internal/transport/http/middleware"
	"github.com/apudiu/quranprism/api/internal/transport/http/response"
	"github.com/apudiu/quranprism/api/internal/transport/http/router"
)

// Handler owns the /v1/me surface (account self-service). The /v1/auth
// surface lives in the auth module.
type Handler struct {
	svc *Service
	jwt *jwt.Service
	log *slog.Logger
}

// NewHandler is the constructor wired into fx.
func NewHandler(svc *Service, jwtSvc *jwt.Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, jwt: jwtSvc, log: log.With("module", "user")}
}

// RegisterRoutes satisfies app.RouteRegistrar. Every endpoint here sits
// behind AuthRequired — the user is always the authenticated principal.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/v1/me", func(r chi.Router) {
		r.Use(httpmw.AuthRequired(h.jwt, h.log))

		r.Get("/", h.get)
		r.Patch("/", h.update)
		r.Delete("/", h.requestDeletion)
		r.Post("/data-export", h.requestDataExport)
	})
}

// Compile-time check that *Handler satisfies router.Registrar.
var _ router.Registrar = (*Handler)(nil)

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(httpmw.MustIdentity(r.Context()).UserID)
	if err != nil {
		httperr.Write(w, h.log, httperr.Unauthorized())
		return
	}
	me, err := h.svc.Me(r.Context(), id)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.JSON(w, http.StatusOK, me)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(httpmw.MustIdentity(r.Context()).UserID)
	if err != nil {
		httperr.Write(w, h.log, httperr.Unauthorized())
		return
	}
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, h.log, httperr.BadRequest("malformed body"))
		return
	}
	updated, err := h.svc.UpdateProfile(r.Context(), id, req.Name, req.AvatarURL)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.JSON(w, http.StatusOK, updated)
}

// requestDeletion stamps deletion_requested_at, kicking off the PRV-4
// 30-day grace window. The cron worker hard-deletes when the window
// closes; a successful login inside the window cancels it.
func (h *Handler) requestDeletion(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(httpmw.MustIdentity(r.Context()).UserID)
	if err != nil {
		httperr.Write(w, h.log, httperr.Unauthorized())
		return
	}
	if err := h.svc.StartDeletionGrace(r.Context(), id); err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.JSON(w, http.StatusAccepted, map[string]any{
		"status":     "deletion_requested",
		"grace_days": 30,
	})
}

// requestDataExport is the PRV-1 entry point. v1 just acknowledges the
// request; the River worker job that actually builds the export archive
// and sends the link by email lands in a later task.
func (h *Handler) requestDataExport(w http.ResponseWriter, r *http.Request) {
	// id is validated to keep the contract honest, even though the job
	// isn't wired yet — the queue.Insert call lands here next.
	if _, err := uuid.Parse(httpmw.MustIdentity(r.Context()).UserID); err != nil {
		httperr.Write(w, h.log, httperr.Unauthorized())
		return
	}
	response.JSON(w, http.StatusAccepted, map[string]any{
		"status": "queued",
	})
}
