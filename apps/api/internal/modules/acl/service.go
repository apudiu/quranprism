package acl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
	"github.com/apudiu/quranprism/api/internal/modules/audit"
	"github.com/apudiu/quranprism/api/internal/transport/http/httperr"
)

// Service owns the ACL domain logic: group CRUD, permission catalog
// reads, user-group membership, and admin user-listing. Every mutating
// method commits the mutation and its audit row inside a single
// transaction.
type Service struct {
	q     *sqlcdb.Queries
	pool  *pgxpool.Pool
	audit *audit.Service
}

// NewService wires the acl Service. *pgxpool.Pool comes from
// platform/db so the service can manage its own transactions for the
// audit-in-tx pattern.
func NewService(q *sqlcdb.Queries, pool *pgxpool.Pool, auditSvc *audit.Service) *Service {
	return &Service{q: q, pool: pool, audit: auditSvc}
}

const (
	groupNameMaxLen        = 100
	groupDescriptionMaxLen = 1000
)

// Existing helpers retained --------------------------------------------

// JoinGroupByName drops a user into a group looked up by name. Used by
// the `qp admin:grant` CLI bootstrap; idempotent.
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
// to, alphabetical.
func (s *Service) ListGroupsForUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	groups, err := s.q.ListGroupsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("acl: list groups for user: %w", err)
	}
	names := make([]string, len(groups))
	for i, g := range groups {
		names[i] = g.Name
	}
	return names, nil
}

// ListPermissionsForUser returns the flat union of perm names from
// every group the user belongs to. Implements middleware.PermissionLister.
func (s *Service) ListPermissionsForUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	perms, err := s.q.ListPermissionsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("acl: list permissions for user: %w", err)
	}
	names := make([]string, len(perms))
	for i, p := range perms {
		names[i] = p.Name
	}
	return names, nil
}

// Admin: group CRUD ---------------------------------------------------

// ListGroups returns a page of groups, alphabetical.
func (s *Service) ListGroups(ctx context.Context, limit, offset int) ([]GroupView, int64, error) {
	rows, err := s.q.ListGroups(ctx, sqlcdb.ListGroupsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("acl: list groups: %w", err)
	}
	items := make([]GroupView, len(rows))
	var total int64
	for i, row := range rows {
		items[i] = GroupView{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}
		total = row.Total
	}
	return items, total, nil
}

// GetGroup returns a group with its permission set populated.
func (s *Service) GetGroup(ctx context.Context, id uuid.UUID) (*GroupView, error) {
	row, err := s.q.GetGroupByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, fmt.Errorf("acl: get group: %w", err)
	}
	perms, err := s.q.ListPermissionsForGroup(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("acl: list perms for group: %w", err)
	}
	names := make([]string, len(perms))
	for i, p := range perms {
		names[i] = p.Name
	}
	return &GroupView{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		Permissions: names,
	}, nil
}

// CreateGroup persists a new group + audit row in one tx. Returns
// ErrGroupNameTaken on unique violation; ErrUnprocessable on bad input.
func (s *Service) CreateGroup(ctx context.Context, actor audit.Actor, req CreateGroupRequest) (*GroupView, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, httperr.Unprocessable("name required")
	}
	if len(name) > groupNameMaxLen {
		return nil, httperr.Unprocessable("name too long")
	}
	if req.Description != nil && len(*req.Description) > groupDescriptionMaxLen {
		return nil, httperr.Unprocessable("description too long")
	}

	var view *GroupView
	err := s.withTx(ctx, func(qtx *sqlcdb.Queries) error {
		row, err := qtx.CreateGroup(ctx, sqlcdb.CreateGroupParams{
			Name:        name,
			Description: req.Description,
		})
		if err != nil {
			if isUniqueViolation(err, "groups_name_unique") {
				return ErrGroupNameTaken
			}
			return fmt.Errorf("acl: create group: %w", err)
		}
		view = &GroupView{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			Permissions: []string{},
		}
		subjectID := row.ID
		return s.audit.RecordTx(ctx, qtx, audit.Params{
			Actor:       actor,
			Action:      "group.create",
			SubjectType: "group",
			SubjectID:   &subjectID,
			Changes:     map[string]any{"name": row.Name, "description": row.Description},
		})
	})
	return view, err
}

// UpdateGroup applies optional name / description changes.
func (s *Service) UpdateGroup(ctx context.Context, actor audit.Actor, id uuid.UUID, req UpdateGroupRequest) (*GroupView, error) {
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return nil, httperr.Unprocessable("name required")
		}
		if len(trimmed) > groupNameMaxLen {
			return nil, httperr.Unprocessable("name too long")
		}
		req.Name = &trimmed
	}
	if req.Description != nil && len(*req.Description) > groupDescriptionMaxLen {
		return nil, httperr.Unprocessable("description too long")
	}

	var view *GroupView
	err := s.withTx(ctx, func(qtx *sqlcdb.Queries) error {
		row, err := qtx.UpdateGroup(ctx, sqlcdb.UpdateGroupParams{
			Name:        req.Name,
			Description: req.Description,
			ID:          id,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrGroupNotFound
			}
			if isUniqueViolation(err, "groups_name_unique") {
				return ErrGroupNameTaken
			}
			return fmt.Errorf("acl: update group: %w", err)
		}
		view = &GroupView{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}
		subjectID := row.ID
		changes := map[string]any{}
		if req.Name != nil {
			changes["name"] = *req.Name
		}
		if req.Description != nil {
			changes["description"] = *req.Description
		}
		return s.audit.RecordTx(ctx, qtx, audit.Params{
			Actor:       actor,
			Action:      "group.update",
			SubjectType: "group",
			SubjectID:   &subjectID,
			Changes:     changes,
		})
	})
	return view, err
}

// DeleteGroup removes the group + cascades. Self-protect guard runs
// first: if the group holds `group:update` and no other group grants
// that perm to a user, the delete is refused with ErrLastAdmin.
func (s *Service) DeleteGroup(ctx context.Context, actor audit.Actor, id uuid.UUID) error {
	return s.withTx(ctx, func(qtx *sqlcdb.Queries) error {
		// Verify exists, snapshot name for the audit row.
		row, err := qtx.GetGroupByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrGroupNotFound
			}
			return fmt.Errorf("acl: get group: %w", err)
		}

		if err := s.guardLastAdminGroupDelete(ctx, qtx, id); err != nil {
			return err
		}

		if err := qtx.DeleteGroup(ctx, id); err != nil {
			return fmt.Errorf("acl: delete group: %w", err)
		}
		subjectID := id
		return s.audit.RecordTx(ctx, qtx, audit.Params{
			Actor:       actor,
			Action:      "group.delete",
			SubjectType: "group",
			SubjectID:   &subjectID,
			Changes:     map[string]any{"name": row.Name},
		})
	})
}

// Admin: group ↔ permission link ---------------------------------------

func (s *Service) AddGroupPermission(ctx context.Context, actor audit.Actor, groupID, permID uuid.UUID) error {
	return s.withTx(ctx, func(qtx *sqlcdb.Queries) error {
		if _, err := qtx.GetGroupByID(ctx, groupID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrGroupNotFound
			}
			return fmt.Errorf("acl: get group: %w", err)
		}
		perm, err := qtx.GetPermissionByID(ctx, permID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPermissionNotFound
			}
			return fmt.Errorf("acl: get permission: %w", err)
		}
		if err := qtx.LinkGroupPermission(ctx, sqlcdb.LinkGroupPermissionParams{
			GroupID:      groupID,
			PermissionID: permID,
		}); err != nil {
			return fmt.Errorf("acl: link group permission: %w", err)
		}
		subjectID := groupID
		return s.audit.RecordTx(ctx, qtx, audit.Params{
			Actor:       actor,
			Action:      "group_permission.add",
			SubjectType: "group",
			SubjectID:   &subjectID,
			Changes:     map[string]any{"permission_id": permID, "permission": perm.Name},
		})
	})
}

func (s *Service) RemoveGroupPermission(ctx context.Context, actor audit.Actor, groupID, permID uuid.UUID) error {
	return s.withTx(ctx, func(qtx *sqlcdb.Queries) error {
		if _, err := qtx.GetGroupByID(ctx, groupID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrGroupNotFound
			}
			return fmt.Errorf("acl: get group: %w", err)
		}
		perm, err := qtx.GetPermissionByID(ctx, permID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPermissionNotFound
			}
			return fmt.Errorf("acl: get permission: %w", err)
		}
		linked, err := qtx.GroupPermissionLinkExists(ctx, sqlcdb.GroupPermissionLinkExistsParams{
			GroupID:      groupID,
			PermissionID: permID,
		})
		if err != nil {
			return fmt.Errorf("acl: check link: %w", err)
		}
		if !linked {
			return ErrPermissionLinkNotFound
		}

		if err := s.guardLastAdminPermUnlink(ctx, qtx, groupID, perm.Name); err != nil {
			return err
		}

		if err := qtx.UnlinkGroupPermission(ctx, sqlcdb.UnlinkGroupPermissionParams{
			GroupID:      groupID,
			PermissionID: permID,
		}); err != nil {
			return fmt.Errorf("acl: unlink group permission: %w", err)
		}
		subjectID := groupID
		return s.audit.RecordTx(ctx, qtx, audit.Params{
			Actor:       actor,
			Action:      "group_permission.remove",
			SubjectType: "group",
			SubjectID:   &subjectID,
			Changes:     map[string]any{"permission_id": permID, "permission": perm.Name},
		})
	})
}

// Admin: user ↔ group membership --------------------------------------

func (s *Service) AddUserToGroup(ctx context.Context, actor audit.Actor, userID, groupID uuid.UUID) error {
	return s.withTx(ctx, func(qtx *sqlcdb.Queries) error {
		if _, err := qtx.GetGroupByID(ctx, groupID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrGroupNotFound
			}
			return fmt.Errorf("acl: get group: %w", err)
		}
		if _, err := qtx.GetUserByID(ctx, userID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrUserNotFound
			}
			return fmt.Errorf("acl: get user: %w", err)
		}
		if err := qtx.JoinUserToGroup(ctx, sqlcdb.JoinUserToGroupParams{
			GroupID: groupID,
			UserID:  userID,
		}); err != nil {
			return fmt.Errorf("acl: join user→group: %w", err)
		}
		subjectID := userID
		return s.audit.RecordTx(ctx, qtx, audit.Params{
			Actor:       actor,
			Action:      "group_membership.add",
			SubjectType: "user",
			SubjectID:   &subjectID,
			Changes:     map[string]any{"group_id": groupID},
		})
	})
}

func (s *Service) RemoveUserFromGroup(ctx context.Context, actor audit.Actor, userID, groupID uuid.UUID) error {
	return s.withTx(ctx, func(qtx *sqlcdb.Queries) error {
		if _, err := qtx.GetGroupByID(ctx, groupID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrGroupNotFound
			}
			return fmt.Errorf("acl: get group: %w", err)
		}
		exists, err := qtx.GroupMembershipExists(ctx, sqlcdb.GroupMembershipExistsParams{
			GroupID: groupID,
			UserID:  userID,
		})
		if err != nil {
			return fmt.Errorf("acl: check membership: %w", err)
		}
		if !exists {
			return ErrMembershipNotFound
		}

		if err := s.guardLastAdminMembershipRemove(ctx, qtx, userID, groupID); err != nil {
			return err
		}

		if err := qtx.RemoveUserFromGroup(ctx, sqlcdb.RemoveUserFromGroupParams{
			GroupID: groupID,
			UserID:  userID,
		}); err != nil {
			return fmt.Errorf("acl: remove user from group: %w", err)
		}
		subjectID := userID
		return s.audit.RecordTx(ctx, qtx, audit.Params{
			Actor:       actor,
			Action:      "group_membership.remove",
			SubjectType: "user",
			SubjectID:   &subjectID,
			Changes:     map[string]any{"group_id": groupID},
		})
	})
}

// Admin: permission catalog (read-only) -------------------------------

func (s *Service) ListPermissions(ctx context.Context, limit, offset int) ([]PermissionView, int64, error) {
	rows, err := s.q.ListPermissions(ctx, sqlcdb.ListPermissionsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("acl: list permissions: %w", err)
	}
	items := make([]PermissionView, len(rows))
	var total int64
	for i, row := range rows {
		items[i] = PermissionView{
			ID:          row.ID,
			Name:        row.Name,
			Subject:     row.Subject,
			Action:      row.Action,
			Description: row.Description,
		}
		total = row.Total
	}
	return items, total, nil
}

func (s *Service) GetPermission(ctx context.Context, id uuid.UUID) (*PermissionView, error) {
	row, err := s.q.GetPermissionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPermissionNotFound
		}
		return nil, fmt.Errorf("acl: get permission: %w", err)
	}
	return &PermissionView{
		ID:          row.ID,
		Name:        row.Name,
		Subject:     row.Subject,
		Action:      row.Action,
		Description: row.Description,
	}, nil
}

// Admin: user listing (read-only) -------------------------------------

func (s *Service) ListUsers(ctx context.Context, limit, offset int, emailLike *string) ([]UserView, int64, error) {
	rows, err := s.q.ListUsers(ctx, sqlcdb.ListUsersParams{
		Limit:     int32(limit),
		Offset:    int32(offset),
		EmailLike: emailLike,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("acl: list users: %w", err)
	}
	items := make([]UserView, len(rows))
	var total int64
	for i, row := range rows {
		items[i] = UserView{
			ID:            row.ID,
			Email:         row.Email,
			Name:          row.Name,
			EmailVerified: row.EmailVerifiedAt.Valid,
			IsDisabled:    row.IsDisabled,
			CreatedAt:     row.CreatedAt,
		}
		total = row.Total
	}
	return items, total, nil
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*UserView, error) {
	row, err := s.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("acl: get user: %w", err)
	}
	groups, err := s.ListGroupsForUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return &UserView{
		ID:            row.ID,
		Email:         row.Email,
		Name:          row.Name,
		EmailVerified: row.EmailVerifiedAt.Valid,
		IsDisabled:    row.IsDisabled,
		CreatedAt:     row.CreatedAt,
		Groups:        groups,
	}, nil
}

// Self-protect guards (Phase 6) ---------------------------------------

// guardLastAdminGroupDelete refuses to delete a group when doing so
// would leave zero users holding `group:update`.
func (s *Service) guardLastAdminGroupDelete(ctx context.Context, qtx *sqlcdb.Queries, groupID uuid.UUID) error {
	has, err := qtx.GroupHasPermission(ctx, sqlcdb.GroupHasPermissionParams{
		GroupID: groupID,
		Name:    grantCapablePermission,
	})
	if err != nil {
		return fmt.Errorf("acl: guard last-admin check: %w", err)
	}
	if !has {
		return nil
	}
	n, err := qtx.CountUsersWithPermissionExcludingGroup(ctx, sqlcdb.CountUsersWithPermissionExcludingGroupParams{
		Name:    grantCapablePermission,
		GroupID: groupID,
	})
	if err != nil {
		return fmt.Errorf("acl: guard count: %w", err)
	}
	if n == 0 {
		return ErrLastAdmin
	}
	return nil
}

// guardLastAdminPermUnlink refuses to unlink `group:update` from a
// group when that would leave zero users holding the perm overall.
func (s *Service) guardLastAdminPermUnlink(ctx context.Context, qtx *sqlcdb.Queries, groupID uuid.UUID, permName string) error {
	if permName != grantCapablePermission {
		return nil
	}
	n, err := qtx.CountUsersWithPermissionExcludingGroup(ctx, sqlcdb.CountUsersWithPermissionExcludingGroupParams{
		Name:    grantCapablePermission,
		GroupID: groupID,
	})
	if err != nil {
		return fmt.Errorf("acl: guard count: %w", err)
	}
	if n == 0 {
		return ErrLastAdmin
	}
	return nil
}

// guardLastAdminMembershipRemove refuses to remove a user from a group
// when that membership is their only path to `group:update`.
func (s *Service) guardLastAdminMembershipRemove(ctx context.Context, qtx *sqlcdb.Queries, userID, groupID uuid.UUID) error {
	has, err := qtx.GroupHasPermission(ctx, sqlcdb.GroupHasPermissionParams{
		GroupID: groupID,
		Name:    grantCapablePermission,
	})
	if err != nil {
		return fmt.Errorf("acl: guard last-admin check: %w", err)
	}
	if !has {
		return nil
	}
	n, err := qtx.CountUsersWithPermissionExcludingMembership(ctx, sqlcdb.CountUsersWithPermissionExcludingMembershipParams{
		Name:    grantCapablePermission,
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		return fmt.Errorf("acl: guard count: %w", err)
	}
	if n == 0 {
		return ErrLastAdmin
	}
	return nil
}

// Internal: tx helper + error helpers ---------------------------------

// withTx runs fn inside a Postgres transaction with a tx-bound queries
// handle. Commits on nil error; rolls back otherwise. Domain sentinels
// (ErrGroupNotFound, ErrLastAdmin, etc.) abort the tx as expected — the
// deferred Rollback is a no-op after a successful Commit.
func (s *Service) withTx(ctx context.Context, fn func(qtx *sqlcdb.Queries) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("acl: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(s.q.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("acl: commit tx: %w", err)
	}
	return nil
}

// isUniqueViolation matches a pgx unique-constraint error against a
// specific constraint name. Same shape as user.isUniqueViolation but
// kept local to avoid cross-module helper coupling.
func isUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "23505" {
		return false
	}
	return pgErr.ConstraintName == constraintName
}
