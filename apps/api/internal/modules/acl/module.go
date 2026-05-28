package acl

import (
	"context"
	"log/slog"

	"go.uber.org/fx"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
)

// Module provides the acl Service and runs the catalog + system-group
// seed as part of the fx startup graph. Seed runs once per process boot
// (fx.Invoke executes constructors at startup, not on first dependency
// access), so a misconfigured catalog fails the api fast.
var Module = fx.Module("acl",
	fx.Provide(NewService),
	fx.Invoke(runSeedOnStart),
)

// runSeedOnStart registers an fx Lifecycle.OnStart that calls Seed.
// Couples seed to process lifetime instead of constructor-invocation
// time so the rest of the platform (DB pool ping, etc.) has booted first.
func runSeedOnStart(lc fx.Lifecycle, q *sqlcdb.Queries, log *slog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return Seed(ctx, q, log)
		},
	})
}
