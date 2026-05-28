// Package api owns embedded migrations exposed to cmd/qp's migrate
// subcommand so a single binary contains everything goose needs to run
// schema migrations in any environment.
package api

import "embed"

// MigrationsFS holds every .sql migration in apps/api/migrations/. The
// directory layout (`migrations/<timestamp>_<name>.sql`) matches what
// goose expects.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
