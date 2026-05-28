package app

import (
	"go.uber.org/fx"

	aclmod "github.com/apudiu/quranprism/api/internal/modules/acl"
	auditmod "github.com/apudiu/quranprism/api/internal/modules/audit"
	authmod "github.com/apudiu/quranprism/api/internal/modules/auth"
	usermod "github.com/apudiu/quranprism/api/internal/modules/user"
	"github.com/apudiu/quranprism/api/internal/platform/cache"
	"github.com/apudiu/quranprism/api/internal/platform/config"
	"github.com/apudiu/quranprism/api/internal/platform/db"
	"github.com/apudiu/quranprism/api/internal/platform/jwt"
	"github.com/apudiu/quranprism/api/internal/platform/logger"
	"github.com/apudiu/quranprism/api/internal/platform/mailer"
	"github.com/apudiu/quranprism/api/internal/platform/pubsub"
	"github.com/apudiu/quranprism/api/internal/platform/queue"
)

// Platform bundles every cross-cutting module the api process needs
// before any business logic loads. Worker / cron entrypoints reuse this
// same option.
var Platform = fx.Options(
	config.Module,
	logger.Module,
	db.Module,
	cache.Module,
	pubsub.Module,
	queue.Module,
	mailer.Module,
	jwt.Module,
)

// Domains is the registry of business modules currently in the api.
// Adding a new module = one line here; nothing else.
var Domains = fx.Options(
	aclmod.Module,
	auditmod.Module,
	usermod.Module,
	authmod.Module,
)

// HTTPApp is the full fx graph the api process runs. cmd/api wires
// `fx.New(app.HTTPApp).Run()` and that's the whole of main.
var HTTPApp = fx.Options(
	Platform,
	Domains,
	fx.Provide(NewRouter),
	fx.Invoke(NewHTTPServer),
)

// WorkerApp boots the same platform + domain graph as the api process
// but skips the HTTP listener. Worker bodies (River workers) land in a
// follow-up task; for now the process boots and idles, useful for
// smoke-testing that the platform + seed run cleanly under cmd/worker.
var WorkerApp = fx.Options(
	Platform,
	Domains,
)

// CronApp mirrors WorkerApp; periodic-task wiring lands later.
var CronApp = fx.Options(
	Platform,
	Domains,
)
