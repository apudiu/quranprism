// Package user is the User domain module: identity, profile, lifecycle
// (verification, deletion grace). Auth-token issuance and credential
// checks live in the sibling auth module.
package user

import (
	"time"

	"github.com/google/uuid"

	sqlcdb "github.com/apudiu/quranprism/api/internal/db/sqlc"
)

// User is the public-facing domain type. Password hash and other internal
// fields are intentionally absent — they never leave the auth/user service
// layer.
type User struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	Name            string     `json:"name"`
	AvatarURL       *string    `json:"avatar_url,omitempty"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	IsDisabled      bool       `json:"is_disabled"`
	CreatedAt       time.Time  `json:"created_at"`
}

func fromRow(r sqlcdb.User) *User {
	u := &User{
		ID:         r.ID,
		Email:      r.Email,
		Name:       r.Name,
		AvatarURL:  r.AvatarUrl,
		IsDisabled: r.IsDisabled,
		CreatedAt:  r.CreatedAt,
	}
	if r.EmailVerifiedAt.Valid {
		t := r.EmailVerifiedAt.Time
		u.EmailVerifiedAt = &t
	}
	return u
}
