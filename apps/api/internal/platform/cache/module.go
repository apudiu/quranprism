package cache

import "go.uber.org/fx"

// Module provides *redis.Client.
var Module = fx.Module("cache",
	fx.Provide(NewClient),
)
