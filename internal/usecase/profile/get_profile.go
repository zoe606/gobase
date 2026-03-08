package profile

import (
	"context"
	"errors"

	profiledto "go-boilerplate/internal/dto/profile"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
)

// GetProfile retrieves a user's profile, creating one if it doesn't exist.
func (uc *UseCase) GetProfile(ctx context.Context, userID uint) (*profiledto.ProfileResponse, error) {
	profile, err := uc.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			// Create empty profile if not exists
			profile = &entity.Profile{UserID: userID}
			if createErr := uc.profileRepo.Create(ctx, profile); createErr != nil {
				return nil, createErr
			}
		} else {
			return nil, err
		}
	}

	var avatarURL string
	if profile.AvatarMediaID != nil {
		url, err := uc.getAvatarURL(ctx, *profile.AvatarMediaID)
		if err != nil {
			uc.l.Warn("Failed to get avatar URL", "error", err, "mediaID", *profile.AvatarMediaID)
		} else {
			avatarURL = url
		}
	}

	return profiledto.FromEntity(profile, avatarURL), nil
}
