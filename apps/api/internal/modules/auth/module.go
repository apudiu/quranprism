package auth

import (
	"go.uber.org/fx"

	"github.com/apudiu/quranprism/api/internal/transport/http/router"
)

// Module wires the auth package into the fx graph.
var Module = fx.Module("auth",
	fx.Provide(
		NewRepository,
		NewService,
		NewHandler,
		fx.Annotate(
			func(h *Handler) router.Registrar { return h },
			fx.ResultTags(`group:"routes"`),
		),
	),
)
