package logger

import "go.uber.org/fx"

// Module provides the application's root *slog.Logger.
var Module = fx.Module("logger",
	fx.Provide(New),
)
