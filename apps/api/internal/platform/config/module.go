package config

import "go.uber.org/fx"

// Module wires *Config into the fx graph. Every other platform package
// declares *Config as a dependency, so this is the root of the DI tree.
var Module = fx.Module("config",
	fx.Provide(Load),
)
