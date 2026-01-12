package profile

import (
	"go-boilerplate/internal/entity"
)

// ProfileResponse represents a user profile in API responses.
type ProfileResponse struct {
	UserID uint            `json:"user_id"`
	Bio    string          `json:"bio"`
	Phone  string          `json:"phone"`
	Avatar *AvatarResponse `json:"avatar,omitempty"`
}

// AvatarResponse represents avatar information in API responses.
type AvatarResponse struct {
	ID  uint   `json:"id"`
	URL string `json:"url"`
}

// FromEntity converts an entity.Profile to ProfileResponse.
func FromEntity(p *entity.Profile, avatarURL string) *ProfileResponse {
	resp := &ProfileResponse{
		UserID: p.UserID,
		Bio:    p.Bio,
		Phone:  p.Phone,
	}

	if p.AvatarMediaID != nil && avatarURL != "" {
		resp.Avatar = &AvatarResponse{
			ID:  *p.AvatarMediaID,
			URL: avatarURL,
		}
	}

	return resp
}
