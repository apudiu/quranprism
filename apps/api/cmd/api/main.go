// cmd/api is the HTTP-facing process. The whole boot is fx.New —
// composition, lifecycle hooks, and graceful shutdown are all handled by
// the framework; main() stays a one-liner so the wiring is
// discoverable from internal/app/module.go.
package main

import (
	"go.uber.org/fx"

	"github.com/apudiu/quranprism/api/internal/app"
)

func main() {
	fx.New(app.HTTPApp).Run()
}
