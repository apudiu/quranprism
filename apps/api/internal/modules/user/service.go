package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
	"github.com/apudiu/quranprism/api/internal/modules/acl"
)

// Service is the public API of the user module. Handlers, the auth
// module, and middleware all depend on this — never on Repository
// directly. Allows future read-through caching / event emission to be
// added in one place.
type Service struct {
	repo *Repository
	acl  *acl.Service
}

// NewService wires the user Service.
func NewService(repo *Repository, aclSvc *acl.Service) *Service {
	return &Service{repo: repo, acl: aclSvc}
}

// --- read paths ----------------------------------------------------

// GetByID returns the public-DTO User.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetCredentialsByEmail returns the row including password_hash. Intended
// for the auth module's login path; handlers must not call this directly.
func (s *Service) GetCredentialsByEmail(ctx context.Context, email string) (*sqlcdb.User, error) {
	return s.repo.GetCredentialsByEmail(ctx, email)
}

// GetCredentialsByID returns the credentialed row for the auth module's
// change-password and similar verify-then-update flows. Handlers must
// not call this directly.
func (s *Service) GetCredentialsByID(ctx context.Context, id uuid.UUID) (*sqlcdb.User, error) {
	return s.repo.GetCredentialsByID(ctx, id)
}

// IsEmailVerified satisfies the middleware.EmailVerifier interface so
// EmailVerifiedRequired can be wired without that package importing user.
// The string form (vs uuid.UUID) matches the Identity.UserID shape.
func (s *Service) IsEmailVerified(ctx context.Context, userIDStr string) (bool, error) {
	id, err := uuid.Parse(userIDStr)
	if err != nil {
		return false, fmt.Errorf("user: parse id: %w", err)
	}
	return s.repo.IsEmailVerified(ctx, id)
}

// MeView is the /v1/me payload: public profile + the names of every group
// the user belongs to. Group names drive the front-end's admin-UI gates.
type MeView struct {
	*User
	Groups []string `json:"groups"`
}

// Me builds the MeView for the currently authenticated user.
func (s *Service) Me(ctx context.Context, id uuid.UUID) (*MeView, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	groups, err := s.acl.ListGroupsForUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return &MeView{User: u, Groups: groups}, nil
}

// --- write paths ---------------------------------------------------

// Create persists a new user with the bcrypt hash supplied by the auth
// module. Wraps Repository.Create for the auth signup flow.
func (s *Service) Create(ctx context.Context, email, name, passwordHash string) (*User, error) {
	return s.repo.Create(ctx, email, name, passwordHash)
}

// UpdateProfile applies optional name/avatar changes.
func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, name, avatarURL *string) (*User, error) {
	return s.repo.UpdateProfile(ctx, id, name, avatarURL)
}

// UpdatePassword swaps the bcrypt hash. Auth module is expected to have
// already verified the current-password challenge.
func (s *Service) UpdatePassword(ctx context.Context, id uuid.UUID, newHash string) error {
	return s.repo.UpdatePassword(ctx, id, newHash)
}

// MarkEmailVerified is idempotent; safe to call from the auth module on
// successful verify-token consumption.
func (s *Service) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	return s.repo.MarkEmailVerified(ctx, id)
}

// StartDeletionGrace begins the PRV-4 30-day window. The cron worker
// hard-deletes the row when the window closes.
func (s *Service) StartDeletionGrace(ctx context.Context, id uuid.UUID) error {
	return s.repo.StartDeletionGrace(ctx, id)
}

// CancelDeletionGrace clears the deletion request. Called from the auth
// module's login path on an in-grace-window login.
func (s *Service) CancelDeletionGrace(ctx context.Context, id uuid.UUID) error {
	return s.repo.CancelDeletionGrace(ctx, id)
}
