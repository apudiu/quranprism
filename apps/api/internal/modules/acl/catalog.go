package acl

// PermissionCatalog is the full set of (subject, action) atoms the
// application knows about today. Keep alphabetised by subject. New
// permissions land here in the PR that introduces them.
//
// The seed upserts every entry on every boot — removing one here will
// NOT delete it from Postgres (would risk dangling group_permission
// rows); for that, write an explicit migration.
var PermissionCatalog = []Permission{
	// Catalog read-only access (Languages, Translators, Reciters, Surahs).
	{Name: "Catalog:view", Subject: "Catalog", Action: "view", Description: "Read the public Quran catalog"},

	// Personal data (Free + Paid; ownership scope enforced at the service
	// layer — these atoms say "is allowed to use the feature").
	{Name: "Playlist:create", Subject: "Playlist", Action: "create"},
	{Name: "Playlist:view", Subject: "Playlist", Action: "view"},
	{Name: "Playlist:update", Subject: "Playlist", Action: "update"},
	{Name: "Playlist:delete", Subject: "Playlist", Action: "delete"},

	{Name: "Bookmark:create", Subject: "Bookmark", Action: "create"},
	{Name: "Bookmark:view", Subject: "Bookmark", Action: "view"},
	{Name: "Bookmark:update", Subject: "Bookmark", Action: "update"},
	{Name: "Bookmark:delete", Subject: "Bookmark", Action: "delete"},

	{Name: "BookmarkCategory:create", Subject: "BookmarkCategory", Action: "create"},
	{Name: "BookmarkCategory:view", Subject: "BookmarkCategory", Action: "view"},
	{Name: "BookmarkCategory:update", Subject: "BookmarkCategory", Action: "update"},
	{Name: "BookmarkCategory:delete", Subject: "BookmarkCategory", Action: "delete"},

	// Catalog management (admin-only).
	{Name: "Language:manage", Subject: "Language", Action: "manage"},
	{Name: "Translator:manage", Subject: "Translator", Action: "manage"},
	{Name: "Reciter:manage", Subject: "Reciter", Action: "manage"},
	{Name: "AudioFile:upload", Subject: "AudioFile", Action: "upload"},
	{Name: "TranslationText:upload", Subject: "TranslationText", Action: "upload"},

	// Internal admin surfaces.
	{Name: "User:manage", Subject: "User", Action: "manage"},
	{Name: "Group:manage", Subject: "Group", Action: "manage"},
	{Name: "Permission:manage", Subject: "Permission", Action: "manage"},
	{Name: "AuditLog:view", Subject: "AuditLog", Action: "view"},
}

// SystemGroups are the seeded, undeletable groups. is_system = true is
// the flag that makes them protected in the admin UI later.
//
//   - DefaultUserGroup is what every new signup joins (acl.md "Default
//     user").
//   - SuperAdminGroup is the break-glass group; granted manually only.
//   - ContentManagerGroup is the bootstrap group for the catalog ops team.
var (
	DefaultUserGroupName    = "Default user"
	SuperAdminGroupName     = "Super Admin"
	ContentManagerGroupName = "Content Manager"
)

// SystemGroups is the seed list. Permissions are referenced by canonical
// name and validated against PermissionCatalog at seed time.
var SystemGroups = []Group{
	{
		Name:        DefaultUserGroupName,
		Description: "Default group for newly registered users.",
		IsSystem:    true,
		Permissions: []string{
			"Catalog:view",
			"Playlist:create", "Playlist:view", "Playlist:update", "Playlist:delete",
			"Bookmark:create", "Bookmark:view", "Bookmark:update", "Bookmark:delete",
			"BookmarkCategory:create", "BookmarkCategory:view", "BookmarkCategory:update", "BookmarkCategory:delete",
		},
	},
	{
		Name:        SuperAdminGroupName,
		Description: "Full administrative access. Granted manually only.",
		IsSystem:    true,
		Permissions: allPermissionNames(),
	},
	{
		Name:        ContentManagerGroupName,
		Description: "Quran catalog operations: languages, translators, reciters, audio uploads.",
		IsSystem:    true,
		Permissions: []string{
			"Catalog:view",
			"Language:manage", "Translator:manage", "Reciter:manage",
			"AudioFile:upload", "TranslationText:upload",
		},
	},
}

// allPermissionNames returns the canonical names of every permission in
// PermissionCatalog. Used to seed Super Admin without restating the list.
func allPermissionNames() []string {
	names := make([]string, len(PermissionCatalog))
	for i, p := range PermissionCatalog {
		names[i] = p.Name
	}
	return names
}
