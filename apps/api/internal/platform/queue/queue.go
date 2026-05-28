// Package queue owns the application's River job-queue client.
//
// River is Postgres-backed: enqueue with `riverClient.InsertTx(ctx, tx,
// args, nil)` and the row commits atomically with the rest of the
// transaction. That's why we keep it instead of moving jobs onto NATS.
//
// The api process is **insert-only**: it builds a client without calling
// Start() so no jobs are consumed here. The worker process imports the
// same package, additionally registers workers, and calls Start().
package queue

import (
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

// Client is the concrete River client type for the pgx/v5 transaction
// type — exported so callers don't have to repeat the generic parameter.
type Client = river.Client[pgx.Tx]

// NewInsertOnly builds a River client suitable for the api process: it
// can InsertTx / Insert jobs but never starts a consumer.
//
// The worker process will build its own client with Workers populated and
// invoke Start() inside cmd/worker.
func NewInsertOnly(pool *pgxpool.Pool, log *slog.Logger) (*Client, error) {
	c, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		// No Queues / no Workers: insert-only, Start() never called.
		Logger: log.With("subsystem", "river"),
	})
	if err != nil {
		return nil, fmt.Errorf("queue: new client: %w", err)
	}
	return c, nil
}
