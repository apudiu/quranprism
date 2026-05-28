package queue

import "go.uber.org/fx"

// Module provides the insert-only River *Client. The worker process
// composes its own module that adds workers + Start().
var Module = fx.Module("queue",
	fx.Provide(NewInsertOnly),
)
