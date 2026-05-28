package mailer

import "go.uber.org/fx"

// Module provides the Mailer interface backed by SMTP.
var Module = fx.Module("mailer",
	fx.Provide(NewSMTP),
)
