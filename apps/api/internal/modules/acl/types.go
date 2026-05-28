// Package acl implements the permission-group-user authorisation model
// described in prd/acl.md.
//
// Permission strings are `resource:action`, lowercase, snake_case for
// multi-word resources. No wildcards. The seed reconciles the catalog;
// groups are created at runtime by admins (no system groups).
package acl

// Permission is the (subject, action) atom enforced at the route layer
// via RequirePermission — e.g. {"group", "create"} matches the
// canonical wire name "group:create".
type Permission struct {
	Name        string
	Subject     string
	Action      string
	Description string
}

// Group is a named bundle of permissions a user can belong to. Multiple
// group memberships union their permissions on the user.
type Group struct {
	Name        string
	Description string
	Permissions []string // canonical Permission.Name values
}
