// Package acl implements the permission-group-user authorisation model
// described in prd/acl.md.
//
// In this phase the module ships scaffold only: schema (covered by the
// migration), the system-group + permission seed, and the small helper
// surface that the auth/user modules need to drop users into the
// "Default user" group on signup. Permission-enforcing middleware lands
// in its own task.
package acl

// Permission is the (subject, action) atom enforced at the route layer
// (later) — e.g. {"Playlist", "create"} matches the canonical wire name
// "Playlist:create".
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
	IsSystem    bool
	Permissions []string // canonical Permission.Name values
}
