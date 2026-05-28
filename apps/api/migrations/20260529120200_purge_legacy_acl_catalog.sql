-- +goose Up
-- T-002 cleanup: the pre-T-002 22-perm PascalCase catalog and the three
-- seeded system groups (Default user / Super Admin / Content Manager)
-- are gone from the code. Any rows left in dev/staging DBs from before
-- this migration are residue — purge them so the API surface matches
-- the new catalog.
--
-- Fresh prod deploys won't have these rows; this migration is a no-op
-- there. Existing dev DBs lose stale data + pivot rows (via CASCADE).
--
-- Whitelist the 7 lowercase perms T-002 ships. Anything else in
-- `permissions` was seeded by the old catalog and is no longer
-- enforced anywhere.
DELETE FROM groups
 WHERE name IN ('Default user', 'Super Admin', 'Content Manager');

DELETE FROM permissions
 WHERE name NOT IN (
     'group:create',
     'group:view',
     'group:update',
     'group:delete',
     'permission:view',
     'user:view',
     'audit_log:view'
 );

-- +goose Down
-- Irreversible: legacy rows are gone; the old seed code that recreated
-- them is gone with the catalog rewrite. A rollback would have to
-- re-seed from the historical PR.
SELECT 1;
