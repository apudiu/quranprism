package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
	"github.com/apudiu/quranprism/api/internal/modules/acl"
)

// seed:run reconciles the permission catalog into Postgres standalone.
//
// The same reconciliation runs automatically as part of every serve:*
// boot (via the fx OnStart hook on aclmod.Module). Having an explicit
// CLI verb lets deploy pipelines run `migrate:up && seed:run &&
// admin:grant` without booting an api process first, which is what
// k8s Jobs and CI bootstrap steps expect.
//
// Idempotent — every catalog entry is upserted by name; re-running is
// a no-op once perms are in sync.
func newSeedRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "seed:run",
		Short: "Reconcile the ACL permission catalog into Postgres",
		RunE: func(*cobra.Command, []string) error {
			return runSeed()
		},
	}
}

func runSeed() error {
	dsn, err := databaseURL()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("pool: %w", err)
	}
	defer pool.Close()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return acl.Seed(ctx, sqlcdb.New(pool), log)
}
