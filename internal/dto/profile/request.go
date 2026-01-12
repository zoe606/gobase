package profile

// UpdateProfileRequest contains the fields for updating a user profile.
type UpdateProfileRequest struct {
	Bio           *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	Phone         *string `json:"phone,omitempty" validate:"omitempty,max=20"`
	AvatarMediaID *uint   `json:"avatar_media_id,omitempty"`
}
