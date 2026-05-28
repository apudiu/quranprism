// Package audit ships the append-only audit log required by PRD ACL-10,
// ADM-12 and ADM-14. Every admin mutation records exactly one row,
// committed inside the same pgx.Tx as the mutation itself so the log
// reflects committed reality only.
//
// Callers usually pass a tx-bound *sqlcdb.Queries to RecordTx so the
// audit row joins the mutation's transaction. Record (without Tx) opens
// its own implicit single-statement transaction — used by callers that
// have no mutation to atomically pair with (e.g. CLI bootstrap).
package audit

import "github.com/google/uuid"

// Actor identifies who performed the action.
//
//   - "user" — a logged-in user; UserID is the JWT subject.
//   - "cli"  — invoked from a qp CLI command; UserID is usually nil.
//   - "system" — internal automation (jobs, schedulers); UserID nil.
type Actor struct {
	UserID *uuid.UUID
	Kind   string
}

// Params is the shape recorded for one mutation. Changes is a free-form
// map serialised to JSONB; keep the keys stable per Action so the log
// is queryable by future tooling.
type Params struct {
	Actor       Actor
	Action      string         // e.g. "group.create", "group_permission.add"
	SubjectType string         // e.g. "group", "user", "permission"
	SubjectID   *uuid.UUID     // optional — nil if the subject isn't UUID-keyed
	Changes     map[string]any // optional — diff or full new state
}

// Kind constants for Actor.Kind, exported so callers don't fat-finger
// the string literal.
const (
	KindUser   = "user"
	KindCLI    = "cli"
	KindSystem = "system"
)
