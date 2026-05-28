package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
	"github.com/apudiu/quranprism/api/internal/modules/acl"
	"github.com/apudiu/quranprism/api/internal/modules/audit"
)

// admin:grant is the bootstrap-admin path: with no system groups, a
// fresh deploy has no users holding `group:update`, so nobody can use
// the admin API. This command opens a tx that:
//
//  1. Runs acl.Seed so the permission catalog is present even if
//     serve:api has never booted. Idempotent (upsert per perm).
//  2. Upserts the named group (default "Admin"; idempotent on re-run).
//  3. Links every permission from acl.PermissionCatalog to it.
//  4. Joins the named user to the group.
//  5. Records an audit_log row with actor_kind=cli.
//
// The whole sequence commits as one tx. Safe to re-run — every step
// uses upsert / ON CONFLICT DO NOTHING semantics.
func newAdminGrantCmd() *cobra.Command {
	var email, group string
	cmd := &cobra.Command{
		Use:   "admin:grant",
		Short: "Grant a user every admin permission (idempotent bootstrap)",
		RunE: func(*cobra.Command, []string) error {
			email = strings.TrimSpace(strings.ToLower(email))
			group = strings.TrimSpace(group)
			if email == "" {
				return errors.New("--email required")
			}
			if group == "" {
				return errors.New("--group cannot be empty")
			}
			return runAdminGrant(email, group)
		},
	}
	cmd.Flags().StringVar(&email, "email", "", "Email of the user to grant admin perms to")
	cmd.Flags().StringVar(&group, "group", "Admin", "Name of the group to create/use")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func runAdminGrant(email, groupName string) error {
	dsn, err := databaseURL()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("pool: %w", err)
	}
	defer pool.Close()

	q := sqlcdb.New(pool)
	auditSvc := audit.NewService(q)

	user, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user not found: %s (sign up + verify first)", email)
		}
		return fmt.Errorf("lookup user: %w", err)
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := q.WithTx(tx)

	// Inline seed so a truly-fresh DB (post-migrate:up, pre-serve:api) is
	// safe to bootstrap. Discards the seed log — admin:grant's stdout is
	// the audit trail for this command.
	if err := acl.Seed(ctx, qtx, silentLogger()); err != nil {
		return fmt.Errorf("seed catalog: %w", err)
	}

	desc := "bootstrap admin"
	group, err := qtx.UpsertGroup(ctx, sqlcdb.UpsertGroupParams{
		Name:        groupName,
		Description: &desc,
	})
	if err != nil {
		return fmt.Errorf("upsert group: %w", err)
	}

	permNames := make([]string, 0, len(acl.PermissionCatalog))
	for _, p := range acl.PermissionCatalog {
		row, err := qtx.GetPermissionByName(ctx, p.Name)
		if err != nil {
			return fmt.Errorf("lookup permission %s: %w", p.Name, err)
		}
		if err := qtx.LinkGroupPermission(ctx, sqlcdb.LinkGroupPermissionParams{
			GroupID:      group.ID,
			PermissionID: row.ID,
		}); err != nil {
			return fmt.Errorf("link permission %s: %w", p.Name, err)
		}
		permNames = append(permNames, p.Name)
	}

	if err := qtx.JoinUserToGroup(ctx, sqlcdb.JoinUserToGroupParams{
		GroupID: group.ID,
		UserID:  user.ID,
	}); err != nil {
		return fmt.Errorf("join user to group: %w", err)
	}

	subjectID := user.ID
	if err := auditSvc.RecordTx(ctx, qtx, audit.Params{
		Actor:       audit.Actor{Kind: audit.KindCLI},
		Action:      "admin.grant",
		SubjectType: "user",
		SubjectID:   &subjectID,
		Changes: map[string]any{
			"group":       group.Name,
			"permissions": permNames,
		},
	}); err != nil {
		return fmt.Errorf("audit: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	fmt.Printf("granted: user=%s group=%s permissions=%d\n", user.Email, group.Name, len(permNames))
	return nil
}

// silentLogger returns a *slog.Logger that discards every record.
// Used by inline acl.Seed calls in one-shot commands where the seed
// output would clutter the operator's terminal.
func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
