package audit

import "go.uber.org/fx"

// Module wires the audit service into the fx graph. No routes are
// registered yet — the listing endpoint (ADM-13) ships in a later task.
var Module = fx.Module("audit",
	fx.Provide(NewService),
)
