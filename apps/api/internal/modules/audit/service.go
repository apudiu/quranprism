package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
)

// Service writes audit_log rows. Stateless aside from the pool-bound
// queries handle; safe to share across goroutines.
type Service struct {
	q *sqlcdb.Queries
}

// NewService is the fx constructor.
func NewService(q *sqlcdb.Queries) *Service {
	return &Service{q: q}
}

// Record writes an audit row using the service's pool-bound queries
// handle. Use when there's no mutation to atomically pair with — e.g.
// the qp admin:grant CLI flow which opens its own transaction and
// passes the tx-bound handle via RecordTx instead.
func (s *Service) Record(ctx context.Context, p Params) error {
	return record(ctx, s.q, p)
}

// RecordTx writes an audit row using the caller's tx-bound queries
// handle. The audit row then commits or rolls back as a unit with the
// mutation it documents.
func (s *Service) RecordTx(ctx context.Context, qtx *sqlcdb.Queries, p Params) error {
	return record(ctx, qtx, p)
}

func record(ctx context.Context, q *sqlcdb.Queries, p Params) error {
	if p.Actor.Kind == "" {
		return fmt.Errorf("audit: actor kind required")
	}
	if p.Action == "" {
		return fmt.Errorf("audit: action required")
	}
	if p.SubjectType == "" {
		return fmt.Errorf("audit: subject_type required")
	}

	var changesJSON []byte
	if p.Changes != nil {
		b, err := json.Marshal(p.Changes)
		if err != nil {
			return fmt.Errorf("audit: marshal changes: %w", err)
		}
		changesJSON = b
	}

	_, err := q.RecordAuditLog(ctx, sqlcdb.RecordAuditLogParams{
		ActorUserID: pgUUID(p.Actor.UserID),
		ActorKind:   p.Actor.Kind,
		Action:      p.Action,
		SubjectType: p.SubjectType,
		SubjectID:   pgUUID(p.SubjectID),
		Changes:     changesJSON,
	})
	if err != nil {
		return fmt.Errorf("audit: record %s: %w", p.Action, err)
	}
	return nil
}

// pgUUID converts the friendly Go-side *uuid.UUID into pgx's pgtype.UUID
// (which models NULL via Valid=false). Centralised here so callers
// build Params with plain pointers and never touch pgtype directly.
func pgUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *u, Valid: true}
}
