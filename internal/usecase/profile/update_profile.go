package profile

import (
	"context"
	"errors"

	"go-boilerplate/internal/dto/profile"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
)

// UpdateProfile updates a user's profile.
func (uc *UseCase) UpdateProfile(ctx context.Context, userID uint, req profiledto.UpdateProfileRequest) (*profiledto.ProfileResponse, error) {
	profile, err := uc.getOrCreateProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := uc.updateProfileFields(ctx, profile, userID, req); err != nil {
		return nil, err
	}

	if err := uc.profileRepo.Upsert(ctx, profile); err != nil {
		return nil, err
	}

	return uc.buildProfileResponse(ctx, profile)
}

// getOrCreateProfile retrieves existing profile or creates a new one.
func (uc *UseCase) getOrCreateProfile(ctx context.Context, userID uint) (*entity.Profile, error) {
	profile, err := uc.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return &entity.Profile{UserID: userID}, nil
		}
		return nil, err
	}
	return profile, nil
}

// updateProfileFields updates the profile fields based on the request.
func (uc *UseCase) updateProfileFields(ctx context.Context, profile *entity.Profile, userID uint, req profiledto.UpdateProfileRequest) error {
	if req.AvatarMediaID != nil {
		if err := uc.validateAndSetAvatar(ctx, profile, userID, *req.AvatarMediaID); err != nil {
			return err
		}
	}

	if req.Bio != nil {
		profile.Bio = *req.Bio
	}

	if req.Phone != nil {
		profile.Phone = *req.Phone
	}

	return nil
}

// validateAndSetAvatar validates media ownership and sets avatar.
func (uc *UseCase) validateAndSetAvatar(ctx context.Context, profile *entity.Profile, userID, mediaID uint) error {
	media, err := uc.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrInvalidMedia
		}
		return err
	}

	if !isMediaOwnedByUser(media, userID) {
		return ErrInvalidMedia
	}

	profile.AvatarMediaID = &mediaID
	return nil
}

// isMediaOwnedByUser checks if media belongs to the user.
func isMediaOwnedByUser(media *entity.Media, userID uint) bool {
	validType := media.AttachableType == "profiles" || media.AttachableType == "users"
	return validType && media.AttachableID == userID
}

// buildProfileResponse creates a profile response with avatar URL.
func (uc *UseCase) buildProfileResponse(ctx context.Context, profile *entity.Profile) (*profiledto.ProfileResponse, error) {
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
