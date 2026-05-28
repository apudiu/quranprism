package jwt

import "go.uber.org/fx"

// Module provides the HS256-backed jwt.Service.
var Module = fx.Module("jwt",
	fx.Provide(New),
)
