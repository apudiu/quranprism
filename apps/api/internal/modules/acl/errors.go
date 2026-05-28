package acl

import (
	"net/http"

	"github.com/apudiu/quranprism/api/internal/transport/http/httperr"
)

// Sentinels exposed as *httperr.E directly, mirroring user/errors.go.
// The handler layer can pass them straight through httperr.Write — no
// translation table required.
//
// Codes are stable machine identifiers; phrasing is user-facing.

var (
	ErrGroupNotFound = httperr.New(
		http.StatusNotFound, "group_not_found", "group not found",
	)

	ErrPermissionNotFound = httperr.New(
		http.StatusNotFound, "permission_not_found", "permission not found",
	)

	ErrUserNotFound = httperr.New(
		http.StatusNotFound, "user_not_found", "user not found",
	)

	ErrGroupNameTaken = httperr.New(
		http.StatusConflict, "group_name_taken", "group name already in use",
	)

	ErrMembershipNotFound = httperr.New(
		http.StatusNotFound, "membership_not_found", "user is not a member of that group",
	)

	ErrPermissionLinkNotFound = httperr.New(
		http.StatusNotFound, "permission_link_not_found", "group does not hold that permission",
	)

	// ErrLastAdmin is the self-protect rejection: a revoke that would
	// orphan `group:update` (the grant-capable perm) is refused. The
	// code is the explicit `last_admin_protected` so the front-end can
	// switch on it rather than the generic `conflict`.
	ErrLastAdmin = httperr.New(
		http.StatusConflict, "last_admin_protected", "would leave system without admin access",
	)
)

// Permission name the self-protect guard counts holders of. Anyone with
// this perm can manage groups (add/remove members, link/unlink perms,
// delete groups), so losing the last holder bricks admin access.
const grantCapablePermission = "group:update"
