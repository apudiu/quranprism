// Package db owns the application's Postgres connection pool.
//
// The pool is built once at startup, pinged before fx lets the rest of the
// graph initialise (fast-fails the process on misconfigured DSN), and
// closed in reverse order on shutdown. Every domain module obtains
// *pgxpool.Pool through fx — no globals.
//
// The sqlc-generated *Queries type is bolted on top of the pool in
// queries.go once `make sqlc` has run.
package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/apudiu/quranprism/api/internal/platform/config"
)

// NewPool constructs a *pgxpool.Pool from the configured DATABASE_URL and
// installs fx lifecycle hooks so the pool is pinged at startup and drained
// at shutdown. The pool itself is goroutine-safe.
func NewPool(lc fx.Lifecycle, cfg *config.Config, log *slog.Logger) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DB.URL)
	if err != nil {
		return nil, fmt.Errorf("db: parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("db: new pool: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := pool.Ping(ctx); err != nil {
				return fmt.Errorf("db: ping: %w", err)
			}
			log.Info("db pool ready", "max_conns", poolCfg.MaxConns)
			return nil
		},
		OnStop: func(_ context.Context) error {
			pool.Close()
			return nil
		},
	})
	return pool, nil
}
