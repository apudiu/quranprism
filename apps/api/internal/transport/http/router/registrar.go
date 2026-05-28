// Package router holds tiny leaf types shared between domain modules and
// the app composition root. The split exists to break the import cycle
// "app imports modules; modules need RouteRegistrar" — neither side
// should depend on the other directly.
package router

import "github.com/go-chi/chi/v5"

// Registrar is the contract every domain module's HTTP handler
// implements. The handler receives a router scope and mounts its
// endpoints — usually under a versioned prefix like `r.Route("/v1/...", ...)`.
type Registrar interface {
	RegisterRoutes(r chi.Router)
}
