package acl

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/apudiu/quranprism/api/internal/modules/audit"
	"github.com/apudiu/quranprism/api/internal/platform/jwt"
	"github.com/apudiu/quranprism/api/internal/transport/http/httperr"
	httpmw "github.com/apudiu/quranprism/api/internal/transport/http/middleware"
	"github.com/apudiu/quranprism/api/internal/transport/http/pagination"
	"github.com/apudiu/quranprism/api/internal/transport/http/response"
	"github.com/apudiu/quranprism/api/internal/transport/http/router"
)

// Handler owns every /v1/admin/* surface. Mounting under a single
// r.Route prefix keeps the chi.Mount collision gotcha out of the way:
// no other module registers /v1/admin so this handler is the sole owner.
type Handler struct {
	svc   *Service
	jwt   *jwt.Service
	users httpmw.EmailVerifier
	log   *slog.Logger
}

// NewHandler is the fx constructor. The EmailVerifier interface is
// satisfied by user.Service; the user module exposes it via an fx
// adapter so this package doesn't have to import user directly.
func NewHandler(svc *Service, jwtSvc *jwt.Service, users httpmw.EmailVerifier, log *slog.Logger) *Handler {
	return &Handler{
		svc:   svc,
		jwt:   jwtSvc,
		users: users,
		log:   log.With("module", "acl"),
	}
}

// RegisterRoutes satisfies router.Registrar.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/v1/admin", func(r chi.Router) {
		// Every admin route requires a verified-email authenticated user;
		// per-route RequirePermission narrows further.
		r.Use(httpmw.AuthRequired(h.jwt, h.log))
		r.Use(httpmw.EmailVerifiedRequired(h.users, h.log))

		// Groups
		r.With(httpmw.RequirePermission(h.svc, h.log, "group:view")).
			Get("/groups", h.listGroups)
		r.With(httpmw.RequirePermission(h.svc, h.log, "group:create")).
			Post("/groups", h.createGroup)
		r.With(httpmw.RequirePermission(h.svc, h.log, "group:view")).
			Get("/groups/{id}", h.getGroup)
		r.With(httpmw.RequirePermission(h.svc, h.log, "group:update")).
			Patch("/groups/{id}", h.updateGroup)
		r.With(httpmw.RequirePermission(h.svc, h.log, "group:delete")).
			Delete("/groups/{id}", h.deleteGroup)

		// Group ↔ permission links
		r.With(httpmw.RequirePermission(h.svc, h.log, "group:update")).
			Post("/groups/{id}/permissions", h.addGroupPermission)
		r.With(httpmw.RequirePermission(h.svc, h.log, "group:update")).
			Delete("/groups/{id}/permissions/{permission_id}", h.removeGroupPermission)

		// User ↔ group memberships
		r.With(httpmw.RequirePermission(h.svc, h.log, "group:update")).
			Post("/users/{user_id}/groups", h.addUserToGroup)
		r.With(httpmw.RequirePermission(h.svc, h.log, "group:update")).
			Delete("/users/{user_id}/groups/{group_id}", h.removeUserFromGroup)

		// Permission catalog (read-only)
		r.With(httpmw.RequirePermission(h.svc, h.log, "permission:view")).
			Get("/permissions", h.listPermissions)
		r.With(httpmw.RequirePermission(h.svc, h.log, "permission:view")).
			Get("/permissions/{id}", h.getPermission)

		// User admin (list + detail only in T-002)
		r.With(httpmw.RequirePermission(h.svc, h.log, "user:view")).
			Get("/users", h.listUsers)
		r.With(httpmw.RequirePermission(h.svc, h.log, "user:view")).
			Get("/users/{id}", h.getUser)
	})
}

var _ router.Registrar = (*Handler)(nil)

// Path-param + actor helpers ------------------------------------------

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	s := chi.URLParam(r, name)
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, httperr.BadRequest("invalid " + name)
	}
	return id, nil
}

func (h *Handler) actor(r *http.Request) (audit.Actor, error) {
	uid, err := uuid.Parse(httpmw.MustIdentity(r.Context()).UserID)
	if err != nil {
		return audit.Actor{}, httperr.Unauthorized()
	}
	return audit.Actor{UserID: &uid, Kind: audit.KindUser}, nil
}

// Groups --------------------------------------------------------------

func (h *Handler) listGroups(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := pagination.Parse(r)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	items, total, err := h.svc.ListGroups(r.Context(), limit, offset)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.JSON(w, http.StatusOK, Page{Items: items, Total: total, Limit: limit, Offset: offset})
}

func (h *Handler) createGroup(w http.ResponseWriter, r *http.Request) {
	actor, err := h.actor(r)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, h.log, httperr.BadRequest("malformed body"))
		return
	}
	view, err := h.svc.CreateGroup(r.Context(), actor, req)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.JSON(w, http.StatusCreated, view)
}

func (h *Handler) getGroup(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	view, err := h.svc.GetGroup(r.Context(), id)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.JSON(w, http.StatusOK, view)
}

func (h *Handler) updateGroup(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	actor, err := h.actor(r)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	var req UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, h.log, httperr.BadRequest("malformed body"))
		return
	}
	view, err := h.svc.UpdateGroup(r.Context(), actor, id, req)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.JSON(w, http.StatusOK, view)
}

func (h *Handler) deleteGroup(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	actor, err := h.actor(r)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	if err := h.svc.DeleteGroup(r.Context(), actor, id); err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.Empty(w, http.StatusNoContent)
}

// Group ↔ permission ---------------------------------------------------

func (h *Handler) addGroupPermission(w http.ResponseWriter, r *http.Request) {
	groupID, err := parseUUIDParam(r, "id")
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	actor, err := h.actor(r)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	var req AddGroupPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, h.log, httperr.BadRequest("malformed body"))
		return
	}
	if req.PermissionID == uuid.Nil {
		httperr.Write(w, h.log, httperr.Unprocessable("permission_id required"))
		return
	}
	if err := h.svc.AddGroupPermission(r.Context(), actor, groupID, req.PermissionID); err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.Empty(w, http.StatusNoContent)
}

func (h *Handler) removeGroupPermission(w http.ResponseWriter, r *http.Request) {
	groupID, err := parseUUIDParam(r, "id")
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	permID, err := parseUUIDParam(r, "permission_id")
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	actor, err := h.actor(r)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	if err := h.svc.RemoveGroupPermission(r.Context(), actor, groupID, permID); err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.Empty(w, http.StatusNoContent)
}

// User ↔ group membership ----------------------------------------------

func (h *Handler) addUserToGroup(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUIDParam(r, "user_id")
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	actor, err := h.actor(r)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	var req AddGroupMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.Write(w, h.log, httperr.BadRequest("malformed body"))
		return
	}
	if req.GroupID == uuid.Nil {
		httperr.Write(w, h.log, httperr.Unprocessable("group_id required"))
		return
	}
	if err := h.svc.AddUserToGroup(r.Context(), actor, userID, req.GroupID); err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.Empty(w, http.StatusNoContent)
}

func (h *Handler) removeUserFromGroup(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUIDParam(r, "user_id")
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	groupID, err := parseUUIDParam(r, "group_id")
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	actor, err := h.actor(r)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	if err := h.svc.RemoveUserFromGroup(r.Context(), actor, userID, groupID); err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.Empty(w, http.StatusNoContent)
}

// Permissions ---------------------------------------------------------

func (h *Handler) listPermissions(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := pagination.Parse(r)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	items, total, err := h.svc.ListPermissions(r.Context(), limit, offset)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.JSON(w, http.StatusOK, Page{Items: items, Total: total, Limit: limit, Offset: offset})
}

func (h *Handler) getPermission(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	view, err := h.svc.GetPermission(r.Context(), id)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.JSON(w, http.StatusOK, view)
}

// Users ---------------------------------------------------------------

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := pagination.Parse(r)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	var emailLike *string
	if s := r.URL.Query().Get("email"); s != "" {
		emailLike = &s
	}
	items, total, err := h.svc.ListUsers(r.Context(), limit, offset, emailLike)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.JSON(w, http.StatusOK, Page{Items: items, Total: total, Limit: limit, Offset: offset})
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	view, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		httperr.Write(w, h.log, err)
		return
	}
	response.JSON(w, http.StatusOK, view)
}
