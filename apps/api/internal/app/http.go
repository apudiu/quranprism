package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"

	"github.com/apudiu/quranprism/api/internal/platform/config"
	httpmw "github.com/apudiu/quranprism/api/internal/transport/http/middleware"
	"github.com/apudiu/quranprism/api/internal/transport/http/router"
)

// routerIn is the fx parameter object the router builder consumes. The
// value group `routes` is populated by every domain module: each one
// declares its handler implements router.Registrar and the fx graph
// collects them here.
type routerIn struct {
	fx.In

	Cfg        *config.Config
	Log        *slog.Logger
	Registrars []router.Registrar `group:"routes"`
}

// NewRouter assembles the chi router: cross-cutting middleware, the
// always-on /hc health route, then every module's RegisterRoutes.
func NewRouter(in routerIn) *chi.Mux {
	r := chi.NewRouter()

	// Cross-cutting middleware. Order matters: RequestID first so every
	// downstream log line carries it; Recoverer last so a panic in any
	// later middleware also goes through it.
	r.Use(chimiddleware.RequestID)
	r.Use(httpmw.Logger(in.Log))
	r.Use(httpmw.CORS(in.Cfg.App.BaseURL))
	r.Use(chimiddleware.Recoverer)

	// Health endpoint, kept at /hc to satisfy the existing k8s probes.
	r.Get("/hc", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Every domain module's handler self-registers here.
	for _, reg := range in.Registrars {
		reg.RegisterRoutes(r)
	}

	return r
}

// NewHTTPServer wires the chi router into an *http.Server and binds its
// lifecycle to the fx graph. Listening starts in OnStart; OnStop runs a
// bounded-graceful Shutdown so in-flight requests get a chance to finish.
func NewHTTPServer(lc fx.Lifecycle, cfg *config.Config, log *slog.Logger, h *chi.Mux) *http.Server {
	srv := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			addr := net.JoinHostPort("", strconv.Itoa(cfg.HTTP.Port))
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				return fmt.Errorf("http: listen: %w", err)
			}
			log.Info("http listening", "addr", ln.Addr().String())
			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					log.Error("http: serve", "err", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			return srv.Shutdown(shutdownCtx)
		},
	})

	return srv
}
