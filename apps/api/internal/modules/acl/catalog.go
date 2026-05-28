package acl

// PermissionCatalog is the full set of resource:action atoms the
// application enforces today. Names are lowercase; multi-word resources
// use snake_case. Keep alphabetised by subject.
//
// New perms land here in the PR that introduces the endpoint enforcing
// them — we never seed perms without a corresponding RequirePermission
// somewhere, because an unenforced perm is dead weight.
//
// The seed upserts every entry on every boot. Removing one here will
// NOT delete it from Postgres (would risk dangling group_permission
// rows); for that, write an explicit migration.
//
// Default behaviour is open: routes without RequirePermission are
// reachable by any authenticated user. Only restricted operations
// carry a perm. Ownership remains a service/data-layer concern.
var PermissionCatalog = []Permission{
	{Name: "audit_log:view", Subject: "audit_log", Action: "view", Description: "Read audit log entries"},

	{Name: "group:create", Subject: "group", Action: "create", Description: "Create a new group"},
	{Name: "group:delete", Subject: "group", Action: "delete", Description: "Delete a group"},
	{Name: "group:update", Subject: "group", Action: "update", Description: "Update a group (name/description, member list, permission set)"},
	{Name: "group:view", Subject: "group", Action: "view", Description: "List or view a group's details"},

	{Name: "permission:view", Subject: "permission", Action: "view", Description: "List or view permission catalog entries"},

	{Name: "user:view", Subject: "user", Action: "view", Description: "List or view user records"},
}
