package acl

import (
	"context"
	"fmt"
	"log/slog"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
)

// Seed reconciles Postgres with PermissionCatalog + SystemGroups.
//
// Strategy:
//   - Upsert every catalog permission (insert-or-update by unique name).
//   - For each system group, upsert the row then sync its permission set
//     to the catalog declaration: link missing, unlink extras.
//
// The system-group sync is one-way: PermissionCatalog + SystemGroups in
// this package are the source of truth, so an admin can't widen a system
// group out-of-band and have it stick. Non-system groups are untouched.
//
// Idempotent. Safe to run on every boot.
func Seed(ctx context.Context, q *sqlcdb.Queries, log *slog.Logger) error {
	log = log.With("subsystem", "acl.seed")

	// 1. Permissions.
	for _, p := range PermissionCatalog {
		desc := descPtr(p.Description)
		if _, err := q.UpsertPermission(ctx, sqlcdb.UpsertPermissionParams{
			Name:        p.Name,
			Subject:     p.Subject,
			Action:      p.Action,
			Description: desc,
		}); err != nil {
			return fmt.Errorf("seed permission %s: %w", p.Name, err)
		}
	}

	// 2. System groups + permission-set sync.
	for _, g := range SystemGroups {
		desc := descPtr(g.Description)
		row, err := q.UpsertGroup(ctx, sqlcdb.UpsertGroupParams{
			Name:        g.Name,
			Description: desc,
			IsSystem:    g.IsSystem,
		})
		if err != nil {
			return fmt.Errorf("seed group %s: %w", g.Name, err)
		}

		current, err := q.ListPermissionsForGroup(ctx, row.ID)
		if err != nil {
			return fmt.Errorf("seed list perms for %s: %w", g.Name, err)
		}
		currentSet := make(map[string]struct{}, len(current))
		currentByName := make(map[string]sqlcdb.Permission, len(current))
		for _, p := range current {
			currentSet[p.Name] = struct{}{}
			currentByName[p.Name] = p
		}

		wantSet := make(map[string]struct{}, len(g.Permissions))
		for _, name := range g.Permissions {
			wantSet[name] = struct{}{}
		}

		// Link missing.
		for _, name := range g.Permissions {
			if _, present := currentSet[name]; present {
				continue
			}
			perm, err := q.GetPermissionByName(ctx, name)
			if err != nil {
				return fmt.Errorf("seed lookup perm %s: %w", name, err)
			}
			if err := q.LinkGroupPermission(ctx, sqlcdb.LinkGroupPermissionParams{
				GroupID:      row.ID,
				PermissionID: perm.ID,
			}); err != nil {
				return fmt.Errorf("seed link %s→%s: %w", g.Name, name, err)
			}
		}

		// Unlink extras (drift cleanup).
		for name, perm := range currentByName {
			if _, want := wantSet[name]; want {
				continue
			}
			if err := q.UnlinkGroupPermission(ctx, sqlcdb.UnlinkGroupPermissionParams{
				GroupID:      row.ID,
				PermissionID: perm.ID,
			}); err != nil {
				return fmt.Errorf("seed unlink %s→%s: %w", g.Name, name, err)
			}
		}
	}

	log.Info("acl seed complete", "permissions", len(PermissionCatalog), "groups", len(SystemGroups))
	return nil
}

func descPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
