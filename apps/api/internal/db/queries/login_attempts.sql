-- name: RecordLoginAttempt :exec
INSERT INTO login_attempts (email, ip, succeeded)
VALUES ($1, $2, $3);

-- name: CountRecentFailedLogins :one
-- Drives the ACC-6 lockout: caller passes the cutoff (NOW() - window).
SELECT COUNT(*)::INT AS failures
FROM login_attempts
WHERE email = $1
  AND succeeded = FALSE
  AND created_at >= $2;

-- name: DeleteOldLoginAttempts :exec
-- Keeps the table from growing unboundedly. Cron-driven; safe to run any time.
DELETE FROM login_attempts WHERE created_at < NOW() - INTERVAL '30 days';
