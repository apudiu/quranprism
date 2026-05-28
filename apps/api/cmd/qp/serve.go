package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/apudiu/quranprism/api/internal/app"
)

// Each serve:* subcommand runs the matching fx graph. Bodies are the
// one-liners previously living in cmd/api / cmd/worker / cmd/cron —
// fx owns lifecycle hooks, signal handling, and graceful shutdown, so
// the cobra layer stays a thin wrapper.

func newServeAPICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve:api",
		Short: "Run the HTTP API server",
		RunE: func(*cobra.Command, []string) error {
			fx.New(app.HTTPApp).Run()
			return nil
		},
	}
}

func newServeWorkerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve:worker",
		Short: "Run the River background-job worker",
		RunE: func(*cobra.Command, []string) error {
			fx.New(app.WorkerApp).Run()
			return nil
		},
	}
}

func newServeCronCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve:cron",
		Short: "Run the periodic-task scheduler",
		RunE: func(*cobra.Command, []string) error {
			fx.New(app.CronApp).Run()
			return nil
		},
	}
}
