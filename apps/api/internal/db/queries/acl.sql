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
INSERT INTO groups (name, description, is_system)
VALUES ($1, $2, $3)
ON CONFLICT (name) DO UPDATE
SET description = EXCLUDED.description,
    is_system   = EXCLUDED.is_system,
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
