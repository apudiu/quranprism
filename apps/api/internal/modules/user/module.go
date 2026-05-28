package user

import (
	"go.uber.org/fx"

	httpmw "github.com/apudiu/quranprism/api/internal/transport/http/middleware"
	"github.com/apudiu/quranprism/api/internal/transport/http/router"
)

// Module wires the user package into the fx graph. NewHandler is
// re-exported through the value group `routes` so app/http.go picks it
// up alongside every other module's registrar.
//
// *Service is also adapted to middleware.EmailVerifier so other
// modules' handlers (e.g. acl) can compose EmailVerifiedRequired
// without importing this package.
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
		// Expose *Service under the middleware.EmailVerifier interface
		// so the acl module's admin handler (and any future caller) can
		// inject the verifier without taking a direct user dependency.
		func(s *Service) httpmw.EmailVerifier { return s },
	),
)
