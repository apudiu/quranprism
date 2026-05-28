// Package cache owns the application's Redis connection.
//
// Used by the HTTP rate-limit middleware (token-bucket counters) and by
// any domain module that needs an opaque KV / TTL store on a hot path.
// For anything durable, use Postgres — Redis here is intentionally a
// fast-path optimisation, not a system of record.
package cache

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/apudiu/quranprism/api/internal/platform/config"
)

// NewClient parses REDIS_URL, opens the connection lazily, and pings on
// fx start so misconfigured DSNs fail fast at boot rather than at first
// use during a request.
func NewClient(lc fx.Lifecycle, cfg *config.Config, log *slog.Logger) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("cache: parse url: %w", err)
	}
	client := redis.NewClient(opts)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := client.Ping(ctx).Err(); err != nil {
				return fmt.Errorf("cache: ping: %w", err)
			}
			log.Info("redis ready", "addr", opts.Addr, "db", opts.DB)
			return nil
		},
		OnStop: func(_ context.Context) error {
			return client.Close()
		},
	})
	return client, nil
}
