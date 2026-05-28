package db

import "go.uber.org/fx"

// Module provides *pgxpool.Pool and the sqlc-generated *Queries layered
// on top of it. Modules typically depend on *Queries for type-safe SQL
// and reach for *pgxpool.Pool only when they need an explicit transaction.
var Module = fx.Module("db",
	fx.Provide(NewPool),
	fx.Provide(NewQueries),
)
