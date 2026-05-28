package user

// UpdateProfileRequest is the body of PATCH /v1/me. Both fields are
// optional; absent means "leave unchanged". AvatarURL can be set to ""
// to clear it (handled at the handler boundary by detecting "" vs nil).
type UpdateProfileRequest struct {
	Name      *string `json:"name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}
