package acl

import (
	"time"

	"github.com/google/uuid"
)

// Request DTOs --------------------------------------------------------

// CreateGroupRequest is POST /v1/admin/groups body. Name is required;
// description is optional (nil = no description).
type CreateGroupRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// UpdateGroupRequest is PATCH /v1/admin/groups/{id} body. Both fields
// are optional pointer fields; nil = leave unchanged.
type UpdateGroupRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// AddGroupPermissionRequest is POST /v1/admin/groups/{id}/permissions body.
type AddGroupPermissionRequest struct {
	PermissionID uuid.UUID `json:"permission_id"`
}

// AddGroupMemberRequest is POST /v1/admin/users/{user_id}/groups body.
type AddGroupMemberRequest struct {
	GroupID uuid.UUID `json:"group_id"`
}

// Response DTOs -------------------------------------------------------

// GroupView is the wire shape of a group. Permissions is populated only
// on the detail endpoint, omitted from the list response.
type GroupView struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Permissions []string  `json:"permissions,omitempty"`
}

// PermissionView is the wire shape of a permission catalog entry.
type PermissionView struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Subject     string    `json:"subject"`
	Action      string    `json:"action"`
	Description *string   `json:"description"`
}

// UserView is the wire shape used by GET /v1/admin/users[?detail]. The
// Groups slice is populated on the detail endpoint only.
type UserView struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	EmailVerified bool      `json:"email_verified"`
	IsDisabled    bool      `json:"is_disabled"`
	CreatedAt     time.Time `json:"created_at"`
	Groups        []string  `json:"groups,omitempty"`
}

// Page wraps a list response with the pagination metadata the front-end
// needs to render page indicators. Items is any so each handler can
// pass its own typed slice.
type Page struct {
	Items  any   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}
