// Package logger constructs the application's root *slog.Logger.
//
// Format follows the CLAUDE.md rule: JSON in production (machine-readable
// for the log pipeline), text in dev (human-readable for `docker compose
// logs`). Verbosity is INFO in production and DEBUG everywhere else.
package logger

import (
	"log/slog"
	"os"

	"github.com/apudiu/quranprism/api/internal/platform/config"
)

// New builds the root logger from environment-driven config. Returned
// logger is goroutine-safe and is intended to be wrapped per-component
// via `slog.With(...)`.
func New(cfg *config.Config) *slog.Logger {
	level := slog.LevelDebug
	if cfg.App.IsProduction() {
		level = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: level}

	var h slog.Handler
	if cfg.App.IsProduction() {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(h)
}
