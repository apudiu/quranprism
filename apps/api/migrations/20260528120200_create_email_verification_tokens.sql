-- +goose Up
-- Per PRV-11, only the SHA-256 hash of the random token is stored. The
-- plaintext lives in the user's inbox; if we never persist it, a DB leak
-- can't be replayed against verify endpoints.
CREATE TABLE email_verification_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT email_verification_tokens_hash_unique UNIQUE (token_hash)
);

CREATE INDEX idx_evt_user_id ON email_verification_tokens (user_id);

-- +goose Down
DROP TABLE email_verification_tokens;
