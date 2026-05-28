-- name: CreateRefreshSession :one
INSERT INTO refresh_sessions (user_id, token_hash, user_agent, ip, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRefreshSessionByTokenHash :one
SELECT * FROM refresh_sessions
WHERE token_hash = $1
  AND revoked_at IS NULL
  AND expires_at > NOW();

-- name: TouchRefreshSession :exec
UPDATE refresh_sessions
SET last_used_at = NOW()
WHERE id = $1;

-- name: RevokeRefreshSession :exec
UPDATE refresh_sessions
SET revoked_at = NOW()
WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeAllRefreshSessionsForUser :exec
UPDATE refresh_sessions
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: RevokeOtherRefreshSessionsForUser :exec
UPDATE refresh_sessions
SET revoked_at = NOW()
WHERE user_id = $1 AND id <> $2 AND revoked_at IS NULL;

-- name: DeleteExpiredRefreshSessions :exec
DELETE FROM refresh_sessions
WHERE expires_at < NOW() OR (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '7 days');
