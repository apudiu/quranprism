-- +goose Up
-- T-002 / PRD ACL-10, ADM-12, ADM-14: append-only audit log for every
-- admin mutation (group CRUD, group_permission link/unlink, group_user
-- add/remove, and the qp admin:grant CLI bootstrap). The shared shape
-- is reused by the catalog admin task once it ships.
--
-- actor_user_id is nullable so CLI / system actors can record without a
-- bound user. actor_kind disambiguates ('user' / 'cli' / 'system').
-- subject_id is nullable for subjects that aren't UUID-keyed.
CREATE TABLE audit_log (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_kind    TEXT NOT NULL CHECK (actor_kind IN ('user', 'cli', 'system')),
    action        TEXT NOT NULL,
    subject_type  TEXT NOT NULL,
    subject_id    UUID,
    changes       JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_actor_user_id   ON audit_log (actor_user_id);
CREATE INDEX idx_audit_log_subject         ON audit_log (subject_type, subject_id);
CREATE INDEX idx_audit_log_created_at_desc ON audit_log (created_at DESC);

-- +goose Down
DROP TABLE audit_log;
