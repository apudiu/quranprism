package acl

import (
	"context"
	"log/slog"

	"go.uber.org/fx"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
	"github.com/apudiu/quranprism/api/internal/transport/http/router"
)

// Module provides the acl Service + Handler and runs the permission
// catalog seed at startup. Groups are NOT seeded — admins create them
// post-deploy via `qp admin:grant`.
var Module = fx.Module("acl",
	fx.Provide(
		NewService,
		NewHandler,
		// Route registration: tag *Handler as a router.Registrar and
		// drop it into the `routes` value group consumed by NewRouter.
		fx.Annotate(
			func(h *Handler) router.Registrar { return h },
			fx.ResultTags(`group:"routes"`),
		),
	),
	fx.Invoke(runSeedOnStart),
)

// runSeedOnStart binds Seed to the fx lifecycle. Running it on OnStart
// (vs. constructor-invocation time) ensures the DB pool has pinged and
// is ready before the catalog upsert runs.
func runSeedOnStart(lc fx.Lifecycle, q *sqlcdb.Queries, log *slog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return Seed(ctx, q, log)
		},
	})
}
