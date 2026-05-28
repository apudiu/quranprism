-- +goose Up
-- Backs the ACC-6 lockout policy: 10 failed attempts inside 15 minutes
-- locks the (email, ip) bucket. Successful attempts are also logged so
-- the cooldown query can prove the user got in cleanly.
CREATE TABLE login_attempts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      CITEXT NOT NULL,
    ip         INET,
    succeeded  BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Partial index on the only hot read path: "did this email rack up >=10
-- failures in the last 15 min?". Successful rows aren't queried by the
-- lockout check so they're left out of the index.
CREATE INDEX idx_login_attempts_recent_failed
    ON login_attempts (email, created_at)
    WHERE succeeded = FALSE;

-- +goose Down
DROP TABLE login_attempts;
