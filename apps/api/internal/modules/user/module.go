package user

import (
	"go.uber.org/fx"

	"github.com/apudiu/quranprism/api/internal/transport/http/router"
)

// Module wires the user package into the fx graph. NewHandler is
// re-exported through the value group `routes` so app/http.go picks it
// up alongside every other module's registrar.
var Module = fx.Module("user",
	fx.Provide(
		NewRepository,
		NewService,
		NewHandler,
		// Route registration: tag the *Handler as an app.RouteRegistrar and
		// drop it into the `routes` value group consumed by NewRouter.
		fx.Annotate(
			func(h *Handler) router.Registrar { return h },
			fx.ResultTags(`group:"routes"`),
		),
	),
)
