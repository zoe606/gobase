package auth

import "go-boilerplate/internal/entity"

// RegisterOutput represents registration output.
type RegisterOutput struct {
	User         *entity.User
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

// ToResponse converts output to API response.
func (o *RegisterOutput) ToResponse() LoginResponse {
	return LoginResponse{
		AccessToken:  o.AccessToken,
		RefreshToken: o.RefreshToken,
		ExpiresAt:    o.ExpiresAt,
		User:         userToResponse(o.User),
	}
}

// LoginOutput represents login output.
type LoginOutput struct {
	User         *entity.User
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

// ToResponse converts output to API response.
func (o *LoginOutput) ToResponse() LoginResponse {
	return LoginResponse{
		AccessToken:  o.AccessToken,
		RefreshToken: o.RefreshToken,
		ExpiresAt:    o.ExpiresAt,
		User:         userToResponse(o.User),
	}
}

// RefreshOutput represents refresh token output.
type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

// ToResponse converts output to API response.
func (o *RefreshOutput) ToResponse() TokenResponse {
	return TokenResponse{
		AccessToken:  o.AccessToken,
		RefreshToken: o.RefreshToken,
		ExpiresAt:    o.ExpiresAt,
	}
}

// UserOutput represents user data output.
type UserOutput struct {
	User *entity.User
}

// ToResponse converts output to API response.
func (o *UserOutput) ToResponse() UserResponse {
	return userToResponse(o.User)
}

// userToResponse converts entity.User to UserResponse.
func userToResponse(u *entity.User) UserResponse {
	return UserResponse{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
		Role:  u.Role.Name,
	}
}
