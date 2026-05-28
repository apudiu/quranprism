-- name: UpsertPermission :one
INSERT INTO permissions (name, subject, action, description)
VALUES ($1, $2, $3, $4)
ON CONFLICT (name) DO UPDATE
SET subject     = EXCLUDED.subject,
    action      = EXCLUDED.action,
    description = EXCLUDED.description,
    updated_at  = NOW()
RETURNING *;

-- name: UpsertGroup :one
INSERT INTO groups (name, description)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE
SET description = EXCLUDED.description,
    updated_at  = NOW()
RETURNING *;

-- name: LinkGroupPermission :exec
INSERT INTO group_permission (group_id, permission_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: UnlinkGroupPermission :exec
DELETE FROM group_permission
WHERE group_id = $1 AND permission_id = $2;

-- name: GetGroupByName :one
SELECT * FROM groups WHERE name = $1;

-- name: GetPermissionByName :one
SELECT * FROM permissions WHERE name = $1;

-- name: ListPermissionsForGroup :many
SELECT p.*
FROM permissions p
JOIN group_permission gp ON gp.permission_id = p.id
WHERE gp.group_id = $1
ORDER BY p.name;

-- name: JoinUserToGroup :exec
INSERT INTO group_user (group_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveUserFromGroup :exec
DELETE FROM group_user
WHERE group_id = $1 AND user_id = $2;

-- name: ListGroupsForUser :many
SELECT g.*
FROM groups g
JOIN group_user gu ON gu.group_id = g.id
WHERE gu.user_id = $1
ORDER BY g.name;

-- name: ListPermissionsForUser :many
SELECT DISTINCT p.*
FROM permissions p
JOIN group_permission gp ON gp.permission_id = p.id
JOIN group_user       gu ON gu.group_id      = gp.group_id
WHERE gu.user_id = $1
ORDER BY p.name;

-- -- Admin CRUD (T-002) -------------------------------------------------

-- name: CreateGroup :one
-- Distinct from UpsertGroup: this surfaces a unique-constraint violation
-- on (name) so the handler can return 409 instead of silently updating.
INSERT INTO groups (name, description)
VALUES ($1, $2)
RETURNING *;

-- name: ListGroups :many
SELECT *, COUNT(*) OVER() AS total
FROM groups
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = $1;

-- name: UpdateGroup :one
-- Optional fields via sqlc.narg; nil = leave unchanged.
UPDATE groups
SET name        = COALESCE(sqlc.narg('name'),        name),
    description = COALESCE(sqlc.narg('description'), description),
    updated_at  = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteGroup :exec
DELETE FROM groups WHERE id = $1;

-- name: ListPermissions :many
SELECT *, COUNT(*) OVER() AS total
FROM permissions
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: GetPermissionByID :one
SELECT * FROM permissions WHERE id = $1;

-- name: ListMembersOfGroup :many
-- Returns the soft-delete-filtered users who belong to a group, ordered
-- by email. Used by GET /v1/admin/groups/:id detail.
SELECT u.id, u.email, u.name, u.email_verified_at, u.is_disabled, u.created_at
FROM users u
JOIN group_user gu ON gu.user_id = u.id
WHERE gu.group_id = $1 AND u.deleted_at IS NULL
ORDER BY u.email;

-- name: CountUsersWithPermission :one
-- Distinct users currently holding the named permission via any group.
SELECT COUNT(DISTINCT gu.user_id)::BIGINT AS n
FROM group_user gu
JOIN group_permission gp ON gp.group_id = gu.group_id
JOIN permissions p       ON p.id = gp.permission_id
WHERE p.name = $1;

-- name: CountUsersWithPermissionExcludingGroup :one
-- Distinct users holding the perm via any group OTHER than the named
-- one. Used by DeleteGroup / RemoveGroupPermission self-protect: if 0
-- after the mutation, the system would orphan the perm.
SELECT COUNT(DISTINCT gu.user_id)::BIGINT AS n
FROM group_user gu
JOIN group_permission gp ON gp.group_id = gu.group_id
JOIN permissions p       ON p.id = gp.permission_id
WHERE p.name = $1
  AND gu.group_id <> $2;

-- name: CountUsersWithPermissionExcludingMembership :one
-- Distinct users holding the perm, excluding the specific (user, group)
-- membership row. Used by RemoveUserFromGroup self-protect.
SELECT COUNT(DISTINCT gu.user_id)::BIGINT AS n
FROM group_user gu
JOIN group_permission gp ON gp.group_id = gu.group_id
JOIN permissions p       ON p.id = gp.permission_id
WHERE p.name = $1
  AND NOT (gu.user_id = $2 AND gu.group_id = $3);

-- name: GroupHasPermission :one
-- Cheap (boolean) test used by the self-protect guards before paying
-- for the count query.
SELECT EXISTS (
    SELECT 1
    FROM group_permission gp
    JOIN permissions p ON p.id = gp.permission_id
    WHERE gp.group_id = $1 AND p.name = $2
)::BOOLEAN AS has;

-- name: GroupMembershipExists :one
SELECT EXISTS (
    SELECT 1 FROM group_user WHERE group_id = $1 AND user_id = $2
)::BOOLEAN AS exists;

-- name: GroupPermissionLinkExists :one
SELECT EXISTS (
    SELECT 1 FROM group_permission WHERE group_id = $1 AND permission_id = $2
)::BOOLEAN AS exists;
