package pubsub

import "go.uber.org/fx"

// Module provides *Conn (NATS connection + JetStream context).
var Module = fx.Module("pubsub",
	fx.Provide(New),
)
