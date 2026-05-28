-- +goose Up
-- T-002: system groups no longer exist; admins create groups post-deploy
-- via `qp admin:grant`. The is_system protection bit is unused — drop it.
ALTER TABLE groups DROP COLUMN is_system;

-- +goose Down
ALTER TABLE groups ADD COLUMN is_system BOOLEAN NOT NULL DEFAULT FALSE;
