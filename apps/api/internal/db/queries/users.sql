-- name: CreateUser :one
INSERT INTO users (email, name, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL;

-- name: IsEmailVerified :one
SELECT (email_verified_at IS NOT NULL)::BOOLEAN AS verified
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserProfile :one
UPDATE users
SET name       = COALESCE(sqlc.narg('name'),       name),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2,
    updated_at    = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: MarkUserEmailVerified :exec
UPDATE users
SET email_verified_at = COALESCE(email_verified_at, NOW()),
    updated_at        = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: StartUserDeletionGrace :exec
UPDATE users
SET deletion_requested_at = NOW(),
    updated_at            = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: CancelUserDeletionGrace :exec
UPDATE users
SET deletion_requested_at = NULL,
    updated_at            = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SetUserDisabled :exec
UPDATE users
SET is_disabled = $2,
    updated_at  = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListUsers :many
-- Admin user-listing. Optional case-insensitive substring filter on
-- email; nil = no filter. Excludes soft-deleted users.
SELECT id, email, name, email_verified_at, is_disabled, created_at,
       COUNT(*) OVER() AS total
FROM users
WHERE deleted_at IS NULL
  AND (sqlc.narg('email_like')::TEXT IS NULL
       OR email ILIKE '%' || sqlc.narg('email_like')::TEXT || '%')
ORDER BY email
LIMIT $1 OFFSET $2;
