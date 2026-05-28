// cmd/cron runs periodic tasks (e.g. login-attempt pruning, token
// cleanup, post-grace user hard-delete). Same platform stack as
// cmd/api, minus the HTTP listener, plus a cron scheduler. Skeleton
// only in this phase.
package main

import (
	"go.uber.org/fx"

	"github.com/apudiu/quranprism/api/internal/app"
)

func main() {
	fx.New(app.CronApp).Run()
}
