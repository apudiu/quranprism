package db

import (
	"github.com/jackc/pgx/v5/pgxpool"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
)

// NewQueries wraps the pool in the sqlc-generated *Queries handle.
//
// Modules depend on *sqlcdb.Queries for type-safe SQL. When a module needs
// to issue multiple statements inside an explicit transaction it takes the
// *pgxpool.Pool dependency too, opens a tx, and calls q.WithTx(tx).
func NewQueries(pool *pgxpool.Pool) *sqlcdb.Queries {
	return sqlcdb.New(pool)
}
