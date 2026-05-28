package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
)

// Repository owns the user table reads/writes. The service layer talks
// only to this — never to *sqlc.Queries directly — so that swapping the
// SQL layer later doesn't ripple through business code.
type Repository struct {
	q *sqlcdb.Queries
}

// NewRepository wires the Repository against the shared *sqlc.Queries.
func NewRepository(q *sqlcdb.Queries) *Repository { return &Repository{q: q} }

// Create inserts a new user. Caller has already bcrypt-hashed the
// password. Returns ErrEmailTaken on unique-constraint collision so the
// signup handler can render a clean 409.
func (r *Repository) Create(ctx context.Context, email, name, passwordHash string) (*User, error) {
	row, err := r.q.CreateUser(ctx, sqlcdb.CreateUserParams{
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if isUniqueViolation(err, "users_email_unique") {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("user: create: %w", err)
	}
	return fromRow(row), nil
}

// GetByID returns the public user record. ErrNotFound for tombstoned or
// missing rows.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("user: get by id: %w", err)
	}
	return fromRow(row), nil
}

// GetCredentialsByEmail returns the row including password_hash. Only
// the auth module's login path calls this; handlers must not.
func (r *Repository) GetCredentialsByEmail(ctx context.Context, email string) (*sqlcdb.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("user: get by email: %w", err)
	}
	return &row, nil
}

// GetCredentialsByID returns the credentialed row by id. Used by the
// auth module's verify-then-update flows (change-password).
func (r *Repository) GetCredentialsByID(ctx context.Context, id uuid.UUID) (*sqlcdb.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("user: get credentials by id: %w", err)
	}
	return &row, nil
}

// IsEmailVerified is the hot-path predicate the EmailVerifiedRequired
// middleware calls on every protected request. Treats a vanished user
// row as "not verified" rather than surfacing pgx.ErrNoRows.
func (r *Repository) IsEmailVerified(ctx context.Context, id uuid.UUID) (bool, error) {
	verified, err := r.q.IsEmailVerified(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("user: is verified: %w", err)
	}
	return verified, nil
}

// UpdateProfile applies optional name/avatar changes. Nil fields are
// left untouched (handled by COALESCE in the SQL).
func (r *Repository) UpdateProfile(ctx context.Context, id uuid.UUID, name, avatarURL *string) (*User, error) {
	row, err := r.q.UpdateUserProfile(ctx, sqlcdb.UpdateUserProfileParams{
		Name:      name,
		AvatarUrl: avatarURL,
		ID:        id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("user: update profile: %w", err)
	}
	return fromRow(row), nil
}

// UpdatePassword swaps in a new bcrypt hash.
func (r *Repository) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	if err := r.q.UpdateUserPassword(ctx, sqlcdb.UpdateUserPasswordParams{
		ID: id, PasswordHash: hash,
	}); err != nil {
		return fmt.Errorf("user: update password: %w", err)
	}
	return nil
}

// MarkEmailVerified flips email_verified_at to NOW() unless already set.
func (r *Repository) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	if err := r.q.MarkUserEmailVerified(ctx, id); err != nil {
		return fmt.Errorf("user: mark verified: %w", err)
	}
	return nil
}

// StartDeletionGrace stamps deletion_requested_at; cancelled on next login.
func (r *Repository) StartDeletionGrace(ctx context.Context, id uuid.UUID) error {
	if err := r.q.StartUserDeletionGrace(ctx, id); err != nil {
		return fmt.Errorf("user: start deletion grace: %w", err)
	}
	return nil
}

// CancelDeletionGrace clears deletion_requested_at on grace-window login.
func (r *Repository) CancelDeletionGrace(ctx context.Context, id uuid.UUID) error {
	if err := r.q.CancelUserDeletionGrace(ctx, id); err != nil {
		return fmt.Errorf("user: cancel deletion grace: %w", err)
	}
	return nil
}

// isUniqueViolation matches a pgx unique-constraint error against a
// specific constraint name. Used to translate generic sqlc errors into
// typed domain errors (e.g. ErrEmailTaken).
func isUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "23505" {
		return false
	}
	return pgErr.ConstraintName == constraintName
}
