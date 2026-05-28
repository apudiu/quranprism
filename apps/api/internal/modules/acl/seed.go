package acl

import (
	"context"
	"fmt"
	"log/slog"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
)

// Seed reconciles Postgres with PermissionCatalog.
//
// Strategy: upsert every catalog permission (insert-or-update by unique
// name). Groups are NOT seeded — admins create them post-deploy via the
// `qp admin:grant` CLI command.
//
// Idempotent. Safe to run on every boot.
func Seed(ctx context.Context, q *sqlcdb.Queries, log *slog.Logger) error {
	log = log.With("subsystem", "acl.seed")

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

	log.Info("acl seed complete", "permissions", len(PermissionCatalog))
	return nil
}

func descPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
