// Package pubsub owns the NATS connection and its JetStream context.
//
// Two roles:
//   - **Pub/sub fan-out** (core NATS) for transient cross-replica signals
//     (e.g. notification.upserted → fan to every connected client).
//   - **JetStream** (durable streams / KV / queue groups) when at-least-once
//     delivery or replay is required.
//
// Background jobs do NOT go through here — they live in platform/queue
// (River) so we keep DB-atomic enqueue. NATS is reserved for fan-out and
// cross-process notification.
package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/fx"

	"github.com/apudiu/quranprism/api/internal/platform/config"
)

// Conn bundles the underlying *nats.Conn and the JetStream API handle.
// Code that only needs core publish/subscribe takes Conn; code that needs
// durable streams takes Conn.JS.
type Conn struct {
	NC *nats.Conn
	JS jetstream.JetStream
}

// New connects to NATS, opens the JetStream API, and registers fx
// lifecycle hooks. Drain (not hard-close) on shutdown so in-flight
// messages on subscriptions get delivered before the conn goes away.
func New(lc fx.Lifecycle, cfg *config.Config, log *slog.Logger) (*Conn, error) {
	nc, err := nats.Connect(
		cfg.NATS.URL,
		nats.Name("qp-api"),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Warn("nats disconnected", "err", err)
		}),
		nats.ReconnectHandler(func(c *nats.Conn) {
			log.Info("nats reconnected", "url", c.ConnectedUrl())
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("pubsub: connect: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("pubsub: jetstream: %w", err)
	}

	c := &Conn{NC: nc, JS: js}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Info("nats ready", "url", nc.ConnectedUrl())
			return nil
		},
		OnStop: func(_ context.Context) error {
			return nc.Drain()
		},
	})
	return c, nil
}
