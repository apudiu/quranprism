// cmd/qp is the single binary that runs every quranprism api process —
// long-running services (serve:api / serve:worker / serve:cron),
// one-shot ops (migrate:up / migrate:down / migrate:status / migrate:queue),
// and admin bootstrap (admin:grant). Subcommands use colon-namespacing
// (`qp admin:grant`) so groups read at a glance and the CLI scales as
// future tasks add their own commands.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:           "qp",
		Short:         "quranprism api control plane",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		newServeAPICmd(),
		newServeWorkerCmd(),
		newServeCronCmd(),
		newMigrateUpCmd(),
		newMigrateDownCmd(),
		newMigrateStatusCmd(),
		newMigrateQueueCmd(),
		newSeedRunCmd(),
		newAdminGrantCmd(),
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "qp:", err)
		os.Exit(1)
	}
}
