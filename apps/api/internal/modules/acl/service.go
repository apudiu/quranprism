package acl

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
)

// Service is the small acl API the auth + user modules consume in this
// phase. Permission-check middleware (RequirePermission(...)) lands in
// the dedicated ACL task.
type Service struct {
	q *sqlcdb.Queries
}

// NewService wires the acl Service against the shared sqlc Queries.
func NewService(q *sqlcdb.Queries) *Service { return &Service{q: q} }

// JoinGroupByName drops a user into a system group looked up by name.
// Idempotent — calling it twice for the same (user, group) is a no-op.
// Errors if the named group doesn't exist.
func (s *Service) JoinGroupByName(ctx context.Context, userID uuid.UUID, groupName string) error {
	g, err := s.q.GetGroupByName(ctx, groupName)
	if err != nil {
		return fmt.Errorf("acl: get group %q: %w", groupName, err)
	}
	if err := s.q.JoinUserToGroup(ctx, sqlcdb.JoinUserToGroupParams{
		GroupID: g.ID,
		UserID:  userID,
	}); err != nil {
		return fmt.Errorf("acl: join user→group: %w", err)
	}
	return nil
}

// ListGroupsForUser returns the names of every group the user belongs
// to. Stable alphabetical order.
func (s *Service) ListGroupsForUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	groups, err := s.q.ListGroupsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("acl: list groups: %w", err)
	}
	names := make([]string, len(groups))
	for i, g := range groups {
		names[i] = g.Name
	}
	return names, nil
}

// ListPermissionsForUser returns the flat union of permission names from
// every group the user belongs to. Used by the (forthcoming)
// RequirePermission middleware; safe to call from any handler.
func (s *Service) ListPermissionsForUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	perms, err := s.q.ListPermissionsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("acl: list permissions: %w", err)
	}
	names := make([]string, len(perms))
	for i, p := range perms {
		names[i] = p.Name
	}
	return names, nil
}
