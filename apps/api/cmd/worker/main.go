// cmd/worker runs the River background-job consumer.
//
// In this phase no worker bodies exist yet — the process boots the same
// platform stack as cmd/api (minus the HTTP listener) and idles waiting
// for the first River worker registration to land.
package main

import (
	"go.uber.org/fx"

	"github.com/apudiu/quranprism/api/internal/app"
)

func main() {
	fx.New(app.WorkerApp).Run()
}
