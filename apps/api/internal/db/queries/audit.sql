-- name: RecordAuditLog :one
INSERT INTO audit_log (actor_user_id, actor_kind, action, subject_type, subject_id, changes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListAuditLog :many
-- Paginated list with total via COUNT(*) OVER(). Ordered newest-first.
-- Filters are optional via sqlc.narg so the same query supports
-- "all", "by actor", "by subject", or "by subject_type" lookups.
SELECT *, COUNT(*) OVER() AS total
FROM audit_log
WHERE (sqlc.narg('actor_user_id')::UUID IS NULL OR actor_user_id = sqlc.narg('actor_user_id'))
  AND (sqlc.narg('subject_type')::TEXT IS NULL OR subject_type = sqlc.narg('subject_type'))
  AND (sqlc.narg('subject_id')::UUID   IS NULL OR subject_id   = sqlc.narg('subject_id'))
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
