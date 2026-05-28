-- +goose Up
-- Each refresh cookie corresponds to one row here. The cookie carries the
-- random plaintext; we persist only its SHA-256 digest so a DB leak can't
-- be replayed. user_agent / ip help the user audit "where am I signed in"
-- in a future devices view.
CREATE TABLE refresh_sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash   TEXT NOT NULL,
    user_agent   TEXT,
    ip           INET,
    expires_at   TIMESTAMPTZ NOT NULL,
    -- Set on rotation, explicit logout, or change-password "revoke all".
    revoked_at   TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT refresh_sessions_token_hash_unique UNIQUE (token_hash)
);

CREATE INDEX idx_refresh_user_id ON refresh_sessions (user_id);

-- +goose Down
DROP TABLE refresh_sessions;
