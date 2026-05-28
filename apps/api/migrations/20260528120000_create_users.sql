-- +goose Up
-- pgcrypto gives us gen_random_uuid() for application-friendly UUIDs;
-- citext keeps the email comparison case-insensitive without juggling
-- LOWER() everywhere.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email                    CITEXT NOT NULL,
    name                     TEXT NOT NULL,
    password_hash            TEXT NOT NULL,
    avatar_url               TEXT,
    -- NULL until the user clicks the verification link in their inbox.
    email_verified_at        TIMESTAMPTZ,
    -- Admin-flippable kill switch (ACC-7). Disabled users cannot sign in,
    -- but their content (playlists, comments, …) remains untouched.
    is_disabled              BOOLEAN NOT NULL DEFAULT FALSE,
    -- Set when the user starts the PRV-4 30-day deletion grace window.
    -- A successful login inside the window clears this back to NULL.
    deletion_requested_at    TIMESTAMPTZ,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Set when the post-grace cron has hard-deleted the row's payload.
    -- The id stays around as a tombstone so FKs from comments/etc. stay
    -- intact and authors render as "[deleted]" (PRV-7).
    deleted_at               TIMESTAMPTZ,
    CONSTRAINT users_email_unique UNIQUE (email)
);

-- Active accounts only: filter out tombstones when uniqueness on email is
-- needed (two real users can't share an inbox; a tombstoned row freeing
-- up its email is fine).
CREATE INDEX idx_users_active ON users (email) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE users;
-- Extensions are left in place: dropping them in down might break other
-- tables that pick up the same dependency later. Database is forward-only
-- in prod anyway.
