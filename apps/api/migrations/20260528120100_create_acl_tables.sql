-- +goose Up
-- Permission-Group-User ACL, mirroring the order-online pattern but with
-- the explicit (subject, action) atoms called for by prd/acl.md so
-- middleware can authorize on Permission.Name = "Subject:action" tuples.
CREATE TABLE permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    subject     TEXT NOT NULL,
    action      TEXT NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT permissions_name_unique UNIQUE (name)
);

CREATE INDEX idx_permissions_subject ON permissions (subject);

CREATE TABLE groups (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT,
    -- is_system protects the seed groups (Default user, Super Admin,
    -- Content Manager) from accidental deletion in the admin UI.
    is_system   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT groups_name_unique UNIQUE (name)
);

CREATE TABLE group_user (
    group_id   UUID NOT NULL REFERENCES groups (id)  ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users  (id)  ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX idx_group_user_user_id ON group_user (user_id);

CREATE TABLE group_permission (
    group_id      UUID NOT NULL REFERENCES groups      (id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, permission_id)
);

CREATE INDEX idx_group_permission_permission_id ON group_permission (permission_id);

-- +goose Down
DROP TABLE group_permission;
DROP TABLE group_user;
DROP TABLE groups;
DROP TABLE permissions;
