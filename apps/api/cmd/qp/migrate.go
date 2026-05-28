package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver for goose
	"github.com/pressly/goose/v3"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/spf13/cobra"

	api "github.com/apudiu/quranprism/api"
)

// migrate:up / :down / :status drive goose against the embedded
// migrations FS so the binary is self-contained. migrate:queue applies
// River's own internal migrations (separate schema, separate tooling).

func newMigrateUpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:up",
		Short: "Apply pending goose migrations",
		RunE: func(*cobra.Command, []string) error {
			return withGoose(func(db *sql.DB) error {
				return goose.Up(db, "migrations")
			})
		},
	}
}

func newMigrateDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:down",
		Short: "Roll back the most-recent goose migration",
		RunE: func(*cobra.Command, []string) error {
			return withGoose(func(db *sql.DB) error {
				return goose.Down(db, "migrations")
			})
		},
	}
}

func newMigrateStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:status",
		Short: "List goose migrations with applied/pending state",
		RunE: func(*cobra.Command, []string) error {
			return withGoose(func(db *sql.DB) error {
				return goose.Status(db, "migrations")
			})
		},
	}
}

func newMigrateQueueCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:queue",
		Short: "Apply River's internal queue migrations",
		RunE: func(*cobra.Command, []string) error {
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
			m, err := rivermigrate.New(riverpgxv5.New(pool), nil)
			if err != nil {
				return fmt.Errorf("river migrator: %w", err)
			}
			if _, err := m.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
				return fmt.Errorf("river migrate up: %w", err)
			}
			fmt.Println("river migrations applied")
			return nil
		},
	}
}

// withGoose handles the cross-cutting work for every goose subcommand:
// load DSN, open a database/sql handle backed by pgx, point goose at
// the embedded migrations FS, run the caller's action, then close.
//
// migrate / admin one-shots intentionally bypass platform/config since
// they only need DATABASE_URL — the serve:* subcommands run the full
// fx graph which validates every required env var.
func withGoose(action func(db *sql.DB) error) error {
	dsn, err := databaseURL()
	if err != nil {
		return err
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Embedded migrations: cmd/qp ships with every .sql baked in. The
	// embed root contains the migrations/ subdir directly, so passing
	// it straight to goose works (goose.Up/Down/Status all take the
	// subpath "migrations" relative to the base FS).
	goose.SetBaseFS(api.MigrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	return action(db)
}

// databaseURL reads DATABASE_URL with a clean error message. Used by
// every one-shot subcommand (migrate, admin) that doesn't need the
// rest of the platform config to load.
func databaseURL() (string, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return "", errors.New("DATABASE_URL not set")
	}
	return dsn, nil
}
